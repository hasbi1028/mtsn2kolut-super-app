import { describe, expect, it, vi } from 'vitest';

import {
	fetchStudentPortalProfile,
	fetchStudentPortalResults,
	fetchStudentPortalSchedule
} from './student-portal';

describe('student portal client helpers', () => {
	it('fetches self-scoped profile, schedule, and results endpoints', async () => {
		const profile = { student: { id: 'student-1', nama: 'Siswa A', class_name: 'VII A' } };
		const schedule = { schedule: [{ id: 'slot-1', subject_name: 'IPA', day_of_week: 1 }] };
		const results = { results: [{ session_id: 'session-1', session_title: 'Ujian IPA', score: 90 }] };
		const fetcher = vi
			.fn()
			.mockResolvedValueOnce(new Response(JSON.stringify({ data: profile }), { status: 200 }))
			.mockResolvedValueOnce(new Response(JSON.stringify({ data: schedule }), { status: 200 }))
			.mockResolvedValueOnce(new Response(JSON.stringify({ data: results }), { status: 200 }));

		await expect(fetchStudentPortalProfile(fetcher)).resolves.toEqual(profile);
		await expect(fetchStudentPortalSchedule(fetcher)).resolves.toEqual(schedule);
		await expect(fetchStudentPortalResults(fetcher)).resolves.toEqual(results);
		expect(fetcher).toHaveBeenNthCalledWith(1, '/api/portal/siswa/profile');
		expect(fetcher).toHaveBeenNthCalledWith(2, '/api/portal/siswa/schedule');
		expect(fetcher).toHaveBeenNthCalledWith(3, '/api/portal/siswa/results');
	});

	it('rejects self API errors with useful messages', async () => {
		const fetcher = vi.fn().mockResolvedValue(new Response(JSON.stringify({ message: 'forbidden' }), { status: 403 }));

		await expect(fetchStudentPortalProfile(fetcher)).rejects.toThrow('forbidden');
	});

});
