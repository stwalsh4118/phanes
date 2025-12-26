# PBI-10: Execution Summary Report

[View in Backlog](../backlog.md#user-content-10)

## Overview

Add a comprehensive execution summary table at the end of provisioning runs that clearly displays what modules were installed, skipped, or failed. This provides immediate visibility into the provisioning outcome without having to scroll through logs.

## Problem Statement

Currently, when Phanes completes a provisioning run, users must:
- Scroll through logs to understand what happened
- Manually count which modules succeeded or failed
- Determine if any modules were skipped due to already being installed
- Parse through error messages to understand failures

This makes it difficult to quickly assess the provisioning outcome, especially for runs with many modules.

## User Stories

- As a developer, I want a summary table at the end of runs so that I can immediately see the overall outcome
- As a developer, I want to see which modules were installed, skipped, or failed so that I know what happened
- As a developer, I want total counts (X installed, Y skipped, Z failed) so that I can quickly assess success
- As a developer, I want color-coded status indicators so that I can visually identify issues
- As a developer, I want the summary in dry-run mode so that I can preview what would happen

## Technical Approach

### Runner Enhancement

Modify `internal/runner/runner.go` to track execution results:

```go
type ModuleResult struct {
    Name        string
    Status      ModuleStatus  // Installed, Skipped, Failed, Error
    Error       error         // Only set if Status is Failed or Error
    Duration    time.Duration // Optional: track execution time
}

type ModuleStatus string

const (
    StatusInstalled ModuleStatus = "installed"
    StatusSkipped   ModuleStatus = "skipped"
    StatusFailed    ModuleStatus = "failed"
    StatusError     ModuleStatus = "error"
)
```

### Result Collection

- Runner collects `ModuleResult` structs during `RunModules()` execution
- Store results in a slice: `results []ModuleResult`
- Track results for both normal execution and dry-run mode

### Summary Display

Create `PrintSummary(results []ModuleResult)` function that:
- Formats results as a table
- Uses color-coded status indicators (green=installed, yellow=skipped, red=failed)
- Shows totals: "X installed, Y skipped, Z failed"
- Displays error messages for failed modules
- Works in both normal and dry-run modes

### Table Format

```
┌─────────────────┬──────────┬─────────────────────────────────┐
│ Module          │ Status   │ Details                         │
├─────────────────┼──────────┼─────────────────────────────────┤
│ baseline        │ ✓ Installed │                               │
│ user            │ ✓ Installed │                               │
│ security        │ ⊘ Skipped   │ Already installed            │
│ docker          │ ✗ Failed    │ Permission denied            │
└─────────────────┴──────────┴─────────────────────────────────┘

Summary: 2 installed, 1 skipped, 1 failed
```

## UX/UI Considerations

- Summary should be visually distinct from regular log output
- Use clear visual separators (lines, boxes) to make the table stand out
- Color coding should follow existing log color scheme:
  - Green for success/installed
  - Yellow for skipped
  - Red for errors/failed
- Summary should appear after all module execution completes
- In dry-run mode, clearly indicate this is a preview

## Acceptance Criteria

1. Runner tracks module execution results (installed/skipped/failed/error) for all modules
2. Summary table is displayed after all modules complete execution
3. Color-coded status indicators clearly show module outcomes
4. Summary shows total counts (X installed, Y skipped, Z failed)
5. Failed modules display error messages in the summary
6. Summary works in both normal and dry-run modes
7. Summary is visually distinct from regular log output

## Dependencies

- PBI-2 (Module Framework) must be complete - runner system must exist

## Open Questions

- Should we track execution time per module? (Nice to have, but not required)
- Should summary be optional via CLI flag? (Default: always show)
- Should we support exporting summary to JSON/CSV? (Future enhancement)

## Related Tasks

See [Tasks](./tasks.md)

