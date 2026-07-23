import type { PageServerLoad } from './$types';
import { apiPublicGetWithFetch } from '$lib/server/api';

type WebsiteContent = {
	id: string;
	kind: 'page' | 'post' | 'announcement';
	title: string;
	slug: string;
	excerpt: string;
	content_html: string;
	cover_image_url: string;
	published_at: string | null;
};

type PublicHomePayload = {
	publicHome: {
		posts: WebsiteContent[];
		featuredPosts: WebsiteContent[];
		announcements: WebsiteContent[];
		profil: WebsiteContent | null;
		ppdbInfo: WebsiteContent | null;
	};
};

const PUBLIC_HOME_CACHE_TTL_MS = 60_000;
const PUBLIC_HOME_EDGE_MAX_AGE_SECONDS = 300;
let publicHomeCache: { expiresAt: number; payload: PublicHomePayload } | null = null;
let publicHomeCachePromise: Promise<PublicHomePayload> | null = null;

async function loadPublicHome(fetcher: typeof fetch): Promise<PublicHomePayload> {
	const [posts, featuredPosts, announcements, profil, ppdbInfo] = await Promise.all([
		apiPublicGetWithFetch<WebsiteContent[]>(fetcher, '/api/public/site/posts?limit=3').catch(() => []),
		apiPublicGetWithFetch<WebsiteContent[]>(fetcher, '/api/public/site/posts/featured?limit=2').catch(() => []),
		apiPublicGetWithFetch<WebsiteContent[]>(fetcher, '/api/public/site/announcements?limit=4').catch(() => []),
		apiPublicGetWithFetch<WebsiteContent>(fetcher, '/api/public/site/pages/profil').catch(() => null),
		apiPublicGetWithFetch<WebsiteContent>(fetcher, '/api/public/site/pages/ppdb-info').catch(() => null),
	]);

	return {
		publicHome: {
			posts,
			featuredPosts,
			announcements,
			profil,
			ppdbInfo,
		},
	};
}

export const load: PageServerLoad = async ({ fetch, locals, setHeaders }) => {
	if (locals.user) return {};

	setHeaders({
		'cache-control': `public, max-age=60, s-maxage=${PUBLIC_HOME_EDGE_MAX_AGE_SECONDS}, stale-while-revalidate=600`,
	});

	const now = Date.now();
	if (publicHomeCache && publicHomeCache.expiresAt > now) return publicHomeCache.payload;

	publicHomeCachePromise ??= loadPublicHome(fetch).then((payload) => {
		publicHomeCache = { expiresAt: Date.now() + PUBLIC_HOME_CACHE_TTL_MS, payload };
		return payload;
	}).finally(() => {
		publicHomeCachePromise = null;
	});

	return publicHomeCachePromise;
};
