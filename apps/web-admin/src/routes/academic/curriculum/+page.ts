import type { PageLoad } from './$types.js';
import { redirect } from '@sveltejs/kit';

export const load: PageLoad = async ({ fetch, url, parent }) => {
	const { user } = await parent();
	const isAdmin = user?.role === 'admin' || user?.roles?.includes('admin');
	if (!isAdmin) throw redirect(302, '/');

	const [profilesRes, rombelsRes] = await Promise.all([
		fetch(`${url.origin}/api/academic/curriculum/profiles`),
		fetch(`${url.origin}/api/academic/rombels`),
	]);
	let profiles: any[] = [];
	let rombels: any[] = [];
	if (profilesRes.ok) {
		const p = await profilesRes.json();
		profiles = p?.data ?? p ?? [];
	}
	if (rombelsRes.ok) {
		const p = await rombelsRes.json();
		rombels = p?.data ?? p ?? [];
	}
	return { profiles, rombels };
};
