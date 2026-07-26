/**
 * Auto-saving draft store for journal create/edit forms.
 * Persists to sessionStorage so form data survives SPA navigation.
 */
import { writable, type Writable } from 'svelte/store';

export type JournalDraft = {
	tanggal: string;
	materi: string;
	kegiatan: string;
	catatan: string;
	guru_hadir: boolean;
};

const EMPTY_DRAFT: JournalDraft = {
	tanggal: new Date().toISOString().slice(0, 10),
	materi: '',
	kegiatan: '',
	catatan: '',
	guru_hadir: true,
};

const STORAGE_KEY = 'journal_draft_create';

function loadDraft(): JournalDraft {
	if (typeof sessionStorage === 'undefined') return { ...EMPTY_DRAFT };
	try {
		const raw = sessionStorage.getItem(STORAGE_KEY);
		if (raw) {
			const parsed = JSON.parse(raw);
			return { ...EMPTY_DRAFT, ...parsed };
		}
	} catch { /* ignore */ }
	return { ...EMPTY_DRAFT };
}

function createDraftStore(): Writable<JournalDraft> & { clear: () => void; restore: (overrides?: Partial<JournalDraft>) => JournalDraft } {
	const initial = loadDraft();
	const { subscribe, set, update } = writable<JournalDraft>(initial);

	// Auto-save on every change
	let autoSaveTimeout: ReturnType<typeof setTimeout> | undefined;
	subscribe((value) => {
		clearTimeout(autoSaveTimeout);
		autoSaveTimeout = setTimeout(() => {
			if (typeof sessionStorage !== 'undefined') {
				try { sessionStorage.setItem(STORAGE_KEY, JSON.stringify(value)); } catch {}
			}
		}, 300); // debounce 300ms
	});

	return {
		subscribe,
		set: (value: JournalDraft) => set(value),
		update: (fn: (state: JournalDraft) => JournalDraft) => update(fn),
		clear: () => {
			const fresh = { ...EMPTY_DRAFT, tanggal: new Date().toISOString().slice(0, 10) };
			set(fresh);
			if (typeof sessionStorage !== 'undefined') {
				try { sessionStorage.removeItem(STORAGE_KEY); } catch {}
			}
		},
		restore: (overrides?: Partial<JournalDraft>) => {
			const draft = loadDraft();
			if (overrides) Object.assign(draft, overrides);
			set(draft);
			return draft;
		},
	};
}

export const journalDraft = createDraftStore();
