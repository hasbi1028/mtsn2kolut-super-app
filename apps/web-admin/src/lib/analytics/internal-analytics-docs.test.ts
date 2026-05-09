import { describe, expect, it } from 'vitest';
import analyticsCatalogDoc from '../../../../../docs/internal-analytics-event-catalog.md?raw';
import analyticsPlanDoc from '../../../../../docs/internal-analytics-plan.md?raw';

describe('internal analytics phase 0 documentation contract', () => {
	it('keeps the internal-only architecture and readiness boundaries explicit', () => {
		for (const phrase of [
			'internal-only, no third-party analytics',
			'privacy-first',
			'PostgreSQL/Core API ownership',
			'analytics.read',
			'analytics.export',
			'analytics.manage',
			'analytics.security_read',
			'separation from audit_logs',
			'phased plan',
			'SvelteKit/BFF',
			'Go Core API',
			'PostgreSQL',
			'Web Admin Dashboard',
			'Tidak ada runtime ingestion table',
			'Tidak ada migration',
			'Tidak ada SDK analytics pihak ketiga'
		]) {
			expect(analyticsPlanDoc).toContain(phrase);
		}
	});

	it('keeps the event allowlist and forbidden sensitive keys explicit', () => {
		for (const phrase of [
			'event allowlist',
			'forbidden sensitive keys',
			'Public Website',
			'Auth',
			'Dashboard',
			'Bank Soal',
			'Asesmen',
			'PUSAKA',
			'Users Dan RBAC',
			'Security',
			'public.page_view',
			'public.cta_click',
			'public.download',
			'auth.login_attempt',
			'dashboard.view',
			'bank_soal.editor_open',
			'asesmen.session_create',
			'pusaka.job_completed',
			'rbac.permission_update',
			'security.sensitive_key_rejected',
			'password',
			'token',
			'cookie',
			'authorization',
			'secret',
			'api_key',
			'nik',
			'nip',
			'nisn',
			'device_fingerprint',
			'credential_pusaka',
			'raw_ip',
			'raw_user_agent'
		]) {
			expect(analyticsCatalogDoc.toLowerCase()).toContain(phrase.toLowerCase());
		}
	});

	it('links the plan and catalog contracts together', () => {
		expect(analyticsPlanDoc).toContain('docs/internal-analytics-event-catalog.md');
		expect(analyticsCatalogDoc).toContain('separation from audit_logs');
		expect(analyticsCatalogDoc).toContain('internal-only, no third-party analytics');
	});
});
