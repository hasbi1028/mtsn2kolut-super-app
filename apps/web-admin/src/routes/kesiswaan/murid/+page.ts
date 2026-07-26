import type { PageLoad } from './$types.js';

export const load: PageLoad = async ({ fetch, url, parent }) => {
	const { user } = await parent();

	const res = await fetch('/api/kesiswaan/murid?limit=5000', {
		headers: { 'Content-Type': 'application/json' },
	});

	let items: any[] = [];
	if (res.ok) {
		const payload = await res.json();
		items = payload?.data ?? payload?.items ?? [];
	}

	return { items, user };
};
