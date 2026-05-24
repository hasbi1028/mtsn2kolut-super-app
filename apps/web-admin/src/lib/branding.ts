export type BrandingSettings = {
	app_name: string;
	short_name: string;
	tagline: string;
	primary_color: string;
	theme_color: string;
	logo_url: string;
	mark_url: string;
	favicon_url: string;
	apple_touch_icon_url: string;
	pwa_icon_192_url: string;
	pwa_icon_512_url: string;
	formal_logo_url: string;
	version: string;
};

export const defaultBranding: BrandingSettings = {
	app_name: 'MTs Negeri 2 Kolaka Utara',
	short_name: 'MTsN 2 Kolut',
	tagline: 'Super App Madrasah',
	primary_color: '#166534',
	theme_color: '#166534',
	logo_url: '/brand/madrasah-mark.svg',
	mark_url: '/brand/madrasah-mark.svg',
	favicon_url: '/favicon.ico',
	apple_touch_icon_url: '/apple-touch-icon.png',
	pwa_icon_192_url: '/pwa-icon-192.png',
	pwa_icon_512_url: '/pwa-icon-512.png',
	formal_logo_url: '/brand/logo-kemenag.png',
	version: 'default'
};

export const appAttribution = {
	productName: 'MTsN 2 Kolut Super App',
	shortLabel: 'by HasbiGML',
	loginLabel: 'Super App MTsN 2 Kolaka Utara · Powered by HasbiGML',
	developerName: 'HasbiGML',
	formalDeveloperLabel: 'Dikembangkan oleh HasbiGML',
	description: 'Platform digital terpadu untuk mendukung layanan MTsN 2 Kolaka Utara.'
} as const;

export function normalizeBranding(value: Partial<BrandingSettings> | null | undefined): BrandingSettings {
	return { ...defaultBranding, ...(value ?? {}) };
}

export function versionedAsset(url: string, version: string) {
	if (!url) return url;
	if (url.includes('?')) return `${url}&v=${encodeURIComponent(version)}`;
	return `${url}?v=${encodeURIComponent(version)}`;
}

export async function fetchPublicBranding(fetcher: typeof fetch = fetch): Promise<BrandingSettings> {
	const res = await fetcher('/api/public/branding');
	if (!res.ok) return defaultBranding;
	return normalizeBranding(await res.json());
}
