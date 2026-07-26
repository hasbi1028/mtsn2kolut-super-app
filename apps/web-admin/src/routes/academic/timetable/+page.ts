import type { PageLoad } from './$types.js';
import { redirect } from '@sveltejs/kit';

export const load: PageLoad = async ({ fetch, parent }) => {
	const { user } = await parent();
	const isAdmin = user?.role === 'admin' || user?.roles?.includes('admin');
	if (!isAdmin) throw redirect(302, '/');

	const res = await fetch('/api/academic/timetable');
	let classes: any[] = [];
	let subjects: any[] = [];
	let teachers: any[] = [];
	let slots: any[] = [];
	let assignments: any[] = [];

	if (res.ok) {
		const payload = await res.json();
		classes = payload.classes ?? [];
		subjects = payload.subjects ?? [];
		teachers = payload.teachers ?? [];
		slots = payload.slots ?? [];
		assignments = payload.assignments ?? [];
	}

	return { classes, subjects, teachers, slots, assignments };
};
