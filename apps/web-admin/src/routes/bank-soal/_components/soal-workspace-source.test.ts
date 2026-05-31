// @vitest-environment node

import { describe, expect, it } from 'vitest';
import { readFileSync } from 'node:fs';
import { dirname, join } from 'node:path';
import { fileURLToPath } from 'node:url';

const currentDir = dirname(fileURLToPath(import.meta.url));
const source = readFileSync(join(currentDir, 'SoalWorkspacePage.svelte'), 'utf8');
const subjectProxySource = readFileSync(
	join(currentDir, '../../..', 'lib/server/cbt-backend-proxy/soal-support/subjects/+server.ts'),
	'utf8'
);

describe('Bank Soal composer source contract', () => {
	it('loads subjects from the narrow academic subjects endpoint for composer dropdown', () => {
		expect(subjectProxySource).toContain("proxy(event).get<AcademicPayload>('/api/academic/subjects')");
		expect(subjectProxySource).not.toContain("proxy(event).get<AcademicPayload>('/api/academic')");
	});

	it('does not let optional asesmen events failure blank the mapel dropdown', () => {
		expect(source).toContain("fetch('/api/bank-soal/soal-support/subjects')");
		expect(source).toContain("fetch('/api/asesmen/events')");
		expect(source).toContain(".catch(() => [] as CbtEvent[])");
	});
});
