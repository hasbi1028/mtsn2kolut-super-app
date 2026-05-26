import type { PageServerLoad } from './$types';
import { listCbtMobileArtifacts } from '$lib/server/cbt-mobile-artifacts';

export const load: PageServerLoad = async (event) => {
	try {
		return {
			cbtArtifacts: await listCbtMobileArtifacts(event),
			cbtArtifactsError: ''
		};
	} catch (error) {
		return {
			cbtArtifacts: null,
			cbtArtifactsError: error instanceof Error ? error.message : 'Artifact APK CBT belum tersedia.'
		};
	}
};
