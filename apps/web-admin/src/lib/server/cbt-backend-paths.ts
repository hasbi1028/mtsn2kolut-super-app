import { apiPath, apiPathWithQuery } from '$lib/server/api';

const BACKEND_API_ROOT = '/api';
const CBT_SEGMENT = 'cbt';
const CBT_PREFIX = `${BACKEND_API_ROOT}/${CBT_SEGMENT}`;

const CBT_TOP_SEGMENTS = new Set(['events', 'sessions', 'proctoring', 'readiness', 'approvals']);

function normalizeInternalPath(path: string): string {
	if (path.includes('?') || path.includes('#') || path.includes('\\')) {
		throw new Error('cbt backend path must be a relative API path without query, hash, or backslash');
	}
	const normalizedPath = path.startsWith('/') ? path : `/${path}`;
	const lowerPath = normalizedPath.toLowerCase();
	if (
		normalizedPath.includes('//')
		|| lowerPath.includes('%2e')
		|| lowerPath.includes('%2f')
		|| lowerPath.includes('%5c')
	) {
		throw new Error('cbt backend path contains unsafe traversal or encoded separators');
	}
	const segments = normalizedPath.slice(1).split('/');
	if (segments.length === 0 || !segments[0] || segments.some((segment) => segment === '.' || segment === '..')) {
		throw new Error('cbt backend path must contain safe non-empty segments');
	}
	return normalizedPath;
}

function dispatchPrefix(path: string): string {
	const normalizedPath = normalizeInternalPath(path);
	const firstSegment = normalizedPath.slice(1).split('/')[0] ?? '';
	if (CBT_TOP_SEGMENTS.has(firstSegment)) {
		return CBT_PREFIX;
	}
	throw new Error(`unsupported cbt backend path segment: ${firstSegment}`);
}

export function cbtBackendPath(path: string): string {
	const normalizedPath = normalizeInternalPath(path);
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
