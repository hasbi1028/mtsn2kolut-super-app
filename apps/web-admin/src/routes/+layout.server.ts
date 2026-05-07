import type { LayoutServerLoad } from './$types';
import { proxy } from '$lib/server/api';
import type { AccountIdentity } from '$lib/client/account';

export const load: LayoutServerLoad = async (event) => {
	let account: AccountIdentity | null = null;

	if (event.locals.user) {
		try {
			account = await proxy(event).get<AccountIdentity>('/api/auth/account');
		} catch {
			account = null;
		}
	}

	return { user: event.locals.user, account };
};
