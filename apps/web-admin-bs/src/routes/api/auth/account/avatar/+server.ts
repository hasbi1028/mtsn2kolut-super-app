import { json } from '@sveltejs/kit';
import type { RequestHandler } from './$types';
import { handleRouteError, jsonProxyResponse, proxy } from '$lib/server/api';

export const POST: RequestHandler = async (event) => {
	if (!event.locals.user) {
		return json({ error: 'Unauthorized' }, { status: 401 });
	}

	try {
		const form = await event.request.formData();
		const response = await proxy(event).fetch('/api/auth/account/avatar', {
			method: 'POST',
			body: form
		});
		return await jsonProxyResponse(response, {
			fallbackMessage: 'Gagal mengunggah foto profil.'
		});
	} catch (e) {
		return handleRouteError(e, 'auth/account/avatar POST');
	}
};

export const DELETE: RequestHandler = async (event) => {
	if (!event.locals.user) {
		return json({ error: 'Unauthorized' }, { status: 401 });
	}

	try {
		const response = await proxy(event).fetch('/api/auth/account/avatar', {
			method: 'DELETE'
		});
		return await jsonProxyResponse(response, {
			fallbackMessage: 'Gagal menghapus foto profil.'
		});
	} catch (e) {
		return handleRouteError(e, 'auth/account/avatar DELETE');
	}
};
