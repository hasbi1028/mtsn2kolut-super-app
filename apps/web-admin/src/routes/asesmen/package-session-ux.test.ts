import { describe, expect, it } from 'vitest';
import { readFileSync } from 'node:fs';
import { resolve } from 'node:path';

const root = resolve(__dirname, '../../..');
const packagePage = readFileSync(resolve(root, 'src/routes/asesmen/paket/+page.svelte'), 'utf8');
const sessionPage = readFileSync(resolve(root, 'src/routes/asesmen/sesi/+page.svelte'), 'utf8');
const newSessionPage = readFileSync(resolve(root, 'src/routes/asesmen/sesi/new/+page.svelte'), 'utf8');

describe('assessment package and session operator UX', () => {
  it('uses Pratinjau/Buka labels and archive-aware package actions', () => {
    expect(packagePage).toContain('Pratinjau');
    expect(packagePage).toContain('Buka');
    expect(packagePage).toContain('Arsipkan');
    expect(packagePage).toContain('Paket tidak dapat dihapus');
    expect(packagePage).toContain('session_count');
    expect(packagePage).toContain('archivePackage');
    expect(packagePage).toContain('Buat Revisi');
    expect(packagePage).toContain('createRevisionPackage');
  });

  it('adds package-like structured filters to the session list while avoiding duplicate quick actions', () => {
    expect(sessionPage).toContain('Cari sesi atau paket');
    expect(sessionPage).toContain('Status sesi');
    expect(sessionPage).toContain('Cakupan');
    expect(sessionPage).toContain('Paket soal');
    expect(sessionPage).toContain('Reset Filter');
    expect(sessionPage).toContain('sessionSearch');
    expect(sessionPage).toContain('sessionStatusFilter');
    expect(sessionPage).toContain('sessionPackageFilter');
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
