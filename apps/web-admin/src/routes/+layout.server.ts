import type { LayoutServerLoad } from './$types';
import { redirect } from '@sveltejs/kit';
import { apiPublicGetWithFetch, proxy } from '$lib/server/api';
import { defaultBranding, normalizeBranding, type BrandingSettings } from '$lib/branding';
import type { AccountIdentity } from '$lib/client/account';
import { canBypassMaintenance, routeAffectedByMaintenance, type MaintenanceStatus } from '$lib/maintenance/modules';

export const load: LayoutServerLoad = async (event) => {
	let account: AccountIdentity | null = null;
	let branding: BrandingSettings = defaultBranding;
	let maintenanceStatus: MaintenanceStatus = {
		active: false,
		mode: 'off',
		server_time: new Date().toISOString()
	};

	try {
		branding = normalizeBranding(await apiPublicGetWithFetch(event.fetch, '/api/public/branding'));
	} catch {
		branding = defaultBranding;
	}

	try {
		maintenanceStatus = await apiPublicGetWithFetch<MaintenanceStatus>(event.fetch, '/api/system/maintenance/status');
	} catch {
		maintenanceStatus = {
			active: false,
			mode: 'off',
			server_time: new Date().toISOString()
		};
	}

	if (event.locals.user) {
		try {
			account = await proxy(event).get<AccountIdentity>('/api/auth/account');
		} catch {
			account = null;
		}
	}

	const maintenanceWindow = maintenanceStatus.window;
	const isMaintenancePage = event.url.pathname === '/maintenance';
	const isAffectedByMaintenance = routeAffectedByMaintenance(maintenanceWindow, event.url.pathname);
	const canBypass = canBypassMaintenance(event.locals.user ?? null, maintenanceWindow);
	const shouldRedirectForMaintenance = maintenanceWindow?.mode === 'global' || maintenanceWindow?.mode === 'module';
	if (maintenanceStatus.active && shouldRedirectForMaintenance && !isMaintenancePage && isAffectedByMaintenance && !canBypass) {
		throw redirect(302, `/maintenance?from=${encodeURIComponent(event.url.pathname + event.url.search)}`);
	}

	return { user: event.locals.user, account, branding, maintenanceStatus };
};
