package main

// AppearanceMode describes the user's selected appearance preference.
type AppearanceMode string

const (
	// AppearanceModeSystem follows the reported system theme.
	AppearanceModeSystem AppearanceMode = "system"
	// AppearanceModeLight forces the light theme.
	AppearanceModeLight AppearanceMode = "light"
	// AppearanceModeDark forces the dark theme.
	AppearanceModeDark AppearanceMode = "dark"
)

// AllAppearanceModes lists AppearanceMode values for Wails enum binding.
var AllAppearanceModes = []struct {
	Value  AppearanceMode
	TSName string
}{
	{AppearanceModeSystem, "System"},
	{AppearanceModeLight, "Light"},
	{AppearanceModeDark, "Dark"},
}

// AppearanceTheme describes a resolved light or dark theme.
type AppearanceTheme string

const (
	// AppearanceThemeLight is the light effective theme.
	AppearanceThemeLight AppearanceTheme = "light"
	// AppearanceThemeDark is the dark effective theme.
	AppearanceThemeDark AppearanceTheme = "dark"
)

// AllAppearanceThemes lists AppearanceTheme values for Wails enum binding.
var AllAppearanceThemes = []struct {
	Value  AppearanceTheme
	TSName string
}{
	{AppearanceThemeLight, "Light"},
	{AppearanceThemeDark, "Dark"},
}

// AppearanceState describes backend-owned runtime appearance state.
type AppearanceState struct {
	Mode           AppearanceMode  `json:"mode"`
	SystemTheme    AppearanceTheme `json:"systemTheme"`
	EffectiveTheme AppearanceTheme `json:"effectiveTheme"`
}

func defaultAppearanceState() AppearanceState {
	return AppearanceState{
		Mode:           AppearanceModeSystem,
		SystemTheme:    AppearanceThemeLight,
		EffectiveTheme: AppearanceThemeLight,
	}
}

func (m AppearanceMode) valid() bool {
	switch m {
	case AppearanceModeSystem, AppearanceModeLight, AppearanceModeDark:
		return true
	default:
		return false
	}
}

func (t AppearanceTheme) valid() bool {
	switch t {
	case AppearanceThemeLight, AppearanceThemeDark:
		return true
	default:
		return false
	}
}

func resolveEffectiveTheme(mode AppearanceMode, systemTheme AppearanceTheme) AppearanceTheme {
	// Zero-value snapshots fall back to light until the frontend reports a theme.
	if !systemTheme.valid() {
		systemTheme = AppearanceThemeLight
	}

	switch mode {
	case AppearanceModeLight:
		return AppearanceThemeLight
	case AppearanceModeDark:
		return AppearanceThemeDark
	case AppearanceModeSystem:
		return systemTheme
	default:
		return systemTheme
	}
}

func normalizeAppearanceState(state AppearanceState) AppearanceState {
	mode := state.Mode
	if !mode.valid() {
		mode = AppearanceModeSystem
	}

	systemTheme := state.SystemTheme
	if !systemTheme.valid() {
		systemTheme = AppearanceThemeLight
	}

	// Recompute the derived field instead of trusting copied snapshot data.
	return AppearanceState{
		Mode:           mode,
		SystemTheme:    systemTheme,
		EffectiveTheme: resolveEffectiveTheme(mode, systemTheme),
	}
}
