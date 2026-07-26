import type { PageLoad } from './$types.js';
import { redirect } from '@sveltejs/kit';

export const load: PageLoad = async ({ fetch, parent }) => {
	const { user } = await parent();
	const isAdmin = user?.role === 'admin' || user?.roles?.includes('admin');
	if (!isAdmin) throw redirect(302, '/');

	const res = await fetch('/api/academic/timetable/weekly');
	let classes: any[] = [];
	let subjects: any[] = [];
	let teachers: any[] = [];
	let slots: any[] = [];
	let assignments: any[] = [];

	if (res.ok) {
		const payload = await res.json();
		// Data dari Go API terbungkus dalam { data: { ... } }
		const d = payload.data ?? payload ?? {};
		classes = d.classes ?? [];
		subjects = d.subjects ?? [];
		teachers = d.teachers ?? [];
		assignments = d.assignments ?? [];
		slots = d.slots ?? [];
	}

	return { classes, subjects, teachers, slots, assignments };
};
