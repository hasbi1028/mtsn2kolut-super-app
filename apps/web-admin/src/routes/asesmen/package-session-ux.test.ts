import { describe, expect, it } from 'vitest';
import { readFileSync } from 'node:fs';
import { resolve } from 'node:path';

const root = resolve(__dirname, '../../..');
const packagePage = readFileSync(resolve(root, 'src/routes/asesmen/paket/+page.svelte'), 'utf8');
const sessionPage = readFileSync(resolve(root, 'src/routes/asesmen/sesi/+page.svelte'), 'utf8');
const newSessionPage = readFileSync(resolve(root, 'src/routes/asesmen/sesi/new/+page.svelte'), 'utf8');

describe('assessment package and session operator UX', () => {
  it('uses clearer package labels for navigation and lifecycle actions', () => {
    expect(packagePage).toContain('Pratinjau paket');
    expect(packagePage).toContain('Buka paket');
    expect(packagePage).toContain('Arsipkan paket');
    expect(packagePage).toContain('Hapus paket');
    expect(packagePage).toContain('Kunci paket');
    expect(packagePage).toContain('Buat revisi paket');
    expect(packagePage).toContain('Paket tidak dapat dihapus');
    expect(packagePage).toContain('session_count');
    expect(packagePage).toContain('archivePackage');
    expect(packagePage).toContain('createRevisionPackage');
  });

  it('adds package-like structured filters to the session list while avoiding duplicate quick actions and terse labels', () => {
    expect(sessionPage).toContain('Cari sesi atau paket');
    expect(sessionPage).toContain('Status sesi');
    expect(sessionPage).toContain('Cakupan');
    expect(sessionPage).toContain('Paket soal');
    expect(sessionPage).toContain('Reset Filter');
    expect(sessionPage).toContain('sessionSearch');
    expect(sessionPage).toContain('sessionStatusFilter');
    expect(sessionPage).toContain('sessionPackageFilter');
    expect(sessionPage).toContain('Daftarkan peserta');
    expect(sessionPage).toContain('Atur jadwal');
    expect(sessionPage).toContain('Buka detail');
    expect(sessionPage).toContain('Pantau sesi');
    expect(sessionPage).toContain('Lihat BA Sesi');
    expect(sessionPage).toContain('Lihat detail');
    expect(sessionPage).toContain('Batalkan sesi');
    expect(sessionPage).toContain('Selesaikan sesi');
    expect(sessionPage).toContain('Hapus sesi');
    expect(sessionPage).not.toContain('Aksi:');
    expect(sessionPage).not.toContain('nextSessionAction');
    expect(sessionPage).not.toContain('scheduleQuickAction');
  });

  it('blocks event-scoped packages from the global session form with operator guidance', () => {
    expect(newSessionPage).toContain('Paket ini terikat ke Kegiatan Ujian');
    expect(newSessionPage).toContain('paket kegiatan');
    expect(newSessionPage).toContain('event_id ikut terkirim');
  });
});
