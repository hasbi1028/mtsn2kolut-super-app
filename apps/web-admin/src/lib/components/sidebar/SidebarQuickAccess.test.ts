import { render, screen } from '@testing-library/svelte';
import userEvent from '@testing-library/user-event';
import { describe, expect, it, vi } from 'vitest';
import SidebarQuickAccess from './SidebarQuickAccess.svelte';

vi.mock('$app/paths', () => ({
	resolve: (href: string) => href
}));

const baseItems: Array<{
	href: string;
	label: string;
	icon: 'grid' | 'clipboard' | 'calendar';
	pinnable?: boolean;
	group: string;
}> = [
	{ href: '/', label: 'Dashboard', icon: 'grid', pinnable: false, group: 'Utama' },
	{ href: '/grades', label: 'Nilai', icon: 'clipboard', group: 'Akademik' },
	{ href: '/jadwal', label: 'Jadwal', icon: 'calendar', group: 'Akademik' }
];

function renderQuickAccess() {
	return render(SidebarQuickAccess, {
		props: {
			items: baseItems,
			desktopExpanded: true,
			isActive: (href: string) => href === '/grades',
			rememberRecent: vi.fn(),
			togglePin: vi.fn(),
			isPinned: (href: string) => href !== '/',
			pinButtonLabel: (item: { label: string }) => `Pin ${item.label}`,
			canMovePinned: (href: string, direction: -1 | 1) => !(href === '/grades' && direction === -1),
			movePinned: vi.fn(),
			railTooltip: (item: { label: string }, group: string) => `${group} · ${item.label}`,
			closeMobile: vi.fn()
		}
	});
}

describe('SidebarQuickAccess', () => {
	it('renders quick access items and hides pin controls for dashboard', () => {
		renderQuickAccess();

		expect(screen.getByText('Akses Cepat')).toBeTruthy();
		expect(screen.getByText('Dashboard')).toBeTruthy();
		expect(screen.getByText('Nilai')).toBeTruthy();
		expect(screen.queryByLabelText('Pin Dashboard')).toBeNull();
	});

	it('triggers move and pin handlers for pinned items', async () => {
		const user = userEvent.setup();
		const movePinned = vi.fn();
		const togglePin = vi.fn();

		render(SidebarQuickAccess, {
			props: {
				items: baseItems,
				desktopExpanded: true,
				isActive: () => false,
				rememberRecent: vi.fn(),
				togglePin,
				isPinned: (href: string) => href !== '/',
				pinButtonLabel: (item: { label: string }) => `Pin ${item.label}`,
				canMovePinned: () => true,
				movePinned,
				railTooltip: (item: { label: string }, group: string) => `${group} · ${item.label}`,
				closeMobile: vi.fn()
			}
		});

		await user.click(screen.getByLabelText('Naikkan Nilai dalam akses cepat'));
		await user.click(screen.getByLabelText('Pin Nilai'));

		expect(movePinned).toHaveBeenCalledWith('/grades', -1);
		expect(togglePin).toHaveBeenCalledWith('/grades');
	});
});
