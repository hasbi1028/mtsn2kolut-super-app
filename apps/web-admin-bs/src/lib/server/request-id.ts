export const REQUEST_ID_HEADER = 'x-request-id';

const requestIds = new WeakMap<Request, string>();
const REQUEST_ID_PATTERN = /^[A-Za-z0-9._:-]{1,96}$/;

export function sanitizeRequestId(value: string | null | undefined): string {
	const trimmed = value?.trim() ?? '';
	return REQUEST_ID_PATTERN.test(trimmed) ? trimmed : '';
}

export function getOrCreateRequestId(input: Request | Headers): string {
	if (input instanceof Request) {
		const cached = requestIds.get(input);
		if (cached) return cached;
		const created = requestIdFromHeaders(input.headers);
		requestIds.set(input, created);
		return created;
	}
	return requestIdFromHeaders(input);
}

function requestIdFromHeaders(headers: Headers): string {
	return sanitizeRequestId(headers.get(REQUEST_ID_HEADER) ?? headers.get('x-correlation-id')) || `req_${crypto.randomUUID()}`;
}
