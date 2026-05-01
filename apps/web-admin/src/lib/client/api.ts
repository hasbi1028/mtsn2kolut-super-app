type ApiErrorPayload = {
	error?: string;
	message?: string;
};

type ApiEnvelope<T> = ApiErrorPayload & {
	data?: T;
	items?: T;
};

type PathValue = string | number | boolean;

function isRecord(value: unknown): value is Record<string, unknown> {
	return typeof value === 'object' && value !== null;
}

export function clientApiPath(strings: TemplateStringsArray, ...values: PathValue[]): string {
	let path = strings[0] ?? '';
	for (let i = 0; i < values.length; i += 1) {
		path += encodeURIComponent(String(values[i]));
		path += strings[i + 1] ?? '';
	}
	return path;
}

export function clientApiPathWithQuery(path: string, params: URLSearchParams | string): string {
	const query = typeof params === 'string' ? params.replace(/^\?/, '') : params.toString();
	return query ? `${path}?${query}` : path;
}

export function clientApiErrorMessage(payload: unknown, status: number): string {
	if (isRecord(payload)) {
		const candidate = payload as ApiErrorPayload;
		return candidate.error ?? candidate.message ?? `HTTP ${status}`;
	}
	return status >= 500 ? 'Backend tidak dapat dihubungi' : `HTTP ${status}`;
}

export async function readClientJson<T>(response: Response): Promise<T> {
	if (response.status === 204) return null as T;
	const payload = await response.json().catch(() => null) as unknown;
	if (!response.ok) {
		throw new Error(clientApiErrorMessage(payload, response.status));
	}
	if (payload === null) {
		throw new Error('Respons backend kosong atau bukan JSON');
	}
	return payload as T;
}

export async function readClientApiData<T>(
	response: Response,
	fallbackMessage = 'Respons backend kosong atau bukan JSON'
): Promise<T> {
	const payload = await readClientJson<ApiEnvelope<T> | T | null>(response);
	if (payload === null) {
		throw new Error(fallbackMessage);
	}
	if (isRecord(payload)) {
		const error = payload.error;
		if (typeof error === 'string' && error.trim()) {
			throw new Error(error);
		}
		if ('data' in payload) {
			const envelope = payload as ApiEnvelope<T>;
			if (envelope.data === undefined) throw new Error(fallbackMessage);
			return envelope.data;
		}
		if ('items' in payload && !('meta' in payload)) {
			const envelope = payload as ApiEnvelope<T>;
			if (envelope.items === undefined) throw new Error(fallbackMessage);
			return envelope.items;
		}
	}
	return payload as T;
}
