package alerts

import (
	"fmt"
	"time"

	"github.com/MaplesMcDepth/canopy/pkg/models"
)

// AlertSender defines the interface for sending alerts.
type AlertSender interface {
	SendAlert(alert *models.Alert) error
}

// LoggerAlertSender logs alerts to standard output.
type LoggerAlertSender struct{}

// SendAlert logs the alert.
func (l *LoggerAlertSender) SendAlert(alert *models.Alert) error {
	fmt.Printf("[ALERT] Budget %.0f%% exceeded at %v: Budget ID %d\n",
		alert.Threshold*100, alert.Timestamp.Format(time.RFC3339), alert.BudgetID)
	return nil
}

// Manager handles alert creation and sending.
type Manager struct {
	Store    models.DBStore
	Sender   AlertSender
}

// NewManager creates a new Alert Manager.
func NewManager(store models.DBStore, sender AlertSender) *Manager {
	return &Manager{Store: store, Sender: sender}
}

// CreateAlert creates a new alert record.
func (a *Manager) CreateAlert(budgetID int64, threshold float64) (*models.Alert, error) {
	alert := &models.Alert{
		BudgetID:  budgetID,
		Threshold: threshold,
		Timestamp: time.Now(),
		Sent:      false,
	}
	if err := a.Store.CreateAlert(alert); err != nil {
		return nil, err
	}
	return alert, nil
}

// GetUnsentAlerts returns alerts that haven't been sent yet.
func (a *Manager) GetUnsentAlerts() ([]models.Alert, error) {
	return a.Store.GetUnsentAlerts()
}

// SendUnsentAlerts sends all unsent alerts and marks them as sent.
func (a *Manager) SendUnsentAlerts() error {
	alerts, err := a.Store.GetUnsentAlerts()
	if err != nil {
		return err
	}
	for _, alert := range alerts {
		if err := a.Sender.SendAlert(&alert); err != nil {
			return err
		}
		if err := a.Store.MarkAlertSent(alert.ID); err != nil {
			return err
		}
	}
	return nil
}