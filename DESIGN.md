---
version: alpha
name: sheetlite
description: Design system for Sheetlite, a lightweight, spreadsheet-first desktop workspace.
colors:
  primary: "#137333"
  accent-light: "#137333"
  accent-dark: "#10b981"
  bg-light: "#f8f9fa"
  bg-dark: "#121212"
  surface-light: "#ffffff"
  surface-dark: "#161616"
  chrome-light: "#f1f3f4"
  chrome-dark: "#1e1e1e"
  text-light: "#202124"
  text-dark: "#e0e0e0"
  muted-light: "#5f6368"
  muted-dark: "#a0a0a0"
  border-light: "#dadce0"
  border-dark: "#2d2d2d"
  gridline-light: "#e2e3e5"
  gridline-dark: "#242424"
  selection-border-light: "#137333"
  selection-border-dark: "#10b981"
  selection-bg-light: "#e7f1eb"
  selection-bg-dark: "#123226"
  focus-ring-light: "#1a73e8"
  focus-ring-dark: "#3b82f6"
  disabled-text-light: "#9ca3af"
  disabled-text-dark: "#5a5a5a"
  disabled-bg-light: "#f1f3f4"
  disabled-bg-dark: "#1a1a1a"
typography:
  headline-md:
    fontFamily: "Inter, -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, Helvetica, Arial, sans-serif"
    fontSize: "14px"
    fontWeight: 600
    lineHeight: 1.2
  body-md:
    fontFamily: "Inter, -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, Helvetica, Arial, sans-serif"
    fontSize: "13px"
    fontWeight: 400
    lineHeight: 1.4
  mono-sm:
    fontFamily: "SFMono-Regular, Consolas, 'Liberation Mono', Menlo, Courier, monospace"
    fontSize: "12px"
    fontWeight: 400
    lineHeight: 1.3
spacing:
  xxs: "2px"
  xs: "4px"
  sm: "8px"
  md: "12px"
  lg: "16px"
  xl: "24px"
rounded:
  none: "0px"
  sm: "2px"
  md: "4px"
  lg: "8px"
  full: "9999px"
components:
  brand-mark:
    backgroundColor: "{colors.primary}"
    rounded: "{rounded.sm}"
  workspace-canvas-light:
    backgroundColor: "{colors.bg-light}"
  workspace-canvas-dark:
    backgroundColor: "{colors.bg-dark}"
  workspace-chrome-light:
    backgroundColor: "{colors.chrome-light}"
    textColor: "{colors.text-light}"
  workspace-chrome-dark:
    backgroundColor: "{colors.chrome-dark}"
    textColor: "{colors.text-dark}"
  workspace-border-light:
    backgroundColor: "{colors.border-light}"
  workspace-border-dark:
    backgroundColor: "{colors.border-dark}"
  grid-cell-light:
    backgroundColor: "{colors.surface-light}"
    textColor: "{colors.text-light}"
    typography: "{typography.body-md}"
  grid-cell-dark:
    backgroundColor: "{colors.surface-dark}"
    textColor: "{colors.text-dark}"
    typography: "{typography.body-md}"
  grid-line-light:
    backgroundColor: "{colors.gridline-light}"
  grid-line-dark:
    backgroundColor: "{colors.gridline-dark}"
  grid-header-light:
    backgroundColor: "{colors.chrome-light}"
    textColor: "{colors.muted-light}"
  grid-header-dark:
    backgroundColor: "{colors.chrome-dark}"
    textColor: "{colors.muted-dark}"
  grid-accent-light:
    backgroundColor: "{colors.accent-light}"
  grid-accent-dark:
    backgroundColor: "{colors.accent-dark}"
  grid-selection-light:
    backgroundColor: "{colors.selection-bg-light}"
  grid-selection-dark:
    backgroundColor: "{colors.selection-bg-dark}"
  grid-selection-border-light:
    backgroundColor: "{colors.selection-border-light}"
  grid-selection-border-dark:
    backgroundColor: "{colors.selection-border-dark}"
  focus-ring-indicator-light:
    backgroundColor: "{colors.focus-ring-light}"
  focus-ring-indicator-dark:
    backgroundColor: "{colors.focus-ring-dark}"
  disabled-surface-light:
    backgroundColor: "{colors.disabled-bg-light}"
  disabled-surface-dark:
    backgroundColor: "{colors.disabled-bg-dark}"
  disabled-label-light:
    textColor: "{colors.disabled-text-light}"
  disabled-label-dark:
    textColor: "{colors.disabled-text-dark}"
---

## Overview

Sheetlite is a lightweight, spreadsheet-first desktop workspace built on Svelte and Wails. It prioritizes the spreadsheet grid as the dominant canvas. The application relies on a native titlebar provided by the operating system (retaining the native Wails window frame) and provides a highly-optimized, low-overhead layout with cohesive light and dark appearance modes.

The visual tone is professional, technical, minimal, and highly functional:

- **Spreadsheet-first**: Emphasizes structural clarity and grid readability.
- **Dense and compact chrome**: Controls, rails, and formula bar occupy minimal vertical and horizontal space to maximize the visible grid viewport.
- **Honest stub states**: Unfinished interactive controls are clearly styled as disabled, preventing fake behavior or confusing placeholders.

## Colors

The color system is divided into explicit Light and Dark token palettes to ensure excellent contrast, theme separation, and grid readability.

### Light Theme

- **App Canvas Background (`{colors.bg-light}`)**: `#f8f9fa` - Used for layout backdrops and margins.
- **Workspace Chrome Background (`{colors.chrome-light}`)**: `#f1f3f4` - Used for toolbars, formula bar, bottom bar, and side rails.
- **Grid Cell Background (`{colors.surface-light}`)**: `#ffffff` - Crisp white surface for spreadsheet rows and columns.
- **Primary Text (`{colors.text-light}`)**: `#202124` - Highly legible near-black for body copy and headings.
- **Muted Text (`{colors.muted-light}`)**: `#5f6368` - Medium gray for label helpers, secondary metrics, and inactive headers.
- **Border Separation (`{colors.border-light}`)**: `#dadce0` - Crisp dividing borders between workspace panels.
- **Gridline (`{colors.gridline-light}`)**: `#e2e3e5` - Thin, unobtrusive, highly readable spreadsheet grid borders.
- **Accent (`{colors.accent-light}`)**: `#137333` - Dedicated brand green for active highlights, callouts, or key interactive indicators.
- **Selection (`{colors.selection-border-light}`)**: `#137333` - Dedicated green border highlighting active cells, tabs, or focus states.

### Dark Theme

- **App Canvas Background (`{colors.bg-dark}`)**: `#121212` - Near-black background.
- **Workspace Chrome Background (`{colors.chrome-dark}`)**: `#1e1e1e` - Deep grey for container shells and command areas.
- **Grid Cell Background (`{colors.surface-dark}`)**: `#161616` - Very dark grey to isolate the grid surface from chrome.
- **Primary Text (`{colors.text-dark}`)**: `#e0e0e0` - Off-white text for high contrast and reduced eye strain.
- **Muted Text (`{colors.muted-dark}`)**: `#a0a0a0` - Warm grey for secondary text labels.
- **Border Separation (`{colors.border-dark}`)**: `#2d2d2d` - Subtle dark divider borders.
- **Gridline (`{colors.gridline-dark}`)**: `#242424` - Dim, low-glow gridline borders for clear visibility without excessive brightness.
- **Accent (`{colors.accent-dark}`)**: `#10b981` - Dedicated glowing emerald green for active highlights, callouts, or key interactive indicators.
- **Selection (`{colors.selection-border-dark}`)**: `#10b981` - Glowing emerald green highlight for active cell boundaries.

## Typography

The typography scale is highly compact and prioritizes UI readability over decorative expression.

- **Workspace Headers & UI Titles (`{typography.headline-md}`)**: Used for app identity, major sections, and bold callouts. Spaced closely and set to standard semibold.
- **Main App & Grid Copy (`{typography.body-md}`)**: Standard body text, optimized for numeric entry and small-scale dense layouts.
- **Code, Formula & Coordinates (`{typography.mono-sm}`)**: Fixed-width font stack for formula bars, cell coordinate readouts, and monospace values.

## Layout

Sheetlite enforces a single-window desktop workspace shell that fits exactly into the 100% viewport dimensions of the native container.

- **Base Layout Grid**: Fits the viewport height exactly and avoids page-level overflow (`overflow: hidden` on root).
- **Chrome Sizing**:
  - Top Chrome: Combined workspace title, menu/toolbar, and formula strip should remain compact.
  - Side Rails: Left and right vertical rails remain locked to a narrow width (`44px` to `56px`) with centered, secondary items.
  - Bottom Bar: Retains a slim height (`36px` to `44px`) to host sheet tabs and zoom/status metrics.

## Elevation & Depth

To avoid visual noise, Sheetlite does not use elevated cards, heavy drop shadows, or fuzzy borders. Depth is communicated strictly through surface layering:

- **Level 0 (Bottom-most)**: App Canvas Background (`bg-light`/`bg-dark`).
- **Level 1 (Structural Chrome)**: Workspace Chrome Background (`chrome-light`/`chrome-dark`) with single-pixel border dividers.
- **Level 2 (Data Canvas)**: Grid Cell Background (`surface-light`/`surface-dark`), indicating the main editable zone.

## Shapes

Corners are predominantly square (`{rounded.none}`) to maintain the structural spreadsheet layout.

- **Sharp (`{rounded.none}`)**: All cells, workspace layout divisions, chrome panels, and input fields.
- **Subtle Rounding (`{rounded.sm}`)**: Small UI badges, inline keyboard focuses, or individual selection highlights.
- **Circular (`{rounded.full}`)**: Avatar stubs or fully rounded toggle pill indicators.

## Components

### Spreadsheet Grid

- **Gridlines**: Set to `1px solid` of `gridline-light` or `gridline-dark`.
- **Row/Column Headers**: Sourced from the workspace chrome surface, with a subtle border in the direction of the cells. Text is centered and utilizes `muted-text` styling.
- **Active Selection**: A `2px solid` border of `selection-border-light` or `selection-border-dark` with an optional `selection-bg` overlay for the active range.

### Inactive Controls & Stubs

- **Disabled Items**: Use semantic `disabled` properties when interactive, or non-interactive tags with the disabled cursor and muted/disabled text colors.
- **Keyboard Traversal**: Non-functional stubs must be removed from the keyboard tab sequence (`tabindex="-1"`).

## Do's and Don'ts

### Do

- Do use CSS custom properties linked to the theme tokens for all color variables.
- Do prioritize cell gridline contrast; make sure they are visible but not visually overwhelming.
- Do ensure the active cell indicator (`A1`) is bright, crisp, and high-contrast.
- Do restrict component-specific styling to the Svelte components themselves; keep the global stylesheet focused strictly on baseline reset, variables, and root sizing.

### Don't

- Don't use raw hex colors or arbitrary rgb values in components; always reference CSS properties.
- Don't add decorative shadows, box-shadow cards, or deep background gradients.
- Don't include unfinished or stubbed controls in the keyboard tab focus loop.
- Don't import bulky third-party visual frameworks, Tailwind configurations, or visual asset loaders.
