import { describe, expect, it } from 'vitest';
import { readFileSync } from 'node:fs';
import { fileURLToPath } from 'node:url';
import { dirname, join } from 'node:path';

const currentDir = dirname(fileURLToPath(import.meta.url));
const source = readFileSync(join(currentDir, '+page.svelte'), 'utf8');

describe('/paket-soal builder source contract', () => {
	it('unwraps soal-support subjects envelope before using it as an array', () => {
		expect(source).toContain("type SubjectsPayload = SubjectOption[] | { subjects?: SubjectOption[] }");
		expect(source).toContain('readClientApiData<SubjectsPayload>');
		expect(source).toContain('subjectOptions = Array.isArray(payload) ? payload : payload.subjects ?? []');
		expect(source).not.toContain('subjectOptions = await readClientApiData<SubjectOption[]>(response)');
	});
});
