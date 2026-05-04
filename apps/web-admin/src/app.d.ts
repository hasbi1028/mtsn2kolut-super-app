import type { AuthUser } from '$lib/server/auth';
import type { TokenPair } from '$lib/server/api';

declare global {
	namespace App {
		interface Locals {
			user?: AuthUser;
			accessToken?: string;
			authRefreshPromise?: Promise<TokenPair>;
			authUserPromise?: Promise<AuthUser | undefined>;
		}
		interface PageData {
			user?: AuthUser;
		}
	}
}

export {};
