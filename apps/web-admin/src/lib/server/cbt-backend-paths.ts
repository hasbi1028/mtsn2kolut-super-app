import { apiPath, apiPathWithQuery } from '$lib/server/api';

const BACKEND_API_ROOT = '/api';
const BACKEND_CBT_SEGMENT = 'cbt';
const BACKEND_CBT_PREFIX = `${BACKEND_API_ROOT}/${BACKEND_CBT_SEGMENT}`;

export function cbtBackendPath(path: string): string {
	const normalizedPath = path.startsWith('/') ? path : `/${path}`;
	return `${BACKEND_CBT_PREFIX}${normalizedPath}`;
}

export function cbtApiPath(
	strings: TemplateStringsArray,
	...values: Array<string | number | boolean>
): string {
	return `${BACKEND_CBT_PREFIX}${apiPath(strings, ...values)}`;
}

export function cbtBackendPathWithQuery(path: string, params: URLSearchParams | string): string {
	return apiPathWithQuery(cbtBackendPath(path), params);
}
