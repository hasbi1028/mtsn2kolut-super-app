import { describe, expect, it } from 'vitest';

const CBT_SURFACES = import.meta.glob('../../routes/asesmen/{kegiatan,sesi}/**/+page.svelte', {
	eager: true,
	query: '?raw',
	import: 'default'
}) as Record<string, string>;

const INCLUDED_SURFACES = [
	'../../routes/asesmen/kegiatan/+page.svelte',
	'../../routes/asesmen/kegiatan/new/+page.svelte',
	'../../routes/asesmen/sesi/+page.svelte',
	'../../routes/asesmen/sesi/new/+page.svelte',
	'../../routes/asesmen/sesi/[id]/+page.svelte',
	'../../routes/asesmen/sesi/[id]/rooms/[rid]/proctoring/+page.svelte'
];

const LIGHT_ONLY_CLASS_PATTERNS = [
	/bg-white\b/,
	/border-white\//,
	/text-slate-/,
	/bg-slate-/,
	/border-slate-/,
	/text-gray-/,
	/bg-gray-/,
	/border-gray-/,
	/accent-emerald-/
];

describe('CBT dark-mode internal surfaces', () => {
	it('uses theme tokens instead of light-only Tailwind classes on CBT operator surfaces', () => {
		for (const surface of INCLUDED_SURFACES) {
			const source = CBT_SURFACES[surface];
			expect(source, `${surface} should be included in dark-mode guard`).toBeDefined();
			for (const pattern of LIGHT_ONLY_CLASS_PATTERNS) {
				expect(source, `${surface} should not contain ${pattern}`).not.toMatch(pattern);
			}
		}
	});
});
