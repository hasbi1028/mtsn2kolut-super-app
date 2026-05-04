import { render, screen, waitFor } from '@testing-library/svelte';
import userEvent from '@testing-library/user-event';
import { afterEach, describe, expect, it, vi } from 'vitest';
import EventsPage from './+page.svelte';

vi.mock('$lib/components/ui/sonner', () => ({
	toast: {
		success: vi.fn(),
		error: vi.fn(),
	},
}));

afterEach(() => {
	vi.unstubAllGlobals();
	vi.restoreAllMocks();
});

describe('/cbt/events page', () => {
	it('persists existing event status changes through the status endpoint', async () => {
		const events = [{
			id: 'event-1',
			title: 'UTS Semester Ganjil',
			exam_type: 'uts',
			scope: 'grade',
			target_levels: ['VII'],
			academic_year_id: 'year-1',
			academic_year_name: '2025/2026',
			status: 'draft',
			created_at: '2026-05-04T00:00:00Z',
			session_count: 0,
		}];
		let eventListReads = 0;
		const fetchMock = vi.fn(async (input: RequestInfo | URL, init?: RequestInit) => {
			const url = String(input);
			const method = init?.method ?? 'GET';

			if (url === '/api/cbt/events' && method === 'GET') {
				const status = eventListReads++ === 0 ? 'draft' : 'active';
				return jsonResponse({ data: [{ ...events[0], status }] });
			}
			if (url === '/api/academic' && method === 'GET') {
				return jsonResponse({ data: { years: [{ id: 'year-1', name: '2025/2026', is_active: true }] } });
			}
			if (url === '/api/cbt/events/event-1' && method === 'PUT') {
				return jsonResponse({ data: { status: 'updated' } });
			}
			if (url === '/api/cbt/events/event-1/status' && method === 'PATCH') {
				return jsonResponse({ data: { status: 'active' } });
			}

			return jsonResponse({ error: `Unhandled ${method} ${url}` }, 500);
		});
		vi.stubGlobal('fetch', fetchMock);

		render(EventsPage);

		await screen.findAllByText('UTS Semester Ganjil');
		await userEvent.click(screen.getAllByRole('button', { name: 'Edit' })[0]);
		await userEvent.selectOptions(screen.getByLabelText('Status'), 'active');
		await userEvent.click(screen.getByRole('button', { name: 'Perbarui' }));

		await waitFor(() => {
			const statusCall = fetchMock.mock.calls.find(([url, init]) =>
				String(url) === '/api/cbt/events/event-1/status' && init?.method === 'PATCH'
			);
			expect(statusCall).toBeTruthy();
		});

		const putCall = fetchMock.mock.calls.find(([url, init]) =>
			String(url) === '/api/cbt/events/event-1' && init?.method === 'PUT'
		);
		const statusCall = fetchMock.mock.calls.find(([url, init]) =>
			String(url) === '/api/cbt/events/event-1/status' && init?.method === 'PATCH'
		);
		expect(JSON.parse(String(putCall?.[1]?.body))).not.toHaveProperty('status');
		expect(JSON.parse(String(statusCall?.[1]?.body))).toEqual({ status: 'active' });
		expect(putCall?.[1]?.method).toBe('PUT');
		expect(statusCall?.[1]?.method).toBe('PATCH');
		expect(fetchMock.mock.invocationCallOrder[fetchMock.mock.calls.indexOf(putCall!)]).toBeLessThan(
			fetchMock.mock.invocationCallOrder[fetchMock.mock.calls.indexOf(statusCall!)]
		);
		expect(eventListReads).toBeGreaterThan(1);
	});
});

function jsonResponse(body: unknown, status = 200) {
	return Promise.resolve(new Response(JSON.stringify(body), {
		status,
		headers: { 'Content-Type': 'application/json' },
	}));
}
