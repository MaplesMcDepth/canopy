package reports

import (
	"encoding/csv"
	"fmt"
	"io"
	"time"

	"github.com/MaplesMcDepth/canopy/pkg/models"
)

// Generator creates cost reports in various formats.
type Generator struct {
	Store models.DBStore
}

// NewGenerator creates a new report generator.
func NewGenerator(store models.DBStore) *Generator {
	return &Generator{Store: store}
}

// GenerateDailyReport creates a CSV report for the last 24 hours.
func (g *Generator) GenerateDailyReport(w io.Writer) error {
	since := time.Now().Add(-24 * time.Hour)
	return g.generateReport(w, since, "Daily")
}

// GenerateWeeklyReport creates a CSV report for the last 7 days.
func (g *Generator) GenerateWeeklyReport(w io.Writer) error {
	since := time.Now().Add(-7 * 24 * time.Hour)
	return g.generateReport(w, since, "Weekly")
}

// GenerateMonthlyReport creates a CSV report for the last 30 days.
func (g *Generator) GenerateMonthlyReport(w io.Writer) error {
	since := time.Now().Add(-30 * 24 * time.Hour)
	return g.generateReport(w, since, "Monthly")
}

// generateReport creates a CSV report for the given time period.
func (g *Generator) generateReport(w io.Writer, since time.Time, period string) error {
	calls, err := g.Store.GetAPICallsSince(since)
	if err != nil {
		return err
	}

	writer := csv.NewWriter(w)
	defer writer.Flush()

	// Write header
	header := []string{
		"Timestamp", "API Key", "Endpoint", "Model", 
		"Input Tokens", "Output Tokens", "Cost (USD)",
	}
	if err := writer.Write(header); err != nil {
		return err
	}

	// Write data rows
	totalCost := 0.0
	for _, call := range calls {
		row := []string{
			call.Timestamp.Format(time.RFC3339),
			call.APIKey,
			call.Endpoint,
			call.Model,
			fmt.Sprintf("%d", call.TokensInput),
			fmt.Sprintf("%d", call.TokensOutput),
			fmt.Sprintf("%.6f", call.Cost),
		}
		if err := writer.Write(row); err != nil {
			return err
		}
		totalCost += call.Cost
	}

	// Write summary row
	summaryRow := []string{
		"TOTAL", "", "", "", "", "", fmt.Sprintf("%.6f", totalCost),
	}
	if err := writer.Write(summaryRow); err != nil {
		return err
	}

	return nil
}

// GenerateSummary creates a text summary of costs.
func (g *Generator) GenerateSummary(w io.Writer, since time.Time) error {
	calls, err := g.Store.GetAPICallsSince(since)
	if err != nil {
		return err
	}

	totalCost := 0.0
	modelCosts := make(map[string]float64)
	apiKeyCosts := make(map[string]float64)

	for _, call := range calls {
		totalCost += call.Cost
		modelCosts[call.Model] += call.Cost
		apiKeyCosts[call.APIKey] += call.Cost
	}

	fmt.Fprintf(w, "Canopy Cost Summary\n")
	fmt.Fprintf(w, "Period: %s to %s\n", since.Format(time.RFC3339), time.Now().Format(time.RFC3339))
	fmt.Fprintf(w, "Total API Calls: %d\n", len(calls))
	fmt.Fprintf(w, "Total Cost: $%.6f\n\n", totalCost)

	fmt.Fprintf(w, "Cost by Model:\n")
	for model, cost := range modelCosts {
		fmt.Fprintf(w, "  %s: $%.6f\n", model, cost)
	}
	fmt.Fprintf(w, "\n")

	fmt.Fprintf(w, "Cost by API Key:\n")
	for apiKey, cost := range apiKeyCosts {
		fmt.Fprintf(w, "  %s: $%.6f\n", apiKey, cost)
	}

	return nil
}