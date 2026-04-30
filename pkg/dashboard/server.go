package dashboard

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/MaplesMcDepth/canopy/pkg/models"
	"github.com/MaplesMcDepth/canopy/pkg/budget"
	"github.com/MaplesMcDepth/canopy/pkg/alerts"
)

// Server handles the web dashboard for Canopy.
type Server struct {
	Store      models.DBStore
	BudgetMgr  *budget.Manager
	AlertMgr   *alerts.Manager
	Router     http.ServeMux
}

// NewServer creates a new dashboard server.
func NewServer(store models.DBStore, budgetMgr *budget.Manager, alertMgr *alerts.Manager) *Server {
	s := &Server{
		Store:      store,
		BudgetMgr:  budgetMgr,
		AlertMgr:   alertMgr,
		Router:     *http.NewServeMux(),
	}
	s.setupRoutes()
	return s
}

// setupRoutes defines the HTTP endpoints for the dashboard.
func (s *Server) setupRoutes() {
	s.Router.HandleFunc("/", s.handleIndex)
	s.Router.HandleFunc("/api/budgets", s.handleBudgets)
	s.Router.HandleFunc("/api/recent-calls", s.handleRecentCalls)
	s.Router.HandleFunc("/api/alerts", s.handleAlerts)
	s.Router.HandleFunc("/api/total-cost", s.handleTotalCost)
}

// Start begins serving the dashboard on the specified address.
func (s *Server) Start(addr string) error {
	return http.ListenAndServe(addr, &s.Router)
}

// handleIndex serves the main dashboard HTML page.
func (s *Server) handleIndex(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "web/index.html")
}

// handleBudgets returns JSON with all budgets and their current usage.
func (s *Server) handleBudgets(w http.ResponseWriter, r *http.Request) {
	budgets, err := s.BudgetMgr.ListBudgets()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(budgets)
}

// handleRecentCalls returns recent API calls.
func (s *Server) handleRecentCalls(w http.ResponseWriter, r *http.Request) {
	// Get calls from the last 24 hours
	since := time.Now().Add(-24 * time.Hour)
	calls, err := s.Store.GetAPICallsSince(since)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(calls)
}

// handleAlerts returns unsent alerts.
func (s *Server) handleAlerts(w http.ResponseWriter, r *http.Request) {
	alerts, err := s.AlertMgr.GetUnsentAlerts()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(alerts)
}

// handleTotalCost returns total cost since a given time (default: 24 hours).
func (s *Server) handleTotalCost(w http.ResponseWriter, r *http.Request) {
	// Default to last 24 hours
	since := time.Now().Add(-24 * time.Hour)
	if r.URL.Query().Get("since") != "" {
		// Parse custom since time (simplified)
		if t, err := time.Parse(time.RFC3339, r.URL.Query().Get("since")); err == nil {
			since = t
		}
	}
	totalCost, err := s.Store.GetTotalCostSince(since)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]float64{"total_cost": totalCost})
}