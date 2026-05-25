import type { RequestHandler } from './$types';
import { proxyCore } from '$lib/api';

export const POST: RequestHandler = async (event) => {
  const id = encodeURIComponent(event.params.participantID);
  return proxyCore(event, `/api/cbt-portal/participants/${id}/submit`, { method: 'POST', body: '{}' });
};
