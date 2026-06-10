# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Added

- BottomBar component with add-sheet button, sheet tabs, and status readouts
  (Ready, A1, 100%)
- SideRail component with disabled icon buttons for collapsed left and right
  workspace rails
- FormulaBar component with name box, visual divider, fx marker, and disabled
  formula display area, integrated into WorkspaceShell
- Workspace shell with CSS Grid layout (top chrome, formula strip, left/right
  rails, grid canvas, bottom bar)
- TopChrome component with brand mark, document title, status indicator,
  disabled menu bar, and disabled toolbar groups with SVG icons
- AppearanceControl segmented component for toggling system/light/dark modes
- Appearance mode support (system/light/dark) with localStorage persistence
- SpreadsheetGrid component with CSS Grid layout, 40 sticky column headers
  (A–AN), 100 sticky row headers (1–100), and active-cell selection styling
