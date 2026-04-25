import { json } from '@sveltejs/kit';
import type { RequestHandler } from '@sveltejs/kit';
import { getPasswordHash, verifyPassword, hashPassword, setPasswordHash } from '$lib/server/auth';

interface Body { current_password: string; new_password: string; }

export const POST: RequestHandler = async ({ request, locals }) => {
	if (!locals.user) return json({ error: 'Unauthorized' }, { status: 401 });

	const { current_password, new_password } = await request.json() as Body;

	if (!current_password || !new_password)
		return json({ error: 'Field tidak boleh kosong' }, { status: 400 });
	if (new_password.length < 6)
		return json({ error: 'Password baru minimal 6 karakter' }, { status: 400 });

	if (!verifyPassword(current_password, getPasswordHash()))
		return json({ error: 'Password saat ini salah' }, { status: 401 });

	setPasswordHash(hashPassword(new_password));
	return json({ ok: true });
};
