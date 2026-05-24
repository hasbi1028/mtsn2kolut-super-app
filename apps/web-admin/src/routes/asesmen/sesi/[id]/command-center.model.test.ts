import { describe, expect, it } from 'vitest';
import { deriveCommandRoomStatus, prioritizeCommandCenterIssues, maskToken } from './command-center.model';

describe('command-center model', () => {
	it('masks sensitive tokens by default', () => {
		expect(maskToken('abcdef1234567890')).toBe('abcd••••7890');
		expect(maskToken('short')).toBe('••••');
		expect(maskToken('')).toBe('••••');
	});

	it('derives room status from operational risk', () => {
		expect(deriveCommandRoomStatus({ participant_count: 24, submitted_count: 24, handover_locked: true })).toBe('final');
		expect(deriveCommandRoomStatus({ participant_count: 24, submitted_count: 12, suspicious_count: 2 })).toBe('attention');
		expect(deriveCommandRoomStatus({ participant_count: 24, submitted_count: 12, missing_seat_count: 1 })).toBe('attention');
		expect(deriveCommandRoomStatus({ participant_count: 24, submitted_count: 12, offline_count: 2, minutes_remaining: 8 })).toBe('critical');
		expect(deriveCommandRoomStatus({ participant_count: 24, submitted_count: 12, minutes_remaining: 8 })).toBe('finishing');
	});

	it('prioritizes critical operational issues before warnings and info', () => {
		const issues = prioritizeCommandCenterIssues([
			{ id: 'handover', severity: 'info', title: 'Handover belum final' },
			{ id: 'offline', severity: 'critical', title: 'Peserta offline belum submit' },
			{ id: 'flagged', severity: 'warning', title: 'Perlu atensi pengawas' }
		]);
		expect(issues.map((issue) => issue.id)).toEqual(['offline', 'flagged', 'handover']);
	});
});
