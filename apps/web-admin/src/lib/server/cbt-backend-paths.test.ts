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

	it('routes retained CBT runtime paths to the legacy CBT API namespace', () => {
		expect(cbtBackendPath('/events')).toBe('/api/cbt/events');
		expect(cbtBackendPath('/events/event-1/results')).toBe('/api/cbt/events/event-1/results');
		expect(cbtBackendPath('/sessions')).toBe('/api/cbt/sessions');
		expect(cbtBackendPath('/sessions/session-1/rooms/readiness')).toBe(
			'/api/cbt/sessions/session-1/rooms/readiness'
		);
		expect(cbtBackendPath('/packages')).toBe('/api/cbt/packages');
		expect(cbtBackendPath('/proctoring/my-rooms')).toBe('/api/cbt/proctoring/my-rooms');
		expect(cbtBackendPath('/approvals')).toBe('/api/cbt/approvals');
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
		expect(cbtBackendPathWithQuery('/sessions', 'status=active')).toBe('/api/cbt/sessions?status=active');
	});

	it('keeps Bank Soal clients on the native namespace and runtime clients on CBT', () => {
		const bankSoalPaths = [
			'/questions',
			'/questions/export',
			'/assets/asset-1/file',
			'/soal-support/subjects'
		];
		const runtimePaths = [
			'/events',
			'/events/event-1/members',
			'/packages',
			'/sessions',
			'/sessions/session-1/proctoring/events',
			'/sessions/session-1/results',
			'/sessions/session-1/rooms/readiness',
			'/proctoring/my-rooms',
			'/approvals'
		];

		for (const path of bankSoalPaths) {
			expect(cbtBackendPath(path), path).toMatch(/^\/api\/bank-soal(?:\/|$)/);
		}
		for (const path of runtimePaths) {
			expect(cbtBackendPath(path), path).toMatch(/^\/api\/cbt(?:\/|$)/);
		}
	});
});
