# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Added

- Stateful App struct with synchronized AppState and view-model methods
  (State, SetActiveSheet, SelectCell, SetScrollPosition, SetZoom),
  with SelectCell synchronizing the selection range
- Inline cell editing via double-click, F2, or character-typing to start,
  with an interactive text editor per cell, Enter to commit, Escape to
  cancel, blur-commit semantics across both the grid and formula bar,
  pending-edit tracking, formula-clearing and style-preserving edit
  semantics, and styled-blank retention on clear
- xlsx workbook loading via excelize with full cell, sheet, and style support
  (OpenWorkbook, OpenWorkbookPath, OpenDroppedFiles)
- SaveWorkbook and SaveWorkbookAs with xlsx save pipeline, dirty-state
  guarding (unsaved-changes dialog before open/drop/close), File menu
  Save/Save As commands, dirty indicator dot in the title bar, and
  testable file-dialog injection
- Cross-platform dirty-prompt normalization: map native Yes/No dialog
  responses to internal Save/Don't Save constants for Linux and Windows
- beforeClose handler that intercepts the Wails window close event
- BottomBar component with interactive sheet tabs rendered from workbook
  sheets, click-to-switch handling, horizontal scroll overflow for many
  sheets, and dynamic status readouts with color-coded kind indicators
- SideRail component with disabled icon buttons for collapsed left and right
  workspace rails
- Interactive FormulaBar with focus-to-edit, Enter to commit, Escape to
  cancel, and blur-commit semantics wired to the grid's edit session
- Workspace shell with CSS Grid layout (top chrome, formula strip, left/right
  rails, grid canvas, bottom bar) wiring all child components to Wails Go
  backend bindings, with drag-and-drop file opening via Wails runtime handler
- TopChrome component with brand mark, document title, status indicator,
  functional File menu popover for opening workbooks, disabled stub menu
  items (Edit, View, etc.), and disabled toolbar groups with SVG icons
- AppearanceControl segmented component for toggling system/light/dark modes
- Appearance mode support (system/light/dark) with Go backend state machine,
  theme resolution, appearance command methods (InitializeAppearance,
  SetAppearanceMode, SetSystemTheme), localStorage persistence, snapshot-safe
  normalization, and appearance-aware workbook style rendering on load
- Dark theme cell text with automatic WCAG 2.1 AA contrast adjustment
  via binary-search color mixing and minimum 4.5:1 contrast ratio
- SpreadsheetGrid component with dynamic grid derived from sheet bounds,
  visual cell styling (fonts, fills, alignment, borders), merged cell
  spanning with value display, selection range highlighting with outline
  overlay, hidden row/column support, dynamic column widths and row
  heights from workbook layout, scroll position tracking,
  keyboard-accessible cell selection, and active-cell styling with inset
  box-shadow
