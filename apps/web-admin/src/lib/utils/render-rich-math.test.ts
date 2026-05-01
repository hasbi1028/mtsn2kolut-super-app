import { describe, expect, it } from 'vitest';
import { renderRichMathHtml } from './render-rich-math';

describe('renderRichMathHtml', () => {
	it('removes active content while preserving safe rich HTML', () => {
		const html = renderRichMathHtml(
			'<p onclick="alert(1)">Aman</p><script>alert(1)</script><img src="/media/soal.png" onerror="alert(1)"><a href="javascript:alert(1)">tautan</a>'
		);

		expect(html).toContain('<p>Aman</p>');
		expect(html).toContain('<img src="/media/soal.png">');
		expect(html).toContain('<a>tautan</a>');
		expect(html).not.toContain('onclick');
		expect(html).not.toContain('onerror');
		expect(html).not.toContain('script');
		expect(html).not.toContain('javascript:');
	});

	it('renders inline and block math after sanitizing the source', () => {
		const html = renderRichMathHtml('<p>Nilai $x^2$</p><p>$$a+b$$</p>');

		expect(html).toContain('katex');
		expect(html).toContain('latex-display');
		expect(html).toContain('Nilai');
	});
});
