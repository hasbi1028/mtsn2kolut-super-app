import { beforeEach, describe, expect, it, vi } from 'vitest';

const apiLoginWithFetchMock = vi.fn();

vi.mock('$lib/server/api', async (importActual) => {
	const actual = await importActual<typeof import('$lib/server/api')>();
	return {
		...actual,
		apiLoginWithFetch: apiLoginWithFetchMock,
	};
});

describe('login action', () => {
	beforeEach(() => {
		apiLoginWithFetchMock.mockReset();
	});

	it('maps backend 403 auth rejection to a controlled disabled-account message', async () => {
		const { ApiError } = await import('$lib/server/api');
		const { actions } = await import('./+page.server');
		apiLoginWithFetchMock.mockRejectedValueOnce(new ApiError(403, 'Akun pengguna sedang dinonaktifkan'));

		const result = await actions.default({
			request: new Request('http://localhost/login', {
				method: 'POST',
				headers: { 'content-type': 'application/x-www-form-urlencoded' },
				body: new URLSearchParams({ username: 'guru', password: 'password' }),
			}),
			cookies: { set: vi.fn() },
			url: new URL('http://localhost/login'),
			getClientAddress: () => '127.0.0.1',
			fetch: vi.fn<typeof fetch>(),
		} as never);

		expect(result).toMatchObject({
			status: 403,
			data: { error: 'Akun pengguna sedang dinonaktifkan' },
		});
	});
});
