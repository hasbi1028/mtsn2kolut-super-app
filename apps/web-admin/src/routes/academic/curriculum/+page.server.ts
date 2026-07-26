import type { PageServerLoad } from './$types.js';
import { redirect } from '@sveltejs/kit';
import { hasAnyRole } from '$lib/server/route-access';

export const load: PageServerLoad = async ({ fetch, url, locals }) => {
	if (!hasAnyRole(locals.user, ['admin'])) throw redirect(302, '/');
	const [profilesRes, rombelsRes] = await Promise.all([
		fetch(`${url.origin}/api/academic/curriculum/profiles`),
		fetch(`${url.origin}/api/academic/rombels`),
	]);

	let profiles: any[] = [];
	if (profilesRes.ok) {
		const p = await profilesRes.json();
		profiles = p.items ?? [];
	}

	let rombels: any[] = [];
	if (rombelsRes.ok) {
		const r = await rombelsRes.json();
		rombels = r.items ?? [];
	}

	return { profiles, rombels };
};
