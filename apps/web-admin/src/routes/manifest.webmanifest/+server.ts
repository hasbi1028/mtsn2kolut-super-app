import type { RequestHandler } from './$types';
import { apiPublicGetWithFetch } from '$lib/server/api';
import { defaultBranding, normalizeBranding, versionedAsset } from '$lib/branding';

export const GET: RequestHandler = async ({ fetch }) => {
	let branding = defaultBranding;
	try {
		branding = normalizeBranding(await apiPublicGetWithFetch(fetch, '/api/public/branding'));
	} catch {
		branding = defaultBranding;
	}
	const manifest = {
		name: branding.app_name,
		short_name: branding.short_name,
		description: branding.tagline || 'Super App Madrasah',
		start_url: '/',
		display: 'standalone',
		background_color: '#ffffff',
		theme_color: branding.theme_color,
		icons: [
			{ src: versionedAsset(branding.pwa_icon_192_url, branding.version), sizes: '192x192', type: 'image/png', purpose: 'any maskable' },
			{ src: versionedAsset(branding.pwa_icon_512_url, branding.version), sizes: '512x512', type: 'image/png', purpose: 'any maskable' }
		]
	};
	return Response.json(manifest, {
		headers: {
			'content-type': 'application/manifest+json',
			'cache-control': 'public, max-age=300, stale-while-revalidate=86400'
		}
	});
};
