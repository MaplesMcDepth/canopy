package models

import (
	"time"
)

// APICall represents a single API call made to an AI service.
type APICall struct {
	ID          int64     `json:"id"`
	Timestamp   time.Time `json:"timestamp"`
	APIKey      string    `json:"api_key"` // hashed or masked for privacy
	Endpoint    string    `json:"endpoint"` // e.g., "https://api.openai.com/v1/chat/completions"
	Model       string    `json:"model"`    // e.g., "gpt-4"
	TokensInput int       `json:"tokens_input"`
	TokensOutput int      `json:"tokens_output"`
	Cost        float64   `json:"cost"`     // in USD
}

// Budget represents a spending limit for a specific API key, project, or team.
type Budget struct {
	ID            int64     `json:"id"`
	Name          string    `json:"name"`
	Limit         float64   `json:"limit"` // in USD
	Period        string    `json:"period"` // "daily", "weekly", "monthly"
	APIKey        string    `json:"api_key,omitempty"` // if empty, applies to all keys
	CurrentUsage  float64   `json:"current_usage"`
	AlertThresholds []float64 `json:"alert_thresholds"` // e.g., [0.5, 0.8, 0.95] for 50%, 80%, 95%
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}

// Alert represents a notification that a budget threshold has been reached.
type Alert struct {
	ID        int64     `json:"id"`
	BudgetID  int64     `json:"budget_id"`
	Threshold float64   `json:"threshold"` // fraction of budget limit that triggered the alert
	Timestamp time.Time `json:"timestamp"`
	Sent      bool      `json:"sent"` // whether the alert has been sent out
}

// DBStore defines the interface for persisting Canopy data.
type DBStore interface {
	Close() error
	CreateAPICall(call *APICall) error
	GetAPICallsSince(since time.Time) ([]APICall, error)
	CreateBudget(budget *Budget) error
	GetBudget(id int64) (*Budget, error)
	UpdateBudget(budget *Budget) error
	ListBudgets() ([]Budget, error)
	CreateAlert(alert *Alert) error
	GetUnsentAlerts() ([]Alert, error)
	MarkAlertSent(id int64) error
	GetTotalCostSince(since time.Time) (float64, error)
}