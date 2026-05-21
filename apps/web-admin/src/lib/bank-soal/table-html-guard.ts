export type TableHtmlIssueSeverity = 'info' | 'warning' | 'blocker';

export type TableHtmlIssue = {
	severity: TableHtmlIssueSeverity;
	code:
		| 'table-without-rows'
		| 'table-without-cells'
		| 'cell-count-dropped'
		| 'row-count-dropped'
		| 'contains-complex-table-attrs'
		| 'suspicious-flat-table-text';
	message: string;
	evidence?: string;
};

export type TableHtmlSummary = {
	tableCount: number;
	rowCount: number;
	cellCount: number;
	hasTable: boolean;
	hasRows: boolean;
	hasCells: boolean;
	hasComplexAttrs: boolean;
	issues: TableHtmlIssue[];
};

export type CompareTableHtmlBeforeSaveArgs = {
	serverHtml: string | null | undefined;
	draftHtml: string | null | undefined;
	fieldLabel: string;
};

const closedTableRegex = /<table\b[^>]*>[\s\S]*?<\/table>/gi;
const tableOpenRegex = /<table\b[^>]*>/gi;
const tableCloseRegex = /<\/table>/gi;
const rowRegex = /<tr\b[^>]*>/gi;
const cellRegex = /<t[dh]\b[^>]*>/gi;
const complexAttributeRegex = /\s(?:colspan|rowspan|width|style)\s*=/gi;

function countMatches(value: string, pattern: RegExp): number {
	const matches = value.match(pattern);
	return matches?.length ?? 0;
}

function pluralTable(count: number): string {
	return count === 1 ? '1 tabel' : `${count} tabel`;
}

function issueMessage(fieldLabel: string, message: string): string {
	return fieldLabel ? `${fieldLabel}: ${message}` : message;
}

export function summarizeTableHtml(html: string | null | undefined): TableHtmlSummary {
	const source = typeof html === 'string' ? html : '';
	const tableCount = countMatches(source, tableOpenRegex);
	const tableCloseCount = countMatches(source, tableCloseRegex);
	const closedTables = source.match(closedTableRegex) ?? [];
	const analysisSource = tableCount > 0 ? source : closedTables.join('\n');
	const rowCount = tableCount > 0 ? countMatches(analysisSource, rowRegex) : 0;
	const cellCount = tableCount > 0 ? countMatches(analysisSource, cellRegex) : 0;
	const complexAttributeCount = tableCount > 0 ? countMatches(analysisSource, complexAttributeRegex) : 0;
	const malformedTableCount = Math.max(0, tableCount - tableCloseCount);
	const issues: TableHtmlIssue[] = [];

	if (tableCount > 0 && (rowCount === 0 || malformedTableCount > 0)) {
		issues.push({
			severity: 'blocker',
			code: 'table-without-rows',
			message: 'Tabel belum memiliki baris (<tr>) atau markup tabel belum tertutup, sehingga berisiko rusak saat disimpan.',
			evidence: malformedTableCount > 0 ? `${pluralTable(malformedTableCount)} tidak tertutup lengkap` : `${pluralTable(tableCount)} tanpa baris`,
		});
	}

	if (tableCount > 0 && (cellCount === 0 || malformedTableCount > 0)) {
		issues.push({
			severity: 'blocker',
			code: 'table-without-cells',
			message: 'Tabel belum memiliki sel (<td> atau <th>) atau markup tabel belum tertutup, sehingga berisiko kehilangan isi.',
			evidence: malformedTableCount > 0 ? `${pluralTable(malformedTableCount)} tidak tertutup lengkap` : `${pluralTable(tableCount)} tanpa sel`,
		});
	}

	if (complexAttributeCount > 0) {
		issues.push({
			severity: 'warning',
			code: 'contains-complex-table-attrs',
			message:
				'Tabel memakai atribut kompleks (colspan, rowspan, width, atau style). Simpan tetap diizinkan, tetapi tampilan perlu diperiksa ulang.',
			evidence: `${complexAttributeCount} atribut kompleks`,
		});
	}

	return {
		tableCount,
		rowCount,
		cellCount,
		hasTable: tableCount > 0,
		hasRows: rowCount > 0,
		hasCells: cellCount > 0,
		hasComplexAttrs: complexAttributeCount > 0,
		issues,
	};
}

export function compareTableHtmlBeforeSave(args: CompareTableHtmlBeforeSaveArgs): TableHtmlIssue[] {
	const fieldLabel = args.fieldLabel.trim();
	const serverSummary = summarizeTableHtml(args.serverHtml);
	const draftSummary = summarizeTableHtml(args.draftHtml);
	const issues = draftSummary.issues.map((issue) => ({ ...issue, message: issueMessage(fieldLabel, issue.message) }));

	if (serverSummary.hasTable && !draftSummary.hasTable) {
		issues.push({
			severity: 'blocker',
			code: 'suspicious-flat-table-text',
			message: issueMessage(fieldLabel, 'Konten tersimpan memiliki tabel, tetapi draf saat ini tidak memiliki tabel.'),
			evidence: `${serverSummary.tableCount} tabel tersimpan menjadi 0 tabel`,
		});
	}

	if (serverSummary.rowCount > 0 && draftSummary.rowCount < serverSummary.rowCount) {
		const dropRatio = (serverSummary.rowCount - draftSummary.rowCount) / serverSummary.rowCount;
		if (dropRatio >= 0.4) {
			issues.push({
				severity: 'blocker',
				code: 'row-count-dropped',
				message: issueMessage(fieldLabel, 'Jumlah baris tabel turun 40% atau lebih dibanding versi tersimpan.'),
				evidence: `${serverSummary.rowCount} baris tersimpan menjadi ${draftSummary.rowCount} baris draf`,
			});
		}
	}

	if (serverSummary.cellCount > 0 && draftSummary.cellCount < serverSummary.cellCount) {
		const dropRatio = (serverSummary.cellCount - draftSummary.cellCount) / serverSummary.cellCount;
		if (dropRatio >= 0.4) {
			issues.push({
				severity: 'blocker',
				code: 'cell-count-dropped',
				message: issueMessage(fieldLabel, 'Jumlah sel tabel turun 40% atau lebih dibanding versi tersimpan.'),
				evidence: `${serverSummary.cellCount} sel tersimpan menjadi ${draftSummary.cellCount} sel draf`,
			});
		}
	}

	return issues;
}

export function hasBlockingTableHtmlIssues(issuesOrSummary: TableHtmlIssue[] | TableHtmlSummary): boolean {
	const issues = Array.isArray(issuesOrSummary) ? issuesOrSummary : issuesOrSummary.issues;
	return issues.some((issue) => issue.severity === 'blocker');
}
