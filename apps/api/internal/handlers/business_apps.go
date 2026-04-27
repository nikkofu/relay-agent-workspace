package handlers

import (
	"net/http"
	"sort"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/nikkofu/relay-agent-workspace/api/internal/db"
	"github.com/nikkofu/relay-agent-workspace/api/internal/domain"
)

type BusinessAppMetadata struct {
	ID          string   `json:"id"`
	Title       string   `json:"title"`
	Description string   `json:"description"`
	Icon        string   `json:"icon"`
	Modes       []string `json:"modes"`
}

var businessApps = []BusinessAppMetadata{
	{
		ID:          "sales",
		Title:       "Sales App",
		Description: "Manage sales orders, customers, and pipelines.",
		Icon:        "DollarSign",
		Modes:       []string{"list", "card_grid", "kanban", "calendar", "stat"},
	},
}

func ListApps(c *gin.Context) {
	SeedSalesOrders()
	c.JSON(http.StatusOK, gin.H{"apps": businessApps})
}

func GetApp(c *gin.Context) {
	id := c.Param("id")
	for _, app := range businessApps {
		if app.ID == id {
			c.JSON(http.StatusOK, gin.H{"app": app})
			return
		}
	}
	c.JSON(http.StatusNotFound, gin.H{"error": "app not found"})
}

func GetAppData(c *gin.Context) {
	SeedSalesOrders()
	appID := c.Param("id")
	if appID != "sales" {
		c.JSON(http.StatusNotFound, gin.H{"error": "app data source not found"})
		return
	}

	mode := c.Query("mode")
	if mode == "" {
		mode = "list"
	}

	// Phase 74: Normalize mode aliases
	if mode == "search" {
		mode = "list"
	}
	if mode == "stats" {
		mode = "stat"
	}

	// Validate mode
	validModes := map[string]bool{
		"list":      true,
		"card_grid": true,
		"kanban":    true,
		"calendar":  true,
		"stat":      true,
	}
	if !validModes[mode] {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid mode"})
		return
	}

	// Filters
	q := c.Query("q")
	stage := c.Query("stage")
	status := c.Query("status")
	ownerID := c.Query("owner_user_id")
	limitStr := c.DefaultQuery("limit", "20")
	limit, _ := strconv.Atoi(limitStr)
	if limit <= 0 || limit > 50 {
		limit = 20
	}
	cursor := c.Query("cursor")

	query := db.DB.Model(&domain.SalesOrder{})

	if q != "" {
		query = query.Where("customer_name LIKE ? OR order_number LIKE ? OR summary LIKE ? OR tags LIKE ?", "%"+q+"%", "%"+q+"%", "%"+q+"%", "%"+q+"%")
	}
	if stage != "" {
		query = query.Where("stage = ?", stage)
	}
	if status != "" {
		query = query.Where("status = ?", status)
	}
	if ownerID != "" {
		query = query.Where("owner_user_id = ?", ownerID)
	}

	// Cursor pagination
	if cursor != "" {
		query = query.Where("id < ?", cursor)
	}

	var orders []domain.SalesOrder
	if err := query.Order("id desc").Limit(limit + 1).Find(&orders).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to load data"})
		return
	}

	nextCursor := ""
	if len(orders) > limit {
		nextCursor = orders[limit-1].ID
		orders = orders[:limit]
	}

	// Phase 74: Calendar event projection
	var calendarEvents []gin.H
	if mode == "calendar" {
		timeField := c.DefaultQuery("calendar_time_field", "expected_close_date")
		validTimeFields := map[string]bool{
			"expected_close_date": true,
			"order_date":          true,
			"due_date":            true,
			"last_activity_at":    true,
		}
		if !validTimeFields[timeField] {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid calendar_time_field"})
			return
		}

		view := c.DefaultQuery("calendar_view", "month")
		validViews := map[string]bool{"day": true, "week": true, "month": true}
		if !validViews[view] {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid calendar_view"})
			return
		}

		calendarEvents = make([]gin.H, 0)
		for _, order := range orders {
			var startTime *time.Time
			switch timeField {
			case "expected_close_date":
				startTime = order.ExpectedCloseDate
			case "order_date", "due_date", "last_activity_at":
				// These fields don't exist yet, but contract allows them
				// We'll treat them as nil for now or add them to model if needed.
				// Based on Task 2, we just need to handle the projection.
			}

			if startTime != nil {
				calendarEvents = append(calendarEvents, gin.H{
					"id":         "cal-" + order.ID,
					"record_id":  order.ID,
					"title":      order.OrderNumber + ": " + order.CustomerName,
					"start":      startTime,
					"end":        startTime.Add(time.Hour), // 1 hour default duration
					"time_field": timeField,
					"stage":      order.Stage,
					"status":     order.Status,
					"amount":     order.Amount,
					"record":     order,
				})
			}
		}
	}

	// Schema metadata
	schema := gin.H{
		"entity": "sales_order",
		"fields": []gin.H{
			{"key": "id", "type": "string", "label": "ID"},
			{"key": "order_number", "type": "string", "label": "Order #"},
			{"key": "customer_name", "type": "string", "label": "Customer"},
			{"key": "amount", "type": "number", "label": "Amount"},
			{"key": "currency", "type": "string", "label": "Currency"},
			{"key": "stage", "type": "string", "label": "Stage"},
			{"key": "status", "type": "string", "label": "Status"},
			{"key": "expected_close_date", "type": "datetime", "label": "Expected Close"},
			{"key": "owner_user_id", "type": "user_id", "label": "Owner"},
		},
	}

	c.JSON(http.StatusOK, gin.H{
		"data":            orders,
		"calendar_events": calendarEvents,
		"next_cursor":     nextCursor,
		"schema":          schema,
	})
}

func GetAppStats(c *gin.Context) {
	SeedSalesOrders()
	appID := c.Param("id")
	if appID != "sales" {
		c.JSON(http.StatusNotFound, gin.H{"error": "app not found"})
		return
	}

	chartStyle := c.DefaultQuery("chart_style", "summary")
	validStyles := map[string]bool{"summary": true, "bar": true, "funnel": true, "timeline": true}
	if !validStyles[chartStyle] {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid chart_style"})
		return
	}

	var stats struct {
		TotalAmount float64 `json:"total_amount"`
		OrderCount  int64   `json:"order_count"`
		ByStage     []struct {
			Stage string  `json:"stage"`
			Count int64   `json:"count"`
			Sum   float64 `json:"sum"`
		} `json:"by_stage"`
		Funnel []struct {
			Stage string  `json:"stage"`
			Count int64   `json:"count"`
			Sum   float64 `json:"sum"`
		} `json:"funnel"`
		TimelineBuckets []struct {
			Label string  `json:"label"`
			Count int64   `json:"count"`
			Sum   float64 `json:"sum"`
		} `json:"timeline_buckets"`
	}

	db.DB.Model(&domain.SalesOrder{}).Count(&stats.OrderCount)
	db.DB.Model(&domain.SalesOrder{}).Select("SUM(amount)").Scan(&stats.TotalAmount)
	db.DB.Model(&domain.SalesOrder{}).
		Select("stage, COUNT(*) as count, SUM(amount) as sum").
		Group("stage").
		Scan(&stats.ByStage)

	// Phase 74: Compute ordered funnel
	stages := []string{"lead", "qualified", "proposal", "negotiation", "closed_won"}
	for _, s := range stages {
		var count int64
		var sum float64
		for _, b := range stats.ByStage {
			if b.Stage == s {
				count = b.Count
				sum = b.Sum
				break
			}
		}
		stats.Funnel = append(stats.Funnel, struct {
			Stage string  `json:"stage"`
			Count int64   `json:"count"`
			Sum   float64 `json:"sum"`
		}{Stage: s, Count: count, Sum: sum})
	}

	// Phase 74: Compute timeline buckets (by month of expected_close_date)
	var orders []domain.SalesOrder
	db.DB.Where("expected_close_date IS NOT NULL").Order("expected_close_date asc").Find(&orders)
	
	buckets := make(map[string]*struct {
		Count int64
		Sum   float64
	})
	
	for _, o := range orders {
		label := o.ExpectedCloseDate.Format("2006-01")
		if buckets[label] == nil {
			buckets[label] = &struct {
				Count int64
				Sum   float64
			}{}
		}
		buckets[label].Count++
		buckets[label].Sum += o.Amount
	}

	// Sort labels
	var labels []string
	for l := range buckets {
		labels = append(labels, l)
	}
	sort.Strings(labels)

	for _, l := range labels {
		stats.TimelineBuckets = append(stats.TimelineBuckets, struct {
			Label string  `json:"label"`
			Count int64   `json:"count"`
			Sum   float64 `json:"sum"`
		}{Label: l, Count: buckets[l].Count, Sum: buckets[l].Sum})
	}

	c.JSON(http.StatusOK, gin.H{"stats": stats})
}

func SeedSalesOrders() {
	var count int64
	db.DB.Model(&domain.SalesOrder{}).Count(&count)
	if count > 0 {
		return
	}

	now := time.Now().UTC()
	orders := []domain.SalesOrder{
		{
			ID:                "so-1",
			OrderNumber:       "SO-2026-001",
			CustomerName:      "TechFlow Systems",
			Amount:            12500,
			Currency:          "USD",
			Stage:             "negotiation",
			Status:            "active",
			ExpectedCloseDate: toPtrTime(now.Add(14 * 24 * time.Hour)),
			OwnerUserID:       "user-1",
			Summary:           "Annual platform subscription and integration service.",
			Tags:              "cloud,integration",
			CreatedAt:         now,
			UpdatedAt:         now,
		},
		{
			ID:                "so-2",
			OrderNumber:       "SO-2026-002",
			CustomerName:      "BioGenix",
			Amount:            42000,
			Currency:          "USD",
			Stage:             "closed_won",
			Status:            "active",
			ExpectedCloseDate: toPtrTime(now.Add(-2 * 24 * time.Hour)),
			OwnerUserID:       "user-1",
			Summary:           "Q2 expansion project for logistics module.",
			Tags:              "retail,logistics",
			CreatedAt:         now,
			UpdatedAt:         now,
		},
		{
			ID:                "so-3",
			OrderNumber:       "SO-2026-003",
			CustomerName:      "InnoMed Labs",
			Amount:            25000,
			Currency:          "USD",
			Stage:             "proposal",
			Status:            "active",
			ExpectedCloseDate: toPtrTime(now.Add(45 * 24 * time.Hour)),
			OwnerUserID:       "user-2",
			Summary:           "New R&D workflow implementation.",
			Tags:              "medical,rnd",
			CreatedAt:         now,
			UpdatedAt:         now,
		},
	}

	for _, o := range orders {
		db.DB.Create(&o)
	}
}

func toPtrTime(t time.Time) *time.Time {
	return &t
}
