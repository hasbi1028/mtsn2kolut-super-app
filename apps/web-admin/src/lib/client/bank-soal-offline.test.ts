import { describe, expect, it } from 'vitest';
import {
	clearBankSoalDraftPayloads,
	enqueueBankSoalQuestionSync,
	isAuthExpiredSyncStatus,
	isLegacyComposerDraftKey,
	listBankSoalQuestionSyncQueue,
	loadBankSoalDraftPayload,
	markBankSoalQuestionSyncFailed,
	migrateLegacyBankSoalDrafts,
	removeBankSoalQuestionSyncItem,
	saveBankSoalDraftPayload,
	type StorageLike,
} from './bank-soal-offline';
import {
	bankSoalQueueAuthRequiredMessage,
	buildBankSoalQuestionSyncInput,
	shouldQueueBankSoalQuestionSave,
} from './bank-soal-composer-offline';

type MemoryStorage = StorageLike & { dump(): Record<string, string> };

function createMemoryStorage(initial: Record<string, string> = {}): MemoryStorage {
	const map = new Map(Object.entries(initial));
	return {
		get length() {
			return map.size;
		},
		getItem(key: string) {
			return map.get(key) ?? null;
		},
		setItem(key: string, value: string) {
			map.set(key, value);
		},
		removeItem(key: string) {
			map.delete(key);
		},
		key(index: number) {
			return Array.from(map.keys())[index] ?? null;
		},
		dump() {
			return Object.fromEntries(map.entries());
		},
	};
}

const noIdb = { indexedDB: null, now: () => '2026-05-07T00:00:00.000Z', randomId: () => 'queue-1' };

describe('bank soal offline storage fallback', () => {
	it('stores and loads composer draft payload without IndexedDB', async () => {
		const storage = createMemoryStorage();
		await saveBankSoalDraftPayload('mtsn2-soal-komposer:default', { stem: '<p>Soal</p>' }, { ...noIdb, storage });

		await expect(loadBankSoalDraftPayload('mtsn2-soal-komposer:default', { ...noIdb, storage })).resolves.toEqual({ stem: '<p>Soal</p>' });
	});

	it('clears fallback and legacy draft keys but leaves unrelated storage alone', async () => {
		const storage = createMemoryStorage({
			'cbt_soal_draft_old': JSON.stringify({ stem: 'legacy' }),
			'mtsn2-bank-soal-offline:draft:mtsn2-soal-komposer:default': JSON.stringify({ payload: { stem: 'new' } }),
			'theme': 'dark',
		});

		const count = await clearBankSoalDraftPayloads({ ...noIdb, storage });

		expect(count).toBe(2);
		expect(storage.dump()).toEqual({ theme: 'dark' });
	});

	it('does not migrate legacy drafts when IndexedDB is unavailable, keeping data safe', async () => {
		const storage = createMemoryStorage({ cbt_soal_draft_old: JSON.stringify({ stem: 'legacy' }) });

		await expect(migrateLegacyBankSoalDrafts({ ...noIdb, storage })).resolves.toBe(0);
		expect(storage.getItem('cbt_soal_draft_old')).toContain('legacy');
	});
});

describe('bank soal sync queue fallback', () => {
	it('upserts queued create/update work by draft key and marks auth-required failures', async () => {
		const storage = createMemoryStorage();
		await enqueueBankSoalQuestionSync({
			draftKey: 'draft-a',
			intent: 'draft',
			method: 'POST',
			endpoint: '/api/bank-soal/questions',
			payload: { stem: 'A' },
		}, { ...noIdb, storage });
		await enqueueBankSoalQuestionSync({
			draftKey: 'draft-a',
			intent: 'review',
			method: 'POST',
			endpoint: '/api/bank-soal/questions',
			payload: { stem: 'B' },
		}, { ...noIdb, storage });

		let queue = await listBankSoalQuestionSyncQueue<{ stem: string }>({ ...noIdb, storage });
		expect(queue).toHaveLength(1);
		expect(queue[0].intent).toBe('review');
		expect(queue[0].payload.stem).toBe('B');

		await markBankSoalQuestionSyncFailed('queue-1', 'Masuk ulang', true, { ...noIdb, storage });
		queue = await listBankSoalQuestionSyncQueue({ ...noIdb, storage });
		expect(queue[0].attempts).toBe(1);
		expect(queue[0].authRequired).toBe(true);

		await removeBankSoalQuestionSyncItem('queue-1', { ...noIdb, storage });
		await expect(listBankSoalQuestionSyncQueue({ ...noIdb, storage })).resolves.toEqual([]);
	});
});

describe('bank soal offline composer helpers', () => {
	it('builds safe sync endpoint and queues only transport/auth failures', () => {
		expect(buildBankSoalQuestionSyncInput({ draftKey: 'd', editingId: 'abc/123', intent: 'draft', payload: { ok: true } })).toMatchObject({
			method: 'PUT',
			endpoint: '/api/bank-soal/questions/abc%2F123',
		});
		expect(buildBankSoalQuestionSyncInput({ draftKey: 'd', editingId: null, intent: 'review', payload: {} })).toMatchObject({
			method: 'POST',
			endpoint: '/api/bank-soal/questions',
		});
		expect(shouldQueueBankSoalQuestionSave({ isOnline: false, responseReceived: false })).toBe(true);
		expect(shouldQueueBankSoalQuestionSave({ isOnline: true, responseReceived: false })).toBe(true);
		expect(shouldQueueBankSoalQuestionSave({ isOnline: true, responseReceived: true, responseStatus: 401 })).toBe(true);
		expect(shouldQueueBankSoalQuestionSave({ isOnline: true, responseReceived: true, responseStatus: 500 })).toBe(false);
		expect(isAuthExpiredSyncStatus(401)).toBe(true);
		expect(isAuthExpiredSyncStatus(403)).toBe(false);
		expect(isLegacyComposerDraftKey('cbt_soal_draft_abc')).toBe(true);
		expect(isLegacyComposerDraftKey('mtsn2-soal-komposer:default')).toBe(true);
		expect(bankSoalQueueAuthRequiredMessage(2)).toContain('2 perubahan');
	});
});
