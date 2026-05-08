import { execFile } from 'node:child_process';
import { existsSync } from 'node:fs';
import { mkdir, mkdtemp, readFile, rm, writeFile } from 'node:fs/promises';
import path from 'node:path';
import { fileURLToPath } from 'node:url';
import { promisify } from 'node:util';

import { afterEach, describe, expect, it } from 'vitest';

import antiCheatRoadmap from '../../../../../docs/flutter-anti-cheat-roadmap.md?raw';
import examApiDoc from '../../../../../docs/exam-api.md?raw';
import phase0Doc from '../../../../../docs/cbt-proposal-integration-phase-0.md?raw';
import phase12Doc from '../../../../../docs/cbt-proposal-integration-phase-1-2.md?raw';
import phase34Doc from '../../../../../docs/cbt-proposal-integration-phase-3-4.md?raw';
import phase5Doc from '../../../../../docs/cbt-proposal-integration-phase-5.md?raw';
import phase6Doc from '../../../../../docs/cbt-proposal-integration-phase-6.md?raw';
import phase7Doc from '../../../../../docs/cbt-proposal-integration-phase-7.md?raw';
import phase8Doc from '../../../../../docs/cbt-proposal-integration-phase-8.md?raw';
import proposalGapAuditDoc from '../../../../../docs/cbt-proposal-gap-audit.md?raw';
import proposalTraceabilityDoc from '../../../../../docs/cbt-proposal-100-percent-traceability.md?raw';
import phase1922Doc from '../../../../../docs/cbt-proposal-integration-phase-19-22.md?raw';
import phase2326Doc from '../../../../../docs/cbt-proposal-integration-phase-23-26.md?raw';
import hotspotDecisionDoc from '../../../../../docs/cbt-hotspot-design-decision.md?raw';
import uploadAnswerPolicyDoc from '../../../../../docs/cbt-upload-answer-policy.md?raw';
import mediaPromptPolicyDoc from '../../../../../docs/cbt-media-prompt-response-policy.md?raw';
import finalReleaseEvidenceDoc from '../../../../../docs/cbt-release-final-evidence.md?raw';
import releaseEvidenceTemplateDoc from '../../../../../docs/cbt-release-evidence-template.md?raw';
import smokeChecklistDoc from '../../../../../docs/cbt-smoke-checklist.md?raw';
import finalReadinessScript from '../../../../../deploy/scripts/cbt-final-readiness.sh?raw';
import releasePreflightScript from '../../../../../deploy/scripts/cbt-release-preflight.sh?raw';
import releaseChecklistDoc from '../../../../../apps/mobile/RELEASE_CHECKLIST.md?raw';
import deviceTestMatrixDoc from '../../../../../apps/mobile/DEVICE_TEST_MATRIX.md?raw';

const execFileAsync = promisify(execFile);
const testFileDir = path.dirname(fileURLToPath(import.meta.url));
const repoRoot = path.resolve(testFileDir, '../../../../..');
const preflightScriptPath = path.join(repoRoot, 'deploy/scripts/cbt-release-preflight.sh');
const finalReadinessScriptPath = path.join(repoRoot, 'deploy/scripts/cbt-final-readiness.sh');
const phase9DocPath = path.join(repoRoot, 'docs/cbt-proposal-integration-phase-9.md');
const phase10DocPath = path.join(repoRoot, 'docs/cbt-proposal-integration-phase-10.md');
const phase11DocPath = path.join(repoRoot, 'docs/cbt-proposal-integration-phase-11.md');
const phase12DocPath = path.join(repoRoot, 'docs/cbt-proposal-integration-phase-12.md');
const phase13DocPath = path.join(repoRoot, 'docs/cbt-proposal-integration-phase-13.md');
const phase14DocPath = path.join(repoRoot, 'docs/cbt-proposal-integration-phase-14.md');
const phase15DocPath = path.join(repoRoot, 'docs/cbt-proposal-integration-phase-15.md');
const makefilePath = path.join(repoRoot, 'Makefile');
const tempOutputDirs: string[] = [];

interface CommandFailure extends Error {
	code?: number | string;
	stdout?: string;
	stderr?: string;
}

interface PreflightCheck {
	name: string;
	status: string;
	detail: string;
	log: string;
}

interface PreflightReport {
	output_dir: string;
	overall_status: string;
	evidence_template: string;
	checks: PreflightCheck[];
}

interface PreflightManifestArtifact {
	path: string;
	kind: string;
	sha256: string;
	bytes: number;
}

interface PreflightManifest {
	output_dir: string;
	algorithm: string;
	artifacts: PreflightManifestArtifact[];
	secret_scan: {
		status: string;
		scanned_files: string[];
	};
}

interface FinalGapAuditReport {
	proposal_source: string;
	status_vocabulary: string[];
	matrix: Array<{
		area: string;
		proposal_section: string;
		status: string;
		next_action: string;
	}>;
	boundary: {
		public_api_cbt_route_tree_required: boolean;
		student_runtime_api: string;
	};
	secret_scan: {
		status: string;
	};
}

interface FinalEvidenceReport {
	output_dir: string;
	commit: string;
	docs: Record<string, { path: string; exists: boolean }>;
	automated_evidence: Array<{
		name: string;
		status: string;
		detail: string;
		log?: string;
	}>;
	manual_evidence: {
		device_matrix: {
			status: string;
			requires_physical_android_devices: boolean;
			minimum_vendors: number;
		};
		operator_rehearsal: {
			status: string;
			requires_live_operator_rehearsal: boolean;
		};
		final_signoff: {
			status: string;
		};
	};
	health_commands: Array<{
		name: string;
		status: string;
		detail: string;
	}>;
	tests_manifest: Array<{
		name: string;
		command: string;
		scope: string;
	}>;
	secret_scan: {
		status: string;
	};
}

interface FinalSignoffReport {
	go_no_go: string;
	production_go: boolean;
	manual_inputs: {
		device_matrix_complete: boolean;
		operator_rehearsal_complete: boolean;
		final_signoff_complete: boolean;
	};
	blockers: string[];
	boundary: {
		public_api_cbt_route_tree_required: boolean;
		student_runtime_api: string;
	};
}

async function makeTempOutputDir() {
	const tempRoot = path.join(repoRoot, 'tmp');
	await mkdir(tempRoot, { recursive: true });
	const outputDir = await mkdtemp(path.join(tempRoot, 'cbt-release-preflight-vitest-'));
	tempOutputDirs.push(outputDir);
	return outputDir;
}

async function runPreflight(args: string[]) {
	return execFileAsync(preflightScriptPath, args, {
		cwd: repoRoot,
		timeout: 30_000,
		maxBuffer: 8 * 1024 * 1024
	});
}

async function runFinalReadiness(args: string[]) {
	return execFileAsync(finalReadinessScriptPath, args, {
		cwd: repoRoot,
		timeout: 30_000,
		maxBuffer: 8 * 1024 * 1024
	});
}

async function expectPreflightFailure(args: string[]) {
	try {
		await runPreflight(args);
	} catch (error) {
		const failure = error as CommandFailure;
		return {
			code: failure.code,
			stdout: failure.stdout ?? '',
			stderr: failure.stderr ?? ''
		};
	}

	throw new Error(`Expected preflight to fail for args: ${args.join(' ')}`);
}

function targetRecipe(makefile: string, target: string) {
	const match = new RegExp(`^${target}:\\n((?:\\t[^\\n]*\\n)+)`, 'm').exec(makefile);
	expect(match, `${target} target should exist with a tab-indented recipe`).not.toBeNull();
	return match?.[1] ?? '';
}

afterEach(async () => {
	await Promise.all(tempOutputDirs.splice(0).map((outputDir) => rm(outputDir, { recursive: true, force: true })));
});

describe('CBT proposal integration documentation guard', () => {

	it('locks Phase 16-18 traceability and question type contract boundaries', () => {
		for (const phrase of [
			'CBT Proposal 100 Percent Traceability Matrix',
			'Implemented runtime',
			'Adapted/out-of-scope',
			'PocketBase / Alpine.js / SQLite proposal stack',
			'Public `/api/cbt/**` route tree',
			'multiple_choice',
			'multiple_answer',
			'essay',
			'short_answer',
			'matching',
			'ordering',
			'true_false',
			'agree_disagree',
			'comma-separated labels in exact sequence',
			'`B,A,C`',
			'Flutter uses `/api/exam/*`'
		]) {
			expect(proposalTraceabilityDoc).toContain(phrase);
		}

		expect(proposalGapAuditDoc).toContain('Phase 16-18 follow-up status');
		expect(proposalGapAuditDoc).toContain('ordering now has an explicit Web Admin/Core API/Flutter contract');
	});

	it('locks Phase 19-22 fixed-pair and deferred high-risk question decisions', () => {
		for (const phrase of [
			'Phase 19-22',
			'true_false',
			'`A=Benar`',
			'`B=Salah`',
			'agree_disagree',
			'`A=Setuju`',
			'`B=Tidak Setuju`',
			'no fake hotspot',
			'hotspot_safe_status: adapted_deferred',
			'upload_answer_safe_status: policy_deferred',
			'recording_answer_safe_status: policy_deferred',
			'Flutter tetap berbicara langsung ke `services/core-api` melalui `/api/exam/*`',
			'Tidak membuat public SvelteKit route tree `/api/cbt/**` baru',
			'Tidak mengubah schema database atau migration'
		]) {
			expect(phase1922Doc).toContain(phrase);
		}

		for (const phrase of [
			'Hotspot Design Decision',
			'Decision: adapted/out-of-scope for BYOD CBT v1',
			'No fake hotspot runtime',
			'hotspot_safe_status: adapted_deferred',
			'coordinate system',
			'image asset requirement',
			'scoring tolerance',
			'Flutter unsupported guard',
			'no schema migration in Phase 20'
		]) {
			expect(hotspotDecisionDoc).toContain(phrase);
		}

		for (const phrase of [
			'CBT Upload/File Answer Policy',
			'upload_answer_safe_status: policy_deferred',
			'No upload-answer runtime is enabled by this policy alone',
			'max size',
			'allowed MIME',
			'storage location',
			'retention',
			'access control',
			'virus/abuse review',
			'manual review'
		]) {
			expect(uploadAnswerPolicyDoc).toContain(phrase);
		}

		for (const phrase of [
			'CBT Media Prompt and Response Policy',
			'Audio/video prompt',
			'stem_audio_url',
			'stimulus_audio_url',
			'one-time play indicator',
			'Video prompt is adapted/deferred',
			'recording_answer_safe_status: policy_deferred',
			'Do not add recording upload unless covered by the upload/file answer policy'
		]) {
			expect(mediaPromptPolicyDoc).toContain(phrase);
		}

		for (const doc of [phase1922Doc, hotspotDecisionDoc, uploadAnswerPolicyDoc, mediaPromptPolicyDoc]) {
			expect(doc).not.toContain('POST /api/cbt/login');
			expect(doc).not.toContain('GET /api/cbt/status');
		}

		expect(proposalTraceabilityDoc).toContain('Phase 19-22 decision status');
		expect(proposalGapAuditDoc).toContain('Phase 19-22 follow-up status');
	});

	it('locks Phase 23-26 proctor evidence, BYOD evidence, analytics, and report boundaries', () => {
		for (const phrase of [
			'Phase 23-26',
			'HEAD `488e105` Phase 19-22',
			'Proctor Dashboard Full Evidence Mode',
			'heartbeat',
			'app background/resume',
			'device mismatch',
			'submit guard',
			'stale connection',
			'warning',
			'force submit',
			'reset access',
			'export/print evidence',
			'no screen preview',
			'no remote desktop',
			'anti_cheat_byod_manual_status: pending_manual_evidence',
			'real_device_pass_claim: forbidden_until_operator_tested',
			'sha256sum build/app/outputs/flutter-apk/app-release.apk',
			'Cronbach alpha: documented_deferred_until_formula_and_dataset_are_tested',
			'No fake psychometrics',
			'Reports PDF/Excel Parity',
			'print HTML / browser PDF',
			'CSV Excel-compatible',
			'No new binary PDF/XLSX endpoint',
			'no broad token spreadsheet export',
			'Tidak membuat public SvelteKit route tree `/api/cbt/**` baru',
			'Flutter tetap berbicara langsung ke `services/core-api` melalui `/api/exam/*`'
		]) {
			expect(phase2326Doc).toContain(phrase);
		}

		for (const phrase of [
			'Phase 23-26 follow-up status',
			'Screen preview and remote desktop remain explicit non-goals',
			'Deterministic build/hash instructions are documented without claiming PASS',
			'Cronbach alpha remains deferred',
			'official template matrix maps PDF to print HTML/browser PDF and Excel to CSV-safe exports'
		]) {
			expect(proposalGapAuditDoc).toContain(phrase);
		}

		for (const phrase of [
			'Phase 23-26 evidence, analytics, and reports',
			'device mismatch',
			'Build/hash instructions do not imply a real-device PASS',
			'Cronbach alpha: documented_deferred_until_formula_and_dataset_are_tested',
			'PDF parity is print HTML/browser PDF',
			'sensitive token reports remain print-only or role-bound'
		]) {
			expect(proposalTraceabilityDoc).toContain(phrase);
		}

		for (const phrase of [
			'Phase 24 Anti-Cheat BYOD Evidence Completion',
			'pending_manual_evidence',
			'No fabricated real-device PASS',
			'sha256sum build/app/outputs/flutter-apk/app-release.apk',
			'minimum two Android vendors'
		]) {
			expect(deviceTestMatrixDoc).toContain(phrase);
			expect(releaseChecklistDoc).toContain(phrase);
		}

		for (const phrase of [
			'Phase 23-26 evidence status',
			'no screen preview and no remote desktop',
			'no fabricated real-device PASS',
			'Cronbach alpha is `documented_deferred_until_formula_and_dataset_are_tested`',
			'no new binary PDF/XLSX endpoint'
		]) {
			expect(finalReleaseEvidenceDoc).toContain(phrase);
		}

		for (const phrase of [
			'Phase 23-26 Evidence, Analytics, and Reports',
			'device mismatch `409` evidence captured without exposing full fingerprint',
			'real-device PASS claimed only after physical Android operator test',
			'Cronbach alpha remains documented/deferred unless real tested implementation exists',
			'no broad token spreadsheet export'
		]) {
			expect(releaseEvidenceTemplateDoc).toContain(phrase);
		}

		for (const doc of [phase2326Doc, finalReleaseEvidenceDoc, releaseEvidenceTemplateDoc]) {
			expect(doc).not.toContain('POST /api/cbt/login');
			expect(doc).not.toContain('GET /api/cbt/status');
		}
	});
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

	it('locks Phase 5 rehearsal, rollout, and post-exam review documentation', () => {
		for (const phrase of [
			'Phase 5 - Rehearsal, Rollout, and Post-Exam Review',
			'docs/cbt-smoke-checklist.md',
			'docs/exam-api.md',
			'apps/mobile/RELEASE_CHECKLIST.md',
			'Bank Soal -> Asesmen Persiapan -> Pelaksanaan/Pengawasan -> Flutter APK -> Hasil/Post-exam review',
			'health check',
			'audit log',
			'server log',
			'backup',
			'migration state',
			'Rollback plan',
			'hentikan sesi baru',
			'pertahankan data backend',
			'distribusikan APK sebelumnya',
			'event member role dan subject scope',
			'export/detail Bank Soal admin/guru/guru non-penulis',
			'berita acara/minutes token visibility',
			'duplicate `(room_id, seat_no)`',
			'endpoint/security boundary checks',
			'Flutter SDK bisa tidak tersedia di host'
		]) {
			expect(phase5Doc).toContain(phrase);
			expect(smokeChecklistDoc).toContain(phrase);
		}
	});

	it('keeps Phase 5 inside existing monorepo runtime and route boundaries', () => {
		for (const doc of [phase5Doc, smokeChecklistDoc, releaseChecklistDoc]) {
			for (const phrase of [
				'PocketBase',
				'SQLite',
				'Alpine',
				'Flutter berbicara langsung ke `services/core-api` melalui `/api/exam/*`',
				'Tidak membuat public SvelteKit route tree `/api/cbt/**` baru',
				'Tidak menjalankan `make db-migrate`',
				'Tidak deploy, tidak PM2 restart'
			]) {
				expect(doc).toContain(phrase);
			}
			expect(doc).not.toContain('POST /api/cbt/login');
			expect(doc).not.toContain('GET /api/cbt/status');
		}
	});

	it('aligns the Flutter release checklist with Phase 5 BYOD rehearsal limits', () => {
		for (const phrase of [
			'Phase 5 rehearsal',
			'perangkat nyata',
			'BYOD tidak setara kiosk penuh',
			'device-owner bukan baseline',
			'login token',
			'heartbeat',
			'answer save',
			'restore',
			'network disturbance',
			'warning',
			'submit',
			'Flutter SDK bisa tidak tersedia di host',
			'Catat sebagai blocker validasi mobile'
		]) {
			expect(releaseChecklistDoc).toContain(phrase);
		}
	});

	it('locks Phase 6 as operational release-evidence automation only', () => {
		for (const phrase of [
			'Phase 6 - Production Readiness Evidence Automation',
			'operational release-evidence automation/guard',
			'roadmap resmi tetap Phase 0 sampai Phase 5',
			'Release Evidence Bundle',
			'Preflight Commands',
			'No-Deploy / No-Migration Boundary',
			'Acceptance Criteria',
			'docs/cbt-release-evidence-template.md',
			'deploy/scripts/cbt-release-preflight.sh',
			'/home/servermtsn2kolut/development/flutter/bin',
			'Tidak deploy',
			'Tidak PM2 restart',
			'Tidak menjalankan `make db-migrate`',
			'Tidak menjalankan migrasi live',
			'Tidak menjalankan ad hoc SQL',
			'Tidak menjalankan `psql` untuk `ALTER`, `UPDATE`, `DELETE`, `INSERT`, atau DDL/DML lain',
			'Flutter berbicara langsung ke `services/core-api` melalui `/api/exam/*`',
			'Tidak membuat public SvelteKit route tree `/api/cbt/**` baru'
		]) {
			expect(phase6Doc).toContain(phrase);
		}
		expect(phase6Doc).not.toContain('POST /api/cbt/login');
		expect(phase6Doc).not.toContain('GET /api/cbt/status');
	});

	it('keeps the CBT release evidence template focused on non-destructive proof', () => {
		for (const phrase of [
			'CBT Release Evidence Template',
			'Commit dan Status Git',
			'Output Preflight',
			'No deploy, no PM2 restart, no `make db-migrate`',
			'No live migration dan no ad hoc SQL',
			'Flutter SDK path',
			'/home/servermtsn2kolut/development/flutter/bin',
			'`/api/exam/*`',
			'Tidak ada runtime siswa melalui `/api/cbt/login` atau `/api/cbt/status`',
			'deploy/scripts/cbt-release-preflight.sh',
			'Markdown evidence',
			'JSON evidence',
			'Keputusan rilis'
		]) {
			expect(releaseEvidenceTemplateDoc).toContain(phrase);
		}
	});

	it('guards the read-only CBT release preflight script', () => {
		for (const phrase of [
			'set -euo pipefail',
			'DEFAULT_FLUTTER_BIN_DIR="/home/servermtsn2kolut/development/flutter/bin"',
			'docs/cbt-release-evidence-template.md',
			'cbt-release-preflight.md',
			'cbt-release-preflight.json',
			'cbt-release-manifest.json',
			'--run-web-docs-guard',
			'--run-web-check',
			'--run-flutter-doctor',
			'--run-flutter-analyze',
			'--run-flutter-test',
			'--run-go-test',
			'--run-go-build',
			'npm run test:unit -- src/lib/cbt/proposal-integration-docs.test.ts',
			'go build -o /dev/null ./cmd/api',
			'redact_stream',
			'scan_generated_evidence_for_secrets',
			'write_manifest',
			'run_optional_command'
		]) {
			expect(releasePreflightScript).toContain(phrase);
		}

		for (const forbiddenPattern of [
			/\bpm2\s+(restart|reload|stop|delete|start)\b/i,
			/\bmake\s+db-migrate\b/i,
			/\bnpm\s+run\s+db:migrate\b/i,
			/\bpsql\b[^\n]*(ALTER|UPDATE|DELETE|INSERT|DROP|TRUNCATE|CREATE)\b/i,
			/\bdeploy\/scripts\/health-check\.sh\b/,
			/\bdeploy\/backup-postgresql\.sh\b/,
			/\bprintenv\b/,
			/\benv\s*\|/
		]) {
			expect(releasePreflightScript).not.toMatch(forbiddenPattern);
		}
	});

	it('locks Phase 7 as verifier and regression-test hardening only', () => {
		for (const phrase of [
			'Phase 7 - Release Preflight Verifier and Test Hardening',
			'verifier/test hardening',
			'bukan product roadmap baru',
			'bukan deploy',
			'deploy/scripts/cbt-release-preflight.sh',
			'apps/web-admin/src/lib/cbt/proposal-integration-docs.test.ts',
			'temporary ignored output directory under `tmp/`',
			'JSON evidence remains parseable',
			'optional checks stay skipped by default',
			'refuses non-empty output directories',
			'refuses unknown flags',
			'Tidak deploy',
			'Tidak PM2 restart',
			'Tidak menjalankan `make db-migrate`',
			'Tidak menjalankan migrasi live',
			'Tidak menjalankan ad hoc SQL',
			'Tidak membuat public SvelteKit route tree `/api/cbt/**` baru',
			'Flutter tetap berbicara langsung ke `services/core-api` melalui `/api/exam/*`'
		]) {
			expect(phase7Doc).toContain(phrase);
		}

		expect(phase7Doc).not.toContain('POST /api/cbt/login');
		expect(phase7Doc).not.toContain('GET /api/cbt/status');
	});

	it('locks Phase 8 as release manifest, checksum, and secret-scan hardening only', () => {
		for (const phrase of [
			'Phase 8 - Release Evidence Manifest and Secret-Scan Hardening',
			'manifest/checksum/secret-scan hardening only',
			'bukan product roadmap baru',
			'bukan deploy',
			'bukan runtime CBT',
			'deploy/scripts/cbt-release-preflight.sh',
			'cbt-release-manifest.json',
			'SHA-256',
			'secret scan',
			'markdown/json/log evidence',
			'Tidak deploy',
			'Tidak PM2 restart',
			'Tidak menjalankan `make db-migrate`',
			'Tidak menjalankan migrasi live',
			'Tidak menjalankan ad hoc SQL',
			'Tidak membuat public SvelteKit route tree `/api/cbt/**` baru',
			'Flutter tetap berbicara langsung ke `services/core-api` melalui `/api/exam/*`'
		]) {
			expect(phase8Doc).toContain(phrase);
		}

		expect(phase8Doc).not.toContain('POST /api/cbt/login');
		expect(phase8Doc).not.toContain('GET /api/cbt/status');
	});

	it('locks Phase 9 as live deploy smoke and Makefile health runbook hardening only', async () => {
		const phase9Doc = await readFile(phase9DocPath, 'utf8');

		for (const phrase of [
			'Phase 9 - Live Deploy Smoke and Runbook Hardening',
			'commit `231d908`',
			'live deploy',
			'`make ops-health` failed because `deploy/scripts/health-check.sh` was not executable',
			'`bash deploy/scripts/health-check.sh all` passed',
			'Makefile health fix',
			'ops-health targets invoke `deploy/scripts/health-check.sh` through `bash`',
			'does not require executable bit',
			'post-deploy smoke',
			'docs/tests/ops script/Makefile only',
			'deploy/scripts/cbt-release-preflight.sh',
			'boundary remains intact',
			'No secrets',
			'Tidak deploy',
			'Tidak PM2 restart',
			'Tidak menjalankan `make db-migrate`',
			'Tidak menjalankan migrasi live',
			'Tidak menjalankan ad hoc SQL',
			'Tidak membuat public SvelteKit route tree `/api/cbt/**` baru',
			'Flutter tetap berbicara langsung ke `services/core-api` melalui `/api/exam/*`'
		]) {
			expect(phase9Doc).toContain(phrase);
		}

		expect(phase9Doc).not.toContain('POST /api/cbt/login');
		expect(phase9Doc).not.toContain('GET /api/cbt/status');
	});

	it('keeps ops-health Makefile targets independent from health-check executable bit', async () => {
		const makefile = await readFile(makefilePath, 'utf8');
		const expectedHealthTargets = new Map([
			['ops-health', 'all'],
			['ops-health-backend', 'backend'],
			['ops-health-frontend', 'frontend'],
			['ops-health-worker', 'worker']
		]);

		for (const [target, service] of expectedHealthTargets) {
			const recipe = targetRecipe(makefile, target);
			expect(recipe).toContain(`bash deploy/scripts/health-check.sh ${service}`);
			expect(recipe).not.toContain(`./deploy/scripts/health-check.sh ${service}`);
		}
	});

	it('locks Phase 10 as post-deploy evidence handoff hardening only', async () => {
		const phase10Doc = await readFile(phase10DocPath, 'utf8');

		for (const phrase of [
			'Phase 10 - Post-Deploy Evidence Handoff Hardening',
			'commit `bf9df69`',
			'handoff package after successful deploy/smoke',
			'commit hash',
			'PM2 status summary placeholders',
			'health smoke outputs',
			'manifest/checksum evidence',
			'rollback owner',
			'follow-up owner',
			'secret hygiene',
			'docs/tests/ops evidence tooling only',
			'No product runtime',
			'No DB schema/migrations',
			'No deploy automation that restarts PM2',
			'No live DB writes',
			'No `/api/cbt` public routes',
			'Tidak deploy',
			'Tidak PM2 restart',
			'Tidak menjalankan `make db-migrate`',
			'Tidak menjalankan migrasi live',
			'Tidak menjalankan ad hoc SQL',
			'Tidak membuat public SvelteKit route tree `/api/cbt/**` baru',
			'Flutter tetap berbicara langsung ke `services/core-api` melalui `/api/exam/*`'
		]) {
			expect(phase10Doc).toContain(phrase);
		}

		for (const phrase of [
			'Phase 9/10 Post-Deploy Handoff',
			'Phase 9 commit `bf9df69`',
			'Phase 10 handoff package',
			'PM2 status summary placeholders',
			'Health smoke outputs',
			'Manifest/checksum evidence',
			'Rollback owner',
			'Follow-up owner',
			'Secret hygiene'
		]) {
			expect(releaseEvidenceTemplateDoc).toContain(phrase);
		}

		expect(phase10Doc).not.toContain('POST /api/cbt/login');
		expect(phase10Doc).not.toContain('GET /api/cbt/status');
	});

	it('locks Phase 11 as evidence archive and retention hardening only', async () => {
		const phase11Doc = await readFile(phase11DocPath, 'utf8');

		for (const phrase of [
			'Phase 11 - Evidence Archive and Retention Hardening',
			'commit `64da343`',
			'archive/retention rules after Phase 10 handoff',
			'ops-controlled storage',
			'not committed tmp',
			'archive path convention',
			'required artifacts',
			'SHA-256 manifest verification',
			'retention owner',
			'retention period placeholder',
			'redaction/no secrets',
			'access control',
			'restore/read-back check',
			'deletion/expiry log placeholder',
			'docs/tests/ops evidence tooling only',
			'No product runtime',
			'No DB schema/migrations',
			'No deploy automation that restarts PM2',
			'No live DB writes',
			'No `/api/cbt` public routes',
			'Tidak deploy',
			'Tidak PM2 restart',
			'Tidak menjalankan `make db-migrate`',
			'Tidak menjalankan migrasi live',
			'Tidak menjalankan ad hoc SQL',
			'Tidak membuat public SvelteKit route tree `/api/cbt/**` baru',
			'Flutter tetap berbicara langsung ke `services/core-api` melalui `/api/exam/*`'
		]) {
			expect(phase11Doc).toContain(phrase);
		}

		for (const phrase of [
			'Phase 11 Evidence Archive and Retention',
			'Phase 11 archive/retention rules after Phase 10 handoff',
			'Ops-controlled archive path',
			'Not committed tmp',
			'Required artifacts',
			'SHA-256 manifest verification',
			'Retention owner',
			'Retention period placeholder',
			'Redaction/no secrets',
			'Access control',
			'Restore/read-back check',
			'Deletion/expiry log placeholder'
		]) {
			expect(releaseEvidenceTemplateDoc).toContain(phrase);
		}

		expect(phase11Doc).not.toContain('POST /api/cbt/login');
		expect(phase11Doc).not.toContain('GET /api/cbt/status');
	});

	it('locks Phase 12 as redacted evidence index and retrieval hardening only', async () => {
		const phase12Doc = await readFile(phase12DocPath, 'utf8');

		for (const phrase of [
			'Phase 12 - Evidence Index and Retrieval Hardening',
			'commit `0d9391f`',
			'redacted evidence index/search and retrieval policy after Phase 11 archive/retention',
			'index fields',
			'release id',
			'commit',
			'date',
			'archive path',
			'manifest sha256',
			'owner',
			'retention status',
			'access classification',
			'no raw secrets',
			'no full logs',
			'no env',
			'stored outside public web roots',
			'not committed tmp data except template docs',
			'lookup/read-back workflow',
			'access request/audit placeholders',
			'stale/missing archive handling',
			'checksum verification before retrieval',
			'docs/tests/ops evidence policy only',
			'No product runtime',
			'No DB schema/migrations',
			'No deploy automation that restarts PM2',
			'No live DB writes',
			'No `/api/cbt` public routes',
			'Tidak deploy',
			'Tidak PM2 restart',
			'Tidak menjalankan `make db-migrate`',
			'Tidak menjalankan migrasi live',
			'Tidak menjalankan ad hoc SQL',
			'Tidak membuat public SvelteKit route tree `/api/cbt/**` baru',
			'Flutter tetap berbicara langsung ke `services/core-api` melalui `/api/exam/*`'
		]) {
			expect(phase12Doc).toContain(phrase);
		}

		for (const phrase of [
			'Phase 12 Evidence Index and Retrieval',
			'Phase 12 evidence index/retrieval policy after Phase 11 archive/retention',
			'Index storage location',
			'Index fields',
			'Release id',
			'Manifest SHA-256',
			'Access classification',
			'Lookup/read-back workflow',
			'Access request/audit placeholders',
			'Stale/missing archive handling',
			'Checksum verification before retrieval'
		]) {
			expect(releaseEvidenceTemplateDoc).toContain(phrase);
		}

		expect(phase12Doc).not.toContain('POST /api/cbt/login');
		expect(phase12Doc).not.toContain('GET /api/cbt/status');
	});

	it('locks Phase 13 as mobile release candidate and device matrix hardening only', async () => {
		const phase13Doc = await readFile(phase13DocPath, 'utf8');

		for (const phrase of [
			'Phase 13 - Mobile Release Candidate and Device Matrix',
			'commit `a3e99bd`',
			'Mobile Release Candidate and Device Matrix',
			'docs/tests/ops readiness only',
			'RC identifier',
			'APK SHA-256 hash',
			'signing mode',
			'`API_BASE_URL`',
			'minimum two Android vendors',
			'background/resume',
			'heartbeat',
			'pending answer',
			'submit guard',
			'device mismatch',
			'screenshot protection / `FLAG_SECURE`',
			'network disturbance',
			'/home/servermtsn2kolut/development/flutter/bin',
			'apps/mobile/DEVICE_TEST_MATRIX.md',
			'apps/mobile/RELEASE_CHECKLIST.md',
			'docs/cbt-release-evidence-template.md',
			'No product runtime',
			'No DB schema/migrations',
			'No deploy automation that restarts PM2',
			'No live DB writes',
			'No `/api/cbt` public routes',
			'Tidak deploy',
			'Tidak PM2 restart',
			'Tidak menjalankan `make db-migrate`',
			'Tidak menjalankan migrasi live',
			'Tidak menjalankan ad hoc SQL',
			'Tidak membuat public SvelteKit route tree `/api/cbt/**` baru',
			'Flutter tetap berbicara langsung ke `services/core-api` melalui `/api/exam/*`'
		]) {
			expect(phase13Doc).toContain(phrase);
		}

		for (const doc of [releaseChecklistDoc, releaseEvidenceTemplateDoc, deviceTestMatrixDoc]) {
			for (const phrase of [
				'Phase 13 Mobile Release Candidate and Device Matrix',
				'RC identifier',
				'APK SHA-256 hash',
				'signing mode',
				'`API_BASE_URL`',
				'minimum two Android vendors',
				'background/resume',
				'heartbeat',
				'pending answer',
				'submit guard',
				'device mismatch',
				'screenshot protection / `FLAG_SECURE`',
				'network disturbance',
				'/home/servermtsn2kolut/development/flutter/bin'
			]) {
				expect(doc).toContain(phrase);
			}
		}

		expect(phase13Doc).not.toContain('POST /api/cbt/login');
		expect(phase13Doc).not.toContain('GET /api/cbt/status');
	});

	it('locks Phase 14 as operator rehearsal and proctor evidence hardening only', async () => {
		const phase14Doc = await readFile(phase14DocPath, 'utf8');

		for (const phrase of [
			'Phase 14 - Operator Rehearsal and Proctor Evidence',
			'operator rehearsal and proctor evidence hardening only',
			'Bank Soal to Asesmen Persiapan to Pelaksanaan/Pengawasan to Flutter APK to Hasil/Post-exam review',
			'proctor evidence',
			'role/scope/token boundary',
			'event/audit evidence',
			'go/no-go rehearsal',
			'Bank Soal',
			'Asesmen Persiapan',
			'Pelaksanaan/Pengawasan',
			'Flutter APK',
			'Hasil/Post-exam review',
			'docs/cbt-smoke-checklist.md',
			'docs/cbt-release-evidence-template.md',
			'No product runtime',
			'No DB schema/migrations',
			'No deploy automation that restarts PM2',
			'No live DB writes',
			'No `/api/cbt` public routes',
			'Tidak deploy',
			'Tidak PM2 restart',
			'Tidak menjalankan `make db-migrate`',
			'Tidak menjalankan migrasi live',
			'Tidak menjalankan ad hoc SQL',
			'Tidak membuat public SvelteKit route tree `/api/cbt/**` baru',
			'Flutter tetap berbicara langsung ke `services/core-api` melalui `/api/exam/*`'
		]) {
			expect(phase14Doc).toContain(phrase);
		}

		for (const doc of [smokeChecklistDoc, releaseEvidenceTemplateDoc]) {
			for (const phrase of [
				'Phase 14 Operator Rehearsal and Proctor Evidence',
				'Bank Soal to Asesmen Persiapan to Pelaksanaan/Pengawasan to Flutter APK to Hasil/Post-exam review',
				'proctor evidence',
				'role/scope/token boundary',
				'event/audit evidence',
				'go/no-go rehearsal'
			]) {
				expect(doc).toContain(phrase);
			}
		}

		expect(phase14Doc).not.toContain('POST /api/cbt/login');
		expect(phase14Doc).not.toContain('GET /api/cbt/status');
	});

	it('locks Phase 15 as final CBT release readiness sign-off only', async () => {
		const phase15Doc = await readFile(phase15DocPath, 'utf8');

		for (const phrase of [
			'Phase 15 - Final CBT Release Readiness Sign-off',
			'Final CBT Release Readiness Sign-off',
			'Phase 0-15',
			'go/no-go',
			'rollback owner',
			'evidence bundle',
			'validation commands',
			'final baseline marker',
			'CBT Phase 15 final baseline',
			'docs/cbt-release-evidence-template.md',
			'git diff --check',
			'npm run test:unit -- src/lib/cbt/proposal-integration-docs.test.ts',
			'npm run check',
			'npm run test:unit',
			'make ops-health',
			'go test ./...',
			'go build -o /dev/null ./cmd/api',
			'/home/servermtsn2kolut/development/flutter/bin/flutter analyze',
			'/home/servermtsn2kolut/development/flutter/bin/flutter test',
			'No product runtime',
			'No DB schema/migrations',
			'No deploy automation that restarts PM2',
			'No live DB writes',
			'No `/api/cbt` public routes',
			'Tidak deploy',
			'Tidak PM2 restart',
			'Tidak menjalankan `make db-migrate`',
			'Tidak menjalankan migrasi live',
			'Tidak menjalankan ad hoc SQL',
			'Tidak membuat public SvelteKit route tree `/api/cbt/**` baru',
			'Flutter tetap berbicara langsung ke `services/core-api` melalui `/api/exam/*`'
		]) {
			expect(phase15Doc).toContain(phrase);
		}

		for (const phrase of [
			'Phase 15 Final CBT Release Readiness Sign-off',
			'Final sign-off summary Phase 0-15',
			'go/no-go',
			'rollback owner',
			'evidence bundle',
			'validation commands',
			'final baseline marker',
			'CBT Phase 15 final baseline'
		]) {
			expect(releaseEvidenceTemplateDoc).toContain(phrase);
		}

		expect(phase15Doc).not.toContain('POST /api/cbt/login');
		expect(phase15Doc).not.toContain('GET /api/cbt/status');
	});

	it('locks the final proposal gap audit and release evidence docs', () => {
		for (const phrase of [
			'CBT Proposal Gap Audit',
			'Proposal source',
			'doc_2954fa0c7a5c_Proposal_Sistem_CBT_MTsN2_Kolaka_Utara.docx',
			'feature-by-feature matrix',
			'Implemented',
			'Partial',
			'Planned manual evidence',
			'Out of scope adapted',
			'Multi-mode assessment',
			'Question types',
			'Anti-cheat layers',
			'Room management',
			'Proctor dashboard',
			'Audit trail',
			'Analytics',
			'Reports',
			'ISO controls',
			'Infrastructure',
			'Risks',
			'PocketBase / Alpine.js / SQLite proposal stack is adapted',
			'Tidak membuat public SvelteKit route tree `/api/cbt/**` baru',
			'Flutter tetap berbicara langsung ke `services/core-api` melalui `/api/exam/*`'
		]) {
			expect(proposalGapAuditDoc).toContain(phrase);
		}

		for (const phrase of [
			'CBT Final Release Evidence',
			'Current commit baseline',
			'488e105',
			'Automated evidence completed on this host',
			'Manual evidence requiring physical Android devices and operator rehearsal',
			'Device matrix',
			'Operator rehearsal',
			'Final go/no-go sign-off',
			'pending_manual_evidence',
			'pending_manual_signoff',
			'deploy/scripts/cbt-final-readiness.sh',
			'cbt-proposal-gap-audit.json',
			'cbt-final-evidence.json',
			'cbt-final-signoff.json',
			'cbt-final-readiness.md',
			'Tidak ada klaim production go tanpa evidence perangkat nyata dan rehearsal operator'
		]) {
			expect(finalReleaseEvidenceDoc).toContain(phrase);
		}

		for (const phrase of [
			'Final audit/readiness artifacts',
			'docs/cbt-proposal-gap-audit.md',
			'docs/cbt-release-final-evidence.md',
			'deploy/scripts/cbt-final-readiness.sh',
			'automated evidence completed',
			'manual evidence requiring physical Android devices and operator rehearsal'
		]) {
			expect(releaseEvidenceTemplateDoc).toContain(phrase);
		}

		for (const doc of [proposalGapAuditDoc, finalReleaseEvidenceDoc, releaseEvidenceTemplateDoc]) {
			expect(doc).not.toContain('POST /api/cbt/login');
			expect(doc).not.toContain('GET /api/cbt/status');
		}
	});

	it('keeps the device matrix explicit about two-vendor manual evidence status', () => {
		for (const phrase of [
			'Manual evidence status',
			'pending_manual_evidence',
			'Two-vendor manual placeholders',
			'Vendor A',
			'Vendor B',
			'Current status',
			'Requires physical Android device',
			'Operator/reviewer',
			'minimum two Android vendors'
		]) {
			expect(deviceTestMatrixDoc).toContain(phrase);
		}
	});

	it('guards the read-only CBT final readiness script', () => {
		for (const phrase of [
			'set -euo pipefail',
			'DEFAULT_OUTPUT_DIR="${REPO_ROOT}/tmp/cbt-final-readiness"',
			'cbt-proposal-gap-audit.json',
			'cbt-final-evidence.json',
			'cbt-final-signoff.json',
			'cbt-final-readiness.md',
			'--manual-device-matrix-complete',
			'--manual-operator-rehearsal-complete',
			'--manual-final-signoff-complete',
			'--run-ops-health',
			'ready_for_rehearsal',
			'pending_manual_signoff',
			'redact_stream',
			'scan_generated_evidence_for_secrets',
			'public_api_cbt_route_tree_required'
		]) {
			expect(finalReadinessScript).toContain(phrase);
		}

		for (const forbiddenPattern of [
			/\bpm2\s+(restart|reload|stop|delete|start)\b/i,
			/\bmake\s+db-migrate\b/i,
			/\bnpm\s+run\s+db:migrate\b/i,
			/\bpsql\b[^\n]*(ALTER|UPDATE|DELETE|INSERT|DROP|TRUNCATE|CREATE)\b/i,
			/\bdeploy\/backup-postgresql\.sh\b/,
			/\bprintenv\b/,
			/\benv\s*\|/
		]) {
			expect(finalReadinessScript).not.toMatch(forbiddenPattern);
		}
	});

	it(
		'executes the CBT final readiness script with pending manual evidence by default',
		async () => {
			const outputDir = await makeTempOutputDir();
			const { stdout, stderr } = await runFinalReadiness(['--output', outputDir]);

			expect(stderr).toBe('');
			expect(stdout).toContain('CBT final readiness evidence written');
			expect(stdout).toContain(path.join(outputDir, 'cbt-proposal-gap-audit.json'));
			expect(stdout).toContain(path.join(outputDir, 'cbt-final-evidence.json'));
			expect(stdout).toContain(path.join(outputDir, 'cbt-final-signoff.json'));
			expect(stdout).toContain(path.join(outputDir, 'cbt-final-readiness.md'));

			const markdown = await readFile(path.join(outputDir, 'cbt-final-readiness.md'), 'utf8');
			const gapAudit = JSON.parse(
				await readFile(path.join(outputDir, 'cbt-proposal-gap-audit.json'), 'utf8')
			) as FinalGapAuditReport;
			const evidence = JSON.parse(
				await readFile(path.join(outputDir, 'cbt-final-evidence.json'), 'utf8')
			) as FinalEvidenceReport;
			const signoff = JSON.parse(
				await readFile(path.join(outputDir, 'cbt-final-signoff.json'), 'utf8')
			) as FinalSignoffReport;

			expect(markdown).toContain('# CBT Final Readiness Evidence');
			expect(markdown).toContain('pending_manual_signoff');
			expect(markdown).toContain('physical Android devices');
			expect(markdown).toContain('operator rehearsal');
			expect(markdown).not.toContain('/api/cbt/login');
			expect(markdown).not.toContain('/api/cbt/status');

			expect(gapAudit.proposal_source).toContain(
				'doc_2954fa0c7a5c_Proposal_Sistem_CBT_MTsN2_Kolaka_Utara.docx'
			);
			expect(gapAudit.status_vocabulary).toEqual([
				'Implemented',
				'Partial',
				'Planned manual evidence',
				'Out of scope adapted'
			]);
			expect(gapAudit.matrix.map((row) => row.area)).toEqual([
				'Multi-mode assessment',
				'Question types',
				'Anti-cheat layers',
				'Room management',
				'Proctor dashboard',
				'Audit trail',
				'Analytics',
				'Reports',
				'ISO controls',
				'Infrastructure',
				'Risks'
			]);
			expect(gapAudit.matrix.every((row) => row.next_action.length > 0)).toBe(true);
			expect(gapAudit.boundary.public_api_cbt_route_tree_required).toBe(false);
			expect(gapAudit.boundary.student_runtime_api).toBe('/api/exam/*');
			expect(gapAudit.secret_scan.status).toBe('pass');

			expect(evidence.output_dir).toBe(outputDir);
			expect(evidence.docs['proposal_gap_audit']?.exists).toBe(true);
			expect(evidence.docs['final_release_evidence']?.exists).toBe(true);
			expect(evidence.docs['device_test_matrix']?.exists).toBe(true);
			expect(evidence.automated_evidence.find((item) => item.name === 'git-diff-check')?.status).toMatch(/^(pass|fail)$/);
			expect(evidence.automated_evidence.find((item) => item.name === 'docs-existence')?.status).toBe('pass');
			expect(evidence.health_commands.find((item) => item.name === 'make ops-health')?.status).toBe('skipped');
			expect(evidence.tests_manifest.map((item) => item.command)).toContain(
				'cd apps/web-admin && npm run test:unit -- src/lib/cbt/proposal-integration-docs.test.ts'
			);
			expect(evidence.manual_evidence.device_matrix.status).toBe('pending_manual_evidence');
			expect(evidence.manual_evidence.device_matrix.requires_physical_android_devices).toBe(true);
			expect(evidence.manual_evidence.device_matrix.minimum_vendors).toBe(2);
			expect(evidence.manual_evidence.operator_rehearsal.status).toBe('pending_manual_evidence');
			expect(evidence.manual_evidence.operator_rehearsal.requires_live_operator_rehearsal).toBe(true);
			expect(evidence.manual_evidence.final_signoff.status).toBe('pending_manual_signoff');
			expect(evidence.secret_scan.status).toBe('pass');

			expect(signoff.go_no_go).toBe('pending_manual_signoff');
			expect(signoff.production_go).toBe(false);
			expect(signoff.manual_inputs.device_matrix_complete).toBe(false);
			expect(signoff.manual_inputs.operator_rehearsal_complete).toBe(false);
			expect(signoff.manual_inputs.final_signoff_complete).toBe(false);
			expect(signoff.blockers).toContain('manual device matrix evidence pending');
			expect(signoff.blockers).toContain('operator rehearsal evidence pending');
			expect(signoff.boundary.public_api_cbt_route_tree_required).toBe(false);
			expect(signoff.boundary.student_runtime_api).toBe('/api/exam/*');
		},
		30_000
	);

	it(
		'allows manual CBT final readiness flags without claiming production go',
		async () => {
			const outputDir = await makeTempOutputDir();
			await runFinalReadiness([
				'--output',
				outputDir,
				'--manual-device-matrix-complete',
				'--manual-operator-rehearsal-complete',
				'--manual-final-signoff-complete'
			]);

			const evidence = JSON.parse(
				await readFile(path.join(outputDir, 'cbt-final-evidence.json'), 'utf8')
			) as FinalEvidenceReport;
			const signoff = JSON.parse(
				await readFile(path.join(outputDir, 'cbt-final-signoff.json'), 'utf8')
			) as FinalSignoffReport;

			expect(evidence.manual_evidence.device_matrix.status).toBe('complete_by_operator_flag');
			expect(evidence.manual_evidence.operator_rehearsal.status).toBe('complete_by_operator_flag');
			expect(evidence.manual_evidence.final_signoff.status).toBe('complete_by_operator_flag');
			expect(signoff.go_no_go).toBe('ready_for_rehearsal');
			expect(signoff.production_go).toBe(false);
			expect(signoff.blockers).toEqual([]);
		},
		30_000
	);

	it(
		'executes the CBT release preflight verifier in a temporary ignored output directory',
		async () => {
			const outputDir = await makeTempOutputDir();
			const { stdout, stderr } = await runPreflight(['--output', outputDir]);

			expect(stderr).toBe('');
			expect(stdout).toContain('CBT release preflight evidence written');
			expect(stdout).toContain(path.join(outputDir, 'cbt-release-preflight.md'));
			expect(stdout).toContain(path.join(outputDir, 'cbt-release-preflight.json'));
			expect(stdout).toContain(path.join(outputDir, 'cbt-release-manifest.json'));
			expect(stdout).toContain(path.join(outputDir, 'logs'));

			const markdown = await readFile(path.join(outputDir, 'cbt-release-preflight.md'), 'utf8');
			const jsonText = await readFile(path.join(outputDir, 'cbt-release-preflight.json'), 'utf8');
			const manifestText = await readFile(path.join(outputDir, 'cbt-release-manifest.json'), 'utf8');
			const report = JSON.parse(jsonText) as PreflightReport;
			const manifest = JSON.parse(manifestText) as PreflightManifest;
			const checksByName = new Map(report.checks.map((check) => [check.name, check]));
			const manifestArtifactsByPath = new Map(
				manifest.artifacts.map((artifact) => [artifact.path, artifact])
			);

			expect(markdown).toContain('# CBT Release Preflight Evidence');
			expect(markdown).toContain('Phase 8 manifest/checksum/secret-scan hardening');
			expect(markdown).toContain('It writes markdown/json/log evidence only under the output directory.');
			expect(markdown).toContain('It does not perform deployment, process manager changes, database migrations, SQL mutation, or runtime state changes.');
			expect(markdown).toContain('`services/core-api` `/api/exam/*`');
			expect(markdown).not.toContain('/api/cbt/login');
			expect(markdown).not.toContain('/api/cbt/status');

			expect(report.output_dir).toBe(outputDir);
			expect(report.overall_status).toBe('pass');
			expect(report.evidence_template).toBe('docs/cbt-release-evidence-template.md');
			expect(report.checks.length).toBeGreaterThan(8);
			expect(checksByName.get('secret-scan')?.status).toBe('pass');
			expect(checksByName.get('secret-scan')?.detail).toContain('generated evidence scanned');

			expect(manifest.output_dir).toBe(outputDir);
			expect(manifest.algorithm).toBe('sha256');
			expect(manifest.secret_scan.status).toBe('pass');
			expect(manifest.secret_scan.scanned_files).toContain('cbt-release-preflight.md');
			expect(manifest.secret_scan.scanned_files).toContain('cbt-release-preflight.json');
			expect(manifest.secret_scan.scanned_files).toContain('logs/git-head.log');

			for (const artifactPath of [
				'cbt-release-preflight.md',
				'cbt-release-preflight.json',
				'logs/git-head.log',
				'logs/git-status.log',
				'logs/git-diff-check.log'
			]) {
				const artifact = manifestArtifactsByPath.get(artifactPath);
				expect(artifact?.kind).toMatch(/^(markdown|json|log)$/);
				expect(artifact?.sha256).toMatch(/^[a-f0-9]{64}$/);
				expect(artifact?.bytes).toBeGreaterThanOrEqual(0);
				expect(existsSync(path.join(outputDir, artifactPath))).toBe(true);
			}

			for (const checkName of ['git-head', 'git-status', 'git-diff-check']) {
				const check = checksByName.get(checkName);
				expect(check?.status).toBe('pass');
				expect(check?.log).toBe(`logs/${checkName}.log`);
				expect(existsSync(path.join(outputDir, `logs/${checkName}.log`))).toBe(true);
			}

			for (const checkName of [
				'web-docs-guard',
				'web-check',
				'go-test',
				'go-build',
				'flutter-doctor',
				'flutter-analyze',
				'flutter-test'
			]) {
				const check = checksByName.get(checkName);
				expect(check?.status).toBe('skipped');
				expect(check?.detail).toContain('not requested');
				expect(check?.log).toBe('');
			}
		},
		30_000
	);

	it(
		'refuses unsafe CBT release preflight invocation shapes',
		async () => {
			const nonEmptyOutputDir = await makeTempOutputDir();
			await writeFile(path.join(nonEmptyOutputDir, 'existing.txt'), 'keep');

			const nonEmptyFailure = await expectPreflightFailure(['--output', nonEmptyOutputDir]);
			expect(nonEmptyFailure.code).toBe(2);
			expect(nonEmptyFailure.stderr).toContain('output directory must be empty');
			expect(existsSync(path.join(nonEmptyOutputDir, 'cbt-release-preflight.md'))).toBe(false);
			expect(existsSync(path.join(nonEmptyOutputDir, 'cbt-release-preflight.json'))).toBe(false);
			expect(existsSync(path.join(nonEmptyOutputDir, 'cbt-release-manifest.json'))).toBe(false);

			const unknownFlagFailure = await expectPreflightFailure(['--unknown-phase-7-flag']);
			expect(unknownFlagFailure.code).toBe(2);
			expect(unknownFlagFailure.stderr).toContain('unknown option: --unknown-phase-7-flag');
			expect(unknownFlagFailure.stderr).toContain('Usage: deploy/scripts/cbt-release-preflight.sh');
		},
		30_000
	);

	it(
		'fails the CBT release preflight when generated evidence contains synthetic secret-like text',
		async () => {
			const outputDir = await makeTempOutputDir();
			const failure = await expectPreflightFailure(['--output', outputDir, '--test-only-write-secret-leak']);

			expect(failure.code).toBe(1);
			expect(failure.stderr).toContain('secret scan failed');
			expect(failure.stderr).toContain('logs/test-only-secret-leak.log');
			expect(existsSync(path.join(outputDir, 'logs/test-only-secret-leak.log'))).toBe(true);
			expect(existsSync(path.join(outputDir, 'cbt-release-manifest.json'))).toBe(true);

			const report = JSON.parse(
				await readFile(path.join(outputDir, 'cbt-release-preflight.json'), 'utf8')
			) as PreflightReport;
			const secretScan = report.checks.find((check) => check.name === 'secret-scan');
			expect(report.overall_status).toBe('failed');
			expect(secretScan?.status).toBe('fail');
		},
		30_000
	);
});
