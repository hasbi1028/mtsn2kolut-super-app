import type { PageServerLoad } from './$types.js';
import { env } from '$env/dynamic/private';
import { redirect } from '@sveltejs/kit';
import { hasAnyRole } from '$lib/server/route-access';

const API_BASE = (env.API_BASE_URL ?? 'http://localhost:8080').replace(/\/$/, '');

export const load: PageServerLoad = async ({ fetch, locals }) => {
	if (!hasAnyRole(locals.user, ['admin'])) throw redirect(302, '/');
	const accessToken = locals.accessToken as string | undefined;
	const headers: Record<string, string> = { 'Content-Type': 'application/json' };
	if (accessToken) headers['Authorization'] = `Bearer ${accessToken}`;

	const res = await fetch(`${API_BASE}/api/academic/timetable/weekly`, { headers });

	let classes: any[] = [];
	let subjects: any[] = [];
	let teachers: any[] = [];
	let assignments: any[] = [];
	let slots: any[] = [];

	if (res.ok) {
		const payload = await res.json();
		const d = payload.data ?? {};
		classes = d.classes ?? [];
		subjects = d.subjects ?? [];
		teachers = d.teachers ?? [];
		assignments = d.assignments ?? [];
		slots = d.slots ?? [];
	}

	return { classes, subjects, teachers, assignments, slots };
};
