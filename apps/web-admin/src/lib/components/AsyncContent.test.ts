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
});
