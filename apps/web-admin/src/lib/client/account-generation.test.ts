import { describe, expect, it, vi } from 'vitest';

import {
	generateParentAccounts,
	generateStudentAccounts,
	previewParentAccounts,
	previewStudentAccounts
} from './account-generation';

const studentResult = {
	role: 'siswa',
	total: 1,
	ready: 1,
	created: 0,
	skipped: 0,
	failed: 0,
	candidates: [
		{
			student_id: 'student-1',
			nis: '123',
			nisn: '00998877',
			nama: 'Siswa A',
			generated_username: '00998877',
			role: 'siswa',
			status: 'ready'
		}
	]
};

const parentResult = {
	role: 'ortu',
	total: 1,
	ready: 0,
	created: 1,
	skipped: 0,
	failed: 0,
	candidates: [
		{
			parent_id: 'parent-1',
			nama: 'Wali A',
			phone: '0812',
			child_count: 1,
			basis_student_id: 'student-1',
			basis_student_nisn: '00998877',
			generated_username: 'ortu00998877',
			temporary_password: 'TempPass456!',
			role: 'ortu',
			status: 'created'
		}
	]
};

describe('student and parent account-generation client helpers', () => {
	it('parses student account preview and generate result envelopes', async () => {
		const fetcher = vi
			.fn()
			.mockResolvedValueOnce(new Response(JSON.stringify({ data: studentResult }), { status: 200 }))
			.mockResolvedValueOnce(new Response(JSON.stringify({ data: { ...studentResult, ready: 0, created: 1 } }), { status: 200 }));

		await expect(previewStudentAccounts(fetcher)).resolves.toEqual(studentResult);
		await expect(generateStudentAccounts(fetcher)).resolves.toMatchObject({ role: 'siswa', created: 1, ready: 0 });
		expect(fetcher).toHaveBeenNthCalledWith(1, '/api/users/student-accounts/preview');
		expect(fetcher).toHaveBeenNthCalledWith(2, '/api/users/student-accounts/generate', { method: 'POST' });
	});

	it('parses parent account preview and generate result envelopes', async () => {
		const fetcher = vi
			.fn()
			.mockResolvedValueOnce(new Response(JSON.stringify({ data: parentResult }), { status: 200 }))
			.mockResolvedValueOnce(new Response(JSON.stringify({ data: parentResult }), { status: 200 }));

		await expect(previewParentAccounts(fetcher)).resolves.toEqual(parentResult);
		await expect(generateParentAccounts(fetcher)).resolves.toEqual(parentResult);
		expect(fetcher).toHaveBeenNthCalledWith(1, '/api/users/parent-accounts/preview');
		expect(fetcher).toHaveBeenNthCalledWith(2, '/api/users/parent-accounts/generate', { method: 'POST' });
	});

	it('rejects non-OK responses with backend messages', async () => {
		const fetcher = vi.fn().mockResolvedValue(new Response(JSON.stringify({ error: 'forbidden' }), { status: 403 }));

		await expect(previewStudentAccounts(fetcher)).rejects.toThrow('forbidden');
	});
});
