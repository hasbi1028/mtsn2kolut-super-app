import katex from 'katex';

const ALLOWED_TAGS = new Set([
	'a',
	'annotation',
	'b',
	'blockquote',
	'br',
	'code',
	'div',
	'em',
	'h1',
	'h2',
	'h3',
	'h4',
	'h5',
	'h6',
	'hr',
	'i',
	'img',
	'li',
	'math',
	'mfrac',
	'mi',
	'mmultiscripts',
	'mn',
	'mo',
	'mover',
	'mpadded',
	'mphantom',
	'mroot',
	'mrow',
	'ms',
	'mspace',
	'msqrt',
	'msub',
	'msubsup',
	'msup',
	'mtable',
	'mtd',
	'mtext',
	'mtr',
	'munder',
	'munderover',
	'ol',
	'p',
	'pre',
	'semantics',
	'span',
	'strong',
	'sub',
	'sup',
	'table',
	'tbody',
	'td',
	'th',
	'thead',
	'tr',
	'u',
	'ul',
]);

const GLOBAL_ATTRS = new Set([
	'aria-hidden',
	'aria-label',
	'class',
	'colspan',
	'dir',
	'height',
	'role',
	'rowspan',
	'style',
	'title',
	'width',
]);

const TAG_ATTRS: Record<string, Set<string>> = {
	a: new Set(['href', 'rel', 'target']),
	annotation: new Set(['encoding']),
	img: new Set(['alt', 'loading', 'src']),
	math: new Set(['display', 'xmlns']),
	span: new Set(['data-align', 'data-color']),
	td: new Set(['colspan', 'rowspan']),
	th: new Set(['colspan', 'rowspan', 'scope']),
};

function skipAncestors(el: Element | null, root: Element): boolean {
	let cur = el;
	while (cur && cur !== root) {
		const tag = cur.tagName?.toLowerCase();
		if (['code', 'pre', 'script', 'style'].includes(tag)) return true;
		if (
			cur.classList?.contains('katex') ||
			cur.classList?.contains('ql-formula') ||
			cur.classList?.contains('ce-math-inline') ||
			cur.classList?.contains('ce-math-block')
		)
			return true;
		cur = cur.parentElement;
	}
	return false;
}

function renderBlock(html: string): string {
	return html.replace(/\$\$([\s\S]+?)\$\$/g, (_, formula) => {
		try {
			const rendered = katex.renderToString(formula.trim(), { throwOnError: false, displayMode: true });
			return `<div class="latex-display">${rendered}</div>`;
		} catch {
			return `$$${formula}$$`;
		}
	});
}

function renderInline(text: string): string {
	return text.replace(/\$([^$\n]+?)\$/g, (_, formula) => {
		try {
			return katex.renderToString(formula.trim(), { throwOnError: false, displayMode: false });
		} catch {
			return `$${formula}$`;
		}
	});
}

function isSafeUrl(value: string, allowDataImage = false): boolean {
	const trimmed = value.trim();
	if (
		!trimmed ||
		trimmed.startsWith('#') ||
		trimmed.startsWith('/') ||
		trimmed.startsWith('./') ||
		trimmed.startsWith('../')
	) {
		return true;
	}
	if (allowDataImage && /^data:image\/(?:gif|png|jpeg|jpg|webp);base64,/i.test(trimmed)) {
		return true;
	}
	try {
		const baseUrl = typeof window === 'undefined' ? 'http://localhost' : window.location.origin;
		const url = new URL(trimmed, baseUrl);
		return url.protocol === 'http:' || url.protocol === 'https:' || url.protocol === 'mailto:';
	} catch {
		return false;
	}
}

function isSafeStyle(value: string): boolean {
	return !/(?:url\s*\(|expression\s*\()/i.test(value);
}

function isAllowedAttr(tag: string, name: string, value: string): boolean {
	if (name.startsWith('on')) return false;
	if (name === 'href') return tag === 'a' && isSafeUrl(value);
	if (name === 'src') return tag === 'img' && isSafeUrl(value, true);
	if (name === 'style') return isSafeStyle(value);
	return GLOBAL_ATTRS.has(name) || TAG_ATTRS[tag]?.has(name) === true;
}

function unwrapElement(el: Element) {
	const parent = el.parentNode;
	if (!parent) return;
	while (el.firstChild) parent.insertBefore(el.firstChild, el);
	parent.removeChild(el);
}

function sanitizeElement(el: Element) {
	const tag = el.tagName.toLowerCase();
	if (['script', 'style', 'template', 'iframe', 'object', 'embed', 'svg'].includes(tag)) {
		el.remove();
		return;
	}
	for (const child of Array.from(el.children)) {
		sanitizeElement(child);
	}
	if (!ALLOWED_TAGS.has(tag)) {
		unwrapElement(el);
		return;
	}
	for (const attr of Array.from(el.attributes)) {
		const name = attr.name.toLowerCase();
		if (!isAllowedAttr(tag, name, attr.value)) {
			el.removeAttribute(attr.name);
		}
	}
	if (tag === 'a' && el.getAttribute('target') === '_blank') {
		el.setAttribute('rel', 'noopener noreferrer');
	}
}

function sanitizeRichHtml(html: string): string {
	const parser = new DOMParser();
	const doc = parser.parseFromString(html, 'text/html');
	for (const child of Array.from(doc.body.children)) {
		sanitizeElement(child);
	}
	return doc.body.innerHTML;
}

export function renderRichMathHtml(html: string): string {
	if (!html) return '';
	if (typeof document === 'undefined') return '';

	const pass1 = renderBlock(sanitizeRichHtml(html));

	const parser = new DOMParser();
	const doc = parser.parseFromString(pass1, 'text/html');
	const body = doc.body;

	const walker = doc.createTreeWalker(body, NodeFilter.SHOW_TEXT);
	const targets: Text[] = [];
	let node = walker.nextNode();
	while (node) {
		const t = node as Text;
		if (t.textContent?.includes('$') && !skipAncestors(t.parentElement, body)) {
			targets.push(t);
		}
		node = walker.nextNode();
	}

	for (const t of targets) {
		if (!t.textContent?.includes('$')) continue;
		const rendered = renderInline(t.textContent ?? '');
		if (rendered !== t.textContent) {
			const span = doc.createElement('span');
			const renderedDoc = parser.parseFromString(rendered, 'text/html');
			for (const child of Array.from(renderedDoc.body.childNodes)) {
				span.appendChild(doc.importNode(child, true));
			}
			t.parentNode?.replaceChild(span, t);
		}
	}

	return sanitizeRichHtml(body.innerHTML);
}
