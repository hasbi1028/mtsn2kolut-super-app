declare global {
	namespace App {
		interface Locals {
			user?: { id: string };
		}
		interface PageData {
			user?: { id: string };
		}
	}
}

export {};
