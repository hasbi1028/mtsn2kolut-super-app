import { describe, expect, it, vi } from 'vitest';
import { fetchSidebarAttention } from './sidebar-attention';

describe('fetchSidebarAttention', () => {
	it('uses canonical PUSAKA job stats route for admin attention badges', async () => {
		const fetchMock = vi.fn<typeof fetch>(async (input) => {
			const url = String(input);
			if (url === '/api/inventory/stats') {
				return new Response(JSON.stringify({ data: { perlu_restok: 2 } }), { status: 200 });
			}
			if (url === '/api/library/stats') {
				return new Response(JSON.stringify({ data: { terlambat: 1, denda_belum_lunas: 3 } }), { status: 200 });
			}
			if (url === '/api/pusaka/jobs/stats') {
				return new Response(JSON.stringify({ failed: 4 }), { status: 200 });
			}
			return new Response(JSON.stringify({ error: 'unexpected route' }), { status: 404 });
		});

		const attention = await fetchSidebarAttention(fetchMock, ['admin']);

		expect(fetchMock).toHaveBeenCalledWith('/api/pusaka/jobs/stats');
		expect(fetchMock).not.toHaveBeenCalledWith('/api/queue/stats');
		expect(attention).toEqual({
			inventory: 2,
			library: 4,
			pusaka: 4
		});
	});

	it('treats malformed stats payloads as zero without rejecting', async () => {
		const fetchMock = vi.fn<typeof fetch>(async () => {
			return new Response('not-json', { status: 200 });
		});

		const attention = await fetchSidebarAttention(fetchMock, ['admin']);

		expect(attention).toEqual({
			inventory: 0,
			library: 0,
			pusaka: 0
		});
	});
});
