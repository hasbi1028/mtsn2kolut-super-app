import katex from 'katex';

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

export function renderRichMathHtml(html: string): string {
	if (!html) return '';
	if (typeof document === 'undefined') return html;

	const pass1 = renderBlock(html);

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
			span.innerHTML = rendered;
			t.parentNode?.replaceChild(span, t);
		}
	}

	return body.innerHTML;
}
