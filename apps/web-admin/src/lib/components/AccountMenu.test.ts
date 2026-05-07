import { render, screen } from '@testing-library/svelte';
import userEvent from '@testing-library/user-event';
import { describe, expect, it } from 'vitest';
import AccountMenu from './AccountMenu.svelte';

describe('AccountMenu', () => {
	it('opens an account menu with profile data, roles, and account link', async () => {
		const user = userEvent.setup();

		const { container } = render(AccountMenu, {
			props: {
				user: { id: 'u1', username: 'guru.ipa', role: 'guru', roles: ['guru'] },
				account: {
					id: 'u1',
					username: 'guru.ipa',
					display_name: 'Guru IPA',
					avatar_url: '/api/auth/account/avatar/u1.jpg',
					roles: ['guru', 'kesiswaan']
				},
				showName: true,
				includeThemeToggle: false
			}
		});

		expect(screen.getByRole('button', { name: /Buka menu akun Guru IPA/ })).toBeTruthy();
		expect(container.querySelector('img')?.getAttribute('src')).toBe('/api/auth/account/avatar/u1.jpg');

		await user.click(screen.getByRole('button', { name: /Buka menu akun Guru IPA/ }));

		expect(screen.getByRole('menu', { name: 'Menu akun' })).toBeTruthy();
		expect(screen.getAllByText('Guru IPA').length).toBeGreaterThan(0);
		expect(screen.getAllByText('@guru.ipa').length).toBeGreaterThan(0);
		expect(screen.getByText('Guru')).toBeTruthy();
		expect(screen.getByText('Kesiswaan')).toBeTruthy();
		expect(screen.getByRole('menuitem', { name: /Profil Saya/ }).getAttribute('href')).toBe('/settings/account');
		expect(screen.getByRole('menuitem', { name: 'Logout' })).toBeTruthy();
	});

	it('falls back to initials when no avatar URL is available', () => {
		render(AccountMenu, {
			props: {
				user: { id: 'u2', username: 'guru.matematika', role: 'guru', roles: ['guru'] },
				account: {
					id: 'u2',
					username: 'guru.matematika',
					display_name: 'Guru Matematika',
					roles: ['guru']
				},
				includeThemeToggle: false
			}
		});

		expect(screen.getByText('GM')).toBeTruthy();
	});

	it('closes with Escape and outside click', async () => {
		const user = userEvent.setup();

		render(AccountMenu, {
			props: {
				user: { id: 'u3', username: 'admin', role: 'admin', roles: ['admin'] },
				includeThemeToggle: false
			}
		});

		const trigger = screen.getByRole('button', { name: /Buka menu akun admin/ });
		await user.click(trigger);
		expect(screen.getByRole('menu', { name: 'Menu akun' })).toBeTruthy();

		await user.keyboard('{Escape}');
		expect(screen.queryByRole('menu', { name: 'Menu akun' })).toBeNull();

		await user.click(trigger);
		expect(screen.getByRole('menu', { name: 'Menu akun' })).toBeTruthy();
		await user.click(document.body);
		expect(screen.queryByRole('menu', { name: 'Menu akun' })).toBeNull();
	});
});
