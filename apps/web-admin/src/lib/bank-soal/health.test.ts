import { describe, expect, it } from 'vitest';

import {
	buildBankSoalHealthModel,
	filterBankSoalHealthActions,
	type BankSoalHealthInput,
} from './health';

const sampleQuestions: BankSoalHealthInput['questions'] = [
	{
		id: 'q-1',
		subject_id: 'math',
		subject_name: 'Matematika',
		workflow_status: 'siap_pakai',
		status: 'published',
		kd_ref: '3.1',
		cp_ref: 'CP-1',
		tp_ref: 'TP-1',
		cognitive_level: 'C3',
		difficulty: 'medium',
		explanation_html: '<p>Bahas</p>',
		package_count: 2,
		asset_count: 1,
		import_batch_id: 'batch-1',
	},
	{
		id: 'q-2',
		subject_id: 'math',
		subject_name: 'Matematika',
		workflow_status: 'diperiksa',
		status: 'draft',
		kd_ref: '',
		cp_ref: 'CP-1',
		tp_ref: '',
		cognitive_level: '',
		difficulty: 'hard',
		explanation_html: '',
		package_count: 0,
		asset_count: 0,
		import_batch_id: 'batch-1',
	},
	{
		id: 'q-3',
		subject_id: 'indo',
		subject_name: 'Bahasa Indonesia',
		workflow_status: 'konsep',
		status: 'draft',
		kd_ref: null,
		cp_ref: '',
		tp_ref: '',
		cognitive_level: 'C2',
		difficulty: '',
		explanation_html: '',
		package_count: 0,
		asset_count: 0,
		import_batch_id: 'batch-2',
	},
];

describe('Bank Soal health model', () => {
	it('builds status, coverage, metadata quality, readiness, and backlog from summary plus sampled questions', () => {
		const model = buildBankSoalHealthModel({
			summary: {
				counts: {
					total: 30,
					published: 8,
					siap_pakai: 10,
					diperiksa: 5,
					konsep: 7,
					archived: 0,
				},
				by_subject: [
					{ subject_id: 'math', subject_name: 'Matematika', total: 20 },
					{ subject_id: 'indo', subject_name: 'Bahasa Indonesia', total: 10 },
				],
			},
			questions: sampleQuestions,
		});

		expect(model.statusCards.map((card) => [card.key, card.value, card.evidence])).toEqual([
			['total', 30, 'summary'],
			['konsep', 7, 'summary'],
			['diperiksa', 5, 'summary'],
			['siap_pakai', 10, 'summary'],
			['published', 8, 'summary'],
			['archived', 0, 'summary'],
		]);
		expect(model.subjectCoverage.value).toBe(2);
		expect(model.subjectCoverage.total).toBe(2);
		expect(model.curriculumCoverage.map((metric) => [metric.key, metric.percent, metric.missing])).toEqual([
			['kd', 33, 2],
			['cp', 67, 1],
			['tp', 33, 2],
		]);
		expect(model.metadataQuality.map((metric) => [metric.key, metric.missing])).toEqual([
			['kd', 2],
			['cognitive_level', 1],
			['difficulty', 1],
			['explanation', 2],
		]);
		expect(model.reviewBacklog.value).toBe(5);
		expect(model.readiness.score).toBeGreaterThanOrEqual(50);
		expect(model.readiness.grade).toMatch(/^[A-D]$/);
		expect(model.warnings.some((warning) => warning.kind === 'asset' && warning.severity === 'warning')).toBe(true);
	});

	it('marks unavailable import, asset, and archived evidence instead of inventing analytics', () => {
		const model = buildBankSoalHealthModel({
			summary: { counts: { total: 2, konsep: 2 } },
			questions: [
				{ id: 'q-1', workflow_status: 'konsep', status: 'draft' },
				{ id: 'q-2', workflow_status: 'konsep', status: 'draft' },
			],
		});

		expect(model.statusCards.find((card) => card.key === 'archived')?.evidence).toBe('missing');
		expect(model.warnings).toEqual(
			expect.arrayContaining([
				expect.objectContaining({ kind: 'import', label: 'Import', evidenceLabel: 'perlu evidence/data' }),
				expect.objectContaining({ kind: 'asset', label: 'Aset', evidenceLabel: 'perlu evidence/data' }),
			])
		);
		expect(model.readiness.evidenceLabel).toContain('sampel 2 soal');
	});

	it('does not report an empty review backlog when review evidence is absent', () => {
		const model = buildBankSoalHealthModel({
			summary: { counts: { total: 0 } },
			questions: [],
		});

		expect(model.reviewBacklog.value).toBeNull();
		expect(model.reviewBacklog.evidenceLabel).toBe('perlu evidence/data');
	});

	it('labels sample-only subject data without pretending full coverage', () => {
		const model = buildBankSoalHealthModel({
			questions: [
				{ id: 'q-1', subject_id: 'math', workflow_status: 'konsep', status: 'draft' },
				{ id: 'q-2', subject_id: 'indo', workflow_status: 'konsep', status: 'draft' },
			],
		});

		expect(model.subjectCoverage.label).toBe('Mapel terdeteksi di sampel');
		expect(model.subjectCoverage.value).toBe(2);
		expect(model.subjectCoverage.total).toBeNull();
		expect(model.subjectCoverage.percent).toBeNull();
		expect(model.subjectCoverage.evidenceLabel).toBe('sampel 2 soal');
	});

	it('filters privileged quick actions by Bank Soal capabilities', () => {
		const model = buildBankSoalHealthModel({
			summary: { counts: { total: 10, diperiksa: 3, konsep: 7 } },
			questions: sampleQuestions,
		});

		expect(filterBankSoalHealthActions(model.actions, { role: 'guru', roles: ['guru'], permissions: [] }).map((action) => action.href)).toEqual([
			'/bank-soal/daftar',
		]);
		expect(filterBankSoalHealthActions(model.actions, { role: 'admin', roles: ['admin'], permissions: [] }).map((action) => action.href)).toEqual(
			expect.arrayContaining(['/bank-soal/verifikasi', '/bank-soal/tambah', '/bank-soal/impor'])
		);
	});
});
