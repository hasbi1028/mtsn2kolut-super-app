import { describe, expect, it } from 'vitest';
import { examSlots, getFlatAssignments, getSupervisorLoad, getWarnings, toCsv } from './uas-genap-2026.model';

describe('uas genap 2026 command center model', () => {
	it('uses corrected dates with no Sunday exam', () => {
		expect(examSlots.map((slot) => `${slot.day}:${slot.date}`)).toContain('Senin:2026-06-08');
		expect(examSlots.map((slot) => `${slot.day}:${slot.date}`)).toContain('Selasa:2026-06-09');
		expect(examSlots.map((slot) => `${slot.day}:${slot.date}`)).toContain('Rabu:2026-06-10');
		expect(examSlots.some((slot) => slot.date === '2026-06-07')).toBe(false);
	});

	it('has 15 slots and 120 room assignments', () => {
		expect(examSlots).toHaveLength(15);
		expect(getFlatAssignments()).toHaveLength(120);
	});

	it('keeps pengawas 23 duplicate as pending review warning', () => {
		const warnings = getWarnings();
		expect(warnings.some((warning) => warning.includes('pengawas 23'))).toBe(true);
		const problematic = examSlots.find((slot) => slot.subject === "Al-Qur'an Hadis");
		expect(problematic?.status).toBe('needs_review');
	});

	it('exports audit-friendly CSV rows', () => {
		const csv = toCsv();
		expect(csv).toContain('"tanggal","hari","jam_mulai"');
		expect(csv).toContain('"2026-06-10","Rabu","09:30","11:00","PKN","R.VIII","16","Susianti S.Pd","ready"');
	});

	it('computes supervisor loads for all 27 supervisors', () => {
		const load = getSupervisorLoad();
		expect(load).toHaveLength(27);
		expect(load.find((item) => item.code === 23)?.count).toBeGreaterThan(0);
	});
});
