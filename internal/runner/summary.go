package runner

import (
	"fmt"
	"os"
	"strings"
)

const (
	// ANSI color codes
	colorReset  = "\033[0m"
	colorGreen  = "\033[32m"
	colorYellow = "\033[33m"
	colorRed    = "\033[31m"
)

// PrintSummary displays a formatted summary table of module execution results.
// The table shows each module's name, status, and error details (if any).
// Status indicators are color-coded: green for installed, yellow for skipped, red for failed/error.
// A summary line shows total counts for each status.
// If dryRun is true, a dry-run indicator is displayed.
func PrintSummary(results []ModuleResult, dryRun bool) {
	if len(results) == 0 {
		return
	}

	// Calculate column widths
	maxNameLen := len("Module")
	maxStatusLen := len("Status")
	maxDetailsLen := len("Details")

	for _, result := range results {
		if len(result.Name) > maxNameLen {
			maxNameLen = len(result.Name)
		}
		statusStr := getStatusString(result.Status)
		if len(statusStr) > maxStatusLen {
			maxStatusLen = len(statusStr)
		}
		if result.Error != nil {
			errorMsg := result.Error.Error()
			// Truncate long error messages for display
			if len(errorMsg) > 50 {
				errorMsg = errorMsg[:47] + "..."
			}
			if len(errorMsg) > maxDetailsLen {
				maxDetailsLen = len(errorMsg)
			}
		}
	}

	// Ensure minimum widths
	if maxNameLen < 6 {
		maxNameLen = 6
	}
	if maxStatusLen < 6 {
		maxStatusLen = 6
	}
	if maxDetailsLen < 7 {
		maxDetailsLen = 7
	}

	// Print separator before summary
	fmt.Fprintf(os.Stdout, "\n")

	// Print table header
	headerLine := fmt.Sprintf("┌─%s─┬─%s─┬─%s─┐",
		strings.Repeat("─", maxNameLen),
		strings.Repeat("─", maxStatusLen),
		strings.Repeat("─", maxDetailsLen))
	fmt.Fprintf(os.Stdout, "%s\n", headerLine)

	fmt.Fprintf(os.Stdout, "│ %-*s │ %-*s │ %-*s │\n",
		maxNameLen, "Module",
		maxStatusLen, "Status",
		maxDetailsLen, "Details")

	separatorLine := fmt.Sprintf("├─%s─┼─%s─┼─%s─┤",
		strings.Repeat("─", maxNameLen),
		strings.Repeat("─", maxStatusLen),
		strings.Repeat("─", maxDetailsLen))
	fmt.Fprintf(os.Stdout, "%s\n", separatorLine)

	// Count totals
	var installedCount, skippedCount, failedCount, wouldInstallCount int

	// Print table rows
	for _, result := range results {
		statusStr := getStatusString(result.Status)
		color := getStatusColor(result.Status)

		var details string
		if result.Error != nil {
			errorMsg := result.Error.Error()
			// Truncate long error messages
			if len(errorMsg) > 50 {
				errorMsg = errorMsg[:47] + "..."
			}
			details = errorMsg
		}

		// Print row with color
		fmt.Fprintf(os.Stdout, "│ %-*s │ %s%-*s%s │ %-*s │\n",
			maxNameLen, result.Name,
			color, maxStatusLen, statusStr, colorReset,
			maxDetailsLen, details)

		// Count totals
		switch result.Status {
		case StatusInstalled:
			installedCount++
		case StatusSkipped:
			skippedCount++
		case StatusWouldInstall:
			wouldInstallCount++
		case StatusFailed, StatusError:
			failedCount++
		}
	}

	// Print table footer
	footerLine := fmt.Sprintf("└─%s─┴─%s─┴─%s─┘",
		strings.Repeat("─", maxNameLen),
		strings.Repeat("─", maxStatusLen),
		strings.Repeat("─", maxDetailsLen))
	fmt.Fprintf(os.Stdout, "%s\n", footerLine)

	// Print summary totals
	summaryParts := []string{}
	if installedCount > 0 {
		summaryParts = append(summaryParts, fmt.Sprintf("%d installed", installedCount))
	}
	if wouldInstallCount > 0 {
		summaryParts = append(summaryParts, fmt.Sprintf("%d would install", wouldInstallCount))
	}
	if skippedCount > 0 {
		summaryParts = append(summaryParts, fmt.Sprintf("%d skipped", skippedCount))
	}
	if failedCount > 0 {
		summaryParts = append(summaryParts, fmt.Sprintf("%d failed", failedCount))
	}

	if len(summaryParts) > 0 {
		summaryLine := fmt.Sprintf("Summary: %s", strings.Join(summaryParts, ", "))
		if dryRun {
			summaryLine += " (dry-run)"
		}
		fmt.Fprintf(os.Stdout, "%s\n", summaryLine)
	} else {
		fmt.Fprintf(os.Stdout, "Summary: No modules processed\n")
	}

	// Print separator after summary
	fmt.Fprintf(os.Stdout, "\n")
}

// getStatusString returns the formatted status string with symbol.
func getStatusString(status ModuleStatus) string {
	switch status {
	case StatusInstalled:
		return "✓ Installed"
	case StatusSkipped:
		return "⊘ Skipped"
	case StatusWouldInstall:
		return "→ Would Install"
	case StatusFailed:
		return "✗ Failed"
	case StatusError:
		return "✗ Error"
	default:
		return string(status)
	}
}

// getStatusColor returns the ANSI color code for the given status.
func getStatusColor(status ModuleStatus) string {
	switch status {
	case StatusInstalled:
		return colorGreen
	case StatusSkipped:
		return colorYellow
	case StatusWouldInstall:
		return colorGreen // Green to indicate positive action, but different symbol distinguishes it
	case StatusFailed, StatusError:
		return colorRed
	default:
		return colorReset
	}
}
