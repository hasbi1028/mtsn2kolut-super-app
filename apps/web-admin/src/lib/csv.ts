const FORMULA_PREFIX = /^[=+\-@\t\r]/;

export function csvCell(value: unknown): string {
	let text = value === null || value === undefined ? '' : String(value);
	if (FORMULA_PREFIX.test(text)) text = `'${text}`;
	if (/[",\n\r]/.test(text)) return `"${text.replaceAll('"', '""')}"`;
	return text;
}

export function csvRow(values: unknown[]): string {
	return values.map(csvCell).join(',');
}
