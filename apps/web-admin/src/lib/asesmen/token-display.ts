export function maskAssessmentToken(value: string | null | undefined): string {
	const token = (value ?? '').trim();
	if (!token) return '—';
	if (token.length <= 4) return '••••';
	const prefix = token.slice(0, 2);
	const suffix = token.slice(-2);
	return `${prefix}${'•'.repeat(Math.max(4, token.length - prefix.length - suffix.length))}${suffix}`;
}
