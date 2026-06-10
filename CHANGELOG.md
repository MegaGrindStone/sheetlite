# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Added

- Stateful App struct with synchronized AppState and view-model methods
  (State, SetActiveSheet, SelectCell, SetScrollPosition, SetZoom)
- xlsx workbook loading via excelize with full cell, sheet, and style support
  (OpenWorkbook, OpenWorkbookPath, OpenDroppedFiles)
- BottomBar component with interactive sheet tabs rendered from workbook
  sheets, click-to-switch handling, horizontal scroll overflow for many
  sheets, and dynamic status readouts with color-coded kind indicators
- SideRail component with disabled icon buttons for collapsed left and right
  workspace rails
- FormulaBar component with name box, visual divider, fx marker, and a
  formula display wired to live cell data (formula or raw value),
  integrated into WorkspaceShell
- Workspace shell with CSS Grid layout (top chrome, formula strip, left/right
  rails, grid canvas, bottom bar) wiring all child components to Wails Go
  backend bindings, with drag-and-drop file opening via Wails runtime handler
- TopChrome component with brand mark, document title, status indicator,
  functional File menu popover for opening workbooks, disabled stub menu
  items (Edit, View, etc.), and disabled toolbar groups with SVG icons
- AppearanceControl segmented component for toggling system/light/dark modes
- Appearance mode support (system/light/dark) with localStorage persistence
- SpreadsheetGrid component with dynamic grid sizing derived from actual
  sheet bounds (40-column/100-row minimum floor), inline cell value
  rendering with text-overflow ellipsis, and active-cell selection styling
