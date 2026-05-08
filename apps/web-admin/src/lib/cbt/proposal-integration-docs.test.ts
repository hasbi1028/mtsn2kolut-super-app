import { describe, expect, it } from 'vitest';

import antiCheatRoadmap from '../../../../../docs/flutter-anti-cheat-roadmap.md?raw';
import phase0Doc from '../../../../../docs/cbt-proposal-integration-phase-0.md?raw';

describe('CBT proposal integration documentation guard', () => {
	it('locks the monorepo runtime ownership for proposal integration', () => {
		for (const phrase of [
			'Web Admin',
			'`apps/web-admin`',
			'Bank Soal',
			'Asesmen',
			'Proctor',
			'Hasil',
			'`services/core-api`',
			'Go',
			'PostgreSQL',
			'owner state',
			'timer',
			'audit',
			'scoring',
			'`apps/mobile`',
			'Flutter',
			'portal ujian siswa utama',
			'lokasi anti-cheat utama'
		]) {
			expect(phase0Doc).toContain(phrase);
		}
	});

	it('keeps proposal-only stacks out of the runtime contract', () => {
		expect(phase0Doc).toContain('PocketBase, SQLite, and Alpine are not runtime architecture for this monorepo.');
		expect(phase0Doc).toContain('PocketBase tidak menjadi API server');
		expect(phase0Doc).toContain('SQLite tidak menjadi database runtime CBT');
		expect(phase0Doc).toContain('Alpine dari proposal tidak menjadi deployment runtime');
		expect(phase0Doc).toContain('Deployment server tetap native PM2');
	});

	it('documents the Phase 0-5 roadmap, mobile API contract, BYOD limits, and validation gates', () => {
		for (const phrase of [
			'Fase 0 - Alignment and Guard',
			'Fase 1 - Contract Stabilization',
			'Fase 2 - Web Admin Workflow Alignment',
			'Fase 3 - Backend Runtime Hardening',
			'Fase 4 - Flutter BYOD Anti-Cheat and Resilience',
			'Fase 5 - Rehearsal, Rollout, and Post-Exam Review',
			'POST /api/exam/login',
			'GET /api/exam/status',
			'POST /api/exam/heartbeat',
			'POST /api/exam/event',
			'POST /api/exam/answer',
			'POST /api/exam/submit',
			'BYOD vs Kiosk/Device-Owner',
			'Keamanan, RBAC, dan Audit',
			'git diff --check',
			'npm run test:unit -- src/lib/cbt/proposal-integration-docs.test.ts'
		]) {
			expect(phase0Doc).toContain(phrase);
		}
	});

	it('keeps the Flutter anti-cheat roadmap focused on BYOD telemetry rather than kiosk claims', () => {
		for (const phrase of [
			'Anti-cheat ownership lives in `apps/mobile`.',
			'Android BYOD',
			'device-owner bukan baseline',
			'FLAG_SECURE',
			'app_backgrounded',
			'app_resumed',
			'heartbeat_failed',
			'stale_connection_escalated',
			'manual_submit_blocked',
			'Core API',
			'Web Admin',
			'Docs dan UI tidak menyebut jaminan kiosk untuk BYOD'
		]) {
			expect(antiCheatRoadmap).toContain(phrase);
		}
	});
});
