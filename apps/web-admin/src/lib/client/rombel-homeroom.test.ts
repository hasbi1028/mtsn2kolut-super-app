import { describe, expect, it } from 'vitest';

import {
	buildHomeroomEmployeeOptions,
	employeeOptionSubtitle,
	employmentTypeLabel,
	filterHomeroomEmployeeOptions
} from './rombel-homeroom';

describe('rombel homeroom employee helpers', () => {
	it('keeps all active employees available even without subject assignments', () => {
		const options = buildHomeroomEmployeeOptions({
			employees: [
				{ id: 'emp-2', nama: 'Nur Aini', pegawai_uid: 'PG-002', nip: '', employment_type: 'pppk', is_active: true },
				{ id: 'emp-1', nama: 'Ahmad Basri', pegawai_uid: '', nip: '19800101', employment_type: 'pns', is_active: true }
			],
			subjectAssignments: [],
			activeHomeroom: null
		});

		expect(options.map((option) => option.id)).toEqual(['emp-1', 'emp-2']);
		expect(employeeOptionSubtitle(options[0])).toBe('NIP 19800101 - PNS');
		expect(employeeOptionSubtitle(options[1])).toBe('UID PG-002 - PPPK');
	});

	it('excludes inactive employees but preserves assignment and active homeroom fallbacks', () => {
		const options = buildHomeroomEmployeeOptions({
			employees: [
				{ id: 'emp-active', nama: 'Guru Aktif', employment_type: 'honorer', is_active: true },
				{ id: 'emp-inactive', nama: 'Guru Nonaktif', employment_type: 'pns', is_active: false }
			],
			subjectAssignments: [
				{ teacher_employee_id: 'emp-mapel', teacher_name: 'Guru Mapel' },
				{ teacher_employee_id: 'emp-active', teacher_name: 'Nama Lama' }
			],
			activeHomeroom: { employee_id: 'emp-wali', employee_name: 'Wali Lama' }
		});

		expect(options.map((option) => option.id)).toEqual(['emp-active', 'emp-mapel', 'emp-wali']);
		expect(options.find((option) => option.id === 'emp-active')).toMatchObject({
			name: 'Guru Aktif',
			isActiveEmployee: true,
			isFallback: false
		});
		expect(options.find((option) => option.id === 'emp-mapel')).toMatchObject({ isFallback: true });
		expect(options.find((option) => option.id === 'emp-wali')).toMatchObject({ isFallback: true });
	});

	it('filters by name, UID/NIP, and employment type labels', () => {
		const options = buildHomeroomEmployeeOptions({
			employees: [
				{ id: 'emp-1', nama: 'Ahmad Basri', pegawai_uid: 'PG-001', nip: '19800101', employment_type: 'pns', is_active: true },
				{ id: 'emp-2', nama: 'Nur Aini', pegawai_uid: 'PG-002', nip: '', employment_type: 'pppk', is_active: true }
			],
			subjectAssignments: [],
			activeHomeroom: null
		});

		expect(filterHomeroomEmployeeOptions(options, 'pg-002').map((option) => option.id)).toEqual(['emp-2']);
		expect(filterHomeroomEmployeeOptions(options, 'PNS').map((option) => option.id)).toEqual(['emp-1']);
		expect(filterHomeroomEmployeeOptions(options, 'aini').map((option) => option.id)).toEqual(['emp-2']);
		expect(employmentTypeLabel('pppk')).toBe('PPPK');
	});
});
