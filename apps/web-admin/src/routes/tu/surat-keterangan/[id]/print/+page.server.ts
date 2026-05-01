import { redirect } from '@sveltejs/kit';
import type { PageServerLoad } from './$types';
import { proxy } from '$lib/server/api';
import type { SchoolProfile } from '$lib/school-profile';

type CertificateDetail = {
	id: string;
	template_name: string;
	template_body: string;
	student_nis: string;
	student_nisn: string;
	student_name: string;
	student_gender: string;
	student_status: string;
	class_name: string;
	parent_name: string;
	nik: string;
	tempat_lahir: string;
	tanggal_lahir: string | null;
	alamat: string;
	agama: string;
	nomor_surat: string;
	classification_code: string;
	tanggal_surat: string;
	purpose: string;
	recipient: string;
	remarks: string;
	status: string;
	created_by_username: string;
};

export const load: PageServerLoad = async (event) => {
	const roles = event.locals.user?.roles ?? (event.locals.user?.role ? [event.locals.user.role] : []);
	if (!roles.includes('admin') && !roles.includes('staf')) {
		const from = encodeURIComponent(event.url.pathname + event.url.search);
		throw redirect(302, `/?from=${from}`);
	}
	const p = proxy(event);
	const [certificate, schoolProfile] = await Promise.all([
		p.get<CertificateDetail>(`/api/tu/surat-keterangan/${event.params.id}`),
		p.get<SchoolProfile>('/api/school-profile'),
	]);
	return { certificate, schoolProfile };
};
