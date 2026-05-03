export function canExportAllQuestions(roles: readonly string[]) {
	return roles.includes('admin');
}

export function questionExportButtonLabel(roles: readonly string[]) {
	return canExportAllQuestions(roles) ? 'Export CSV' : 'Export Soal Saya';
}

export function questionExportSuccessMessage(roles: readonly string[]) {
	return canExportAllQuestions(roles)
		? 'Export CSV bank soal berhasil dibuat'
		: 'Export CSV soal saya berhasil dibuat';
}
