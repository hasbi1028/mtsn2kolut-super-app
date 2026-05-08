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
import releaseEvidenceTemplateDoc from '../../../../../docs/cbt-release-evidence-template.md?raw';
import smokeChecklistDoc from '../../../../../docs/cbt-smoke-checklist.md?raw';
import releasePreflightScript from '../../../../../deploy/scripts/cbt-release-preflight.sh?raw';
import releaseChecklistDoc from '../../../../../apps/mobile/RELEASE_CHECKLIST.md?raw';

const execFileAsync = promisify(execFile);
const testFileDir = path.dirname(fileURLToPath(import.meta.url));
const repoRoot = path.resolve(testFileDir, '../../../../..');
const preflightScriptPath = path.join(repoRoot, 'deploy/scripts/cbt-release-preflight.sh');
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

afterEach(async () => {
	await Promise.all(tempOutputDirs.splice(0).map((outputDir) => rm(outputDir, { recursive: true, force: true })));
});

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
