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

export const load: PageServerLoad = async ({ fetch, locals }) => {
	if (locals.user) return {};

	const [posts, featuredPosts, announcements, profil, ppdbInfo] = await Promise.all([
		apiPublicGetWithFetch<WebsiteContent[]>(fetch, '/api/public/site/posts?limit=3'),
		apiPublicGetWithFetch<WebsiteContent[]>(fetch, '/api/public/site/posts/featured?limit=2').catch(() => []),
		apiPublicGetWithFetch<WebsiteContent[]>(fetch, '/api/public/site/announcements?limit=4'),
		apiPublicGetWithFetch<WebsiteContent>(fetch, '/api/public/site/pages/profil').catch(() => null),
		apiPublicGetWithFetch<WebsiteContent>(fetch, '/api/public/site/pages/ppdb-info').catch(() => null),
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
};
