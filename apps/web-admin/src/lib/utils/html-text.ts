const BLOCK_TAGS = new Set([
	'address',
	'article',
	'aside',
	'blockquote',
	'dd',
	'div',
	'dl',
	'dt',
	'figcaption',
	'figure',
	'footer',
	'h1',
	'h2',
	'h3',
	'h4',
	'h5',
	'h6',
	'header',
	'hr',
	'li',
	'main',
	'nav',
	'ol',
	'p',
	'pre',
	'section',
	'table',
	'tbody',
	'td',
	'tfoot',
	'th',
	'thead',
	'tr',
	'ul',
]);

function normalizePlainText(value: string): string {
	return value.replace(/\s+/g, ' ').trim();
}

function fallbackStripHtml(html: string): string {
	return normalizePlainText(
		html
			.replace(/<br\s*\/?>/gi, ' ')
			.replace(/<\/(?:address|article|aside|blockquote|dd|div|dl|dt|figcaption|figure|footer|h[1-6]|header|hr|li|main|nav|ol|p|pre|section|table|tbody|td|tfoot|th|thead|tr|ul)>/gi, ' ')
			.replace(/<[^>]+>/g, ' ')
	);
}

function nodePlainText(node: Node): string {
	if (node.nodeType === 3) {
		return node.textContent ?? '';
	}
	if (node.nodeType !== 1) {
		return '';
	}

	const element = node as Element;
	const tag = element.tagName.toLowerCase();
	if (tag === 'br') return ' ';

	const text = Array.from(element.childNodes).map(nodePlainText).join('');
	if (BLOCK_TAGS.has(tag)) {
		return ` ${text} `;
	}
	return text;
}

export function htmlToPlainText(html: string): string {
	if (!html) return '';
	if (typeof DOMParser === 'undefined') return fallbackStripHtml(html);

	const doc = new DOMParser().parseFromString(html, 'text/html');
	return normalizePlainText(Array.from(doc.body.childNodes).map(nodePlainText).join(''));
}
