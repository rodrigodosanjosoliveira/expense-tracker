package domain

import "time"

// NotificationChannel representa o canal de notificação
type NotificationChannel string

const (
	ChannelSlack   NotificationChannel = "slack"
	ChannelDiscord NotificationChannel = "discord"
	ChannelWebhook NotificationChannel = "webhook"
)

// NotificationPriority representa a prioridade da notificação
type NotificationPriority string

const (
	PriorityLow    NotificationPriority = "low"
	PriorityMedium NotificationPriority = "medium"
	PriorityHigh   NotificationPriority = "high"
)

// Notification representa uma notificação a ser enviada
type Notification struct {
	ID        string                 `json:"id"`
	UserID    string                 `json:"user_id"`
	Channel   NotificationChannel    `json:"channel"`
	Priority  NotificationPriority   `json:"priority"`
	Title     string                 `json:"title"`
	Message   string                 `json:"message"`
	Data      map[string]interface{} `json:"data,omitempty"`
	SentAt    *time.Time             `json:"sent_at,omitempty"`
	Error     string                 `json:"error,omitempty"`
	CreatedAt time.Time              `json:"created_at"`
}

// UserAlertSettings representa configurações de alertas do usuário
type UserAlertSettings struct {
	UserID                  string    `json:"user_id"`
	DailyLimit              float64   `json:"daily_limit"`
	WeeklyLimit             float64   `json:"weekly_limit"`
	MonthlyLimit            float64   `json:"monthly_limit"`
	HighExpenseThreshold    float64   `json:"high_expense_threshold"` // Valor acima do qual é considerado despesa alta
	EnableDailyAlerts       bool      `json:"enable_daily_alerts"`
	EnableWeeklyAlerts      bool      `json:"enable_weekly_alerts"`
	EnableMonthlyAlerts     bool      `json:"enable_monthly_alerts"`
	EnableHighExpenseAlerts bool      `json:"enable_high_expense_alerts"`
	UpdatedAt               time.Time `json:"updated_at"`
}

// IsLimitReached verifica se um limite foi atingido
func (s *UserAlertSettings) IsLimitReached(period string, currentTotal float64) bool {
	switch period {
	case "daily":
		return s.EnableDailyAlerts && s.DailyLimit > 0 && currentTotal >= s.DailyLimit
	case "weekly":
		return s.EnableWeeklyAlerts && s.WeeklyLimit > 0 && currentTotal >= s.WeeklyLimit
	case "monthly":
		return s.EnableMonthlyAlerts && s.MonthlyLimit > 0 && currentTotal >= s.MonthlyLimit
	default:
		return false
	}
}

// IsHighExpense verifica se uma despesa é considerada alta
func (s *UserAlertSettings) IsHighExpense(amount float64) bool {
	return s.EnableHighExpenseAlerts && s.HighExpenseThreshold > 0 && amount >= s.HighExpenseThreshold
}
