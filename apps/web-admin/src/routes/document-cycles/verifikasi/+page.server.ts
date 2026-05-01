import { redirect } from '@sveltejs/kit';
import type { PageServerLoad } from './$types';

export const load: PageServerLoad = async ({ locals, url }) => {
	const roles = locals.user?.roles ?? (locals.user?.role ? [locals.user.role] : []);
	if (!roles.includes('admin') && !roles.includes('staf')) {
		const from = encodeURIComponent(url.pathname + url.search);
		throw redirect(302, `/?from=${from}`);
	}
	return { user: locals.user };
};
