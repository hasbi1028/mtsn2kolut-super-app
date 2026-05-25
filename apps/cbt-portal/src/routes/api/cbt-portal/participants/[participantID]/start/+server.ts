import type { RequestHandler } from './$types';
import { proxyCore, readJsonBody } from '$lib/api';

export const POST: RequestHandler = async (event) => {
  const body = await readJsonBody(event);
  if (typeof body !== 'string') return body;
  const id = encodeURIComponent(event.params.participantID);
  return proxyCore(event, `/api/cbt-portal/participants/${id}/start`, { method: 'POST', body });
};
