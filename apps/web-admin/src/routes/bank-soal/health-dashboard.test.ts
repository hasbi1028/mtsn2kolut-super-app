import { readFileSync } from 'node:fs';
import path from 'node:path';
import { render, screen, waitFor } from '@testing-library/svelte';
import { afterEach, describe, expect, it, vi } from 'vitest';

import BankSoalHealthDashboard from './_components/BankSoalHealthDashboard.svelte';

vi.mock('$app/paths', () => ({
	resolve: (href: string) => href,
}));

afterEach(() => {
	vi.unstubAllGlobals();
	vi.restoreAllMocks();
});

describe('/bank-soal health dashboard', () => {
	it('keeps /bank-soal root as a single health dashboard surface without the legacy dashboard duplicate', () => {
		const pageSource = readFileSync(path.resolve(process.cwd(), 'src/routes/bank-soal/+page.svelte'), 'utf8');

		expect(pageSource).toContain('BankSoalHealthDashboard');
		expect(pageSource).not.toContain('BankSoalListPage');
		expect(pageSource).not.toContain('mode="dashboard"');
	});

	it('loads summary and sampled questions through Bank Soal BFF routes only', async () => {
		const fetchMock = vi.fn(async (input: RequestInfo | URL) => {
			const url = String(input);
			if (url === '/api/bank-soal/summary') {
				return jsonResponse({
					counts: {
						total: 12,
						published: 4,
						approved: 3,
						review: 2,
						draft: 2,
						revision: 1,
					},
					by_subject: [{ subject_id: 'math', subject_name: 'Matematika', total: 12 }],
				});
			}
			if (url === '/api/bank-soal/questions?limit=50&offset=0') {
				return jsonResponse({
					items: [
						{
							id: 'q-1',
							subject_id: 'math',
							subject_name: 'Matematika',
							workflow_status: 'review',
							status: 'draft',
							kd_ref: '',
							cp_ref: 'CP-1',
							tp_ref: '',
							cognitive_level: '',
							difficulty: '',
							explanation_html: '',
						},
					],
					meta: { total: 12, limit: 50, offset: 0 },
				});
			}
			return jsonResponse({ error: `Unhandled ${url}` }, 500);
		});
		vi.stubGlobal('fetch', fetchMock);

		render(BankSoalHealthDashboard, {
			props: {
				data: {
					user: { role: 'guru', roles: ['guru'], permissions: ['bank_soal.read'] },
				},
			},
		});

		expect(await screen.findByRole('heading', { name: 'Bank Soal' })).toBeTruthy();
		expect(await screen.findByText('Menunggu verifikasi')).toBeTruthy();
		expect(screen.getByText('Data Bank Soal')).toBeTruthy();
		await waitFor(() => {
			expect(fetchMock).toHaveBeenCalledWith('/api/bank-soal/summary');
			expect(fetchMock).toHaveBeenCalledWith('/api/bank-soal/questions?limit=50&offset=0');
		});
		expect(fetchMock.mock.calls.map(([url]) => String(url)).every((url) => !url.startsWith('/api/cbt'))).toBe(true);
		expect(screen.queryByRole('link', { name: /Tambah Soal/i })).toBeNull();
		expect(screen.queryByRole('link', { name: /Impor/i })).toBeNull();
		expect(screen.queryByRole('link', { name: /Review/i })).toBeNull();
	});

	it('shows privileged quick actions only when the current user has matching capabilities', async () => {
		vi.stubGlobal('fetch', vi.fn(async (input: RequestInfo | URL) => {
			const url = String(input);
			if (url === '/api/bank-soal/summary') {
				return jsonResponse({ counts: { total: 10, review: 3, draft: 7 } });
			}
			if (url === '/api/bank-soal/questions?limit=50&offset=0') {
				return jsonResponse({ items: [{ id: 'q-1', workflow_status: 'review', status: 'draft' }], meta: { total: 10 } });
			}
			return jsonResponse({ error: `Unhandled ${url}` }, 500);
		}));

		render(BankSoalHealthDashboard, {
			props: {
				data: {
					user: { role: 'admin', roles: ['admin'], permissions: [] },
				},
			},
		});

		expect((await screen.findByRole('link', { name: /Buka Verifikasi/i })).getAttribute('href')).toBe('/bank-soal/verifikasi');
		expect(screen.getAllByRole('link', { name: /Tambah Soal/i }).some((link) => link.getAttribute('href') === '/bank-soal/tambah')).toBe(true);
		expect(screen.getByRole('link', { name: /Impor/i }).getAttribute('href')).toBe('/bank-soal/impor');
	});
});

function jsonResponse(body: unknown, status = 200) {
	return Promise.resolve(new Response(JSON.stringify(body), {
		status,
		headers: { 'Content-Type': 'application/json' },
	}));
}
