import { describe, expect, it } from 'vitest';
import { existsSync, readFileSync } from 'node:fs';
import { resolve } from 'node:path';

const root = resolve(__dirname, '../../..');

describe('assessment route contract', () => {
	it('keeps public BFF aliases for backend assessment session gaps', () => {
		const requiredFiles = [
			'src/routes/api/asesmen/sessions/[id]/+server.ts',
			'src/routes/api/asesmen/sessions/[id]/participants/[pid]/answers/+server.ts',
		];
		for (const file of requiredFiles) {
			expect(existsSync(resolve(root, file)), file).toBe(true);
		}
	});

	it('documents resolved route alias decisions', () => {
		const routeMap = readFileSync(resolve(root, '../../docs/contracts/asesmen-cbt-route-map.md'), 'utf8');
		expect(routeMap).toContain('GET | `/api/asesmen/sessions/{id}` | public-ui-used');
		expect(routeMap).toContain('GET | `/api/asesmen/sessions/{id}/participants/{pid}/answers` | public-ui-used');
		expect(routeMap).not.toContain('missing-alias-to-decide');
	});
});
