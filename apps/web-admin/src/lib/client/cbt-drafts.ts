const CBT_DRAFT_PREFIX = 'cbt_soal_draft_';

export function clearCbtComposerDrafts(storage: Pick<Storage, 'key' | 'removeItem' | 'length'> = localStorage) {
	const keys: string[] = [];
	for (let index = 0; index < storage.length; index += 1) {
		const key = storage.key(index);
		if (key?.startsWith(CBT_DRAFT_PREFIX)) keys.push(key);
	}
	for (const key of keys) storage.removeItem(key);
	return keys.length;
}
