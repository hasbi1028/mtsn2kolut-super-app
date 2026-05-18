export type HomeroomEmployeeRow = {
	id: string;
	pegawai_uid?: string | null;
	nip?: string | null;
	nama: string;
	employment_type?: string | null;
	is_active?: boolean | null;
};

export type HomeroomSubjectAssignmentRow = {
	teacher_employee_id: string;
	teacher_name: string;
};

export type HomeroomAssignmentRow = {
	employee_id: string;
	employee_name: string;
};

export type HomeroomEmployeeOption = {
	id: string;
	name: string;
	pegawaiUid: string;
	nip: string;
	employmentType: string;
	isActiveEmployee: boolean;
	isFallback: boolean;
};

export function employmentTypeLabel(value: string | null | undefined) {
	const normalized = (value ?? '').trim().toLowerCase();
	const labels: Record<string, string> = {
		pns: 'PNS',
		pppk: 'PPPK',
		honorer: 'Honorer',
		lainnya: 'Lainnya'
	};
	return labels[normalized] ?? (normalized ? normalized.toUpperCase() : '');
}

export function employeeOptionSubtitle(option: HomeroomEmployeeOption) {
	const parts: string[] = [];
	if (option.nip) parts.push(`NIP ${option.nip}`);
	else if (option.pegawaiUid) parts.push(`UID ${option.pegawaiUid}`);
	const employmentLabel = employmentTypeLabel(option.employmentType);
	if (employmentLabel) parts.push(employmentLabel);
	if (option.isFallback && parts.length === 0) parts.push('Dari data rombel');
	return parts.join(' - ') || 'NIP belum tersedia';
}

export function buildHomeroomEmployeeOptions(args: {
	employees: HomeroomEmployeeRow[];
	subjectAssignments: HomeroomSubjectAssignmentRow[];
	activeHomeroom: HomeroomAssignmentRow | null;
}): HomeroomEmployeeOption[] {
	const options = new Map<string, HomeroomEmployeeOption>();

	for (const employee of args.employees) {
		if (!employee.id || !employee.nama || employee.is_active !== true) continue;
		options.set(employee.id, {
			id: employee.id,
			name: employee.nama,
			pegawaiUid: employee.pegawai_uid?.trim() ?? '',
			nip: employee.nip?.trim() ?? '',
			employmentType: employee.employment_type?.trim() ?? '',
			isActiveEmployee: true,
			isFallback: false
		});
	}

	for (const assignment of args.subjectAssignments) {
		if (!assignment.teacher_employee_id || !assignment.teacher_name) continue;
		if (options.has(assignment.teacher_employee_id)) continue;
		options.set(assignment.teacher_employee_id, {
			id: assignment.teacher_employee_id,
			name: assignment.teacher_name,
			pegawaiUid: '',
			nip: '',
			employmentType: '',
			isActiveEmployee: false,
			isFallback: true
		});
	}

	if (args.activeHomeroom?.employee_id && args.activeHomeroom.employee_name && !options.has(args.activeHomeroom.employee_id)) {
		options.set(args.activeHomeroom.employee_id, {
			id: args.activeHomeroom.employee_id,
			name: args.activeHomeroom.employee_name,
			pegawaiUid: '',
			nip: '',
			employmentType: '',
			isActiveEmployee: false,
			isFallback: true
		});
	}

	return Array.from(options.values()).sort((a, b) => a.name.localeCompare(b.name, 'id-ID'));
}

export function filterHomeroomEmployeeOptions(options: HomeroomEmployeeOption[], search: string) {
	const query = search.trim().toLowerCase();
	if (!query) return options;
	return options.filter((option) => {
		const haystack = [
			option.name,
			option.pegawaiUid,
			option.nip,
			option.employmentType,
			employmentTypeLabel(option.employmentType),
			option.isFallback ? 'guru mapel rombel wali kelas' : 'pegawai aktif'
		].join(' ').toLowerCase();
		return haystack.includes(query);
	});
}
