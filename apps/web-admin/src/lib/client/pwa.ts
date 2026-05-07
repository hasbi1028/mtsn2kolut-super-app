export type WebAdminPwaRegistrationContext = {
	isAuthenticated: boolean;
	isPublicSite: boolean;
	isLogin: boolean;
	location?: Pick<Location, 'protocol'>;
	serviceWorker?: ServiceWorkerContainer;
};

function defaultRegistrationContext(): WebAdminPwaRegistrationContext {
	return {
		isAuthenticated: true,
		isPublicSite: false,
		isLogin: false,
		location: typeof window !== 'undefined' ? window.location : undefined,
		serviceWorker: typeof navigator !== 'undefined' ? navigator.serviceWorker : undefined,
	};
}

export function shouldRegisterWebAdminPwa(context: WebAdminPwaRegistrationContext): boolean {
	if (!context.isAuthenticated || context.isPublicSite || context.isLogin) return false;
	if (!context.serviceWorker) return false;
	const protocol = context.location?.protocol;
	return protocol === 'https:' || protocol === 'http:';
}

export async function registerWebAdminPwa(
	context: WebAdminPwaRegistrationContext = defaultRegistrationContext()
): Promise<ServiceWorkerRegistration | null> {
	if (!shouldRegisterWebAdminPwa(context)) return null;
	try {
		return await context.serviceWorker!.register('/sw.js', { scope: '/' });
	} catch (error) {
		console.warn('Web-admin PWA registration skipped', error);
		return null;
	}
}
