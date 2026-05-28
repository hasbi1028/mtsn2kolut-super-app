import { describe, expect, it } from 'vitest';
import { readFileSync } from 'node:fs';

const layoutSource = readFileSync('src/routes/+layout.svelte', 'utf8');
const publicPolicySource = readFileSync('src/lib/routes/public-policy.ts', 'utf8');
const studentPortalSource = readFileSync('src/routes/ujian/+page.svelte', 'utf8');
const proctorPortalSource = readFileSync('src/routes/pengawas-ujian/+page.svelte', 'utf8');

describe('simple CBT mobile web portals', () => {
	it('keeps student and proctor portals as public standalone shells outside admin chrome', () => {
		expect(layoutSource).toContain("page.url.pathname === '/ujian' || page.url.pathname === '/pengawas-ujian'");
		expect(publicPolicySource).toContain("'/ujian'");
		expect(publicPolicySource).toContain("'/pengawas-ujian'");
		expect(publicPolicySource).toContain("'/api/exam/'");
		expect(publicPolicySource).toContain("'/api/cbt-portal/'");
	});

	it('forces light mobile app surfaces for ujian and pengawas-ujian', () => {
		for (const source of [studentPortalSource, proctorPortalSource]) {
			expect(source).toContain('<meta name="color-scheme" content="light" />');
			expect(source).toContain('exam-light-scope');
			expect(source).toContain('color-scheme: light;');
		}
	});

	it('keeps the student demo as a one-question-at-a-time click-through flow', () => {
		expect(studentPortalSource).toContain("demoMode && !payload) activateDemo()");
		expect(studentPortalSource).toContain("portalStep = 'confirm'");
		expect(studentPortalSource).toContain('Soal {Math.min(activeQuestionIndex + 1, questions.length)} dari {questions.length}');
		expect(studentPortalSource).toContain('Sebelumnya');
		expect(studentPortalSource).toContain('Ragu-ragu');
		expect(studentPortalSource).toContain('Berikutnya');
		expect(studentPortalSource).toContain('Kirim Jawaban');
	});

	it('shows lightweight participant watermarks and keeps only non-answer content non-selectable', () => {
		expect(studentPortalSource).toContain('let watermarkTime = $state');
		expect(studentPortalSource).toContain('const watermarkTimer = setInterval(updateWatermarkTime, 30000)');
		expect(studentPortalSource).toContain("timeZone: 'Asia/Makassar'");
		expect(studentPortalSource).toContain(" + ' WITA'");
		expect(studentPortalSource).toContain('seatLabel');
		expect(studentPortalSource).toContain('participantWatermark');
		expect(studentPortalSource).toContain('{watermarkLine}');
		expect(studentPortalSource).toContain('{studentName} · {classLabel} · {roomLabel} · {seatLabel} · {participantWatermark}');
		expect(studentPortalSource).toContain('data-exam-anti-cheat-scope="active"');
		expect(studentPortalSource).toContain('whitespace-pre-wrap text-base font-semibold leading-7 select-none');
		expect(studentPortalSource).toContain('flex min-h-14 select-none items-start');
		expect(studentPortalSource).toContain('class="select-none"');
		expect(studentPortalSource).toContain('textarea class="mt-4 min-h-40 w-full select-text');
	});

	it('keeps the proctor demo simple and actionable without admin-only unlock controls', () => {
		expect(proctorPortalSource).toContain("activeTab = $state<'ruang' | 'peringatan' | 'peserta'>('ruang')");
		expect(proctorPortalSource).toContain('Ruang = buka/tutup ujian, Peringatan = cek masalah peserta, Peserta = lihat status per siswa.');
		expect(proctorPortalSource).toContain('markDemoEventHandled');
		expect(proctorPortalSource).not.toContain('disabled={Boolean(actionBusy) || demoMode}');
		expect(proctorPortalSource).not.toContain('unlockParticipant');
		expect(proctorPortalSource).not.toContain('/unlock');
		expect(proctorPortalSource).not.toContain('/reset');
		expect(proctorPortalSource).not.toContain('force-submit');
		expect(proctorPortalSource).not.toContain('onclick={() => void unlockParticipant');
	});

	it('groups proctor warnings into simple Phase 4 buckets with friendly labels', () => {
		expect(proctorPortalSource).toContain("alertFilter = $state<'all' | 'technical' | 'cheating' | 'supervision' | 'red' | 'yellow' | 'unchecked'>('all')");
		expect(proctorPortalSource).toContain('groupedVisibleEvents');
		expect(proctorPortalSource).toContain('Masalah teknis');
		expect(proctorPortalSource).toContain('Indikasi tata tertib');
		expect(proctorPortalSource).toContain('Tindak lanjut pengawas');
		expect(proctorPortalSource).toContain('Ringkasan');
		expect(proctorPortalSource).toContain('Label ramah pengawas');
		expect(proctorPortalSource).toContain('Clipboard');
		expect(proctorPortalSource).toContain('Focus/visibility');
		expect(proctorPortalSource).toContain('Stale/pending sync');
		expect(proctorPortalSource).toContain('Device mismatch');
		expect(proctorPortalSource).toContain('Token reuse');
		expect(proctorPortalSource).toContain('Screenshot');
		expect(proctorPortalSource).toContain('Proctor follow-up');
	});
});
