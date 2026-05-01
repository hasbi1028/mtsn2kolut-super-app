import { render, screen } from '@testing-library/svelte';
import { describe, expect, it, vi } from 'vitest';
import AsyncContentHarness from './AsyncContentHarness.svelte';

describe('AsyncContent', () => {
	it('renders the pending snippet while the promise is unresolved', () => {
		render(AsyncContentHarness, {
			props: {
				promise: new Promise(() => {})
			}
		});

		expect(screen.getByText('Memuat data...')).toBeTruthy();
	});

	it('renders fulfilled values from the await block', async () => {
		render(AsyncContentHarness, {
			props: {
				promise: Promise.resolve('siap')
			}
		});

		expect(await screen.findByText('Nilai: siap')).toBeTruthy();
	});

	it('renders the failed snippet when the promise rejects', async () => {
		render(AsyncContentHarness, {
			props: {
				promise: Promise.reject(new Error('server gagal'))
			}
		});

		expect((await screen.findByRole('alert')).textContent).toContain('server gagal');
	});

	it('uses svelte boundary failed fallback for render errors', async () => {
		const onerror = vi.fn();

		render(AsyncContentHarness, {
			props: {
				promise: Promise.resolve('siap'),
				shouldThrow: true,
				onerror
			}
		});

		expect((await screen.findByRole('alert')).textContent).toContain('render failed');
		expect(onerror).toHaveBeenCalledTimes(1);
	});

	it('renders a default recovery panel when no failed snippet is provided for rejected promises', async () => {
		render(AsyncContentHarness, {
			props: {
				promise: Promise.reject(new Error('server gagal')),
				withFailed: false
			}
		});

		expect((await screen.findByText('Terjadi Kendala')).textContent).toContain('Terjadi Kendala');
		expect(screen.getByText('server gagal')).toBeTruthy();
	});

	it('renders a default recovery panel when no failed snippet is provided for render errors', async () => {
		const onerror = vi.fn();

		render(AsyncContentHarness, {
			props: {
				promise: Promise.resolve('siap'),
				shouldThrow: true,
				withFailed: false,
				onerror
			}
		});

		expect(await screen.findByText('render failed')).toBeTruthy();
		expect(screen.getByRole('button', { name: 'Coba Lagi' })).toBeTruthy();
		expect(onerror).toHaveBeenCalledTimes(1);
	});
});
