import { describe, expect, it } from 'vitest';

import {
	availableRombelSubjectOptions,
	buildRombelSubjectOptions,
	subjectOptionLabel
} from './rombel-subject-assignments';

describe('rombel subject assignment helpers', () => {
	it('uses active subjects and preserves assigned fallback subjects', () => {
		const options = buildRombelSubjectOptions({
			subjects: [
				{ id: 'subject-2', name: 'Bahasa Indonesia', code: 'BIN', is_active: true },
				{ id: 'subject-1', name: 'Matematika', code: 'MTK', is_active: true },
				{ id: 'subject-inactive', name: 'Mapel Nonaktif', code: 'OFF', is_active: false }
			],
			assignments: [
				{ id: 'assignment-1', subject_id: 'subject-inactive', subject_name: 'Mapel Lama', subject_code: 'ML' },
				{ id: 'assignment-2', subject_id: 'subject-1', subject_name: 'Matematika Lama', subject_code: 'OLD' }
			]
		});

		expect(options.map((option) => option.id)).toEqual(['subject-2', 'subject-inactive', 'subject-1']);
		expect(options.find((option) => option.id === 'subject-1')).toMatchObject({
			name: 'Matematika',
			isActive: true,
			isFallback: false
		});
		expect(options.find((option) => option.id === 'subject-inactive')).toMatchObject({
			name: 'Mapel Lama',
			isActive: false,
			isFallback: true
		});
	});

	it('filters already-assigned subjects except for the assignment being edited', () => {
		const options = buildRombelSubjectOptions({
			subjects: [
				{ id: 'subject-1', name: 'Matematika', is_active: true },
				{ id: 'subject-2', name: 'IPA', is_active: true },
				{ id: 'subject-3', name: 'IPS', is_active: true }
			],
			assignments: [
				{ id: 'assignment-1', subject_id: 'subject-1', subject_name: 'Matematika' },
				{ id: 'assignment-2', subject_id: 'subject-2', subject_name: 'IPA' }
			]
		});

		expect(availableRombelSubjectOptions(options, [
			{ id: 'assignment-1', subject_id: 'subject-1', subject_name: 'Matematika' },
			{ id: 'assignment-2', subject_id: 'subject-2', subject_name: 'IPA' }
		], '').map((option) => option.id)).toEqual(['subject-3']);

		expect(availableRombelSubjectOptions(options, [
			{ id: 'assignment-1', subject_id: 'subject-1', subject_name: 'Matematika' },
			{ id: 'assignment-2', subject_id: 'subject-2', subject_name: 'IPA' }
		], 'assignment-2').map((option) => option.id)).toEqual(['subject-2', 'subject-3']);
	});

	it('formats subject option labels with codes', () => {
		expect(subjectOptionLabel({ id: 'subject-1', name: 'Matematika', code: 'MTK', isActive: true, isFallback: false })).toBe('Matematika (MTK)');
		expect(subjectOptionLabel({ id: 'subject-2', name: 'IPS', code: '', isActive: true, isFallback: false })).toBe('IPS');
	});
});
