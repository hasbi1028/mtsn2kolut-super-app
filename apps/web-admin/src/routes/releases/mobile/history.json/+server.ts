import type { RequestHandler } from './$types';
import { serveMobileReleaseFile } from '$lib/server/mobile-release';

export const GET: RequestHandler = () => {
	return serveMobileReleaseFile('history/history.json', {
		contentType: 'application/json; charset=utf-8',
		downloadName: 'history.json',
		inline: true
	});
};
