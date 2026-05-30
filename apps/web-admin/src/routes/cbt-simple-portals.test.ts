import { readFileSync } from 'node:fs';
import { join } from 'node:path';
import { describe, expect, it } from 'vitest';

const layoutSource = readFileSync(join(process.cwd(), 'src/routes/+layout.svelte'), 'utf8');
const studentPortalSource = readFileSync(join(process.cwd(), 'src/routes/ujian/+page.svelte'), 'utf8');
const proctorPortalSource = readFileSync(join(process.cwd(), 'src/routes/pengawas-ujian/+page.svelte'), 'utf8');
const publicPolicySource = readFileSync(join(process.cwd(), 'src/lib/routes/public-policy.ts'), 'utf8');

describe('simple CBT student and proctor portals', () => {
	it('bypasses the admin/sidebar layout for standalone CBT portals', () => {
		expect(layoutSource).toContain("page.url.pathname === '/ujian'");
		expect(layoutSource).toContain("page.url.pathname === '/pengawas-ujian'");
		expect(publicPolicySource).toContain("'/ujian'");
		expect(publicPolicySource).toContain("'/pengawas-ujian'");
	});

	it('keeps /ujian as a simple mobile-web exam flow', () => {
		for (const copy of ['Portal Ujian Peserta', 'QR/PIN', 'Ruang Tunggu', 'satu soal per layar', 'Minta Konfirmasi Pengawas']) {
			expect(studentPortalSource).toContain(copy);
		}
		expect(studentPortalSource).toContain('color-scheme: light');
		expect(studentPortalSource).toContain('Sebelumnya');
		expect(studentPortalSource).toContain('Ragu-ragu');
		expect(studentPortalSource).toContain('Berikutnya');
	});

	it('keeps /pengawas-ujian focused on room, alerts, participants, and help queue', () => {
		for (const copy of ['Portal Pengawasan Ruang', 'Ruang', 'Peringatan', 'Peserta', 'Hubungi Admin', 'Antrian bantuan admin']) {
			expect(proctorPortalSource).toContain(copy);
		}
		expect(proctorPortalSource).toContain('color-scheme: light');
		expect(proctorPortalSource).toContain('Sudah Dicek');
		expect(proctorPortalSource).toContain('Beri Peringatan');
		expect(proctorPortalSource).not.toMatch(/unlock/i);
	});
});
