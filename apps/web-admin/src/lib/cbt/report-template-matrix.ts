export type CbtReportTemplateKey =
	| 'participant_list'
	| 'exam_cards'
	| 'session_minutes'
	| 'room_attendance'
	| 'session_results'
	| 'item_analysis'
	| 'proctor_event_recap'
	| 'evidence_bundle_index';

export type CbtReportTemplate = {
	key: CbtReportTemplateKey;
	title: string;
	canonicalRoute: string;
	pdfMode: 'print_html_or_browser_pdf' | 'generated_markdown_json_archive';
	excelMode: 'csv_excel_compatible' | 'not_applicable_sensitive_tokens' | 'generated_json_manifest';
	redaction: string;
	boundary: string;
};

export const CBT_REPORT_TEMPLATE_MATRIX: CbtReportTemplate[] = [
	{
		key: 'participant_list',
		title: 'Daftar peserta',
		canonicalRoute: '/asesmen/sesi/[id]/rooms/[rid]/print-pack',
		pdfMode: 'print_html_or_browser_pdf',
		excelMode: 'not_applicable_sensitive_tokens',
		redaction: 'Token peserta appears only in authorized operational print pack; no broad CSV token dump.',
		boundary: 'No public CBT API runtime route; Web Admin uses /api/asesmen/* BFF to Core API.'
	},
	{
		key: 'exam_cards',
		title: 'Kartu peserta/token',
		canonicalRoute: '/asesmen/kegiatan/[id]/exam-cards',
		pdfMode: 'print_html_or_browser_pdf',
		excelMode: 'not_applicable_sensitive_tokens',
		redaction: 'Token output is print-only operational material for authorized roles.',
		boundary: 'No public CBT API runtime route; token generation remains Core API owned.'
	},
	{
		key: 'session_minutes',
		title: 'Berita acara sesi',
		canonicalRoute: '/asesmen/sesi/[id]/minutes',
		pdfMode: 'print_html_or_browser_pdf',
		excelMode: 'not_applicable_sensitive_tokens',
		redaction: 'Minutes follow role-bound token visibility and are not exposed as public export data.',
		boundary: 'No public CBT API runtime route; printable HTML stays in Web Admin.'
	},
	{
		key: 'room_attendance',
		title: 'Absensi ruang',
		canonicalRoute: '/asesmen/sesi/[id]/rooms/[rid]/print-pack',
		pdfMode: 'print_html_or_browser_pdf',
		excelMode: 'not_applicable_sensitive_tokens',
		redaction: 'Attendance with tokens stays in print pack scope, not broad spreadsheet export.',
		boundary: 'No public CBT API runtime route; room data is proxied through /api/asesmen/*.'
	},
	{
		key: 'session_results',
		title: 'Rekap hasil sesi',
		canonicalRoute: '/asesmen/sesi/[id]',
		pdfMode: 'print_html_or_browser_pdf',
		excelMode: 'csv_excel_compatible',
		redaction: 'Result CSV contains participant score fields and no answer key.',
		boundary: 'No public CBT API runtime route; scoring remains Core API owned.'
	},
	{
		key: 'item_analysis',
		title: 'Analisis butir',
		canonicalRoute: '/asesmen/sesi/[id]',
		pdfMode: 'print_html_or_browser_pdf',
		excelMode: 'csv_excel_compatible',
		redaction: 'Item analysis observes answer key role redaction and avoids fake psychometric metrics.',
		boundary: 'No public CBT API runtime route; data comes from existing /api/asesmen/sessions/[id]/item-analysis.'
	},
	{
		key: 'proctor_event_recap',
		title: 'Rekap event pengawas/audit',
		canonicalRoute: '/asesmen/sesi/[id]/rooms/[rid]/proctoring',
		pdfMode: 'print_html_or_browser_pdf',
		excelMode: 'csv_excel_compatible',
		redaction: 'Evidence CSV contains no raw token, password, answer key, or full device fingerprint.',
		boundary: 'No public CBT API runtime route; evidence comes from room-scoped /api/asesmen/* telemetry.'
	},
	{
		key: 'evidence_bundle_index',
		title: 'Final evidence bundle index',
		canonicalRoute: 'docs/cbt-release-evidence-template.md',
		pdfMode: 'generated_markdown_json_archive',
		excelMode: 'generated_json_manifest',
		redaction: 'Generated evidence secret scan rejects DB DSN, env dumps, tokens, passwords, and answer keys.',
		boundary: 'No public CBT API runtime route; bundle is ops-controlled evidence, not product data.'
	}
];

export function reportTemplateByKey(key: CbtReportTemplateKey) {
	return CBT_REPORT_TEMPLATE_MATRIX.find((item) => item.key === key);
}
