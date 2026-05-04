import { describe, expect, it } from 'vitest';

import { clearCbtComposerDrafts } from './cbt-drafts';

class MemoryStorage implements Pick<Storage, 'key' | 'removeItem' | 'length'> {
	private readonly values: Map<string, string>;

	constructor(entries: Array<[string, string]>) {
		this.values = new Map(entries);
	}

	get length() {
		return this.values.size;
	}

	key(index: number) {
		return Array.from(this.values.keys())[index] ?? null;
	}

	removeItem(key: string) {
		this.values.delete(key);
	}

	has(key: string) {
		return this.values.has(key);
	}
}

describe('CBT local draft cleanup', () => {
	it('removes only CBT composer draft keys', () => {
		const storage = new MemoryStorage([
			['cbt_soal_draft_new_event-1', '{}'],
			['cbt_soal_draft_q-1_event-1', '{}'],
			['unrelated', 'keep'],
		]);

		expect(clearCbtComposerDrafts(storage)).toBe(2);
		expect(storage.has('cbt_soal_draft_new_event-1')).toBe(false);
		expect(storage.has('cbt_soal_draft_q-1_event-1')).toBe(false);
		expect(storage.has('unrelated')).toBe(true);
	});
});
