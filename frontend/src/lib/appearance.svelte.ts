export type AppearanceMode = 'light' | 'dark' | 'system';
export type AppearanceTheme = 'light' | 'dark';
export type EffectiveTheme = AppearanceTheme;

export interface AppearanceOption {
	value: AppearanceMode;
	label: string;
}

export interface BackendAppearance {
	mode?: string;
	systemTheme?: string;
	effectiveTheme?: string;
}

export const appearanceOptions: AppearanceOption[] = [
	{ value: 'system', label: 'System' },
	{ value: 'light', label: 'Light' },
	{ value: 'dark', label: 'Dark' }
];

const STORAGE_KEY = 'sheetlite_appearance_mode';
const MEDIA_QUERY_DARK = '(prefers-color-scheme: dark)';

export function isAppearanceMode(value: unknown): value is AppearanceMode {
	return value === 'light' || value === 'dark' || value === 'system';
}

export function isAppearanceTheme(value: unknown): value is AppearanceTheme {
	return value === 'light' || value === 'dark';
}

export function readPersistedAppearanceMode(): AppearanceMode {
	try {
		if (typeof localStorage !== 'undefined') {
			const value = localStorage.getItem(STORAGE_KEY);
			if (isAppearanceMode(value)) {
				return value;
			}
		}
	} catch {
		// Silently fallback in restricted/sandboxed environments.
	}

	return 'system';
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
		return 'light';
	}

	return window.matchMedia(MEDIA_QUERY_DARK).matches ? 'dark' : 'light';
}

export function subscribeToSystemThemeChanges(
	onChange: (theme: AppearanceTheme) => void
): () => void {
	if (!supportsMatchMedia()) {
		return () => {};
	}

	const mediaQuery = window.matchMedia(MEDIA_QUERY_DARK);
	const handleMediaChange = (event: MediaQueryListEvent): void => {
		onChange(event.matches ? 'dark' : 'light');
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

	const mode = isAppearanceMode(appearance.mode) ? appearance.mode : 'system';
	const effectiveTheme = isAppearanceTheme(appearance.effectiveTheme)
		? appearance.effectiveTheme
		: 'light';

	document.documentElement.dataset.theme = effectiveTheme;
	document.documentElement.dataset.appearance = mode;
}
