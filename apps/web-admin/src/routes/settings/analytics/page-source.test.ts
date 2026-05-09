import { readFileSync } from 'node:fs';
import { resolve } from 'node:path';
import { describe, expect, it } from 'vitest';

const pagePath = resolve(__dirname, '+page.svelte');

describe('internal analytics settings page source contract', () => {
	it('renders only aggregate analytics through BFF read endpoints', () => {
		const source = readFileSync(pagePath, 'utf8');

		expect(source).toContain('/api/internal-analytics/summary');
		expect(source).toContain('/api/internal-analytics/daily');
		expect(source).toContain('AsyncContent');
		expect(source).toContain('Skeleton');
		expect(source).toContain('Analytics Internal');
		expect(source).not.toContain('metadata');
		expect(source).not.toContain('actor_user_id');
		expect(source).not.toContain('raw_user_agent');
		expect(source).not.toContain('export');
	});
});
