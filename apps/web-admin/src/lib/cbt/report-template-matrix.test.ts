import { describe, expect, it } from 'vitest';

import { CBT_REPORT_TEMPLATE_MATRIX, reportTemplateByKey } from './report-template-matrix';

describe('CBT report template matrix', () => {
	it('defines the official Phase 26 report set', () => {
		expect(CBT_REPORT_TEMPLATE_MATRIX.map((item) => item.key)).toEqual([
			'participant_list',
			'exam_cards',
			'session_minutes',
			'room_attendance',
			'session_results',
			'item_analysis',
			'proctor_event_recap',
			'evidence_bundle_index'
		]);
	});

	it('keeps PDF and Excel parity mapped to existing safe print/CSV patterns', () => {
		const itemAnalysis = reportTemplateByKey('item_analysis');
		expect(itemAnalysis?.pdfMode).toBe('print_html_or_browser_pdf');
		expect(itemAnalysis?.excelMode).toBe('csv_excel_compatible');
		expect(itemAnalysis?.canonicalRoute).toBe('/asesmen/sesi/[id]');

		const cards = reportTemplateByKey('exam_cards');
		expect(cards?.pdfMode).toBe('print_html_or_browser_pdf');
		expect(cards?.excelMode).toBe('not_applicable_sensitive_tokens');
	});

	it('documents redaction boundaries for sensitive report rows', () => {
		for (const report of CBT_REPORT_TEMPLATE_MATRIX) {
			expect(report.redaction).toBeTruthy();
			expect(report.boundary).toContain('No public CBT API runtime route');
		}
		expect(reportTemplateByKey('proctor_event_recap')?.redaction).toContain('no raw token');
		expect(reportTemplateByKey('item_analysis')?.redaction).toContain('answer key role redaction');
	});
});
