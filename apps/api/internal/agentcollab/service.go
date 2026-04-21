package agentcollab

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
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

type MemberProfile struct {
	Name         string `json:"name"`
	Role         string `json:"role"`
	Specialty    string `json:"specialty"`
	PrimaryTools string `json:"primary_tools"`
}

type CommLogEntry struct {
	ID        string `json:"id"`
	Date      string `json:"date"`
	Title     string `json:"title"`
	From      string `json:"from"`
	To        string `json:"to,omitempty"`
	Content   string `json:"content"`
	CreatedAt string `json:"created_at"`
}

type Snapshot struct {
	ActiveSuperpowers []ActiveSuperpowerItem `json:"active_superpowers"`
	CommLog           []CommLogEntry         `json:"comm_log"`
	Members           []MemberProfile        `json:"members"`
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
	snapshot, err := ReadSnapshot(s.path)
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
			"comm_log":           snapshot.CommLog,
			"members":            snapshot.Members,
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

func ReadSnapshot(path string) (Snapshot, error) {
	content, err := os.ReadFile(path)
	if err != nil {
		return Snapshot{}, err
	}

	return ParseMarkdown(content)
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
	memberRows := extractOptionalSectionTable(lines, "Member Profiles")

	taskBoard, err := parseTaskBoard(taskRows)
	if err != nil {
		return Snapshot{}, err
	}

	activeSuperpowers, err := parseActiveSuperpowers(activeRows)
	if err != nil {
		return Snapshot{}, err
	}
	members, err := parseMembers(memberRows)
	if err != nil {
		return Snapshot{}, err
	}

	return Snapshot{
		ActiveSuperpowers: activeSuperpowers,
		CommLog:           parseCommLog(lines),
		Members:           members,
		TaskBoard:         taskBoard,
	}, nil
}

func AppendCommLogEntry(path string, from, to, title, content string, now time.Time) (CommLogEntry, error) {
	from = strings.TrimSpace(from)
	to = strings.TrimSpace(to)
	title = strings.TrimSpace(title)
	content = strings.TrimSpace(content)
	if from == "" {
		return CommLogEntry{}, fmt.Errorf("from is required")
	}
	if title == "" {
		return CommLogEntry{}, fmt.Errorf("title is required")
	}
	if content == "" {
		return CommLogEntry{}, fmt.Errorf("content is required")
	}

	raw, err := os.ReadFile(path)
	if err != nil {
		return CommLogEntry{}, err
	}

	now = now.UTC()
	entry := CommLogEntry{
		ID:        "comm-" + now.Format("20060102150405.000000"),
		Date:      now.Format("2006-01-02"),
		Title:     title,
		From:      from,
		To:        to,
		Content:   content,
		CreatedAt: now.Format(time.RFC3339Nano),
	}

	block := formatCommLogBlock(entry)
	updated := insertCommLogBlock(string(raw), block)
	if err := os.WriteFile(path, []byte(updated), 0o644); err != nil {
		return CommLogEntry{}, err
	}
	return entry, nil
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

func extractOptionalSectionTable(lines []string, headingNeedle string) [][]string {
	rows, err := extractSectionTable(lines, headingNeedle)
	if err != nil {
		return [][]string{}
	}
	return rows
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

func parseMembers(rows [][]string) ([]MemberProfile, error) {
	items := make([]MemberProfile, 0, len(rows))
	for _, row := range rows {
		if len(row) != 4 {
			return nil, fmt.Errorf("member profile row has %d columns, expected 4", len(row))
		}
		items = append(items, MemberProfile{
			Name:         normalizeCell(row[0]),
			Role:         normalizeCell(row[1]),
			Specialty:    normalizeCell(row[2]),
			PrimaryTools: normalizeCell(row[3]),
		})
	}
	return items, nil
}

func parseCommLog(lines []string) []CommLogEntry {
	start := -1
	for i, line := range lines {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "## ") && strings.Contains(trimmed, "Communication Log") {
			start = i + 1
			break
		}
	}
	if start == -1 {
		return []CommLogEntry{}
	}

	entries := make([]CommLogEntry, 0, 8)
	var currentTitle string
	var currentDate string
	for i := start; i < len(lines); i++ {
		line := strings.TrimSpace(lines[i])
		if strings.HasPrefix(line, "## ") {
			break
		}
		if strings.HasPrefix(line, "### ") {
			currentDate, currentTitle = parseCommHeading(strings.TrimPrefix(line, "### "))
			continue
		}
		if strings.HasPrefix(line, "- **") {
			if entry, ok := parseCommBullet(line, currentDate, currentTitle, len(entries)); ok {
				entries = append(entries, entry)
			}
		}
	}
	return entries
}

func parseCommHeading(heading string) (string, string) {
	parts := strings.SplitN(strings.TrimSpace(heading), " - ", 2)
	if len(parts) != 2 {
		return "", strings.TrimSpace(heading)
	}
	return strings.TrimSpace(parts[0]), strings.TrimSpace(parts[1])
}

func parseCommBullet(line, date, title string, index int) (CommLogEntry, bool) {
	body := strings.TrimPrefix(line, "- **")
	parts := strings.SplitN(body, "**:", 2)
	if len(parts) != 2 {
		return CommLogEntry{}, false
	}
	participants := normalizeCell(parts[0])
	content := strings.TrimSpace(parts[1])
	content = strings.Trim(content, `"`)
	from := participants
	to := ""
	if split := strings.SplitN(participants, "→", 2); len(split) == 2 {
		from = strings.TrimSpace(split[0])
		to = strings.TrimSpace(split[1])
	}
	idParts := []string{date, title, from, to, strconv.Itoa(index)}
	return CommLogEntry{
		ID:      "comm-" + stableToken(strings.Join(idParts, "-")),
		Date:    date,
		Title:   title,
		From:    from,
		To:      to,
		Content: content,
	}, true
}

func formatCommLogBlock(entry CommLogEntry) string {
	target := entry.From
	if entry.To != "" {
		target += " → " + entry.To
	}
	return fmt.Sprintf("### %s - %s\n- **%s**: %q\n\n", entry.Date, entry.Title, target, entry.Content)
}

func insertCommLogBlock(content, block string) string {
	lines := strings.Split(content, "\n")
	for i, line := range lines {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "## ") && strings.Contains(trimmed, "Communication Log") {
			insertAt := i + 1
			for insertAt < len(lines) && strings.TrimSpace(lines[insertAt]) == "" {
				insertAt++
			}
			updated := make([]string, 0, len(lines)+4)
			updated = append(updated, lines[:insertAt]...)
			updated = append(updated, strings.Split(strings.TrimSuffix(block, "\n"), "\n")...)
			updated = append(updated, "")
			updated = append(updated, lines[insertAt:]...)
			return strings.Join(updated, "\n")
		}
	}
	separator := "\n"
	if strings.HasSuffix(content, "\n") {
		separator = ""
	}
	return content + separator + "## 💬 Communication Log\n\n" + block
}

func stableToken(value string) string {
	value = strings.ToLower(value)
	replacer := strings.NewReplacer(" ", "-", "/", "-", ":", "", ".", "", "→", "to")
	value = replacer.Replace(value)
	value = strings.Trim(value, "-")
	if value == "" {
		return "entry"
	}
	return value
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
