import type { PageLoad } from './$types.js';
import { redirect } from '@sveltejs/kit';

export const load: PageLoad = async ({ fetch, parent }) => {
	const { user } = await parent();
	const isAdmin = user?.role === 'admin' || user?.roles?.includes('admin');
	if (!isAdmin) throw redirect(302, '/');

	const [profilesRes, rombelsRes] = await Promise.all([
		fetch('/api/academic/curriculum/profiles'),
		fetch('/api/academic/rombels'),
	]);
	let profiles: any[] = [];
	let rombels: any[] = [];
	if (profilesRes.ok) {
		const p = await profilesRes.json();
		// BFF returns { items: [...] }
		profiles = p?.items ?? [];
	}
	if (rombelsRes.ok) {
		const p = await rombelsRes.json();
		// BFF returns { items: [...] }
		rombels = p?.items ?? [];
	}
	return { profiles, rombels };
};
