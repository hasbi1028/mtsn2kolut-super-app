import { json, type RequestEvent } from '@sveltejs/kit';

const CORE_API_URL = process.env.CORE_API_URL ?? process.env.PUBLIC_CORE_API_URL ?? 'http://127.0.0.1:8080';

export async function proxyCore(event: RequestEvent, path: string, init: RequestInit = {}) {
  const headers = new Headers(init.headers);
  const auth = event.request.headers.get('authorization');
  if (auth) headers.set('authorization', auth);
  if (!headers.has('content-type') && init.body) headers.set('content-type', 'application/json');
  const response = await fetch(`${CORE_API_URL}${path}`, { ...init, headers });
  const text = await response.text();
  const contentType = response.headers.get('content-type') ?? 'application/json';
  return new Response(text, { status: response.status, headers: { 'content-type': contentType } });
}

export async function readJsonBody(event: RequestEvent, maxLength = 8192) {
  const body = await event.request.text();
  if (body.length > maxLength) return json({ error: 'payload too large' }, { status: 413 });
  return body || '{}';
}
