/**
 * Reusable persistent state utilities for the entire app.
 * 
 * Two patterns:
 *   1. `persistState(key, default)` — sessionStorage-backed Svelte 5 rune-like store
 *   2. `useUrlParam(name, default)` — URL search param state (no extra imports needed)
 * 
 * Usage:
 *   import { persistState } from '$lib/stores/persistent';
 *   const filterStore = persistState('employee_filter', '');
 *   $filterStore  // reactive value
 *   $filterStore = 'new value'  // auto-saves
 * 
 *   import { readUrlParam, writeUrlParam } from '$lib/stores/persistent';
 *   let tab = $state(readUrlParam('tab', 'journal'));
 *   writeUrlParam('tab', tab);
 */

// ─── SessionStorage-backed writable store ───

import { writable, type Writable } from 'svelte/store';

function getStorage(): Storage | null {
	if (typeof sessionStorage === 'undefined') return null;
	return sessionStorage;
}

function getLocalStorage(): Storage | null {
	if (typeof localStorage === 'undefined') return null;
	return localStorage;
}

/**
 * Creates a sessionStorage-backed persistent writable store.
 * Data survives SPA navigation but is cleared when tab closes.
 * 
 * @param key - unique storage key (prefix with module name, e.g. 'employee_filter')
 * @param defaultValue - fallback if nothing stored yet
 * @param useLocal - set true for localStorage (survives tab close)
 */
export function persistState<T>(
	key: string,
	defaultValue: T,
	useLocal: boolean = false
): Writable<T> & { clear: () => void } {
	const storage = useLocal ? getLocalStorage() : getStorage();

	let initial: T = defaultValue;
	if (storage) {
		try {
			const raw = storage.getItem(key);
			if (raw !== null) {
				initial = JSON.parse(raw) as T;
			}
		} catch { /* corrupt data */ }
	}

	const store = writable<T>(initial);

	// Auto-save on every change (debounced)
	let saveTimer: ReturnType<typeof setTimeout> | undefined;
	store.subscribe((value) => {
		clearTimeout(saveTimer);
		saveTimer = setTimeout(() => {
			if (storage) {
				try {
					storage.setItem(key, JSON.stringify(value));
				} catch { /* quota */ }
			}
		}, 200);
	});

	return {
		subscribe: store.subscribe,
		set: store.set,
		update: store.update,
		clear: () => {
			store.set(defaultValue);
			if (storage) {
				try { storage.removeItem(key); } catch {}
			}
		},
	};
}

// ─── URL search param helpers ───

/**
 * Read a value from the current URL search params.
 * Works in both SSR and CSR.
 */
export function readUrlParam(name: string, fallback: string = ''): string {
	if (typeof window === 'undefined') return fallback;
	const params = new URLSearchParams(window.location.search);
	return params.get(name) ?? fallback;
}

/**
 * Write a value to URL search params without navigation.
 * Uses history.replaceState to avoid polluting browser history.
 */
export function writeUrlParam(name: string, value: string | null) {
	if (typeof window === 'undefined') return;
	const url = new URL(window.location.href);
	if (value === null || value === '') {
		url.searchParams.delete(name);
	} else {
		url.searchParams.set(name, value);
	}
	window.history.replaceState({}, '', url);
}

/**
 * Write multiple URL params at once (more efficient than individual calls).
 */
export function writeUrlParams(entries: Record<string, string | null>) {
	if (typeof window === 'undefined') return;
	const url = new URL(window.location.href);
	for (const [name, value] of Object.entries(entries)) {
		if (value === null || value === '') {
			url.searchParams.delete(name);
		} else {
			url.searchParams.set(name, value);
		}
	}
	window.history.replaceState({}, '', url);
}

// ─── Draft form auto-save helper ───

type Draftable = Record<string, unknown>;

/**
 * Creates a form draft store with sessionStorage auto-save.
 * Good for multi-step or easily-lost forms.
 * 
 * @param formName - unique name for this form (e.g. 'employee_create')
 * @param defaults - default values
 * @param debounceMs - save debounce in ms (default 300)
 */
export function createDraft<T extends Draftable>(
	formName: string,
	defaults: T,
	debounceMs: number = 300
) {
	const key = `draft_${formName}`;
	const storage = getStorage();

	function load(): T {
		if (!storage) return { ...defaults };
		try {
			const raw = storage.getItem(key);
			if (raw) return { ...defaults, ...JSON.parse(raw) };
		} catch {}
		return { ...defaults };
	}

	function save(data: T) {
		if (!storage) return;
		try { storage.setItem(key, JSON.stringify(data)); } catch {}
	}

	function clear() {
		if (!storage) return;
		try { storage.removeItem(key); } catch {}
	}

	// Return composable helpers (no store — works with plain $state)
	return { load, save, clear, key };
}

// ─── Sidebar collapse state (localStorage so it survives tab close) ───

export const sidebarCollapsed = persistState<boolean>('sidebar_collapsed', false, true);

// ─── Last visited paths for breadcrumb / back navigation ───

export const lastVisitedPath = persistState<string>('last_visited_path', '/', true);
