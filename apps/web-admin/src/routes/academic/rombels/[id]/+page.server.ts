import type { PageServerLoad } from './$types.js';

type RombelDetail = {
	id: string;
	code: string;
	name: string;
	level: string;
	is_active: boolean;
	academic_year_name: string;
	homeroom_teacher_name: string;
	total_students: number;
	total_subject_teachers: number;
	total_subject_assignments: number;
	total_timetable_slots: number;
};

type StudentRow = {
	student_id: string;
	nis: string;
	nisn: string;
	student_name: string;
	gender: string;
	is_active: boolean;
	status: string;
};

export const load: PageServerLoad = async ({ fetch, params }) => {
	const id = params.id;

	const [detailRes, studentsRes] = await Promise.all([
		fetch(`${params.id ? `/api/academic/rombels/${id}` : ''}`),
		fetch(`/api/academic/rombels/${id}/students`),
	]);

	let detail: RombelDetail | null = null;
	if (detailRes.ok) {
		const payload = await detailRes.json();
		detail = payload ?? null;
	}

	let students: StudentRow[] = [];
	if (studentsRes.ok) {
		const payload = await studentsRes.json();
		students = payload.items ?? [];
	}

	return { detail, students };
};
