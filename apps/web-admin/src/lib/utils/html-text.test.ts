import { describe, expect, it } from 'vitest';
import { htmlToPlainText } from './html-text';

describe('htmlToPlainText', () => {
	it('extracts readable text without injecting the html into a live element', () => {
		expect(htmlToPlainText('<p>Nilai <strong>rapor</strong></p><p>Semester &amp; akhir</p>')).toBe(
			'Nilai rapor Semester & akhir'
		);
	});

	it('keeps block and line break boundaries readable', () => {
		expect(htmlToPlainText('<ul><li>A</li><li>B<br>C</li></ul>')).toBe('A B C');
	});
});
