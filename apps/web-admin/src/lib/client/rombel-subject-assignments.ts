export type RombelSubjectRow = {
	id: string;
	code?: string | null;
	name: string;
	is_active?: boolean | null;
};

export type RombelSubjectAssignmentRow = {
	id: string;
	subject_id: string;
	subject_name: string;
	subject_code?: string | null;
};

export type RombelSubjectOption = {
	id: string;
	name: string;
	code: string;
	isActive: boolean;
	isFallback: boolean;
};

export function buildRombelSubjectOptions(args: {
	subjects: RombelSubjectRow[];
	assignments: RombelSubjectAssignmentRow[];
}): RombelSubjectOption[] {
	const options = new Map<string, RombelSubjectOption>();

	for (const subject of args.subjects) {
		if (!subject.id || !subject.name || subject.is_active !== true) continue;
		options.set(subject.id, {
			id: subject.id,
			name: subject.name,
			code: subject.code?.trim() ?? '',
			isActive: true,
			isFallback: false
		});
	}

	for (const assignment of args.assignments) {
		if (!assignment.subject_id || !assignment.subject_name || options.has(assignment.subject_id)) continue;
		options.set(assignment.subject_id, {
			id: assignment.subject_id,
			name: assignment.subject_name,
			code: assignment.subject_code?.trim() ?? '',
			isActive: false,
			isFallback: true
		});
	}

	return Array.from(options.values()).sort((a, b) => a.name.localeCompare(b.name, 'id-ID'));
}

export function availableRombelSubjectOptions(
	options: RombelSubjectOption[],
	assignments: RombelSubjectAssignmentRow[],
	editingAssignmentId: string
) {
	const assignedSubjectIds = new Set(
		assignments
			.filter((assignment) => assignment.id !== editingAssignmentId)
			.map((assignment) => assignment.subject_id)
	);
	return options.filter((option) => !assignedSubjectIds.has(option.id));
}

export function subjectOptionLabel(option: RombelSubjectOption) {
	const suffix = option.code ? ` (${option.code})` : '';
	return `${option.name}${suffix}`;
}
