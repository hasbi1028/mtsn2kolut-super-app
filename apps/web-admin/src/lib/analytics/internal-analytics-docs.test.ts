import { describe, expect, it } from 'vitest';
import analyticsCatalogDoc from '../../../../../docs/internal-analytics-event-catalog.md?raw';
import analyticsPlanDoc from '../../../../../docs/internal-analytics-plan.md?raw';
import analyticsRunbookDoc from '../../../../../docs/internal-analytics-runbook.md?raw';

describe('internal analytics documentation contract', () => {
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
			'Fase 0 tidak menambahkan runtime ingestion table',
			'Fase 0 tidak menambahkan migration',
			'Tidak ada SDK analytics pihak ketiga'
		]) {
			expect(analyticsPlanDoc).toContain(phrase);
		}
	});

	it('keeps Phase 1 limited to schema, migration draft, and sqlc query contract', () => {
		for (const phrase of [
			'Fase 1 - Schema design dan migration draft',
			'Fase 1 menambahkan schema analytics internal',
			'083_internal_analytics_schema.sql',
			'internal_analytics.sql',
			'create analytics event',
			'get event by id',
			'list events for rollup',
			'delete expired events',
			'upsert daily aggregate',
			'list daily aggregates',
			'tidak menambahkan ingestion API',
			'tidak menambahkan BFF route',
			'tidak menambahkan frontend tracking',
			'tidak menjalankan migration live',
			'tidak deploy',
			'tidak restart PM2'
		]) {
			expect(analyticsPlanDoc.toLowerCase()).toContain(phrase.toLowerCase());
		}
	});

	it('documents Phase 2 as Core API ingestion minimum only', () => {
		for (const phrase of [
			'Fase 2 - Core API ingestion minimum',
			'JWT protected',
			'event allowlist',
			'forbidden sensitive keys',
			'body cap',
			'tidak menambahkan BFF route',
			'tidak menambahkan frontend tracking',
			'tidak menjalankan migration live',
			'tidak deploy',
			'tidak restart PM2'
		]) {
			expect(analyticsPlanDoc.toLowerCase()).toContain(phrase.toLowerCase());
		}
	});

	it('documents Phase 3 through Phase 6 implementation boundaries', () => {
		for (const phrase of [
			'Fase 3 - BFF proxy contract',
			'/api/internal-analytics/events',
			'forward JWT',
			'Fase 4 - Public website instrumentation',
			'public runtime deferred',
			'fail-closed',
			'tidak membuka unauthenticated public collector',
			'Fase 5 - Internal app instrumentation',
			'source_surface web_admin',
			'tidak mengirim raw URL query',
			'tidak mengirim raw user agent',
			'Fase 6 - Dashboard read model',
			'/api/internal-analytics/summary',
			'/api/internal-analytics/daily',
			'analytics.read',
			'tidak expose raw event metadata',
			'tidak menambahkan raw event export'
		]) {
			expect(analyticsPlanDoc.toLowerCase()).toContain(phrase.toLowerCase());
		}
	});

	it('documents Phase 7 through Phase 10 readiness boundaries', () => {
		for (const phrase of [
			'Status: Fase 10 selesai',
			'Fase 7 - Export dan reporting',
			'CSV aggregate summary/daily counts only',
			'analytics.export',
			'security.export_requested',
			'CSV injection safe',
			'Fase 8 - Retention, rollup, dan cleanup',
			'manual admin/ops invocation only',
			'expired_event_backlog_count',
			'oldest_expired_event_at',
			'Fase 9 - Security review dan abuse hardening',
			'trusted-proxy-aware rate limit',
			'body cap tetap 16 KiB',
			'Fase 10 - Operational readiness dan handoff',
			'docs/internal-analytics-runbook.md',
			'tidak deploy',
			'tidak restart PM2',
			'tidak menjalankan live migration',
			'tidak menjalankan cleanup terhadap live DB',
			'tidak membuka public unauthenticated collector',
			'tidak menambahkan raw event export',
			'tidak menambahkan third-party analytics'
		]) {
			expect(analyticsPlanDoc.toLowerCase()).toContain(phrase.toLowerCase());
		}
	});

	it('keeps the operations runbook explicit and 100 percent internal', () => {
		for (const phrase of [
			'Internal Analytics Runbook',
			'Owner',
			'Web Admin -> BFF -> Go Core API -> PostgreSQL',
			'analytics.read',
			'analytics.export',
			'Export policy',
			'aggregate_date,event_group,event_name,source_surface,role,result,count',
			'Retention and cleanup',
			'Smoke checklist',
			'Rollback',
			'Recovery',
			'100% internal evidence',
			'no public collector',
			'no raw event export',
			'no third-party analytics'
		]) {
			expect(analyticsRunbookDoc.toLowerCase()).toContain(phrase.toLowerCase());
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
			'deviceFingerprint',
			'credential_pusaka',
			'raw_ip',
			'raw_user_agent',
			'rawUserAgent',
			'query_string',
			'rawQuery',
			'full_url'
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
