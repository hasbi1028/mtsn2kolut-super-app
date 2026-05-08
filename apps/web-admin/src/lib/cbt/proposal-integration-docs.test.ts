import { describe, expect, it } from 'vitest';

import antiCheatRoadmap from '../../../../../docs/flutter-anti-cheat-roadmap.md?raw';
import examApiDoc from '../../../../../docs/exam-api.md?raw';
import phase0Doc from '../../../../../docs/cbt-proposal-integration-phase-0.md?raw';
import phase12Doc from '../../../../../docs/cbt-proposal-integration-phase-1-2.md?raw';
import phase34Doc from '../../../../../docs/cbt-proposal-integration-phase-3-4.md?raw';

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

	it('locks Phase 1 exam API contract stabilization for Flutter runtime', () => {
		for (const phrase of [
			'`services/core-api`',
			'Source of truth token ujian, timer, autosave jawaban, submit final, audit/event, scoring',
			'`apps/mobile`',
			'Portal ujian siswa utama dan lokasi anti-cheat BYOD utama',
			'`POST /api/exam/login`',
			'`GET /api/exam/status`',
			'`POST /api/exam/heartbeat`',
			'`POST /api/exam/event`',
			'`POST /api/exam/answer`',
			'`POST /api/exam/submit`',
			'`401`: request token-scoped tanpa participant context yang sah',
			'`403`: sesi belum aktif, sesi belum mulai, window sudah tertutup',
			'`409`: token sudah terikat perangkat lain atau ujian sudah submitted',
			'services/core-api/internal/handler/exam_test.go',
			'services/core-api/cmd/api/routes_contract_test.go'
		]) {
			expect(phase12Doc).toContain(phrase);
		}
	});

	it('documents the Flutter BYOD event taxonomy used by the exam API', () => {
		for (const phrase of [
			'Taxonomy Event BYOD Flutter',
			'`app_switch`',
			'`warning`',
			'`screenshot_attempt`',
			'`resume_exam`',
			'`repeat_resume_attempt`',
			'`answer_saved_local_only`',
			'`submit_blocked_pending_sync`',
			'`auto_submit_blocked_pending_sync`',
			'`submit_blocked_degraded_mode`',
			'`degraded_mode_entered`',
			'`stale_connection_attention`',
			'`stale_connection_escalated`',
			'`back_button_attempt`',
			'`manual_submit`',
			'token ujian mentah',
			'answer key'
		]) {
			expect(examApiDoc).toContain(phrase);
			expect(phase12Doc).toContain(phrase);
		}
	});

	it('locks Phase 2 Web Admin workflow and canonical route taxonomy', () => {
		for (const phrase of [
			'Workflow operator tetap memakai route canonical',
			'/bank-soal/tambah',
			'/bank-soal/impor',
			'/bank-soal/verifikasi',
			'/asesmen/persiapan',
			'/asesmen/paket',
			'/asesmen/kegiatan',
			'/asesmen/sesi',
			'/asesmen/pengawasan',
			'/asesmen/pelaksanaan',
			'/asesmen/hasil',
			'/asesmen/aplikasi-siswa',
			'/api/bank-soal/*',
			'/api/asesmen/*',
			'Tidak membuat public SvelteKit route tree `/api/cbt/**` baru',
			'`/api/cbt/*` hanya compatibility/deprecated route yang sudah ada',
			'PocketBase, SQLite, Alpine',
			'apps/web-admin/src/lib/server/cbt-backend-paths.test.ts',
			'apps/web-admin/src/lib/components/sidebar/sidebar-config.test.ts',
			'apps/web-admin/src/lib/server/route-access.test.ts'
		]) {
			expect(phase12Doc).toContain(phrase);
		}
	});

	it('locks Phase 3 backend runtime hardening and Phase 4 Flutter BYOD resilience', () => {
		for (const phrase of [
			'Phase 3 - Backend Runtime Hardening',
			'Body limit mobile-facing',
			'`POST /api/exam/login`: 4 KiB',
			'`POST /api/exam/event`: 16 KiB',
			'`POST /api/exam/answer`: 64 KiB serialized JSON',
			'`POST /api/exam/heartbeat` dan `POST /api/exam/submit`: empty body atau `{}` saja, maksimum 1 KiB',
			'`413 request body too large`',
			'`401 unauthorized` untuk token/fingerprint yang belum membentuk participant context sah',
			'`409 token already bound to another device` hanya untuk mismatch',
			'Phase 4 - Flutter BYOD Anti-Cheat and Resilience',
			'apps/mobile/lib/src/exam_events.dart',
			'Helper event menyaring field sensitif seperti token, password, dan answer key',
			'Resume gate tidak lagi dibuka bila status refresh gagal',
			'Restore dari snapshot masuk ke exam shell dengan `initialResumeCheckRequired: true`',
			'Pending-answer flush yang mendapat `409 exam already submitted` diperlakukan sebagai terminal server state',
			'`409 token already bound to another device` tetap menjaga jawaban pending lokal agar tidak hilang',
			'Batas jawaban uraian Flutter dikunci konservatif pada 15.000 karakter',
			'`ExamApiClient` menghitung exact serialized UTF-8 JSON body',
			'Tidak mengklaim BYOD sebagai kiosk penuh'
		]) {
			expect(phase34Doc).toContain(phrase);
		}
		expect(examApiDoc).toContain('Request JSON runtime harus berupa satu JSON object tanpa trailing payload');
		expect(examApiDoc).toContain('`POST /api/exam/answer` | 64 KiB serialized JSON');
		expect(examApiDoc).toContain('`413 Payload Too Large`');
		expect(examApiDoc).toContain('android:student-phone:install-...');
	});
});
