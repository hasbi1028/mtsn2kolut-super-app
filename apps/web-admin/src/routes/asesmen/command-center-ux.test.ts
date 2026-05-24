import { describe, expect, it } from 'vitest';
import { readFileSync } from 'node:fs';

const pageSource = readFileSync('src/routes/asesmen/sesi/[id]/+page.svelte', 'utf8');

describe('CBT Command Center Hari-H UX', () => {
	it('exposes the day-of-exam command center shell for eight rooms', () => {
		expect(pageSource).toContain('Command Center Hari-H');
		expect(pageSource).toContain('Grid 8 Ruang');
		expect(pageSource).toContain('Masalah Aktif');
		expect(pageSource).toContain('Cetak Semua Paket Ruang');
	});

	it('does not render raw room token in the session detail screen', () => {
		expect(pageSource).not.toContain('Token rahasia {room.room_token');
		expect(pageSource).not.toContain('{room.room_token ||');
		expect(pageSource).toContain('maskToken');
	});
});
