import type { PageServerLoad } from './$types.js';

export const load: PageServerLoad = async ({ fetch, url }) => {
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
