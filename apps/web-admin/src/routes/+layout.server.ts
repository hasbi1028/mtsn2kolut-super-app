import type { LayoutServerLoad } from './$types';
import { apiPublicGetWithFetch, proxy } from '$lib/server/api';
import { defaultBranding, normalizeBranding, type BrandingSettings } from '$lib/branding';
import type { AccountIdentity } from '$lib/client/account';

export const load: LayoutServerLoad = async (event) => {
	let account: AccountIdentity | null = null;
	let branding: BrandingSettings = defaultBranding;

	try {
		branding = normalizeBranding(await apiPublicGetWithFetch(event.fetch, '/api/public/branding'));
	} catch {
		branding = defaultBranding;
	}

	if (event.locals.user) {
		try {
			account = await proxy(event).get<AccountIdentity>('/api/auth/account');
		} catch {
			account = null;
		}
	}

	return { user: event.locals.user, account, branding };
};
