import { describe, expect, it } from 'vitest';
import { POST, _parseRestartCommand } from '../../routes/api/pusaka/worker/restart/+server';

describe('PUSAKA worker restart route', () => {
	it('uses the PM2 worker command by default without shell execution', () => {
		expect(_parseRestartCommand()).toEqual({
			file: 'pm2',
			args: ['restart', 'mtsn2kolut-pusaka-worker']
		});
	});

	it('parses a simple configured executable and arguments', () => {
		expect(_parseRestartCommand('/usr/bin/pm2 restart custom-worker --update-env')).toEqual({
			file: '/usr/bin/pm2',
			args: ['restart', 'custom-worker', '--update-env']
		});
	});

	it('rejects shell syntax in configured restart commands', () => {
		expect(() => _parseRestartCommand('pm2 restart worker; curl http://example.test')).toThrow(
			'unsupported shell syntax'
		);
	});

	it('enforces an admin guard in the handler', async () => {
		const response = await POST({
			locals: { user: undefined }
		} as Parameters<typeof POST>[0]);

		expect(response.status).toBe(403);
		await expect(response.json()).resolves.toEqual({ error: 'forbidden: admin role required' });
	});
});
