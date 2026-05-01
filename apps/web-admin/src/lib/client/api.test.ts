import { describe, expect, it } from 'vitest';

import {
	clientApiErrorMessage,
	clientApiPath,
	clientApiPathWithQuery,
	readClientApiData,
	readClientJson
} from './api';

describe('client API helpers', () => {
	it('encodes dynamic client path segments', () => {
		const path = clientApiPath`/api/cbt/sessions/${'session 1/2026'}/participants/${'siswa?1'}/seat`;

		expect(path).toBe('/api/cbt/sessions/session%201%2F2026/participants/siswa%3F1/seat');
	});

	it('appends client query strings only when present', () => {
		expect(clientApiPathWithQuery('/api/academic', new URLSearchParams())).toBe('/api/academic');
		expect(clientApiPathWithQuery('/api/academic', new URLSearchParams({
			entity: 'timetables',
			id: 'slot 1/2026'
		}))).toBe('/api/academic?entity=timetables&id=slot+1%2F2026');
		expect(clientApiPathWithQuery('/api/cbt/questions', '?limit=20')).toBe('/api/cbt/questions?limit=20');
	});

	it('reads successful JSON payloads without assuming an envelope', async () => {
		const response = new Response(JSON.stringify([{ id: 'row-1' }]), { status: 200 });

		await expect(readClientJson(response)).resolves.toEqual([{ id: 'row-1' }]);
	});

	it('uses backend error fields for failed responses', async () => {
		const response = new Response(JSON.stringify({ error: 'Tidak diizinkan' }), { status: 403 });

		await expect(readClientJson(response)).rejects.toThrow('Tidak diizinkan');
	});

	it('falls back to message fields when error is absent', () => {
		expect(clientApiErrorMessage({ message: 'Perlu login ulang' }, 401)).toBe('Perlu login ulang');
	});

	it('masks non-JSON 5xx failures with a generic connectivity message', async () => {
		const response = new Response('upstream stack trace', { status: 502 });

		await expect(readClientJson(response)).rejects.toThrow('Backend tidak dapat dihubungi');
	});

	it('treats 204 responses as successful null payloads', async () => {
		const response = new Response(null, { status: 204 });

		await expect(readClientJson(response)).resolves.toBeNull();
	});

	it('rejects successful empty or malformed bodies with a controlled error', async () => {
		const response = new Response('', { status: 200 });

		await expect(readClientJson(response)).rejects.toThrow('Respons backend kosong atau bukan JSON');
	});

	it('unwraps API data envelopes for client screens', async () => {
		const response = new Response(JSON.stringify({ data: { total: 3 } }), { status: 200 });

		await expect(readClientApiData(response, 'Gagal memuat data')).resolves.toEqual({ total: 3 });
	});

	it('unwraps API items envelopes for list screens', async () => {
		const response = new Response(JSON.stringify({ items: [{ id: 'item-1' }] }), { status: 200 });

		await expect(readClientApiData(response, 'Gagal memuat daftar')).resolves.toEqual([{ id: 'item-1' }]);
	});

	it('preserves paginated list payloads that include items and meta', async () => {
		const payload = { items: [{ id: 'item-1' }], meta: { total: 1, limit: 10, offset: 0 } };
		const response = new Response(JSON.stringify(payload), { status: 200 });

		await expect(readClientApiData(response, 'Gagal memuat daftar')).resolves.toEqual(payload);
	});

	it('rejects successful API envelopes that still carry an error field', async () => {
		const response = new Response(JSON.stringify({ error: 'Operasi ditolak' }), { status: 200 });

		await expect(readClientApiData(response, 'Gagal menjalankan operasi')).rejects.toThrow('Operasi ditolak');
	});

	it('rejects empty 204 responses when a data payload is required', async () => {
		const response = new Response(null, { status: 204 });

		await expect(readClientApiData(response, 'Data wajib ada')).rejects.toThrow('Data wajib ada');
	});
});
