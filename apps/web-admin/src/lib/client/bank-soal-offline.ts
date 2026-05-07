export const BANK_SOAL_OFFLINE_DB_NAME = 'mtsn2-web-admin-offline';
export const BANK_SOAL_OFFLINE_DB_VERSION = 1;
export const BANK_SOAL_DRAFT_STORE = 'bank_soal_drafts';
export const BANK_SOAL_SYNC_STORE = 'bank_soal_sync_queue';
export const CURRENT_COMPOSER_DRAFT_PREFIX = 'mtsn2-soal-komposer:';
export const LEGACY_CBT_COMPOSER_DRAFT_PREFIX = 'cbt_soal_draft_';

const FALLBACK_DRAFT_PREFIX = 'mtsn2-bank-soal-offline:draft:';
const FALLBACK_QUEUE_KEY = 'mtsn2-bank-soal-offline:sync-queue';

export type StorageLike = Pick<Storage, 'getItem' | 'setItem' | 'removeItem' | 'key' | 'length'>;

export type BankSoalDraftSnapshot<TPayload = unknown> = {
	draftKey: string;
	payload: TPayload;
	savedAt: string;
	updatedAt: string;
	source: 'composer' | 'legacy-cbt';
};

export type BankSoalQuestionSyncIntent = 'draft' | 'review';
export type BankSoalQuestionSyncMethod = 'POST' | 'PUT';

export type BankSoalQuestionSyncItem<TPayload = unknown> = {
	id: string;
	draftKey: string;
	intent: BankSoalQuestionSyncIntent;
	method: BankSoalQuestionSyncMethod;
	endpoint: string;
	questionId?: string;
	payload: TPayload;
	createdAt: string;
	updatedAt: string;
	attempts: number;
	lastError?: string;
	authRequired?: boolean;
};

export type BankSoalQuestionSyncInput<TPayload = unknown> = {
	draftKey: string;
	intent: BankSoalQuestionSyncIntent;
	method: BankSoalQuestionSyncMethod;
	endpoint: string;
	questionId?: string;
	payload: TPayload;
	lastError?: string;
	authRequired?: boolean;
};

export type BankSoalStorageOptions = {
	indexedDB?: IDBFactory | null;
	storage?: StorageLike | null;
	now?: () => string;
	randomId?: () => string;
};

function nowIso(options?: BankSoalStorageOptions) {
	return options?.now?.() ?? new Date().toISOString();
}

function randomId(options?: BankSoalStorageOptions) {
	if (options?.randomId) return options.randomId();
	if (typeof crypto !== 'undefined' && 'randomUUID' in crypto) return crypto.randomUUID();
	return `offline-${Date.now().toString(36)}-${Math.random().toString(36).slice(2, 10)}`;
}

function defaultStorage(): StorageLike | null {
	try {
		return typeof localStorage === 'undefined' ? null : localStorage;
	} catch {
		return null;
	}
}

function storageFromOptions(options?: BankSoalStorageOptions): StorageLike | null {
	if (options && 'storage' in options) return options.storage ?? null;
	return defaultStorage();
}

function idbFactoryFromOptions(options?: BankSoalStorageOptions): IDBFactory | null {
	if (options && 'indexedDB' in options) return options.indexedDB ?? null;
	try {
		return typeof indexedDB === 'undefined' ? null : indexedDB;
	} catch {
		return null;
	}
}

function requestResult<T>(request: IDBRequest<T>): Promise<T> {
	return new Promise((resolve, reject) => {
		request.onsuccess = () => resolve(request.result);
		request.onerror = () => reject(request.error ?? new Error('IndexedDB request failed'));
	});
}

function transactionDone(transaction: IDBTransaction): Promise<void> {
	return new Promise((resolve, reject) => {
		transaction.oncomplete = () => resolve();
		transaction.onabort = () => reject(transaction.error ?? new Error('IndexedDB transaction aborted'));
		transaction.onerror = () => reject(transaction.error ?? new Error('IndexedDB transaction failed'));
	});
}

async function openBankSoalOfflineDb(options?: BankSoalStorageOptions): Promise<IDBDatabase | null> {
	const factory = idbFactoryFromOptions(options);
	if (!factory) return null;
	return new Promise((resolve) => {
		const request = factory.open(BANK_SOAL_OFFLINE_DB_NAME, BANK_SOAL_OFFLINE_DB_VERSION);
		request.onupgradeneeded = () => {
			const db = request.result;
			if (!db.objectStoreNames.contains(BANK_SOAL_DRAFT_STORE)) {
				db.createObjectStore(BANK_SOAL_DRAFT_STORE, { keyPath: 'draftKey' });
			}
			if (!db.objectStoreNames.contains(BANK_SOAL_SYNC_STORE)) {
				const store = db.createObjectStore(BANK_SOAL_SYNC_STORE, { keyPath: 'id' });
				store.createIndex('draftKey', 'draftKey', { unique: false });
				store.createIndex('createdAt', 'createdAt', { unique: false });
			}
		};
		request.onsuccess = () => resolve(request.result);
		request.onerror = () => resolve(null);
		request.onblocked = () => resolve(null);
	});
}

async function idbPut<TValue>(storeName: string, value: TValue, options?: BankSoalStorageOptions): Promise<boolean> {
	const db = await openBankSoalOfflineDb(options);
	if (!db) return false;
	try {
		const transaction = db.transaction(storeName, 'readwrite');
		transaction.objectStore(storeName).put(value);
		await transactionDone(transaction);
		return true;
	} catch {
		return false;
	} finally {
		db.close();
	}
}

async function idbGet<TValue>(storeName: string, key: IDBValidKey, options?: BankSoalStorageOptions): Promise<TValue | null> {
	const db = await openBankSoalOfflineDb(options);
	if (!db) return null;
	try {
		const transaction = db.transaction(storeName, 'readonly');
		const value = await requestResult<TValue | undefined>(transaction.objectStore(storeName).get(key));
		await transactionDone(transaction);
		return value ?? null;
	} catch {
		return null;
	} finally {
		db.close();
	}
}

async function idbGetAll<TValue>(storeName: string, options?: BankSoalStorageOptions): Promise<TValue[] | null> {
	const db = await openBankSoalOfflineDb(options);
	if (!db) return null;
	try {
		const transaction = db.transaction(storeName, 'readonly');
		const values = await requestResult<TValue[]>(transaction.objectStore(storeName).getAll());
		await transactionDone(transaction);
		return values;
	} catch {
		return null;
	} finally {
		db.close();
	}
}

async function idbDelete(storeName: string, key: IDBValidKey, options?: BankSoalStorageOptions): Promise<boolean> {
	const db = await openBankSoalOfflineDb(options);
	if (!db) return false;
	try {
		const transaction = db.transaction(storeName, 'readwrite');
		transaction.objectStore(storeName).delete(key);
		await transactionDone(transaction);
		return true;
	} catch {
		return false;
	} finally {
		db.close();
	}
}

async function idbClear(storeName: string, options?: BankSoalStorageOptions): Promise<boolean> {
	const db = await openBankSoalOfflineDb(options);
	if (!db) return false;
	try {
		const transaction = db.transaction(storeName, 'readwrite');
		transaction.objectStore(storeName).clear();
		await transactionDone(transaction);
		return true;
	} catch {
		return false;
	} finally {
		db.close();
	}
}

function fallbackDraftKey(draftKey: string) {
	return `${FALLBACK_DRAFT_PREFIX}${draftKey}`;
}

function fallbackQueueItems<TPayload>(storage: StorageLike | null): BankSoalQuestionSyncItem<TPayload>[] {
	if (!storage) return [];
	const raw = storage.getItem(FALLBACK_QUEUE_KEY);
	if (!raw) return [];
	try {
		const parsed = JSON.parse(raw) as unknown;
		return Array.isArray(parsed) ? (parsed as BankSoalQuestionSyncItem<TPayload>[]) : [];
	} catch {
		return [];
	}
}

function saveFallbackQueueItems<TPayload>(items: BankSoalQuestionSyncItem<TPayload>[], storage: StorageLike | null) {
	if (!storage) return;
	storage.setItem(FALLBACK_QUEUE_KEY, JSON.stringify(items));
}

export function isLegacyComposerDraftKey(key: string): boolean {
	return key.startsWith(CURRENT_COMPOSER_DRAFT_PREFIX) || key.startsWith(LEGACY_CBT_COMPOSER_DRAFT_PREFIX);
}

export async function saveBankSoalDraftPayload<TPayload>(
	draftKey: string,
	payload: TPayload,
	options?: BankSoalStorageOptions
): Promise<BankSoalDraftSnapshot<TPayload>> {
	const timestamp = nowIso(options);
	const snapshot: BankSoalDraftSnapshot<TPayload> = {
		draftKey,
		payload,
		savedAt: typeof payload === 'object' && payload !== null && 'savedAt' in payload && typeof payload.savedAt === 'string'
			? payload.savedAt
			: timestamp,
		updatedAt: timestamp,
		source: draftKey.startsWith(LEGACY_CBT_COMPOSER_DRAFT_PREFIX) ? 'legacy-cbt' : 'composer',
	};
	if (await idbPut(BANK_SOAL_DRAFT_STORE, snapshot, options)) return snapshot;
	const storage = storageFromOptions(options);
	storage?.setItem(fallbackDraftKey(draftKey), JSON.stringify(snapshot));
	return snapshot;
}

export async function loadBankSoalDraftPayload<TPayload>(
	draftKey: string,
	options?: BankSoalStorageOptions
): Promise<TPayload | null> {
	const snapshot = await idbGet<BankSoalDraftSnapshot<TPayload>>(BANK_SOAL_DRAFT_STORE, draftKey, options);
	if (snapshot) return snapshot.payload;
	const storage = storageFromOptions(options);
	const raw = storage?.getItem(fallbackDraftKey(draftKey)) ?? storage?.getItem(draftKey);
	if (!raw) return null;
	try {
		const parsed = JSON.parse(raw) as unknown;
		if (parsed && typeof parsed === 'object' && 'payload' in parsed) {
			return (parsed as BankSoalDraftSnapshot<TPayload>).payload;
		}
		return parsed as TPayload;
	} catch {
		return null;
	}
}

export async function deleteBankSoalDraftPayload(draftKey: string, options?: BankSoalStorageOptions): Promise<void> {
	await idbDelete(BANK_SOAL_DRAFT_STORE, draftKey, options);
	const storage = storageFromOptions(options);
	storage?.removeItem(fallbackDraftKey(draftKey));
	storage?.removeItem(draftKey);
}

export async function clearBankSoalDraftPayloads(options?: BankSoalStorageOptions): Promise<number> {
	const snapshots = await idbGetAll<BankSoalDraftSnapshot>(BANK_SOAL_DRAFT_STORE, options);
	const idbCount = snapshots?.length ?? 0;
	await idbClear(BANK_SOAL_DRAFT_STORE, options);
	const storage = storageFromOptions(options);
	let fallbackCount = 0;
	if (storage) {
		const keys: string[] = [];
		for (let index = 0; index < storage.length; index += 1) {
			const key = storage.key(index);
			if (!key) continue;
			if (key.startsWith(FALLBACK_DRAFT_PREFIX) || isLegacyComposerDraftKey(key)) keys.push(key);
		}
		for (const key of keys) {
			storage.removeItem(key);
			fallbackCount += 1;
		}
	}
	return idbCount + fallbackCount;
}

export async function migrateLegacyBankSoalDrafts(options?: BankSoalStorageOptions): Promise<number> {
	const storage = storageFromOptions(options);
	if (!storage || !idbFactoryFromOptions(options)) return 0;
	const keys: string[] = [];
	for (let index = 0; index < storage.length; index += 1) {
		const key = storage.key(index);
		if (key && isLegacyComposerDraftKey(key)) keys.push(key);
	}
	let migrated = 0;
	for (const key of keys) {
		const raw = storage.getItem(key);
		if (!raw) continue;
		try {
			const payload = JSON.parse(raw) as unknown;
			const snapshot = await saveBankSoalDraftPayload(key, payload, options);
			if (snapshot) {
				storage.removeItem(key);
				migrated += 1;
			}
		} catch {
			/* keep malformed legacy data untouched */
		}
	}
	return migrated;
}

export async function enqueueBankSoalQuestionSync<TPayload>(
	input: BankSoalQuestionSyncInput<TPayload>,
	options?: BankSoalStorageOptions
): Promise<BankSoalQuestionSyncItem<TPayload>> {
	const timestamp = nowIso(options);
	const items = await listBankSoalQuestionSyncQueue<TPayload>(options);
	const existing = items.find((item) => item.draftKey === input.draftKey || (input.questionId && item.questionId === input.questionId));
	const item: BankSoalQuestionSyncItem<TPayload> = {
		id: existing?.id ?? randomId(options),
		draftKey: input.draftKey,
		intent: input.intent,
		method: input.method,
		endpoint: input.endpoint,
		questionId: input.questionId,
		payload: input.payload,
		createdAt: existing?.createdAt ?? timestamp,
		updatedAt: timestamp,
		attempts: existing?.attempts ?? 0,
		lastError: input.lastError,
		authRequired: input.authRequired,
	};
	if (await idbPut(BANK_SOAL_SYNC_STORE, item, options)) return item;
	const storage = storageFromOptions(options);
	const nextItems = existing ? items.map((queued) => (queued.id === existing.id ? item : queued)) : [...items, item];
	saveFallbackQueueItems(nextItems, storage);
	return item;
}

export async function listBankSoalQuestionSyncQueue<TPayload>(
	options?: BankSoalStorageOptions
): Promise<BankSoalQuestionSyncItem<TPayload>[]> {
	const items = await idbGetAll<BankSoalQuestionSyncItem<TPayload>>(BANK_SOAL_SYNC_STORE, options);
	if (items) return items.sort((a, b) => a.createdAt.localeCompare(b.createdAt));
	return fallbackQueueItems<TPayload>(storageFromOptions(options)).sort((a, b) => a.createdAt.localeCompare(b.createdAt));
}

export async function markBankSoalQuestionSyncFailed(
	id: string,
	message: string,
	authRequired = false,
	options?: BankSoalStorageOptions
): Promise<void> {
	const items = await listBankSoalQuestionSyncQueue(options);
	const item = items.find((queued) => queued.id === id);
	if (!item) return;
	const nextItem: BankSoalQuestionSyncItem = {
		...item,
		attempts: item.attempts + 1,
		lastError: message,
		authRequired,
		updatedAt: nowIso(options),
	};
	if (await idbPut(BANK_SOAL_SYNC_STORE, nextItem, options)) return;
	saveFallbackQueueItems(items.map((queued) => (queued.id === id ? nextItem : queued)), storageFromOptions(options));
}

export async function removeBankSoalQuestionSyncItem(id: string, options?: BankSoalStorageOptions): Promise<void> {
	await idbDelete(BANK_SOAL_SYNC_STORE, id, options);
	const storage = storageFromOptions(options);
	const items = fallbackQueueItems(storage).filter((item) => item.id !== id);
	saveFallbackQueueItems(items, storage);
}

export async function clearBankSoalQuestionSyncQueue(options?: BankSoalStorageOptions): Promise<number> {
	const items = await listBankSoalQuestionSyncQueue(options);
	await idbClear(BANK_SOAL_SYNC_STORE, options);
	storageFromOptions(options)?.removeItem(FALLBACK_QUEUE_KEY);
	return items.length;
}

export async function clearBankSoalOfflineState(options?: BankSoalStorageOptions): Promise<{ drafts: number; queue: number }> {
	const drafts = await clearBankSoalDraftPayloads(options);
	const queue = await clearBankSoalQuestionSyncQueue(options);
	return { drafts, queue };
}

export function isAuthExpiredSyncStatus(status: number): boolean {
	return status === 401;
}
