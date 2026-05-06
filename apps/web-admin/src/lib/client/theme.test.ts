import { describe, expect, it, vi } from 'vitest';
import {
	THEME_STORAGE_KEY,
	applyThemePreference,
	effectiveThemeLabel,
	getSystemPrefersDark,
	nextThemePreference,
	normalizeThemePreference,
	readThemePreference,
	resolveThemePreference,
	themePreferenceLabel,
	writeThemePreference
} from './theme';

function memoryStorage(initial?: string) {
	const data = new Map<string, string>();
	if (initial) data.set(THEME_STORAGE_KEY, initial);
	return {
		getItem: vi.fn((key: string) => data.get(key) ?? null),
		setItem: vi.fn((key: string, value: string) => {
			data.set(key, value);
		})
	};
}

describe('theme helpers', () => {
	it('normalizes unsupported preferences to system', () => {
		expect(normalizeThemePreference('dark')).toBe('dark');
		expect(normalizeThemePreference('light')).toBe('light');
		expect(normalizeThemePreference('system')).toBe('system');
		expect(normalizeThemePreference('unknown')).toBe('system');
		expect(normalizeThemePreference(null)).toBe('system');
	});

	it('resolves system preference from media state', () => {
		expect(resolveThemePreference('system', true)).toBe('dark');
		expect(resolveThemePreference('system', false)).toBe('light');
		expect(resolveThemePreference('dark', false)).toBe('dark');
		expect(resolveThemePreference('light', true)).toBe('light');
	});

	it('cycles through system, light, and dark labels', () => {
		expect(nextThemePreference('system')).toBe('light');
		expect(nextThemePreference('light')).toBe('dark');
		expect(nextThemePreference('dark')).toBe('system');
		expect(themePreferenceLabel('system')).toBe('Sistem');
		expect(themePreferenceLabel('light')).toBe('Terang');
		expect(themePreferenceLabel('dark')).toBe('Gelap');
		expect(effectiveThemeLabel('dark')).toBe('Gelap');
	});

	it('reads and writes the persisted preference safely', () => {
		const storage = memoryStorage('dark');
		expect(readThemePreference(storage)).toBe('dark');
		expect(writeThemePreference('light', storage)).toBe('light');
		expect(storage.setItem).toHaveBeenCalledWith(THEME_STORAGE_KEY, 'light');
	});

	it('falls back to system when storage throws', () => {
		const storage = {
			getItem: vi.fn(() => {
				throw new Error('blocked');
			}),
			setItem: vi.fn(() => {
				throw new Error('blocked');
			})
		};

		expect(readThemePreference(storage)).toBe('system');
		expect(writeThemePreference('dark', storage)).toBe('dark');
	});

	it('reads system dark preference from matchMedia', () => {
		expect(getSystemPrefersDark(() => ({ matches: true }))).toBe(true);
		expect(getSystemPrefersDark(() => ({ matches: false }))).toBe(false);
		expect(getSystemPrefersDark(() => {
			throw new Error('unsupported');
		})).toBe(false);
	});

	it('applies dark class and color-scheme to the root element', () => {
		const root = document.createElement('html');
		expect(applyThemePreference('system', { root, systemPrefersDark: true })).toBe('dark');
		expect(root.classList.contains('dark')).toBe(true);
		expect(root.dataset.theme).toBe('system');
		expect(root.dataset.resolvedTheme).toBe('dark');
		expect(root.style.colorScheme).toBe('dark');

		expect(applyThemePreference('light', { root, systemPrefersDark: true })).toBe('light');
		expect(root.classList.contains('dark')).toBe(false);
		expect(root.dataset.theme).toBe('light');
		expect(root.dataset.resolvedTheme).toBe('light');
		expect(root.style.colorScheme).toBe('light');
	});
});
