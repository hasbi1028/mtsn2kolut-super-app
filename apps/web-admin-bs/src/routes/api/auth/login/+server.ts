import { json } from '@sveltejs/kit';
import { dev } from '$app/environment';
import type { RequestHandler } from './$types';
import { apiLoginWithFetch, type TokenPair } from '$lib/server/api';

export const POST: RequestHandler = async ({ fetch, request, cookies }) => {
  const body = await request.json();
  const username = body.username || '(empty)';
  console.log(`[login] Attempt: username="${username}"`);

  const userAgent = request.headers.get('user-agent') || '';
  const ip = request.headers.get('x-forwarded-for') || request.headers.get('x-real-ip') || '';

  try {
    const data: TokenPair = await apiLoginWithFetch(fetch, body.username ?? '', body.password ?? '', {
      userAgent,
      ipAddress: ip,
    });

    console.log(`[login] Success! Has access_token: ${!!data.access_token}`);

    if (data.access_token) {
      cookies.set('access_token', data.access_token, {
        path: '/',
        httpOnly: true,
        sameSite: 'lax',
        secure: !dev,
        maxAge: 3600,
      });
      console.log('[login] Set access_token cookie');
    }
    if (data.refresh_token) {
      cookies.set('refresh_token', data.refresh_token, {
        path: '/',
        httpOnly: true,
        sameSite: 'lax',
        secure: !dev,
        maxAge: 7 * 24 * 3600,
      });
      console.log('[login] Set refresh_token cookie');
    }

    return json(data);
  } catch (e: any) {
    const status = e?.status ?? 503;
    const message = e?.message || 'Backend tidak dapat dihubungi';
    console.error(`[login] Error (status=${status}): ${message}`);
    return json({ error: message }, { status });
  }
};
