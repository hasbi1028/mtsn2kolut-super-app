import { describe, expect, it } from 'vitest';
import {
	canExportAllQuestions,
	questionExportButtonLabel,
	questionExportSuccessMessage
} from './question-export-ui';

describe('CBT question export UI helpers', () => {
	it('keeps admin export copy as full bank soal export', () => {
		const roles = ['admin'];

		expect(canExportAllQuestions(roles)).toBe(true);
		expect(questionExportButtonLabel(roles)).toBe('Export CSV');
		expect(questionExportSuccessMessage(roles)).toBe('Export CSV bank soal berhasil dibuat');
	});

	it('labels teacher export as own-question scoped export', () => {
		const roles = ['guru'];

		expect(canExportAllQuestions(roles)).toBe(false);
		expect(questionExportButtonLabel(roles)).toBe('Export Soal Saya');
		expect(questionExportSuccessMessage(roles)).toBe('Export CSV soal saya berhasil dibuat');
	});

	it('treats multi-role admin as full export capable', () => {
		const roles = ['guru', 'admin'];

		expect(canExportAllQuestions(roles)).toBe(true);
		expect(questionExportButtonLabel(roles)).toBe('Export CSV');
	});
});
