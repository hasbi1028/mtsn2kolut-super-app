import { render, screen } from '@testing-library/svelte';
import userEvent from '@testing-library/user-event';
import { describe, expect, it, vi } from 'vitest';
import SidebarCommandPalette from './SidebarCommandPalette.svelte';

const paletteItems = [
	{ href: '/', label: 'Dashboard', group: 'Utama', groupSection: '0', section: '0', pinned: true },
	{ href: '/grades', label: 'Nilai', group: 'Akademik', groupSection: '3', section: '5.1.1', pinned: true },
	{ href: '/bank-soal', label: 'Daftar Soal', group: 'Bank Soal', groupSection: '6', ancestorSections: ['6.1'], ancestors: ['Kelola Soal'], section: '6.1.1', pinned: false },
	{ href: '/bank-soal/tambah', label: 'Tambah Soal', group: 'Bank Soal', groupSection: '6', ancestorSections: ['6.1'], ancestors: ['Kelola Soal'], section: '6.1.2', pinned: false },
	{ href: '/bank-soal/verifikasi', label: 'Verifikasi Soal', group: 'Bank Soal', groupSection: '6', ancestorSections: ['6.1'], ancestors: ['Kelola Soal'], section: '6.1.4', pinned: false },
	{ href: '/library/loans', label: 'Peminjaman', group: 'Perpustakaan', groupSection: '9', section: '9.1.3', pinned: false },
	{ href: '/inventory/items', label: 'Daftar Barang', group: 'Inventaris', groupSection: '9', section: '9.2.2', pinned: false }
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

	it('searches standalone Bank Soal routes without legacy CBT or mode links', async () => {
		const user = userEvent.setup();

		renderPalette();

		await user.type(screen.getByPlaceholderText('Mis. Persiapan, Pelaksanaan, Inventaris, atau PUSAKA'), 'bank soal');

		expect(screen.getByText('Tambah Soal')).toBeTruthy();
		expect(screen.getByText('6 Bank Soal › 6.1 Kelola Soal › 6.1.2 Tambah Soal · /bank-soal/tambah')).toBeTruthy();
		expect(screen.getByText('6.1.2')).toBeTruthy();
		expect(document.body.textContent).not.toContain('/cbt/soal');
		expect(document.body.textContent).not.toContain('/cbt/questions');
		expect(document.body.textContent).not.toContain('?mode=');
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

		await user.type(screen.getByPlaceholderText('Mis. Persiapan, Pelaksanaan, Inventaris, atau PUSAKA'), 'inventaris');

		expect(screen.getByText('Daftar Barang')).toBeTruthy();
		expect(screen.queryByText('Peminjaman')).toBeNull();

		await user.clear(screen.getByPlaceholderText('Mis. Persiapan, Pelaksanaan, Inventaris, atau PUSAKA'));
		await user.click(screen.getByText('Bersihkan'));

		expect(clearRecent).toHaveBeenCalledTimes(1);
	});
});
