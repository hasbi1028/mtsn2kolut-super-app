import { apiGet } from '$lib/server/api';
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

export const load: PageServerLoad = async ({ params }) => {
	const announcement = await apiGet<WebsiteContent>(`/api/public/site/announcements/${params.slug}`);
	return { announcement };
};
