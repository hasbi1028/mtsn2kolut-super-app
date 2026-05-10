import type { RequestHandler } from './$types';
import { serveMobileReleaseFile } from '$lib/server/mobile-release';

export const GET: RequestHandler = () => {
	return serveMobileReleaseFile('latest-qr.svg', {
		contentType: 'image/svg+xml; charset=utf-8',
		downloadName: 'latest-qr.svg',
		inline: true
	});
};
