import type { AuthUser } from '$lib/server/auth';
import type { TokenPair } from '$lib/server/api';

declare global {
	namespace App {
		interface Locals {
			user?: AuthUser;
			accessToken?: string;
			authRefreshPromise?: Promise<TokenPair>;
		}
		interface PageData {
			user?: AuthUser;
		}
	}
}

export {};
