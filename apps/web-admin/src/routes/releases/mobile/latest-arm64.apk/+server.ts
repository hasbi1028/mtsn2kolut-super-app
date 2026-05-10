import type { RequestHandler } from './$types';
import { serveMobileReleaseFile } from '$lib/server/mobile-release';

export const GET: RequestHandler = () => {
	return serveMobileReleaseFile('latest-arm64.apk', {
		contentType: 'application/vnd.android.package-archive',
		downloadName: 'mtsn2kolut-cbt-latest-arm64.apk'
	});
};
