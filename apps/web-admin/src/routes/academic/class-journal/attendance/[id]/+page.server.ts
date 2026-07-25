import type { PageServerLoad } from './$types.js';

export const load: PageServerLoad = async ({ fetch, params }) => {
	const id = params.id;

	const [sessionRes] = await Promise.all([
		fetch(`/api/class-journal/sessions/${id}`),
	]);

	let session: any = null;
	if (sessionRes.ok) {
		session = await sessionRes.json();
	}

	return { session };
};
