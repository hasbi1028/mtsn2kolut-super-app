import { describe, expect, it } from 'vitest';
import {
	makassarDateKey,
	normalizeSesiCbtRow,
	sessionStatusLabel,
	summarizeSessionRows
} from './sesi-cbt';

describe('sesi cbt ui model', () => {
	it('adds blockers for incomplete draft sessions', () => {
		const row = normalizeSesiCbtRow({ id: 's1', event_id: 'e1', status: 'draft' }, { title: 'PAT Genap' });

		expect(row.exam_title).toBe('PAT Genap');
		expect(row.ready_to_activate).toBe(false);
		expect(row.blockers).toContain('Paket soal belum dipilih');
		expect(row.blockers).toContain('Peserta belum masuk');
	});

	it('labels active sessions for display', () => {
		expect(sessionStatusLabel('active')).toBe('Aktif');
		expect(sessionStatusLabel('running')).toBe('Aktif');
		expect(sessionStatusLabel('archived')).toBe('Arsip/Batal');
		expect(sessionStatusLabel('cancelled')).toBe('Arsip/Batal');
	});

	it('summarizes active, finished, cancelled, and incident counts', () => {
		const rows = [
			normalizeSesiCbtRow({ id: 's1', status: 'active', incident_count: 2 }),
			normalizeSesiCbtRow({ id: 's2', status: 'scheduled' }),
			normalizeSesiCbtRow({ id: 's3', status: 'finished' }),
			normalizeSesiCbtRow({ id: 's4', status: 'cancelled', insiden: 1 })
		];

		expect(summarizeSessionRows(rows)).toMatchObject({
			total: 4,
			active: 2,
			running: 1,
			finished: 1,
			cancelled: 1,
			incident: 3
		});
	});

	it('maps legacy archived status to backend-safe cancelled', () => {
		expect(normalizeSesiCbtRow({ id: 's1', status: 'archived' }).status).toBe('cancelled');
	});

	it('uses Asia/Makassar date keys instead of UTC slices', () => {
		expect('2026-05-01T17:05:00.000Z'.slice(0, 10)).toBe('2026-05-01');
		expect(makassarDateKey('2026-05-01T17:05:00.000Z')).toBe('2026-05-02');
	});

	it('normalizes missing optional fields without crashing', () => {
		const row = normalizeSesiCbtRow(null);

		expect(row.id).toBe('');
		expect(row.room_labels).toEqual([]);
		expect(row.participant_count).toBe(0);
	});
});
