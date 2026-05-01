import { redirect } from '@sveltejs/kit';
import type { PageServerLoad } from './$types';

export const load: PageServerLoad = ({ locals }) => {
	const roles = locals.user?.roles ?? [];
	const canManage = roles.includes('admin') || roles.includes('kesiswaan');
	const canRead = canManage || roles.includes('guru');

	if (!canRead) {
		throw redirect(302, '/');
	}

	return {
		roles,
		canManage,
	};
};
