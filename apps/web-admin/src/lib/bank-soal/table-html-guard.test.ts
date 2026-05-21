import { describe, expect, it } from 'vitest';

import {
	compareTableHtmlBeforeSave,
	hasBlockingTableHtmlIssues,
	summarizeTableHtml,
} from './table-html-guard';

describe('table-html-guard', () => {
	it('accepts valid simple table html', () => {
		const summary = summarizeTableHtml('<table><tbody><tr><th>No</th><td>Isi</td></tr></tbody></table>');

		expect(summary.tableCount).toBe(1);
		expect(summary.rowCount).toBe(1);
		expect(summary.cellCount).toBe(2);
		expect(summary.hasTable).toBe(true);
		expect(summary.hasRows).toBe(true);
		expect(summary.hasCells).toBe(true);
		expect(summary.issues).toEqual([]);
		expect(hasBlockingTableHtmlIssues(summary)).toBe(false);
	});

	it('does not create blockers when html has no table', () => {
		const summary = summarizeTableHtml('<p>Teks biasa tanpa tabel.</p>');

		expect(summary.tableCount).toBe(0);
		expect(summary.hasTable).toBe(false);
		expect(summary.issues).toEqual([]);
		expect(hasBlockingTableHtmlIssues(summary)).toBe(false);
	});

	it('blocks a draft table without rows, including malformed unclosed tables', () => {
		for (const html of ['<table><tbody></tbody></table>', '<table>teks tabel rusak']) {
			const summary = summarizeTableHtml(html);

			expect(summary.issues).toEqual(
				expect.arrayContaining([
					expect.objectContaining({ severity: 'blocker', code: 'table-without-rows' }),
				])
			);
			expect(hasBlockingTableHtmlIssues(summary)).toBe(true);
		}
	});

	it('blocks a draft table without cells', () => {
		const summary = summarizeTableHtml('<table><tbody><tr></tr></tbody></table>');

		expect(summary.rowCount).toBe(1);
		expect(summary.cellCount).toBe(0);
		expect(summary.issues).toEqual(
			expect.arrayContaining([
				expect.objectContaining({ severity: 'blocker', code: 'table-without-cells' }),
			])
		);
	});


	it('blocks malformed table fragments even when another table is valid', () => {
		const summary = summarizeTableHtml('<table><tr><td>A</td></tr></table><table>rusak');

		expect(summary.tableCount).toBe(2);
		expect(summary.issues).toEqual(
			expect.arrayContaining([
				expect.objectContaining({ severity: 'blocker', code: 'table-without-rows' }),
				expect.objectContaining({ severity: 'blocker', code: 'table-without-cells' }),
			])
		);
	});

	it('blocks save when server table is lost in draft', () => {
		const issues = compareTableHtmlBeforeSave({
			fieldLabel: 'Pertanyaan',
			serverHtml: '<table><tr><td>A</td></tr></table>',
			draftHtml: '<p>Tabel terhapus.</p>',
		});

		expect(issues).toEqual(
			expect.arrayContaining([
				expect.objectContaining({ severity: 'blocker', code: 'suspicious-flat-table-text' }),
			])
		);
		expect(hasBlockingTableHtmlIssues(issues)).toBe(true);
	});

	it('blocks save when row count drops by at least 40 percent', () => {
		const issues = compareTableHtmlBeforeSave({
			fieldLabel: 'Pertanyaan',
			serverHtml: '<table><tr><td>1</td></tr><tr><td>2</td></tr><tr><td>3</td></tr><tr><td>4</td></tr><tr><td>5</td></tr></table>',
			draftHtml: '<table><tr><td>1</td></tr><tr><td>2</td></tr><tr><td>3</td></tr></table>',
		});

		expect(issues).toEqual(
			expect.arrayContaining([
				expect.objectContaining({ severity: 'blocker', code: 'row-count-dropped' }),
			])
		);
	});

	it('blocks save when cell count drops by at least 40 percent', () => {
		const issues = compareTableHtmlBeforeSave({
			fieldLabel: 'Pertanyaan',
			serverHtml: '<table><tr><td>1</td><td>2</td><td>3</td><td>4</td><td>5</td></tr></table>',
			draftHtml: '<table><tr><td>1</td><td>2</td><td>3</td></tr></table>',
		});

		expect(issues).toEqual(
			expect.arrayContaining([
				expect.objectContaining({ severity: 'blocker', code: 'cell-count-dropped' }),
			])
		);
	});

	it('warns but does not block on complex table attributes', () => {
		const summary = summarizeTableHtml(
			'<table style="width: 100%"><tr><td colspan="2" rowspan="1" width="50%">A</td></tr></table>'
		);

		expect(summary.hasComplexAttrs).toBe(true);
		expect(summary.issues).toEqual(
			expect.arrayContaining([
				expect.objectContaining({ severity: 'warning', code: 'contains-complex-table-attrs' }),
			])
		);
		expect(hasBlockingTableHtmlIssues(summary)).toBe(false);
	});
});
