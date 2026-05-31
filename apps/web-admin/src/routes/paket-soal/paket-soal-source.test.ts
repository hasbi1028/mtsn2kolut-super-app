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

	it('shows published-question counts and defaults new builder to an available subject', () => {
		expect(source).toContain('let subjectPublishedCounts = $state<Record<string, number>>({});');
		expect(source).toContain("workflow_status: 'published'");
		expect(source).toContain('function defaultBuilderSubjectId()');
		expect(source).toContain('subject_id: defaultBuilderSubjectId(),');
		expect(source).toContain('{subjectOptionLabel(subject)}');
	});

	it('keeps readiness guards before saving or locking packages', () => {
		expect(source).toContain("builderError = 'Pilih minimal 1 soal dari Bank Soal.';");
		expect(source).toContain("builderError = 'Paket sudah terkunci. Gunakan Clone/Revisi untuk mengubah.';");
		expect(source).toContain("builderError = 'Paket belum siap dikunci. Lengkapi target soal dan metadata dulu.';");
		expect(source).toContain('disabled={packageActionBusy !== null || activeLocked || !activeReadiness?.ready}');
	});
});
