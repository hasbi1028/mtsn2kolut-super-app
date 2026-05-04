import { describe, expect, it } from 'vitest';
import { cbtRoomSetupErrorMessage, roomReadinessMessage, roomReadinessTone, type CbtRoomReadiness } from './cbt-room-readiness';

const ready: CbtRoomReadiness = {
	room_count: 2,
	total_capacity: 60,
	participant_count: 50,
	assigned_participant_count: 50,
	unassigned_participant_count: 0,
	missing_seat_count: 0,
	rooms_without_proctor: 0,
	proctor_assignment_count: 2,
};

describe('CBT room readiness helpers', () => {
	it('keeps readiness warning while seats are missing', () => {
		expect(roomReadinessTone({ ...ready, missing_seat_count: 3 })).toBe('warning');
		expect(roomReadinessMessage({ ...ready, missing_seat_count: 3 })).toContain('3 peserta belum punya nomor meja');
	});

	it('warns for missing proctors and insufficient capacity', () => {
		expect(roomReadinessTone({ ...ready, rooms_without_proctor: 1 })).toBe('warning');
		expect(roomReadinessTone({ ...ready, total_capacity: 40 })).toBe('warning');
	});

	it('returns success only when room, seat, capacity, and proctor counts are complete', () => {
		expect(roomReadinessTone(ready)).toBe('success');
	});

	it('normalizes room setup permission failures for operators', () => {
		expect(cbtRoomSetupErrorMessage(new Error('403 forbidden'), 'fallback')).toBe('Akses pengaturan sesi hanya untuk admin/operator CBT.');
		expect(cbtRoomSetupErrorMessage(new Error('validasi kapasitas gagal'), 'fallback')).toBe('validasi kapasitas gagal');
	});
});
