import type { PageLoad } from './$types.js';
import { redirect } from '@sveltejs/kit';

export const load: PageLoad = async ({ parent }) => {
	const { user } = await parent();
	const roles = user?.roles ?? (user?.role ? [user.role] : []);
	const isAdmin = roles.includes('admin');
	if (!isAdmin) throw redirect(302, '/');

	return { user };
};
