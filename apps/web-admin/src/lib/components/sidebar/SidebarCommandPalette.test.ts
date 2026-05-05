import { render, screen } from '@testing-library/svelte';
import userEvent from '@testing-library/user-event';
import { describe, expect, it, vi } from 'vitest';
import SidebarCommandPalette from './SidebarCommandPalette.svelte';

const paletteItems = [
	{ href: '/', label: 'Dashboard', group: 'Utama', pinned: true },
	{ href: '/grades', label: 'Nilai', group: 'Akademik', pinned: true },
	{ href: '/library/loans', label: 'Peminjaman', group: 'Perpustakaan', pinned: false },
	{ href: '/inventory/items', label: 'Daftar Barang', group: 'Inventaris', pinned: false }
];

function renderPalette() {
	return render(SidebarCommandPalette, {
		props: {
			open: true,
			items: paletteItems,
			recentHrefs: ['/library/loans'],
			runCommand: vi.fn(),
			clearRecent: vi.fn(),
			isActive: (href: string) => href === '/grades'
		}
	});
}

describe('SidebarCommandPalette', () => {
	it('renders grouped sections for pinned, recent, and all menu', () => {
		renderPalette();

		expect(screen.getByText('Akses Cepat')).toBeTruthy();
		expect(screen.getByText('Terakhir Dibuka')).toBeTruthy();
		expect(screen.getByText('Semua Menu')).toBeTruthy();
		expect(screen.getAllByText('Nilai')[0]).toBeTruthy();
		expect(screen.getByText('Peminjaman')).toBeTruthy();
	});

	it('filters items from the search input and can clear recent items', async () => {
		const user = userEvent.setup();
		const clearRecent = vi.fn();

		render(SidebarCommandPalette, {
			props: {
				open: true,
				items: paletteItems,
				recentHrefs: ['/library/loans'],
				runCommand: vi.fn(),
				clearRecent,
				isActive: () => false
			}
		});

		await user.type(screen.getByPlaceholderText('Mis. Bank Soal, Hasil & Analisis, Inventaris, atau PUSAKA'), 'inventaris');

		expect(screen.getByText('Daftar Barang')).toBeTruthy();
		expect(screen.queryByText('Peminjaman')).toBeNull();

		await user.clear(screen.getByPlaceholderText('Mis. Bank Soal, Hasil & Analisis, Inventaris, atau PUSAKA'));
		await user.click(screen.getByText('Bersihkan'));

		expect(clearRecent).toHaveBeenCalledTimes(1);
	});
});
