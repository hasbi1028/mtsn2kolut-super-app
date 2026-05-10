import type { RequestHandler } from './$types';
import { serveMobileReleaseFile } from '$lib/server/mobile-release';

export const GET: RequestHandler = () => {
	return serveMobileReleaseFile('latest-arm64.sha256', {
		contentType: 'text/plain; charset=utf-8',
		downloadName: 'latest-arm64.sha256',
		inline: true
	});
};
