import { redirect } from '@sveltejs/kit';
import type { PageServerLoad } from './$types';

const operatorPermissions = [
	'asesmen.operator',
	'asesmen.event_manage',
	'asesmen.package_manage',
	'asesmen.session_manage',
	'asesmen.participant_manage'
];

export const load: PageServerLoad = ({ locals }) => {
	const roles = locals.user?.roles ?? (locals.user?.role ? [locals.user.role] : []);
	const permissions = locals.user?.permissions ?? [];
	const hasOperatorLane = roles.includes('admin') || operatorPermissions.some((permission) => permissions.includes(permission));
	const hasFieldLane = permissions.includes('asesmen.proctor');
	const hasResultLane = permissions.includes('asesmen.result_read') || permissions.includes('asesmen.result_manage');

	if (hasOperatorLane) {
		throw redirect(307, '/asesmen/ringkas');
	}
	if (hasFieldLane) {
		throw redirect(307, '/asesmen/ruang-saya');
	}
	if (hasResultLane) {
		throw redirect(307, '/asesmen/hasil');
	}

	throw redirect(307, '/');
};
