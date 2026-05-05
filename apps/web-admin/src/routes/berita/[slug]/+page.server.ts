import { apiPath, apiPublicGetWithFetch } from '$lib/server/api';
import type { PageServerLoad } from './$types';

type WebsiteContent = {
	id: string;
	title: string;
	slug: string;
	excerpt: string;
	content_html: string;
	cover_image_url: string;
	meta_title: string;
	meta_description: string;
	published_at: string | null;
};

export const load: PageServerLoad = async ({ fetch, params }) => {
	const post = await apiPublicGetWithFetch<WebsiteContent>(fetch, apiPath`/api/public/site/posts/${params.slug}`);
	return { post };
};
