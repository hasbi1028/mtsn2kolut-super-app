/**
 * SPA Migration — semua +page.server.ts dengan load only
 * di-convert ke +page.ts (client-side load).
 * 
 * Form actions tetap di +page.server.ts jika ada.
 */

import type { PageServerLoad } from './$types';

// Pages that only use load (no actions) — simple removal
// The load functions will run client-side via +page.ts equivalents

// Helper: check if a file has any actions export
export function hasActions(content: string): boolean {
	return content.includes('export const actions') || content.includes('export const action');
}
