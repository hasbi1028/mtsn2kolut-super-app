import type { LayoutLoad } from './$types';
import { redirect } from '@sveltejs/kit';
import { defaultBranding, normalizeBranding, type BrandingSettings } from '$lib/branding';
import type { AccountIdentity } from '$lib/client/account';

export const load: LayoutLoad = async ({ fetch, url }) => {
	let user: Record<string, unknown> | null = null;
	let account: AccountIdentity | null = null;
	let branding: BrandingSettings = defaultBranding;

	// Fetch branding (public)
	try {
		const r = await fetch('/api/public/branding');
		if (r.ok) branding = normalizeBranding(await r.json());
	} catch { /* keep default */ }

	// Fetch maintenance status
	try {
		const r = await fetch('/api/system/maintenance/status');
		if (r.ok) {
			const payload = await r.json();
			// just consume, not used in layout currently
		}
	} catch { /* ignore */ }

	// Fetch current user (auth check via cookies)
	try {
		const r = await fetch('/api/auth/account');
		if (r.ok) {
			const payload = await r.json();
			account = payload as AccountIdentity;
			// Derive user from account — include display info
			user = {
				id: account.id,
				username: account.username,
				display_name: account.display_name,
				profile_nama: account.profile_nama,
				profile_type: account.profile_type,
				role: account.role,
				roles: account.roles ?? [account.role],
			};
		}
	} catch { /* not authenticated */ }

	// Redirect to login if not authenticated (skip login page itself)
	if (!user && url.pathname !== '/login') {
		throw redirect(302, `/login?from=${encodeURIComponent(url.pathname + url.search)}`);
	}

	return { user, account, branding } as Record<string, unknown>;
};
