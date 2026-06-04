import { describe, expect, it } from 'vitest';
import {
	makassarDateKey,
	normalizeSesiCbtRow,
	sessionIsActiveWindow,
	sessionMonitorHref,
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

	it('builds monitor href without exposing raw token text as the label contract', () => {
		const rowWithUrl = normalizeSesiCbtRow({
			id: 's1',
			monitor_token: 'secret-token',
			proctor_monitor_url: '/asesmen/pelaksanaan?session_id=s1'
		});
		const rowWithTokenOnly = normalizeSesiCbtRow({ id: 's2', monitor_token: 'token with spaces' });
		const rowWithUnsafeUrl = normalizeSesiCbtRow({ id: 's3', monitor_url: 'javascript:alert(1)' });

		expect(sessionMonitorHref(rowWithUrl)).toBe('/asesmen/pelaksanaan?session_id=s1');
		expect(sessionMonitorHref(rowWithTokenOnly)).toBe('/asesmen/pelaksanaan?monitor_token=token%20with%20spaces');
		expect(sessionMonitorHref(rowWithUnsafeUrl)).toBeUndefined();
	});

	it('detects active windows using absolute time while date keys stay WITA', () => {
		const row = normalizeSesiCbtRow({
			id: 's1',
			status: 'scheduled',
			scheduled_start: '2026-05-01T23:30:00.000Z',
			scheduled_end: '2026-05-02T01:00:00.000Z'
		});

		expect(makassarDateKey(row.scheduled_start)).toBe('2026-05-02');
		expect(sessionIsActiveWindow(row, new Date('2026-05-02T00:00:00.000Z'))).toBe(true);
		expect(sessionIsActiveWindow(row, new Date('2026-05-02T02:00:00.000Z'))).toBe(false);
	});

	it('normalizes missing optional fields without crashing', () => {
		const row = normalizeSesiCbtRow(null);

		expect(row.id).toBe('');
		expect(row.room_labels).toEqual([]);
		expect(row.participant_count).toBe(0);
	});
});
