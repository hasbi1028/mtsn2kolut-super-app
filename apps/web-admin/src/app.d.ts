import type { AuthUser } from '$lib/server/auth';

declare global {
	namespace App {
		interface Locals {
			user?: AuthUser;
			accessToken?: string;
		}
		interface PageData {
			user?: AuthUser;
		}
	}
}

export {};
