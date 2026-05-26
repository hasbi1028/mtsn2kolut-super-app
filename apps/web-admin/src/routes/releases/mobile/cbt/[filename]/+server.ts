import type { RequestHandler } from './$types';
import { serveCbtMobileArtifact } from '$lib/server/cbt-mobile-artifacts';

export const GET: RequestHandler = ({ params }) => {
	return serveCbtMobileArtifact(params.filename);
};
