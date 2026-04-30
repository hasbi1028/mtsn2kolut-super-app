declare global {
	namespace App {
		interface Locals {
			user?: {
				id: string;
				username: string;
				role: string;
				roles?: string[];
				session_id?: string;
				employee_id?: string;
				student_id?: string;
				parent_id?: string;
			};
			accessToken?: string;
		}
		interface PageData {
			user?: {
				id: string;
				username: string;
				role: string;
				roles?: string[];
				session_id?: string;
				employee_id?: string;
				student_id?: string;
				parent_id?: string;
			};
		}
	}
}

export {};
