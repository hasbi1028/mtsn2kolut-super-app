import { apiPublicGetWithFetch } from '$lib/server/api';
import type { PageServerLoad } from './$types';

type WebsiteContent = {
	id: string;
	title: string;
	slug: string;
	excerpt: string;
	content_html: string;
	cover_image_url: string;
	published_at: string | null;
};

export const load: PageServerLoad = async ({ fetch }) => {
	const items = await apiPublicGetWithFetch<WebsiteContent[]>(fetch, '/api/public/site/posts?limit=24');
	return { items };
};
