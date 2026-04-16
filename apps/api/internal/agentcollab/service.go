package agentcollab

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/fsnotify/fsnotify"

	"github.com/nikkofu/relay-agent-workspace/api/internal/realtime"
)

const collabChannelID = "ch-collab"

type TaskBoardItem struct {
	Status      string `json:"status"`
	Task        string `json:"task"`
	AssignedTo  string `json:"assigned_to"`
	Deadline    string `json:"deadline"`
	Description string `json:"description"`
}

type ActiveSuperpowerItem struct {
	Agent        string `json:"agent"`
	CurrentSkill string `json:"current_skill"`
	ActiveTask   string `json:"active_task"`
	Progress     string `json:"progress"`
}

type Snapshot struct {
	ActiveSuperpowers []ActiveSuperpowerItem `json:"active_superpowers"`
	TaskBoard         []TaskBoardItem        `json:"task_board"`
}

type Service struct {
	path    string
	hub     *realtime.Hub
	watcher *fsnotify.Watcher
}

func NewService(path string, hub *realtime.Hub) *Service {
	return &Service{path: path, hub: hub}
}

func DefaultPath() string {
	return filepath.Join("..", "..", "docs", "AGENT-COLLAB.md")
}

func (s *Service) Start() error {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return err
	}

	if err := watcher.Add(s.path); err != nil {
		_ = watcher.Close()
		return err
	}

	s.watcher = watcher

	if err := s.SyncFile(); err != nil {
		return err
	}

	go s.loop()
	return nil
}

func (s *Service) Close() error {
	if s.watcher == nil {
		return nil
	}
	return s.watcher.Close()
}

func (s *Service) SyncFile() error {
	content, err := os.ReadFile(s.path)
	if err != nil {
		return err
	}

	snapshot, err := ParseMarkdown(content)
	if err != nil {
		return err
	}

	return s.hub.Broadcast(realtime.Event{
		ID:        "evt_" + time.Now().Format("20060102150405.000000"),
		Type:      "agent_collab.sync",
		ChannelID: collabChannelID,
		TS:        time.Now().UTC().Format(time.RFC3339Nano),
		Payload: map[string]any{
			"active_superpowers": snapshot.ActiveSuperpowers,
			"task_board":         snapshot.TaskBoard,
		},
	})
}

func (s *Service) loop() {
	for {
		select {
		case event, ok := <-s.watcher.Events:
			if !ok {
				return
			}
			if event.Has(fsnotify.Write) || event.Has(fsnotify.Create) {
				_ = s.SyncFile()
			}
		case _, ok := <-s.watcher.Errors:
			if !ok {
				return
			}
		}
	}
}

func ParseMarkdown(content []byte) (Snapshot, error) {
	lines := scanLines(string(content))

	taskRows, err := extractSectionTable(lines, "Task Board")
	if err != nil {
		return Snapshot{}, err
	}

	activeRows, err := extractSectionTable(lines, "Active Superpowers")
	if err != nil {
		return Snapshot{}, err
	}

	taskBoard, err := parseTaskBoard(taskRows)
	if err != nil {
		return Snapshot{}, err
	}

	activeSuperpowers, err := parseActiveSuperpowers(activeRows)
	if err != nil {
		return Snapshot{}, err
	}

	return Snapshot{
		ActiveSuperpowers: activeSuperpowers,
		TaskBoard:         taskBoard,
	}, nil
}

func extractSectionTable(lines []string, headingNeedle string) ([][]string, error) {
	sectionIndex := -1
	for i, line := range lines {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "## ") && strings.Contains(trimmed, headingNeedle) {
			sectionIndex = i
			break
		}
	}
	if sectionIndex == -1 {
		return nil, fmt.Errorf("section not found: %s", headingNeedle)
	}

	for i := sectionIndex + 1; i < len(lines); i++ {
		line := strings.TrimSpace(lines[i])
		if !strings.HasPrefix(line, "|") {
			continue
		}
		if i+1 >= len(lines) || !strings.HasPrefix(strings.TrimSpace(lines[i+1]), "|") {
			return nil, fmt.Errorf("markdown table separator missing for section: %s", headingNeedle)
		}

		rows := make([][]string, 0, 8)
		for j := i + 2; j < len(lines); j++ {
			rowLine := strings.TrimSpace(lines[j])
			if !strings.HasPrefix(rowLine, "|") {
				break
			}
			rows = append(rows, parseMarkdownRow(rowLine))
		}
		return rows, nil
	}

	return nil, fmt.Errorf("table not found for section: %s", headingNeedle)
}

func parseTaskBoard(rows [][]string) ([]TaskBoardItem, error) {
	items := make([]TaskBoardItem, 0, len(rows))
	for _, row := range rows {
		if len(row) != 5 {
			return nil, fmt.Errorf("task board row has %d columns, expected 5", len(row))
		}
		items = append(items, TaskBoardItem{
			Status:      normalizeCell(row[0]),
			Task:        normalizeCell(row[1]),
			AssignedTo:  normalizeCell(row[2]),
			Deadline:    normalizeCell(row[3]),
			Description: normalizeCell(row[4]),
		})
	}
	return items, nil
}

func parseActiveSuperpowers(rows [][]string) ([]ActiveSuperpowerItem, error) {
	items := make([]ActiveSuperpowerItem, 0, len(rows))
	for _, row := range rows {
		if len(row) != 4 {
			return nil, fmt.Errorf("active superpowers row has %d columns, expected 4", len(row))
		}
		items = append(items, ActiveSuperpowerItem{
			Agent:        normalizeCell(row[0]),
			CurrentSkill: normalizeCell(row[1]),
			ActiveTask:   normalizeCell(row[2]),
			Progress:     normalizeCell(row[3]),
		})
	}
	return items, nil
}

func parseMarkdownRow(line string) []string {
	trimmed := strings.TrimSpace(line)
	trimmed = strings.TrimPrefix(trimmed, "|")
	trimmed = strings.TrimSuffix(trimmed, "|")

	parts := strings.Split(trimmed, "|")
	for i := range parts {
		parts[i] = strings.TrimSpace(parts[i])
	}
	return parts
}

func normalizeCell(value string) string {
	replacer := strings.NewReplacer("**", "", "`", "", "\r", "")
	return strings.TrimSpace(replacer.Replace(value))
}

func scanLines(content string) []string {
	scanner := bufio.NewScanner(strings.NewReader(content))
	lines := make([]string, 0, 64)
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}
	return lines
}
