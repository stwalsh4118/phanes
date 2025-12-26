# PBI-16: TUI Version

[View in Backlog](../backlog.md#user-content-16)

## Overview

Create a Terminal User Interface (TUI) version of Phanes that provides an interactive way to configure and run provisioning without editing YAML config files. This makes Phanes more accessible to users who prefer visual interfaces.

## Problem Statement

Currently, using Phanes requires:
- Editing YAML configuration files
- Understanding config file structure
- Using command-line flags
- Reading documentation for available options

This creates a barrier for users who prefer interactive interfaces. A TUI would:
- Provide visual module selection
- Allow interactive configuration editing
- Show real-time progress
- Display results in a user-friendly format

## User Stories

- As a developer, I want an interactive TUI so that I can configure Phanes without editing files
- As a developer, I want to select modules visually so that I can see what's available
- As a developer, I want to edit configuration interactively so that I can set options easily
- As a developer, I want to see progress in real-time so that I know what's happening
- As a developer, I want to save configurations so that I can reuse them later

## Technical Approach

### TUI Library

Use a Go TUI library:
- **Bubble Tea** (`github.com/charmbracelet/bubbletea`) - Popular, well-maintained
- **tview** (`github.com/rivo/tview`) - Alternative option

**Recommendation**: Bubble Tea for modern, component-based approach.

### TUI Structure

Multi-screen TUI with:

1. **Main Menu**
   - Select modules (checkboxes)
   - Choose profile or custom selection
   - Edit configuration
   - Start provisioning
   - Load/save config

2. **Module Selection Screen**
   - List all available modules
   - Checkboxes for selection
   - Module descriptions
   - Profile shortcuts

3. **Configuration Screen**
   - Form-based configuration editing
   - Grouped by category (User, System, Services)
   - Validation feedback
   - Default values shown

4. **Execution Screen**
   - Real-time progress display
   - Current module being executed
   - Status indicators (installed/skipped/failed)
   - Scrollable log output

5. **Summary Screen**
   - Execution summary table
   - Module results
   - Next steps/suggestions

### CLI Integration

Add TUI command:

```bash
phanes tui              # Launch TUI
phanes --tui            # Alternative flag
phanes tui --config file.yaml  # Load existing config
```

### Configuration Loading/Saving

- TUI can load existing `config.yaml` files
- TUI can save configuration to `config.yaml`
- TUI validates configuration before execution
- TUI shows validation errors inline

### Real-Time Updates

- Use Bubble Tea's update model for real-time progress
- Stream module execution output to TUI
- Update status indicators as modules complete
- Show execution summary when complete

### Key Bindings

- `Tab` / `Shift+Tab`: Navigate between fields
- `Space`: Toggle checkboxes
- `Enter`: Confirm/select
- `Esc`: Go back/cancel
- `Ctrl+C`: Exit TUI
- `Ctrl+S`: Save configuration

## UX/UI Considerations

- Clean, modern interface design
- Clear visual hierarchy
- Color-coded status indicators (consistent with CLI)
- Responsive to terminal size
- Help text available (press `?`)
- Keyboard navigation intuitive
- Visual feedback for all actions

## Acceptance Criteria

1. TUI allows interactive module selection with checkboxes
2. TUI allows configuration editing for all config options
3. TUI shows real-time progress during module execution
4. TUI displays execution summary table
5. TUI works alongside existing CLI (accessible via `phanes tui` or `--tui` flag)
6. TUI can load and save config files
7. TUI validates configuration before execution
8. TUI handles errors gracefully with clear messages

## Dependencies

- PBI-1 (Foundation) must be complete - config package needed
- PBI-2 (Module Framework) must be complete - runner system needed
- PBI-8 (CLI) should be complete - for consistency
- Go TUI library (Bubble Tea or tview)

## Open Questions

- Which TUI library should we use? (Bubble Tea recommended)
- Should TUI support all CLI features? (Yes - full feature parity)
- Should TUI support dry-run mode? (Yes)
- Should TUI support remote execution? (Future enhancement)
- Should TUI be the default interface? (No - CLI remains primary)

## Related Tasks

See [Tasks](./tasks.md)

