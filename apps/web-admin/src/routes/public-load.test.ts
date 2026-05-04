import { beforeEach, describe, expect, it, vi } from 'vitest';

const apiPublicGetWithFetchMock = vi.fn();

vi.mock('$lib/server/api', () => ({
	apiPublicGetWithFetch: apiPublicGetWithFetchMock
}));

function event(fetch = vi.fn<typeof globalThis.fetch>(), locals: Record<string, unknown> = {}) {
	return { fetch, locals };
}

describe('public website server loads', () => {
	beforeEach(() => {
		apiPublicGetWithFetchMock.mockReset();
	});

	it('home skips public content aggregation for authenticated dashboard users', async () => {
		const mod = await import('./+page.server');

		await expect(mod.load(event(undefined, { user: { id: '1' } }) as never)).resolves.toEqual({});
		expect(apiPublicGetWithFetchMock).not.toHaveBeenCalled();
	});

	it('home keeps optional featured/profile/PPDB content defensive when unavailable', async () => {
		const mod = await import('./+page.server');
		const fetchMock = vi.fn<typeof globalThis.fetch>();
		apiPublicGetWithFetchMock.mockImplementation((_fetch: typeof globalThis.fetch, path: string) => {
			if (path === '/api/public/site/posts?limit=3') return Promise.resolve([{ id: 'post-1' }]);
			if (path === '/api/public/site/announcements?limit=4') return Promise.resolve([{ id: 'ann-1' }]);
			return Promise.reject(new Error(`missing optional ${path}`));
		});

		await expect(mod.load(event(fetchMock) as never)).resolves.toEqual({
			publicHome: {
				posts: [{ id: 'post-1' }],
				featuredPosts: [],
				announcements: [{ id: 'ann-1' }],
				profil: null,
				ppdbInfo: null
			}
		});
		expect(apiPublicGetWithFetchMock).toHaveBeenCalledWith(fetchMock, '/api/public/site/posts?limit=3');
		expect(apiPublicGetWithFetchMock).toHaveBeenCalledWith(fetchMock, '/api/public/site/posts/featured?limit=2');
		expect(apiPublicGetWithFetchMock).toHaveBeenCalledWith(fetchMock, '/api/public/site/announcements?limit=4');
		expect(apiPublicGetWithFetchMock).toHaveBeenCalledWith(fetchMock, '/api/public/site/pages/profil');
		expect(apiPublicGetWithFetchMock).toHaveBeenCalledWith(fetchMock, '/api/public/site/pages/ppdb-info');
	});

	it('public listing and profile loads use the public API fetch boundary', async () => {
		const posts = await import('./berita/+page.server');
		const announcements = await import('./pengumuman/+page.server');
		const profile = await import('./profil/+page.server');
		const fetchMock = vi.fn<typeof globalThis.fetch>();
		apiPublicGetWithFetchMock
			.mockResolvedValueOnce([{ id: 'post-1' }])
			.mockResolvedValueOnce([{ id: 'ann-1' }])
			.mockResolvedValueOnce({ id: 'profil' });

		await expect(posts.load(event(fetchMock) as never)).resolves.toEqual({ items: [{ id: 'post-1' }] });
		await expect(announcements.load(event(fetchMock) as never)).resolves.toEqual({ items: [{ id: 'ann-1' }] });
		await expect(profile.load(event(fetchMock) as never)).resolves.toEqual({ page: { id: 'profil' } });
		expect(apiPublicGetWithFetchMock).toHaveBeenNthCalledWith(1, fetchMock, '/api/public/site/posts?limit=24');
		expect(apiPublicGetWithFetchMock).toHaveBeenNthCalledWith(2, fetchMock, '/api/public/site/announcements?limit=24');
		expect(apiPublicGetWithFetchMock).toHaveBeenNthCalledWith(3, fetchMock, '/api/public/site/pages/profil');
	});
});
