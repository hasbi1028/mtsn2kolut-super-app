import { redirect } from '@sveltejs/kit';
import type { PageServerLoad } from './$types';

export const load: PageServerLoad = async ({ locals, url }) => {
	const roles = locals.user?.roles ?? (locals.user?.role ? [locals.user.role] : []);
	if (!roles.includes('guru') && !roles.includes('siswa') && !roles.includes('ortu')) {
		const from = encodeURIComponent(url.pathname + url.search);
		throw redirect(302, `/?from=${from}`);
	}
	return {};
};
