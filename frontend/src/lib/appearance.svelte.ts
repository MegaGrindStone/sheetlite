import { main } from '$lib/wailsjs/go/models';

export type AppearanceMode = main.AppearanceMode;
export type AppearanceTheme = main.AppearanceTheme;
export type EffectiveTheme = AppearanceTheme;

export interface AppearanceOption {
	value: AppearanceMode;
	label: string;
}

export interface BackendAppearance {
	mode?: unknown;
	systemTheme?: unknown;
	effectiveTheme?: unknown;
}

export const appearanceOptions: AppearanceOption[] = [
	{ value: main.AppearanceMode.System, label: 'System' },
	{ value: main.AppearanceMode.Light, label: 'Light' },
	{ value: main.AppearanceMode.Dark, label: 'Dark' }
];

const STORAGE_KEY = 'sheetlite_appearance_mode';
const MEDIA_QUERY_DARK = '(prefers-color-scheme: dark)';

export function isAppearanceMode(value: unknown): value is AppearanceMode {
	return (
		value === main.AppearanceMode.Light ||
		value === main.AppearanceMode.Dark ||
		value === main.AppearanceMode.System
	);
}

export function isAppearanceTheme(value: unknown): value is AppearanceTheme {
	return value === main.AppearanceTheme.Light || value === main.AppearanceTheme.Dark;
}

export function normalizeAppearanceMode(value: unknown): AppearanceMode {
	return isAppearanceMode(value) ? value : main.AppearanceMode.System;
}

export function normalizeAppearanceTheme(value: unknown): AppearanceTheme {
	return isAppearanceTheme(value) ? value : main.AppearanceTheme.Light;
}

export function readPersistedAppearanceMode(): AppearanceMode {
	try {
		if (typeof localStorage !== 'undefined') {
			return normalizeAppearanceMode(localStorage.getItem(STORAGE_KEY));
		}
	} catch {
		// Silently fallback in restricted/sandboxed environments.
	}

	return main.AppearanceMode.System;
}

export function writePersistedAppearanceMode(mode: AppearanceMode): void {
	try {
		if (typeof localStorage !== 'undefined') {
			localStorage.setItem(STORAGE_KEY, mode);
		}
	} catch {
		// Silently fail in restricted/sandboxed environments.
	}
}

function supportsMatchMedia(): boolean {
	return typeof window !== 'undefined' && typeof window.matchMedia === 'function';
}

export function detectSystemTheme(): AppearanceTheme {
	if (!supportsMatchMedia()) {
		return main.AppearanceTheme.Light;
	}

	return window.matchMedia(MEDIA_QUERY_DARK).matches
		? main.AppearanceTheme.Dark
		: main.AppearanceTheme.Light;
}

export function subscribeToSystemThemeChanges(
	onChange: (theme: AppearanceTheme) => void
): () => void {
	if (!supportsMatchMedia()) {
		return () => {};
	}

	const mediaQuery = window.matchMedia(MEDIA_QUERY_DARK);
	const handleMediaChange = (event: MediaQueryListEvent): void => {
		onChange(event.matches ? main.AppearanceTheme.Dark : main.AppearanceTheme.Light);
	};

	mediaQuery.addEventListener('change', handleMediaChange);

	return () => {
		mediaQuery.removeEventListener('change', handleMediaChange);
	};
}

export function applyAppearanceAttributes(appearance?: BackendAppearance | null): void {
	if (typeof document === 'undefined' || !appearance) {
		return;
	}

	const mode = normalizeAppearanceMode(appearance.mode);
	const effectiveTheme = normalizeAppearanceTheme(appearance.effectiveTheme);

	document.documentElement.dataset.theme = effectiveTheme;
	document.documentElement.dataset.appearance = mode;
}
