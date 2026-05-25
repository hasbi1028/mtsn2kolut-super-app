import type { RequestHandler } from './$types';
import { proxyCore, readJsonBody } from '$lib/api';

export const POST: RequestHandler = async (event) => {
  const body = await readJsonBody(event, 65536);
  if (typeof body !== 'string') return body;
  const id = encodeURIComponent(event.params.participantID);
  return proxyCore(event, `/api/cbt-portal/participants/${id}/answer`, { method: 'POST', body });
};
