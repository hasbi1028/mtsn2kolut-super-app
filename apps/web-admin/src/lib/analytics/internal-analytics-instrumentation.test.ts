import { readFileSync } from 'node:fs';
import { resolve } from 'node:path';
import { describe, expect, it } from 'vitest';

const repoRoot = resolve(__dirname, '../../../../..');

function source(path: string) {
	return readFileSync(resolve(repoRoot, path), 'utf8');
}

describe('internal analytics instrumentation source guards', () => {
	it('instruments only high-value internal web-admin page views', () => {
		const targets = [
			['apps/web-admin/src/routes/+page.svelte', 'dashboard.view'],
			['apps/web-admin/src/routes/bank-soal/+page.svelte', 'bank_soal.list_view'],
			['apps/web-admin/src/routes/pusaka/+page.svelte', 'pusaka.dashboard_view'],
			['apps/web-admin/src/routes/settings/users/+page.svelte', 'users.list_view'],
			['apps/web-admin/src/routes/settings/rbac/+page.svelte', 'rbac.roles_view'],
			['apps/web-admin/src/routes/settings/analytics/+page.svelte', 'security.analytics_view']
		] as const;

		for (const [path, eventName] of targets) {
			const text = source(path);
			expect(text, `${path} should import analytics helper`).toContain('$lib/analytics/internal-analytics');
			expect(text, `${path} should record ${eventName}`).toContain(eventName);
			expect(text, `${path} should keep analytics client-side only`).toContain('onMount');
		}
	});

	it('does not add third-party analytics packages or browser beacon tracking', () => {
		const packageJson = source('apps/web-admin/package.json').toLowerCase();
		for (const forbidden of ['posthog', 'plausible', 'google-analytics', '@vercel/analytics', 'hotjar']) {
			expect(packageJson).not.toContain(forbidden);
		}
		const helper = source('apps/web-admin/src/lib/analytics/internal-analytics.ts');
		expect(helper).not.toContain('navigator.sendBeacon');
		expect(helper).not.toContain('sendBeacon(');
	});

	it('activates public website analytics only through the first-party collector', () => {
		const shell = source('apps/web-admin/src/lib/components/PublicSiteShell.svelte');
		const helper = source('apps/web-admin/src/lib/analytics/public-analytics.ts');
		const route = source('apps/web-admin/src/routes/api/public/analytics/events/+server.ts');

		for (const eventName of ['public.page_view', 'public.cta_click', 'public.download', 'public.search', 'public.form_start', 'public.form_submit']) {
			expect(helper).toContain(eventName);
		}
		expect(shell).toContain('trackPublicPageView');
		expect(shell).toContain('handlePublicClick');
		expect(route).toContain('/api/internal-analytics/public-events');
		expect(route).toContain('X-Internal-Key');
		expect(route).not.toContain('X-Forwarded-For');
		expect(route).not.toContain('user-agent');
	});
});
