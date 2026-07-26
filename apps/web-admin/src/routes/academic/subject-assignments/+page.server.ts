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

	const res = await fetch(`${API_BASE}/api/academic/subject-assignments`, { headers });

	let classes: any[] = [];
	let subjects: any[] = [];
	let teachers: any[] = [];
	let cells: any[] = [];

	if (res.ok) {
		const payload = await res.json();
		classes = payload.classes ?? [];
		subjects = payload.subjects ?? [];
		teachers = payload.teachers ?? [];
		cells = payload.cells ?? [];
	}

	return { classes, subjects, teachers, cells };
};
