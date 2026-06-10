export type AppearanceMode = 'light' | 'dark' | 'system';
export type EffectiveTheme = 'light' | 'dark';

export interface AppearanceOption {
	value: AppearanceMode;
	label: string;
}

export const appearanceOptions: AppearanceOption[] = [
	{ value: 'system', label: 'System' },
	{ value: 'light', label: 'Light' },
	{ value: 'dark', label: 'Dark' }
];

const STORAGE_KEY = 'sheetlite_appearance_mode';
const MEDIA_QUERY_DARK = '(prefers-color-scheme: dark)';

function getStoredMode(): AppearanceMode {
	try {
		if (typeof localStorage !== 'undefined') {
			const val = localStorage.getItem(STORAGE_KEY);
			if (val === 'light' || val === 'dark' || val === 'system') {
				return val;
			}
		}
	} catch {
		// Silently fallback in restricted/sandboxed environments
	}
	return 'system';
}

function setStoredMode(mode: AppearanceMode): void {
	try {
		if (typeof localStorage !== 'undefined') {
			localStorage.setItem(STORAGE_KEY, mode);
		}
	} catch {
		// Silently fail in restricted/sandboxed environments
	}
}

function supportsMatchMedia(): boolean {
	return typeof window !== 'undefined' && typeof window.matchMedia === 'function';
}

export class AppearanceState {
	mode = $state<AppearanceMode>('system');
	resolvedTheme = $state<EffectiveTheme>('light');

	setMode(newMode: AppearanceMode) {
		this.mode = newMode;
		setStoredMode(newMode);
		this.updateTheme();
	}

	updateTheme() {
		if (typeof window === 'undefined') return;

		let effective: EffectiveTheme;
		if (this.mode === 'system') {
			if (supportsMatchMedia()) {
				const mediaQuery = window.matchMedia(MEDIA_QUERY_DARK);
				effective = mediaQuery.matches ? 'dark' : 'light';
			} else {
				effective = 'light';
			}
		} else {
			effective = this.mode === 'light' ? 'light' : 'dark';
		}

		this.resolvedTheme = effective;

		// Apply root attributes
		if (typeof document !== 'undefined') {
			document.documentElement.dataset.theme = effective;
			document.documentElement.dataset.appearance = this.mode;
		}
	}
}

// Export a singleton instance
export const appearanceState = new AppearanceState();

export function initializeAppearance() {
	if (typeof window === 'undefined') {
		return () => {};
	}

	// 1. Load persisted mode
	appearanceState.mode = getStoredMode();

	// 2. Initial application of the theme
	appearanceState.updateTheme();

	// 3. Listen to media query changes (prefers-color-scheme)
	if (!supportsMatchMedia()) {
		return () => {};
	}

	const mediaQuery = window.matchMedia(MEDIA_QUERY_DARK);
	const handleMediaChange = () => {
		if (appearanceState.mode === 'system') {
			appearanceState.updateTheme();
		}
	};

	mediaQuery.addEventListener('change', handleMediaChange);

	// Return a cleanup function
	return () => {
		mediaQuery.removeEventListener('change', handleMediaChange);
	};
}
