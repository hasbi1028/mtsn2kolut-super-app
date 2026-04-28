declare global {
	namespace App {
		interface Locals {
			user?: { id: string; username: string; role: string; employee_id?: string };
			accessToken?: string;
		}
		interface PageData {
			user?: { id: string; username: string; role: string; employee_id?: string };
		}
	}
}

export {};
