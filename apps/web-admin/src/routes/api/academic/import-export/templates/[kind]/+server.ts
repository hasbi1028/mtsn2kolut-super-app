import type { RequestEvent } from '@sveltejs/kit';
import { ApiError, handleRouteError, requiredRouteParam } from '$lib/server/api';

const templates: Record<string, { filename: string; header: string[] }> = {
	siswa: {
		filename: 'template-import-siswa.csv',
		header: ['NIS', 'NISN', 'Nama', 'Jenis Kelamin', 'Kode Rombel', 'Status']
	},
	rombel: {
		filename: 'template-import-rombel.csv',
		header: ['Kode Rombel', 'Nama Rombel', 'Tingkat', 'Tahun Ajaran', 'Aktif']
	},
	guru_mapel: {
		filename: 'template-import-guru-mapel.csv',
		header: ['Kode Rombel', 'Kode Mapel', 'NIP Guru', 'Nama Guru']
	},
	jadwal: {
		filename: 'template-import-jadwal-pelajaran.csv',
		header: ['Kode Rombel', 'Kode Mapel', 'NIP Guru', 'Nama Guru', 'Hari', 'Jam Mulai', 'Jam Selesai', 'Ruang', 'Catatan']
	}
};

function csvRow(values: string[]): string {
	return values.map((value) => `"${value.replaceAll('"', '""')}"`).join(',');
}

export const GET = async (event: RequestEvent) => {
	try {
		const kind = requiredRouteParam(event.params.kind, 'kind');
		const template = templates[kind];
		if (!template) throw new ApiError(404, 'Template akademik tidak ditemukan');
		const body = `${csvRow(template.header)}\n`;
		return new Response(body, {
			headers: {
				'content-type': 'text/csv; charset=utf-8',
				'content-disposition': `attachment; filename="${template.filename}"`,
				'x-content-type-options': 'nosniff'
			}
		});
	} catch (e) {
		return handleRouteError(e, 'academic import template GET');
	}
};
