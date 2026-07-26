/**
 * Global store for journal wizard flow across pages.
 * Persists during browser session (survives SPA navigation, lost on tab close).
 */
import { writable, type Writable } from 'svelte/store';

export type JournalFlowState = {
	/** Last selected class ID (step 1 → step 2) */
	selectedClassId: string | null;
	/** Last selected class code for display */
	selectedClassCode: string | null;
	/** Last active tab on the journal detail page */
	lastTab: 'journal' | 'rekap';
};

const STORAGE_KEY = 'journal_flow';

function loadFromStorage(): JournalFlowState {
	if (typeof sessionStorage === 'undefined') return { selectedClassId: null, selectedClassCode: null, lastTab: 'journal' };
	try {
		const raw = sessionStorage.getItem(STORAGE_KEY);
		if (raw) return JSON.parse(raw);
	} catch { /* ignore */ }
	return { selectedClassId: null, selectedClassCode: null, lastTab: 'journal' };
}

function saveToStorage(state: JournalFlowState) {
	if (typeof sessionStorage === 'undefined') return;
	try {
		sessionStorage.setItem(STORAGE_KEY, JSON.stringify(state));
	} catch { /* quota exceeded */ }
}

function createJournalFlowStore(): Writable<JournalFlowState> & { persist: () => void } {
	const initial = loadFromStorage();
	const { subscribe, set, update } = writable<JournalFlowState>(initial);

	return {
		subscribe,
		set: (value: JournalFlowState) => {
			saveToStorage(value);
			set(value);
		},
		update: (fn: (state: JournalFlowState) => JournalFlowState) => {
			update((current) => {
				const next = fn(current);
				saveToStorage(next);
				return next;
			});
		},
		persist: () => {
			// Manually trigger save current state
			update((state) => {
				saveToStorage(state);
				return state;
			});
		},
	};
}

export const journalFlow = createJournalFlowStore();
