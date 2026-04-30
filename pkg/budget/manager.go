package budget

import (
	"time"

	"github.com/MaplesMcDepth/canopy/pkg/models"
)

// Manager handles budget creation, updating, and checking.
type Manager struct {
	Store models.DBStore
}

// NewManager creates a new Budget Manager.
func NewManager(store models.DBStore) *Manager {
	return &Manager{Store: store}
}

// CreateBudget creates a new budget.
func (m *Manager) CreateBudget(name string, limit float64, period string, apiKey string) (*models.Budget, error) {
	budget := &models.Budget{
		Name:          name,
		Limit:         limit,
		Period:        period,
		APIKey:        apiKey,
		CurrentUsage:  0.0,
		AlertThresholds: []float64{0.5, 0.8, 0.95}, // Default thresholds
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
	}
	if err := m.Store.CreateBudget(budget); err != nil {
		return nil, err
	}
	return budget, nil
}

// GetBudget retrieves a budget by ID.
func (m *Manager) GetBudget(id int64) (*models.Budget, error) {
	return m.Store.GetBudget(id)
}

// UpdateBudget updates an existing budget.
func (m *Manager) UpdateBudget(budget *models.Budget) error {
	budget.UpdatedAt = time.Now()
	return m.Store.UpdateBudget(budget)
}

// ListBudgets returns all budgets.
func (m *Manager) ListBudgets() ([]models.Budget, error) {
	return m.Store.ListBudgets()
}

// AddUsage adds cost to a budget's current usage and checks for threshold alerts.
// Returns a slice of thresholds that were crossed in this call.
func (m *Manager) AddUsage(budgetID int64, cost float64) ([]float64, error) {
	budget, err := m.Store.GetBudget(budgetID)
	if err != nil {
		return nil, err
	}

	// Update usage
	budget.CurrentUsage += cost

	// Check which thresholds are crossed
	var crossed []float64
	for _, threshold := range budget.AlertThresholds {
		if budget.CurrentUsage/budget.Limit >= threshold {
			crossed = append(crossed, threshold)
		}
	}

	// Update budget in store
	if err := m.Store.UpdateBudget(budget); err != nil {
		return nil, err
	}

	return crossed, nil
}

// IsWithinBudget checks if adding a cost would exceed the budget.
func (m *Manager) IsWithinBudget(budgetID int64, cost float64) (bool, error) {
	budget, err := m.Store.GetBudget(budgetID)
	if err != nil {
		return false, err
	}
	return (budget.CurrentUsage + cost) <= budget.Limit, nil
}

// ResetBudget resets the current usage of a budget (e.g., at the start of a new period).
func (m *Manager) ResetBudget(id int64) error {
	budget, err := m.Store.GetBudget(id)
	if err != nil {
		return err
	}
	budget.CurrentUsage = 0.0
	budget.UpdatedAt = time.Now()
	return m.Store.UpdateBudget(budget)
}