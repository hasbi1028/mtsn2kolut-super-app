import { json } from '@sveltejs/kit';

export class ValidationError extends Error {
	constructor(message: string) {
		super(message);
	}
}

export function requireStringField(body: Record<string, unknown>, field: string, label = field): string {
	const value = body[field];
	if (typeof value !== 'string') throw new ValidationError(`${label} tidak valid`);
	const trimmed = value.trim();
	if (!trimmed) throw new ValidationError(`${label} wajib diisi`);
	return trimmed;
}

export function optionalStringField(body: Record<string, unknown>, field: string, label = field): string {
	const value = body[field];
	if (value === undefined || value === null) return '';
	if (typeof value !== 'string') throw new ValidationError(`${label} tidak valid`);
	return value.trim();
}

export function requireUuidField(body: Record<string, unknown>, field: string, label = field): string {
	const value = requireStringField(body, field, label);
	if (!/^[0-9a-f]{8}-[0-9a-f]{4}-[1-5][0-9a-f]{3}-[89ab][0-9a-f]{3}-[0-9a-f]{12}$/i.test(value)) {
		throw new ValidationError(`${label} tidak valid`);
	}
	return value;
}

export function validationErrorResponse(error: unknown): Response | undefined {
	if (error instanceof ValidationError) {
		return json({ error: error.message }, { status: 400 });
	}
	return undefined;
}
