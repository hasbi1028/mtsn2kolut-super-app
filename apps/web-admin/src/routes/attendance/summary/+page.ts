import { redirect } from '@sveltejs/kit';
import type { PageLoad } from './$types';

export const load: PageServerLoad = async () => {
	throw redirect(307, '/pusaka/summary');
};
