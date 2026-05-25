import type { RequestHandler } from './$types';
import { proxyCore, readJsonBody } from '$lib/api';

export const POST: RequestHandler = async (event) => {
  const body = await readJsonBody(event);
  if (typeof body !== 'string') return body;
  return proxyCore(event, '/api/cbt-portal/login', { method: 'POST', body });
};
