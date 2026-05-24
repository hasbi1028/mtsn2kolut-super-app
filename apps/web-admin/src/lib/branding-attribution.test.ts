import { readFileSync } from 'node:fs';
import path from 'node:path';
import { describe, expect, it } from 'vitest';
import { appAttribution } from './branding';

const srcRoot = path.resolve(process.cwd(), 'src');

function readSource(relativePath: string) {
	return readFileSync(path.join(srcRoot, relativePath), 'utf8');
}

describe('HasbiGML app attribution', () => {
	it('keeps attribution centralized and intentionally understated', () => {
		expect(appAttribution).toEqual({
			productName: 'MTsN 2 Kolut Super App',
			shortLabel: 'by HasbiGML',
			loginLabel: 'Super App MTsN 2 Kolaka Utara · Powered by HasbiGML',
			developerName: 'HasbiGML',
			formalDeveloperLabel: 'Dikembangkan oleh HasbiGML',
			description: 'Platform digital terpadu untuk mendukung layanan MTsN 2 Kolaka Utara.'
		});
	});

	it('places attribution only in non-document UI surfaces', () => {
		const sidebar = readSource('lib/components/Sidebar.svelte');
		const login = readSource('routes/login/+page.svelte');
		const brandingSettings = readSource('routes/settings/branding/+page.svelte');

		expect(sidebar).toContain('appAttribution.productName');
		expect(sidebar).toContain('appAttribution.shortLabel');
		expect(sidebar).toContain('{#if desktopExpanded}');
		expect(login).toContain('appAttribution.loginLabel');
		expect(brandingSettings).toContain('Tentang Aplikasi');
		expect(brandingSettings).toContain('appAttribution.developerName');
	});

	it('does not add personal attribution to official print/document pages', () => {
		const officialDocumentSurfaces = [
			'routes/tu/surat-keterangan/[id]/print/+page.svelte',
			'routes/asesmen/sesi/[id]/rooms/[rid]/print-pack/+page.svelte',
			'routes/asesmen/kegiatan/[id]/exam-cards/+page.svelte',
			'routes/asesmen/sesi/[id]/minutes/+page.svelte'
		];

		for (const relativePath of officialDocumentSurfaces) {
			expect(readSource(relativePath), relativePath).not.toContain('HasbiGML');
		}
	});
});
