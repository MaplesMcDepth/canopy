# Canopy - AI Agent Cost Guardian

Canopy is a lightweight, high-performance tool designed to monitor and control AI API costs. It helps prevent budget overruns by intercepting API calls, tracking expenses, setting budgets, and providing real-time alerts and dashboards.

## Features

- **API Interceptor**: Middleware that wraps AI API calls to track usage and costs.
- **Budget Manager**: Set spending limits per API key, project, or team with flexible periods (daily, weekly, monthly).
- **Real-time Dashboard**: Web UI showing current spend, budgets, and recent API calls.
- **Alert System**: Notifications when budget thresholds are reached.
- **Auto-fallback**: (Planned) Switch to local models when budget is exceeded.
- **Cost Reports**: Generate daily, weekly, and monthly cost reports in CSV or text format.

## Architecture

- **Language**: Go (for performance and concurrency)
- **Database**: SQLite (for simplicity and zero-configuration deployment)
- **Dashboard**: Simple HTML/JavaScript served by Go HTTP server
- **Modular Design**: Separated packages for interceptor, budgeting, alerts, dashboard, reports, and storage.

## Packages

- `cmd/canopy`: Main application entry point
- `pkg/interceptor`: HTTP middleware to intercept and track API calls
- `pkg/budget`: Budget creation, tracking, and threshold checking
- `pkg/alerts`: Alert generation and notification (currently logs to stdout)
- `pkg/dashboard`: Web server for the real-time dashboard
- `pkg/reports`: Report generation (CSV and text summaries)
- `pkg/models`: Data structures for API calls, budgets, and alerts
- `pkg/store`: SQLite implementation of the storage interface
- `web`: Static assets for the dashboard

## Installation

### Prerequisites

- Go 1.18 or higher
- SQLite3 (via CGO, included in the standard library with the go-sqlite3 driver)

### Build from Source

```bash
git clone https://github.com/MaplesMcDepth/canopy.git
cd canopy
go build -o canopy ./cmd/canopy
```

## Usage

### Running the Dashboard Server

```bash
./canopy -cmd=serve -dsn="canopy.db" -addr=":8080"
```

Then open your browser to `http://localhost:8080` to see the dashboard.

### Generating Reports

```bash
# Daily report
./canopy -cmd=report -report=daily

# Weekly report
./canopy -cmd=report -report=weekly

# Monthly report
./canopy -cmd=report -report=monthly
```

### Checking Version

```bash
./canopy -cmd=version
```

## Configuration

Canopy is configured via command-line flags:

- `-dsn`: Data source name for SQLite database (default: "canopy.db")
- `-addr`: Address to listen on for the dashboard (default: ":8080")
- `-cmd`: Command to run (`serve`, `report`, `version`) (default: "serve")
- `-report`: Type of report to generate (`daily`, `weekly`, `monthly`) (default: "daily")

## How It Works

1. **API Interceptor**: Wrap your AI API calls with Canopy's interceptor. It tracks each call, calculates cost (based on token usage for supported APIs), and stores the data.
2. **Budget Management**: Define budgets with limits and periods. Canopy automatically tracks spending against these budgets.
3. **Alerts**: When spending crosses defined thresholds (e.g., 50%, 80%, 95% of budget), alerts are generated.
4. **Dashboard**: View real-time spending, budget status, and recent API calls.
5. **Reports**: Generate detailed reports for analysis and record-keeping.

## Extending Canopy

- **Custom Cost Calculation**: Modify the `costCalculator` function in `main.go` to implement pricing models for different AI providers.
- **Alert Notifications**: Implement the `alerts.AlertSender` interface to send alerts via email, Slack, SMS, etc.
- **Storage Backends**: Implement the `models.DBStore` interface to use PostgreSQL, MySQL, or other databases.

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## Acknowledgments

- Inspired by the need for better cost visibility in AI-powered applications.
- Built with Go for its performance and simplicity in creating concurrent services.