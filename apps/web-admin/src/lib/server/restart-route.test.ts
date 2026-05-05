import { describe, expect, it } from 'vitest';
import { POST } from '../../routes/api/pusaka/worker/restart/+server';

describe('PUSAKA worker restart route', () => {
	it('enforces an admin guard in the handler', async () => {
		const response = await POST({
			locals: { user: undefined }
		} as Parameters<typeof POST>[0]);

		expect(response.status).toBe(403);
		await expect(response.json()).resolves.toEqual({ error: 'forbidden: admin role required' });
	});

	it('refuses to control the worker process from the BFF', async () => {
		const response = await POST({
			locals: { user: { role: 'admin', roles: ['admin'] } }
		} as Parameters<typeof POST>[0]);

		expect(response.status).toBe(410);
		await expect(response.json()).resolves.toMatchObject({
			error: expect.stringContaining('tidak dijalankan dari web-admin')
		});
	});
});
