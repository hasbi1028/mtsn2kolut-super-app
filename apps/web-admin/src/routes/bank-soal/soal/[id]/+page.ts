import type { PageLoad } from './$types';

export const load: PageLoad = ({ params }) => ({ questionId: params.id });
