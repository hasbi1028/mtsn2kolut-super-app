import { describe, expect, it, vi } from 'vitest';

import {
	fetchParentPortalChildProfile,
	fetchParentPortalChildResults,
	fetchParentPortalChildSchedule,
	fetchParentPortalChildren
} from './parent-portal';

describe('parent portal client helpers', () => {
	it('fetches parent-scoped children and child detail endpoints', async () => {
		const children = { children: [{ id: 'student-1', nama: 'Siswa A', class_name: 'VII A' }] };
		const profile = { student: { id: 'student-1', nama: 'Siswa A', class_name: 'VII A' } };
		const schedule = { schedule: [{ id: 'slot-1', subject_name: 'IPA', day_of_week: 1 }] };
		const results = { results: [{ session_id: 'session-1', session_title: 'Ujian IPA', score: 90 }] };
		const fetcher = vi
			.fn()
			.mockResolvedValueOnce(new Response(JSON.stringify({ data: children }), { status: 200 }))
			.mockResolvedValueOnce(new Response(JSON.stringify({ data: profile }), { status: 200 }))
			.mockResolvedValueOnce(new Response(JSON.stringify({ data: schedule }), { status: 200 }))
			.mockResolvedValueOnce(new Response(JSON.stringify({ data: results }), { status: 200 }));

		await expect(fetchParentPortalChildren(fetcher)).resolves.toEqual(children);
		await expect(fetchParentPortalChildProfile('student-1', fetcher)).resolves.toEqual(profile);
		await expect(fetchParentPortalChildSchedule('student-1', fetcher)).resolves.toEqual(schedule);
		await expect(fetchParentPortalChildResults('student-1', fetcher)).resolves.toEqual(results);
		expect(fetcher).toHaveBeenNthCalledWith(1, '/api/portal/orang-tua/children');
		expect(fetcher).toHaveBeenNthCalledWith(2, '/api/portal/orang-tua/children/student-1/profile');
		expect(fetcher).toHaveBeenNthCalledWith(3, '/api/portal/orang-tua/children/student-1/schedule');
		expect(fetcher).toHaveBeenNthCalledWith(4, '/api/portal/orang-tua/children/student-1/results');
	});

	it('encodes guessed child identifiers and rejects forbidden responses clearly', async () => {
		const fetcher = vi.fn().mockResolvedValue(new Response(JSON.stringify({ error: 'forbidden' }), { status: 403 }));

		await expect(fetchParentPortalChildProfile('student/guess', fetcher)).rejects.toThrow('forbidden');
		expect(fetcher).toHaveBeenCalledWith('/api/portal/orang-tua/children/student%2Fguess/profile');
	});
});
