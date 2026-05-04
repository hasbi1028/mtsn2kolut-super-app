import { describe, expect, it } from 'vitest';

import { csvCell, csvRow } from './csv';

describe('CSV helpers', () => {
	it('escapes quotes commas and newlines', () => {
		expect(csvRow(['A', 'B,C', 'D"E', 'baris\nbaru'])).toBe('A,"B,C","D""E","baris\nbaru"');
	});

	it('prefixes formula-like values before export', () => {
		expect(csvCell('=SUM(A1:A2)')).toBe("'=SUM(A1:A2)");
		expect(csvCell('+628123')).toBe("'+628123");
		expect(csvCell('@cmd')).toBe("'@cmd");
	});
});
