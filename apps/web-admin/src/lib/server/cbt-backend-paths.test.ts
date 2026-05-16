import { describe, expect, it } from 'vitest';

import { cbtApiPath, cbtBackendPath, cbtBackendPathWithQuery } from './cbt-backend-paths';

describe('cbt backend path dispatch', () => {
	it('routes Bank Soal-owned paths to the native Bank Soal API namespace', () => {
		expect(cbtBackendPath('/questions')).toBe('/api/bank-soal/questions');
		expect(cbtBackendPath('/questions/123')).toBe('/api/bank-soal/questions/123');
		expect(cbtApiPath`/questions/${'abc 123'}`).toBe('/api/bank-soal/questions/abc%20123');
		expect(cbtBackendPath('/assets')).toBe('/api/bank-soal/assets');
		expect(cbtBackendPath('/assets/asset-1/file')).toBe('/api/bank-soal/assets/asset-1/file');
		expect(cbtBackendPath('/soal-support')).toBe('/api/bank-soal/soal-support');
	});

	it('routes Asesmen-owned paths to the native Asesmen API namespace', () => {
		expect(cbtBackendPath('/events')).toBe('/api/asesmen/events');
		expect(cbtBackendPath('/events/event-1/results')).toBe('/api/asesmen/events/event-1/results');
		expect(cbtBackendPath('/sessions')).toBe('/api/asesmen/sessions');
		expect(cbtBackendPath('/sessions/session-1/rooms/readiness')).toBe(
			'/api/asesmen/sessions/session-1/rooms/readiness'
		);
		expect(cbtBackendPath('/packages')).toBe('/api/asesmen/packages');
		expect(cbtBackendPath('/non-test-assessments')).toBe('/api/asesmen/non-test-assessments');
		expect(cbtBackendPath('/non-test-assessments/nta-1/sync-grade')).toBe(
			'/api/asesmen/non-test-assessments/nta-1/sync-grade'
		);
		expect(cbtBackendPath('/proctoring/my-rooms')).toBe('/api/asesmen/proctoring/my-rooms');
		expect(cbtBackendPath('/approvals')).toBe('/api/asesmen/approvals');
	});

	it('rejects unsafe or unsupported backend paths before dispatching', () => {
		for (const path of [
			'/users',
			'/questions/../users',
			'/questions/%2e%2e/users',
			'/questions/%2F/users',
			'/questions/%5C/users',
			'/questions//asset',
			'/questions?limit=1',
			'/questions#fragment',
			'\\questions'
		]) {
			expect(() => cbtBackendPath(path)).toThrow();
		}
	});

	it('preserves query strings after dispatching to the selected backend namespace', () => {
		expect(cbtBackendPathWithQuery('/questions', new URLSearchParams({ limit: '10' }))).toBe(
			'/api/bank-soal/questions?limit=10'
		);
		expect(cbtBackendPathWithQuery('/sessions', 'status=active')).toBe(
			'/api/asesmen/sessions?status=active'
		);
	});

	it('never dispatches active Bank Soal or Asesmen clients back to the legacy CBT API namespace', () => {
		const activeWorkflowPaths = [
			'/questions',
			'/questions/export',
			'/assets/asset-1/file',
			'/soal-support/subjects',
			'/events',
			'/events/event-1/members',
			'/packages',
			'/sessions',
			'/sessions/session-1/proctoring/events',
			'/sessions/session-1/results',
			'/sessions/session-1/rooms/readiness',
			'/proctoring/my-rooms',
			'/non-test-assessments',
			'/approvals'
		];

		for (const path of activeWorkflowPaths) {
			expect(cbtBackendPath(path), path).not.toMatch(/^\/api\/cbt(?:\/|$)/);
		}
	});
});
