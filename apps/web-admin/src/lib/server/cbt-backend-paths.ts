import { apiPath, apiPathWithQuery } from '$lib/server/api';

const BACKEND_API_ROOT = '/api';
const ASESMEN_SEGMENT = 'asesmen';
const BANK_SOAL_SEGMENT = 'bank-soal';
const ASESMEN_PREFIX = `${BACKEND_API_ROOT}/${ASESMEN_SEGMENT}`;
const BANK_SOAL_PREFIX = `${BACKEND_API_ROOT}/${BANK_SOAL_SEGMENT}`;

// Top-level segments owned by Bank Soal in the Go backend. Anything else
// dispatched through this helper goes to the Asesmen namespace.
const BANK_SOAL_TOP_SEGMENTS = new Set(['questions', 'assets', 'soal-support']);

function dispatchPrefix(path: string): string {
	const normalizedPath = path.startsWith('/') ? path : `/${path}`;
	const firstSegment = normalizedPath.slice(1).split(/[/?#]/)[0] ?? '';
	if (BANK_SOAL_TOP_SEGMENTS.has(firstSegment)) {
		return BANK_SOAL_PREFIX;
	}
	return ASESMEN_PREFIX;
}

export function cbtBackendPath(path: string): string {
	const normalizedPath = path.startsWith('/') ? path : `/${path}`;
	return `${dispatchPrefix(normalizedPath)}${normalizedPath}`;
}

export function cbtApiPath(
	strings: TemplateStringsArray,
	...values: Array<string | number | boolean>
): string {
	const rendered = apiPath(strings, ...values);
	return `${dispatchPrefix(rendered)}${rendered}`;
}

export function cbtBackendPathWithQuery(path: string, params: URLSearchParams | string): string {
	return apiPathWithQuery(cbtBackendPath(path), params);
}
