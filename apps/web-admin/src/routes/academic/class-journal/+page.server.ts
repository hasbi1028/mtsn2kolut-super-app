import type { PageServerLoad } from './$types.js';
import { env } from '$env/dynamic/private';

const API_BASE = (env.API_BASE_URL ?? 'http://localhost:8080').replace(/\/$/, '');

export const load: PageServerLoad = async ({ fetch, locals }) => {
	const accessToken = locals.accessToken as string | undefined;
	const headers: Record<string, string> = { 'Content-Type': 'application/json' };
	if (accessToken) headers['Authorization'] = `Bearer ${accessToken}`;

	// Load assignments available for timetable (already have class_subject_assignments)
	// Use the timetable weekly endpoint to get assignments
	const res = await fetch(`${API_BASE}/api/academic/timetable/weekly`, { headers });

	let classes: any[] = [];
	let assignments: any[] = [];

	if (res.ok) {
		const payload = await res.json();
		const d = payload.data ?? {};
		classes = d.classes ?? [];
		assignments = d.assignments ?? [];
	}

	return { classes, assignments };
};
