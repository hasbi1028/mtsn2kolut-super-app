export const THEME_STORAGE_KEY = 'mtsn2kolut:web-admin:theme';
export const THEME_QUERY = '(prefers-color-scheme: dark)';

export type ThemePreference = 'system' | 'light' | 'dark';
export type EffectiveTheme = 'light' | 'dark';
export type ThemeStorage = Pick<Storage, 'getItem' | 'setItem'>;

export function normalizeThemePreference(value: unknown): ThemePreference {
	if (value === 'light' || value === 'dark' || value === 'system') {
		return value;
	}
	return 'system';
}

export function resolveThemePreference(preference: ThemePreference, systemPrefersDark: boolean): EffectiveTheme {
	return preference === 'dark' || (preference === 'system' && systemPrefersDark) ? 'dark' : 'light';
}

export function nextThemePreference(preference: ThemePreference): ThemePreference {
	const normalized = normalizeThemePreference(preference);
	if (normalized === 'system') return 'light';
	if (normalized === 'light') return 'dark';
	return 'system';
}

export function themePreferenceLabel(preference: ThemePreference) {
	const normalized = normalizeThemePreference(preference);
	if (normalized === 'light') return 'Terang';
	if (normalized === 'dark') return 'Gelap';
	return 'Sistem';
}

export function effectiveThemeLabel(theme: EffectiveTheme) {
	return theme === 'dark' ? 'Gelap' : 'Terang';
}

function browserStorage(): ThemeStorage | null {
	if (typeof window === 'undefined') return null;
	return window.localStorage;
}

export function readThemePreference(storage: ThemeStorage | null = browserStorage()): ThemePreference {
	if (!storage) return 'system';
	try {
		return normalizeThemePreference(storage.getItem(THEME_STORAGE_KEY));
	} catch {
		return 'system';
	}
}

export function writeThemePreference(
	preference: ThemePreference,
	storage: ThemeStorage | null = browserStorage()
): ThemePreference {
	const normalized = normalizeThemePreference(preference);
	if (!storage) return normalized;
	try {
		storage.setItem(THEME_STORAGE_KEY, normalized);
	} catch {
		// localStorage may be unavailable in hardened browser contexts.
	}
	return normalized;
}

export function getSystemPrefersDark(
	matchMediaFn: ((query: string) => Pick<MediaQueryList, 'matches'>) | null =
		typeof window === 'undefined' ? null : window.matchMedia.bind(window)
): boolean {
	if (!matchMediaFn) return false;
	try {
		return matchMediaFn(THEME_QUERY).matches;
	} catch {
		return false;
	}
}

export function applyThemePreference(
	preference: ThemePreference,
	options: {
		root?: HTMLElement | null;
		systemPrefersDark?: boolean;
	} = {}
): EffectiveTheme {
	const normalized = normalizeThemePreference(preference);
	const effective = resolveThemePreference(
		normalized,
		options.systemPrefersDark ?? getSystemPrefersDark()
	);
	const root = options.root ?? (typeof document === 'undefined' ? null : document.documentElement);

	if (root) {
		root.classList.toggle('dark', effective === 'dark');
		root.dataset.theme = normalized;
		root.dataset.resolvedTheme = effective;
		root.style.colorScheme = effective;
	}

	return effective;
}
