import { describe, expect, it } from 'vitest';
import { readFileSync } from 'node:fs';

const pageSource = readFileSync('src/routes/asesmen/sesi/[id]/+page.svelte', 'utf8');

describe('Ujian Digital Command Center Hari-H UX', () => {
	it('exposes the day-of-exam command center shell with dynamic room wording', () => {
		expect(pageSource).toContain('Command Center Hari-H');
		expect(pageSource).toContain('Distribusi Ruang Sesi');
		expect(pageSource).toContain('ruang dinamis per sesi');
		expect(pageSource).toContain('Masalah Aktif');
		expect(pageSource).not.toContain('Cetak Paket');
	});

	it('does not render raw room token in the session detail screen', () => {
		expect(pageSource).not.toContain('Token rahasia {room.room_token');
		expect(pageSource).not.toContain('{room.room_token ||');
		expect(pageSource).toContain('maskToken');
	});
});
