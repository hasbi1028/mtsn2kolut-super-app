export type DisplayNameCandidate = {
	display_name?: string | null;
	displayName?: string | null;
	name?: string | null;
	nama?: string | null;
	full_name?: string | null;
	fullName?: string | null;
	label?: string | null;
	username?: string | null;
	id?: string | null;
};

function clean(value: unknown) {
	return typeof value === 'string' ? value.trim() : '';
}

export function displayName(
	value: DisplayNameCandidate | string | null | undefined,
	fallback = 'Tidak diketahui'
) {
	if (typeof value === 'string') return clean(value) || fallback;
	if (!value) return fallback;
	return (
		clean(value.display_name) ||
		clean(value.displayName) ||
		clean(value.nama) ||
		clean(value.name) ||
		clean(value.full_name) ||
		clean(value.fullName) ||
		clean(value.label) ||
		clean(value.username) ||
		clean(value.id) ||
		fallback
	);
}

export function actorDisplayName(prefix: string, row: Record<string, unknown>, fallback = 'Tidak diketahui') {
	return displayName(
		{
			display_name: clean(row[`${prefix}_display_name`]),
			name: clean(row[`${prefix}_name`]),
			nama: clean(row[`${prefix}_nama`]),
			full_name: clean(row[`${prefix}_full_name`]),
			label: clean(row[`${prefix}_label`]),
			username: clean(row[`${prefix}_username`]),
			id: clean(row[`${prefix}_id`])
		},
		fallback
	);
}

export function optionLabel<T extends DisplayNameCandidate>(value: T, fallback = 'Tidak diketahui') {
	return displayName(value, fallback);
}
