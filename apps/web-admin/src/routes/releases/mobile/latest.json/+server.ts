import type { RequestHandler } from './$types';
import { serveMobileReleaseFile } from '$lib/server/mobile-release';

export const GET: RequestHandler = () => {
	return serveMobileReleaseFile('latest.json', {
		contentType: 'application/json; charset=utf-8',
		downloadName: 'latest.json',
		inline: true
	});
};
