import { describe, expect, it } from 'vitest';

import {
	PROCTOR_EVIDENCE_CATEGORIES,
	buildProctorEvidenceCsvRows,
	classifyProctorEvent,
	proctorEventLabel,
	proctorRiskGroup,
	proctorRiskGroupLabel,
	summarizeProctorEvidence
} from './proctor-evidence';

describe('CBT proctor evidence helpers', () => {
	it('declares the full Phase 23 evidence coverage without screen preview claims', () => {
		expect(PROCTOR_EVIDENCE_CATEGORIES).toEqual([
			'heartbeat',
			'app_background_resume',
			'device_mismatch',
			'submit_guard',
			'stale_connection',
			'warning',
			'anti_cheat',
			'force_submit',
			'reset_access',
			'export_print'
		]);
	});

	it('classifies known runtime and BYOD warning events', () => {
		expect(classifyProctorEvent({ event_type: 'heartbeat', event_data: null })).toBe('heartbeat');
		expect(classifyProctorEvent({ event_type: 'app_switch', event_data: { state: 'paused' } })).toBe(
			'app_background_resume'
		);
		expect(classifyProctorEvent({ event_type: 'warning', event_data: { reason: 'resume_exam' } })).toBe(
			'app_background_resume'
		);
		expect(
			classifyProctorEvent({
				event_type: 'warning',
				event_data: { reason: 'submit_blocked_pending_sync' }
			})
		).toBe('submit_guard');
		expect(
			classifyProctorEvent({
				event_type: 'warning',
				event_data: { reason: 'stale_connection_escalated' }
			})
		).toBe('stale_connection');
		expect(classifyProctorEvent({ event_type: 'device_mismatch', event_data: { status: 409 } })).toBe(
			'device_mismatch'
		);
		expect(classifyProctorEvent({ event_type: 'proctor_force_submit', event_data: {} })).toBe('force_submit');
		expect(classifyProctorEvent({ event_type: 'proctor_reset_access', event_data: {} })).toBe('reset_access');
	});



	it('classifies backend canonical proctor event names for reports', () => {
		expect(classifyProctorEvent({ event_type: 'focus_lost_short', event_data: {} })).toBe('app_background_resume');
		expect(classifyProctorEvent({ event_type: 'app_switch_repeated', event_data: {} })).toBe('app_background_resume');
		expect(classifyProctorEvent({ event_type: 'background_over_threshold', event_data: {} })).toBe('app_background_resume');
		expect(classifyProctorEvent({ event_type: 'screenshot_attempt_ambiguous', event_data: {} })).toBe('anti_cheat');
		expect(classifyProctorEvent({ event_type: 'screenshot_attempt_valid', event_data: {} })).toBe('anti_cheat');
		expect(classifyProctorEvent({ event_type: 'split_screen_detected', event_data: {} })).toBe('anti_cheat');
		expect(classifyProctorEvent({ event_type: 'pip_detected', event_data: {} })).toBe('anti_cheat');
		expect(classifyProctorEvent({ event_type: 'root_emulator_strong', event_data: {} })).toBe('anti_cheat');
		expect(classifyProctorEvent({ event_type: 'device_mismatch_strong', event_data: {} })).toBe('device_mismatch');
		expect(classifyProctorEvent({ event_type: 'token_reuse_confirmed', event_data: {} })).toBe('device_mismatch');
		expect(classifyProctorEvent({ event_type: 'submit_held_pending_sync', event_data: {} })).toBe('submit_guard');
		expect(classifyProctorEvent({ event_type: 'pending_sync', event_data: {} })).toBe('submit_guard');
		expect(classifyProctorEvent({ event_type: 'offline_short', event_data: {} })).toBe('stale_connection');
		expect(classifyProctorEvent({ event_type: 'web_connection_degraded', event_data: {} })).toBe('stale_connection');
		expect(proctorEventLabel({ event_type: 'screenshot_attempt_valid', event_data: {} })).toBe('Percobaan tangkap layar tervalidasi');
	});

	it('groups proctor events into simple technical and conduct buckets', () => {
		expect(proctorRiskGroup({ event_type: 'web_connection_degraded', event_data: {} })).toBe('technical');
		expect(proctorRiskGroup({ event_type: 'pending_sync', event_data: {} })).toBe('technical');
		expect(proctorRiskGroup({ event_type: 'paste_attempt', event_data: {} })).toBe('cheating');
		expect(proctorRiskGroup({ event_type: 'web_focus_lost', event_data: {} })).toBe('cheating');
		expect(proctorRiskGroup({ event_type: 'proctor_acknowledge', event_data: {} })).toBe('supervision');
		expect(proctorRiskGroupLabel('technical')).toBe('Masalah teknis');
		expect(proctorRiskGroupLabel('cheating')).toBe('Indikasi tata tertib');
		expect(proctorRiskGroupLabel('supervision')).toBe('Tindak lanjut pengawas');
	});

	it('decodes JSON event data returned from Go byte slices as base64 strings', () => {
		const encoded = btoa(JSON.stringify({ reason: 'stale_connection_attention' }));
		expect(classifyProctorEvent({ event_type: 'warning', event_data: encoded })).toBe('stale_connection');
		expect(proctorEventLabel({ event_type: 'warning', event_data: encoded })).toBe('Koneksi perlu dicek');
	});

	it('summarizes heartbeat, stale connection, warning, actions, and export evidence deterministically', () => {
		const now = new Date('2026-05-08T08:10:00Z');
		const summary = summarizeProctorEvidence({
			now,
			hasPrintPack: true,
			participants: [
				{
					participant_id: 'p1',
					nama: 'Aminah',
					nis: '24001',
					submitted_at: null,
					last_heartbeat: '2026-05-08T08:09:00Z',
					app_switch_count: 0,
					screenshot_attempt: 0,
					suspicious_flag: false,
					answered_count: 3,
					score: null
				},
				{
					participant_id: 'p2',
					nama: 'Budi',
					nis: '24002',
					submitted_at: null,
					last_heartbeat: '2026-05-08T08:01:00Z',
					app_switch_count: 1,
					screenshot_attempt: 0,
					suspicious_flag: true,
					answered_count: 2,
					score: null
				}
			],
			events: [
				{ event_type: 'heartbeat', event_data: null },
				{ event_type: 'warning', event_data: { reason: 'submit_blocked_degraded_mode' } },
				{ event_type: 'proctor_force_submit', event_data: { actor: 'admin' } },
				{ event_type: 'proctor_reset_access', event_data: { actor: 'admin' } }
			]
		});

		expect(summary.counts.heartbeat).toBe(3);
		expect(summary.counts.stale_connection).toBe(1);
		expect(summary.counts.submit_guard).toBe(1);
		expect(summary.counts.warning).toBe(1);
		expect(summary.counts.force_submit).toBe(1);
		expect(summary.counts.reset_access).toBe(1);
		expect(summary.counts.export_print).toBe(1);
		expect(summary.staleParticipantCount).toBe(1);
		expect(summary.missingCategories).toContain('device_mismatch');
		expect(summary.missingCategories).not.toContain('export_print');
	});

	it('builds token-free CSV rows for the evidence bundle', () => {
		const rows = buildProctorEvidenceCsvRows({
			generatedAt: new Date('2026-05-08T08:10:00Z'),
			sessionTitle: 'PAT IPA',
			roomName: 'Ruang 1',
			proctors: ['Pengawas A'],
			participants: [
				{
					participant_id: 'p1',
					nama: 'Aminah',
					nis: '24001',
					submitted_at: null,
					last_heartbeat: '2026-05-08T08:09:00Z',
					app_switch_count: 0,
					screenshot_attempt: 0,
					suspicious_flag: false,
					answered_count: 3,
					score: null
				}
			],
			events: [{ event_type: 'warning', event_data: { reason: 'resume_exam' }, created_at: '2026-05-08T08:00:00Z' }]
		});

		expect(rows[0]).toEqual(['section', 'category', 'timestamp', 'participant', 'nis', 'detail']);
		expect(rows.flat().join(' ')).toContain('PAT IPA');
		expect(rows.flat().join(' ')).toContain('Masuk kembali ke ujian');
		expect(rows.flat().join(' ')).not.toMatch(/token/i);
	});

	it('maps warning reason labels for the room timeline', () => {
		expect(proctorEventLabel({ event_type: 'warning', event_data: { reason: 'stale_connection_attention' } })).toBe(
			'Koneksi perlu dicek'
		);
		expect(proctorEventLabel({ event_type: 'warning', event_data: { reason: 'manual_submit' } })).toBe(
			'Pengiriman manual'
		);
	});

	it('provides C2 operator-facing evidence labels, summaries, and guidance without raw technical wording', async () => {
		const { proctorEvidenceCategoryLabel, proctorEvidenceCategorySummary, proctorOperatorGuidance } = await import(
			'./proctor-evidence'
		);
		expect(proctorEvidenceCategoryLabel('heartbeat')).toBe('Koneksi aktif');
		expect(proctorEvidenceCategorySummary('stale_connection')).toContain('Kontak perangkat terlambat');
		expect(proctorEvidenceCategoryLabel('app_background_resume')).toBe('Aplikasi ditinggalkan');
		expect(proctorEvidenceCategoryLabel('device_mismatch')).toBe('Perangkat tidak sesuai');
		expect(proctorEvidenceCategoryLabel('force_submit')).toBe('Jawaban dikirim oleh pengawas');
		expect(proctorEvidenceCategoryLabel(null)).toBe('Kejadian lain');
		expect(proctorOperatorGuidance.map((item) => item.title)).toEqual([
			'Kapan memperingatkan siswa',
			'Kapan reset akses',
			'Kapan paksa kirim'
		]);
		expect(JSON.stringify(proctorOperatorGuidance)).not.toMatch(/token\s*=|heartbeat|background|stale|screenshot|app switch|anti-cheat/i);
	});

	it('formats event detail with formal labels instead of raw key/value incident terms', async () => {
		const { proctorEventDetail, proctorEventLabel, proctorIncidentReasonLabel } = await import('./proctor-evidence');
		const detail = proctorEventDetail({
			event_type: 'proctor_incident_action',
			event_data: { action: 'warning_given', reason: 'split_screen', actor: 'admin', command_type: 'reconnect' }
		});

		expect(proctorEventLabel({ event_type: 'unknown_internal_event', event_data: null })).toBe('Kejadian pengawasan');
		expect(proctorIncidentReasonLabel('device_mismatch')).toBe('Perangkat tidak sesuai');
		expect(detail).toContain('Tindakan: Peringatan diberikan');
		expect(detail).toContain('Alasan: Layar terbagi');
		expect(detail).toContain('Petugas: Operator/admin');
		expect(detail).not.toMatch(/warning_given|split_screen|actor=|command_type=/);
	});

	it('redacts sensitive evidence values and escapes CSV injection cells', () => {
		const rows = buildProctorEvidenceCsvRows({
			generatedAt: new Date('2026-05-08T08:10:00Z'),
			sessionTitle: '=HYPERLINK("https://evil.test","click")',
			roomName: '+SUM(1,1)',
			proctors: ['@Pengawas'],
			participants: [
				{
					participant_id: 'p1',
					nama: '-Aminah',
					nis: '24001',
					submitted_at: null,
					last_heartbeat: '2026-05-08T08:09:00Z',
					app_switch_count: 0,
					screenshot_attempt: 0,
					suspicious_flag: false,
					answered_count: 3,
					score: null
				}
			],
			events: [
				{
					event_type: 'warning',
					created_at: '2026-05-08T08:00:00Z',
					nama: '\tBudi',
					nis: '\r24002',
					event_data: {
						reason: 'manual_submit token=inline-token-should-not-leak',
						actor: 'Bearer actor-token-should-not-leak',
						token: 'EXAM-TOKEN-SHOULD-NOT-LEAK',
						password: 'super-secret-password',
						api_key: 'secret-api-key',
						authorization: 'Bearer auth-token-should-not-leak',
						access_token: 'access-token-should-not-leak',
						device_fingerprint: 'raw-device-fingerprint'
					}
				}
			]
		});

		const flattened = rows.flat().join(' ');
		expect(flattened).not.toContain('EXAM-TOKEN-SHOULD-NOT-LEAK');
		expect(flattened).not.toContain('super-secret-password');
		expect(flattened).not.toContain('secret-api-key');
		expect(flattened).not.toContain('raw-device-fingerprint');
		expect(flattened).not.toContain('actor-token-should-not-leak');
		expect(flattened).not.toContain('inline-token-should-not-leak');
		expect(flattened).not.toContain('auth-token-should-not-leak');
		expect(flattened).not.toContain('access-token-should-not-leak');
		expect(flattened).toContain('[redacted]');

		for (const value of rows.flat()) {
			if (typeof value !== 'string') continue;
			expect(value).not.toMatch(/^[=+\-@\t\r\n]/);
		}
		expect(flattened).toContain("'=HYPERLINK");
		expect(flattened).toContain("'+SUM");
		expect(flattened).toContain("'@Pengawas");
	});

});
