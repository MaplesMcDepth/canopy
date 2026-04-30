package store

import (
	"database/sql"
	"fmt"
	"time"

	_ "github.com/mattn/go-sqlite3"

	"github.com/MaplesMcDepth/canopy/pkg/models"
)

// SQLiteStore implements the DBStore interface using SQLite.
type SQLiteStore struct {
	db *sql.DB
}

// NewSQLiteStore creates a new SQLite store and initializes the database schema.
func NewSQLiteStore(dataSourceName string) (*SQLiteStore, error) {
	db, err := sql.Open("sqlite3", dataSourceName)
	if err != nil {
		return nil, fmt.Errorf("opening sqlite database: %w", err)
	}
	// Test the connection
	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("pinging sqlite database: %w", err)
	}

	// Create tables if they don't exist
	if err := createTables(db); err != nil {
		return nil, fmt.Errorf("creating tables: %w", err)
	}

	return &SQLiteStore{db: db}, nil
}

// Close closes the database connection.
func (s *SQLiteStore) Close() error {
	return s.db.Close()
}

// createTables creates the necessary tables if they don't exist.
func createTables(db *sql.DB) error {
	// APICalls table
	apiCallsTable := `
	CREATE TABLE IF NOT EXISTS api_calls (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		timestamp DATETIME NOT NULL,
		api_key TEXT NOT NULL,
		endpoint TEXT NOT NULL,
		model TEXT,
		tokens_input INTEGER DEFAULT 0,
		tokens_output INTEGER DEFAULT 0,
		cost REAL NOT NULL
	);`
	if _, err := db.Exec(apiCallsTable); err != nil {
		return err
	}

	// Budgets table
	budgetsTable := `
	CREATE TABLE IF NOT EXISTS budgets (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		name TEXT NOT NULL,
		limit REAL NOT NULL,
		period TEXT NOT NULL,
		api_key TEXT,
		current_usage REAL DEFAULT 0.0,
		alert_thresholds TEXT, -- JSON array of floats
		created_at DATETIME NOT NULL,
		updated_at DATETIME NOT NULL
	);`
	if _, err := db.Exec(budgetsTable); err != nil {
		return err
	}

	// Alerts table
	alertsTable := `
	CREATE TABLE IF NOT EXISTS alerts (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		budget_id INTEGER NOT NULL,
		threshold REAL NOT NULL,
		timestamp DATETIME NOT NULL,
		sent BOOLEAN NOT NULL DEFAULT 0,
		FOREIGN KEY (budget_id) REFERENCES budgets(id)
	);`
	if _, err := db.Exec(alertsTable); err != nil {
		return err
	}

	// Create indexes for better performance
	if _, err := db.Exec(`CREATE INDEX IF NOT EXISTS idx_api_calls_timestamp ON api_calls(timestamp);`); err != nil {
		return err
	}
	if _, err := db.Exec(`CREATE INDEX IF NOT EXISTS idx_budgets_api_key ON budgets(api_key);`); err != nil {
		return err
	}
	if _, err := db.Exec(`CREATE INDEX IF NOT EXISTS idx_alerts_budget_id ON alerts(budget_id);`); err != nil {
		return err
	}
	if _, err := db.Exec(`CREATE INDEX IF NOT EXISTS idx_alerts_sent ON alerts(sent);`); err != nil {
		return err
	}

	return nil
}

// CreateAPICall stores a new API call.
func (s *SQLiteStore) CreateAPICall(call *models.APICall) error {
	result, err := s.db.Exec(
		`INSERT INTO api_calls (timestamp, api_key, endpoint, model, tokens_input, tokens_output, cost) 
		 VALUES (?, ?, ?, ?, ?, ?, ?)`,
		call.Timestamp, call.APIKey, call.Endpoint, call.Model, call.TokensInput, call.TokensOutput, call.Cost,
	)
	if err != nil {
		return err
	}
	id, err := result.LastInsertId()
	if err != nil {
		return err
	}
	call.ID = id
	return nil
}

// GetAPICallsSince returns all API calls since the given time.
func (s *SQLiteStore) GetAPICallsSince(since time.Time) ([]models.APICall, error) {
	rows, err := s.db.Query(
		`SELECT id, timestamp, api_key, endpoint, model, tokens_input, tokens_output, cost 
		 FROM api_calls 
		 WHERE timestamp >= ? 
		 ORDER BY timestamp DESC`,
		since,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var calls []models.APICall
	for rows.Next() {
		var call models.APICall
		if err := rows.Scan(
			&call.ID, &call.Timestamp, &call.APIKey, &call.Endpoint, &call.Model,
			&call.TokensInput, &call.TokensOutput, &call.Cost,
		); err != nil {
			return nil, err
		}
		calls = append(calls, call)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return calls, nil
}

// CreateBudget stores a new budget.
func (s *SQLiteStore) CreateBudget(budget *models.Budget) error {
	result, err := s.db.Exec(
		`INSERT INTO budgets (name, limit, period, api_key, current_usage, alert_thresholds, created_at, updated_at) 
		 VALUES (?, ?, ?, ?, ?, ?, ?, ?)`,
		budget.Name, budget.Limit, budget.Period, budget.APIKey, budget.CurrentUsage,
		// We'll store alert_thresholds as a JSON string for simplicity
		"[]", budget.CreatedAt, budget.UpdatedAt,
	)
	if err != nil {
		return err
	}
	id, err := result.LastInsertId()
	if err != nil {
		return err
	}
	budget.ID = id
	return nil
}

// GetBudget retrieves a budget by ID.
func (s *SQLiteStore) GetBudget(id int64) (*models.Budget, error) {
	row := s.db.QueryRow(
		`SELECT id, name, limit, period, api_key, current_usage, alert_thresholds, created_at, updated_at 
		 FROM budgets 
		 WHERE id = ?`,
		id,
	)
	var budget models.Budget
	var alertThresholds string
	if err := row.Scan(
		&budget.ID, &budget.Name, &budget.Limit, &budget.Period, &budget.APIKey,
		&budget.CurrentUsage, &alertThresholds, &budget.CreatedAt, &budget.UpdatedAt,
	); err != nil {
		return nil, err
	}
	// For now, we leave AlertThresholds empty; in a real app, we'd parse the JSON.
	// We'll set a default if empty.
	if alertThresholds == "" {
		budget.AlertThresholds = []float64{0.5, 0.8, 0.95}
	} else {
		// TODO: parse JSON string into []float64
		// For simplicity, we'll use the default.
		budget.AlertThresholds = []float64{0.5, 0.8, 0.95}
	}
	return &budget, nil
}

// UpdateBudget updates an existing budget.
func (s *SQLiteStore) UpdateBudget(budget *models.Budget) error {
	_, err := s.db.Exec(
		`UPDATE budgets 
		 SET name = ?, limit = ?, period = ?, api_key = ?, current_usage = ?, 
		     alert_thresholds = ?, updated_at = ?
		 WHERE id = ?`,
		budget.Name, budget.Limit, budget.Period, budget.APIKey, budget.CurrentUsage,
		"[]", budget.UpdatedAt, budget.ID,
	)
	return err
}

// ListBudgets returns all budgets.
func (s *SQLiteStore) ListBudgets() ([]models.Budget, error) {
	rows, err := s.db.Query(
		`SELECT id, name, limit, period, api_key, current_usage, alert_thresholds, created_at, updated_at 
		 FROM budgets 
		 ORDER BY created_at DESC`,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var budgets []models.Budget
	for rows.Next() {
		var budget models.Budget
		var alertThresholds string
		if err := rows.Scan(
			&budget.ID, &budget.Name, &budget.Limit, &budget.Period, &budget.APIKey,
			&budget.CurrentUsage, &alertThresholds, &budget.CreatedAt, &budget.UpdatedAt,
		); err != nil {
			return nil, err
		}
		if alertThresholds == "" {
			budget.AlertThresholds = []float64{0.5, 0.8, 0.95}
		} else {
			// TODO: parse JSON
			budget.AlertThresholds = []float64{0.5, 0.8, 0.95}
		}
		budgets = append(budgets, budget)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return budgets, nil
}

// CreateAlert stores a new alert.
func (s *SQLiteStore) CreateAlert(alert *models.Alert) error {
	result, err := s.db.Exec(
		`INSERT INTO alerts (budget_id, threshold, timestamp, sent) 
		 VALUES (?, ?, ?, ?)`,
		alert.BudgetID, alert.Threshold, alert.Timestamp, alert.Sent,
	)
	if err != nil {
		return err
	}
	id, err := result.LastInsertId()
	if err != nil {
		return err
	}
	alert.ID = id
	return nil
}

// GetUnsentAlerts returns alerts that have not been sent.
func (s *SQLiteStore) GetUnsentAlerts() ([]models.Alert, error) {
	rows, err := s.db.Query(
		`SELECT id, budget_id, threshold, timestamp, sent 
		 FROM alerts 
		 WHERE sent = 0 
		 ORDER BY timestamp ASC`,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var alerts []models.Alert
	for rows.Next() {
		var alert models.Alert
		if err := rows.Scan(
			&alert.ID, &alert.BudgetID, &alert.Threshold, &alert.Timestamp, &alert.Sent,
		); err != nil {
			return nil, err
		}
		alerts = append(alerts, alert)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return alerts, nil
}

// MarkAlertSent marks an alert as sent.
func (s *SQLiteStore) MarkAlertSent(id int64) error {
	_, err := s.db.Exec(
		`UPDATE alerts SET sent = 1 WHERE id = ?`,
		id,
	)
	return err
}

// GetTotalCostSince returns the total cost of API calls since the given time.
func (s *SQLiteStore) GetTotalCostSince(since time.Time) (float64, error) {
	var total float64
	err := s.db.QueryRow(
		`SELECT SUM(cost) FROM api_calls WHERE timestamp >= ?`,
		since,
	).Scan(&total)
	if err != nil {
		return 0, err
	}
	if total == 0 {
		return 0, nil
	}
	return total, nil
}