import { error } from '@sveltejs/kit';
import type { RequestHandler } from './$types';
import { serveMobileReleaseFile } from '$lib/server/mobile-release';

function contentType(filename: string) {
	if (filename.endsWith('.apk')) return 'application/vnd.android.package-archive';
	if (filename.endsWith('.sha256')) return 'text/plain; charset=utf-8';
	if (filename.endsWith('.json')) return 'application/json; charset=utf-8';
	throw error(404, 'Mobile release artifact tidak didukung');
}

export const GET: RequestHandler = ({ params }) => {
	const filename = params.filename;
	if (!/^[A-Za-z0-9._-]+\.(apk|sha256|json)$/.test(filename)) {
		throw error(404, 'Mobile release artifact tidak ditemukan');
	}
	return serveMobileReleaseFile(`history/${filename}`, {
		contentType: contentType(filename),
		downloadName: filename,
		inline: !filename.endsWith('.apk')
	});
};
