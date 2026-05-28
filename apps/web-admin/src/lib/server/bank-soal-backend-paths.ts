import { apiPath, apiPathWithQuery } from '$lib/server/api';

const BANK_SOAL_PREFIX = '/api/bank-soal';
const BANK_SOAL_TOP_SEGMENTS = new Set(['questions', 'assets', 'soal-support']);

function normalizeBankSoalPath(path: string): string {
	if (path.includes('?') || path.includes('#') || path.includes('\\')) {
		throw new Error('bank-soal backend path must be a relative API path without query, hash, or backslash');
	}
	const normalizedPath = path.startsWith('/') ? path : `/${path}`;
	const lowerPath = normalizedPath.toLowerCase();
	if (
		normalizedPath.includes('//')
		|| lowerPath.includes('%2e')
		|| lowerPath.includes('%2f')
		|| lowerPath.includes('%5c')
	) {
		throw new Error('bank-soal backend path contains unsafe traversal or encoded separators');
	}
	const segments = normalizedPath.slice(1).split('/');
	const firstSegment = segments[0] ?? '';
	if (segments.length === 0 || !firstSegment || segments.some((segment) => segment === '.' || segment === '..')) {
		throw new Error('bank-soal backend path must contain safe non-empty segments');
	}
	if (!BANK_SOAL_TOP_SEGMENTS.has(firstSegment)) {
		throw new Error(`unsupported bank-soal backend path segment: ${firstSegment}`);
	}
	return normalizedPath;
}

export function bankSoalBackendPath(path: string): string {
	return `${BANK_SOAL_PREFIX}${normalizeBankSoalPath(path)}`;
}

export function bankSoalApiPath(
	strings: TemplateStringsArray,
	...values: Array<string | number | boolean>
): string {
	const rendered = apiPath(strings, ...values);
	return bankSoalBackendPath(rendered);
}

export function bankSoalBackendPathWithQuery(path: string, params: URLSearchParams | string): string {
	return apiPathWithQuery(bankSoalBackendPath(path), params);
}
