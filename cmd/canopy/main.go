package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/MaplesMcDepth/canopy/pkg/alerts"
	"github.com/MaplesMcDepth/canopy/pkg/budget"
	"github.com/MaplesMcDepth/canopy/pkg/dashboard"
	"github.com/MaplesMcDepth/canopy/pkg/interceptor"
	"github.com/MaplesMcDepth/canopy/pkg/models"
	"github.com/MaplesMcDepth/canopy/pkg/reports"
	"github.com/MaplesMcDepth/canopy/pkg/store"
)

// Version and build info (set via ldflags)
var (
	version   = "dev"
	commit    = "none"
	date      = "unknown"
	builtBy   = "unknown"
)

// Commands
const (
	cmdServe     = "serve"
	cmdReport    = "report"
	cmdVersion   = "version"
)

func main() {
	// Define flags
	dataSource := flag.String("dsn", "canopy.db", "Data source name for SQLite database")
	listenAddr := flag.String("addr", ":8080", "Address to listen on for the dashboard")
	command := flag.String("cmd", "serve", fmt.Sprintf("Command to run (%s, %s, %s)", cmdServe, cmdReport, cmdVersion))
	reportType := flag.String("report", "daily", "Type of report to generate (daily, weekly, monthly)")
	flag.Parse()

	// Initialize store
	sqliteStore, err := store.NewSQLiteStore(*dataSource)
	if err != nil {
		log.Fatalf("Failed to initialize store: %v", err)
	}
	defer sqliteStore.Close()

	// Initialize managers
	budgetMgr := budget.NewManager(sqliteStore)
	alertSender := &alerts.LoggerAlertSender{}
	alertMgr := alerts.NewManager(sqliteStore, alertSender)

	// For demonstration, we'll use a simple cost calculator that returns a fixed cost.
	// In a real implementation, this would parse the response and calculate based on token usage.
	costCalculator := func(req *http.Request, resp *http.Response) (float64, error) {
		// Placeholder: In reality, we'd extract token usage and calculate cost per model.
		// For now, we return a fixed small cost to demonstrate.
		return 0.002, nil
	}

	// Create interceptor
	apiInterceptor := &interceptor.Interceptor{
		Store:        sqliteStore,
		CostCalculator: costCalculator,
		RoundTripper: http.DefaultTransport,
	}

	// Create dashboard server
	dashServer := dashboard.NewServer(sqliteStore, budgetMgr, alertMgr)

	// Create report generator
	reportGen := reports.NewGenerator(sqliteStore)

	// Handle commands
	switch *command {
	case cmdServe:
		// Start the dashboard server
		log.Printf("Starting Canopy dashboard on %s", *listenAddr)
		log.Printf("Version: %s (commit: %s, built by %s on %s)", version, commit, builtBy, date)
		if err := dashServer.Start(*listenAddr); err != nil {
			log.Fatalf("Failed to start server: %v", err)
		}
	case cmdReport:
		// Generate a report
		var err error
		switch *reportType {
		case "daily":
			err = reportGen.GenerateDailyReport(os.Stdout)
		case "weekly":
			err = reportGen.GenerateWeeklyReport(os.Stdout)
		case "monthly":
			err = reportGen.GenerateMonthlyReport(os.Stdout)
		default:
			log.Fatalf("Unknown report type: %s", *reportType)
		}
		if err != nil {
			log.Fatalf("Failed to generate report: %v", err)
		}
	case cmdVersion:
		fmt.Printf("Canopy version %s\n", version)
		fmt.Printf("Commit: %s\n", commit)
		fmt.Printf("Built by: %s\n", builtBy)
		fmt.Printf("Built on: %s\n", date)
	default:
		log.Fatalf("Unknown command: %s. Use %s, %s, or %s", *command, cmdServe, cmdReport, cmdVersion)
	}
}