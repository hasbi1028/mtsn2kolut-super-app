import { render, screen } from '@testing-library/svelte';
import { describe, expect, it } from 'vitest';
import ErrorPage from './+error.svelte';

describe('global error page', () => {
	it('renders a controlled 503 recovery message for auth validation outages', () => {
		render(ErrorPage, {
			props: {
				status: 503,
				error: { message: 'Layanan validasi sesi sedang bermasalah. Silakan coba beberapa saat lagi.' }
			}
		});

		expect(screen.getByText('Layanan sementara bermasalah')).toBeTruthy();
		expect(screen.getByText('Layanan validasi sesi sedang bermasalah. Silakan coba beberapa saat lagi.')).toBeTruthy();
		expect(screen.getByRole('button', { name: 'Coba Muat Ulang' })).toBeTruthy();
	});
});
