import { readClientApiData } from '$lib/client/api';

export type AccountGenerationStatus = 'ready' | 'created' | 'skipped' | 'failed';
export type FetchLike = typeof fetch;

export type StudentAccountGenerationCandidate = {
	student_id: string;
	nis: string;
	nisn: string;
	nama: string;
	generated_username?: string;
	temporary_password?: string;
	role: string;
	status: AccountGenerationStatus;
	reason?: string;
	user_id?: string;
	existing_user_id?: string;
};

export type StudentAccountGenerationResult = {
	role: string;
	total: number;
	ready: number;
	created: number;
	skipped: number;
	failed: number;
	candidates: StudentAccountGenerationCandidate[];
};

export type ParentAccountGenerationCandidate = {
	parent_id: string;
	nama: string;
	phone?: string;
	child_count: number;
	basis_student_id?: string;
	basis_student_nisn?: string;
	generated_username?: string;
	temporary_password?: string;
	role: string;
	status: AccountGenerationStatus;
	reason?: string;
	user_id?: string;
	existing_user_id?: string;
};

export type ParentAccountGenerationResult = {
	role: string;
	total: number;
	ready: number;
	created: number;
	skipped: number;
	failed: number;
	candidates: ParentAccountGenerationCandidate[];
};

export async function previewStudentAccounts(fetcher: FetchLike = fetch) {
	const res = await fetcher('/api/users/student-accounts/preview');
	return readClientApiData<StudentAccountGenerationResult>(res, 'Gagal memuat preview akun siswa.');
}

export async function generateStudentAccounts(fetcher: FetchLike = fetch) {
	const res = await fetcher('/api/users/student-accounts/generate', { method: 'POST' });
	return readClientApiData<StudentAccountGenerationResult>(res, 'Gagal generate akun siswa.');
}

export async function previewParentAccounts(fetcher: FetchLike = fetch) {
	const res = await fetcher('/api/users/parent-accounts/preview');
	return readClientApiData<ParentAccountGenerationResult>(res, 'Gagal memuat preview akun orang tua.');
}

export async function generateParentAccounts(fetcher: FetchLike = fetch) {
	const res = await fetcher('/api/users/parent-accounts/generate', { method: 'POST' });
	return readClientApiData<ParentAccountGenerationResult>(res, 'Gagal generate akun orang tua.');
}
