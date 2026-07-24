import { json } from '@sveltejs/kit';
import type { RequestEvent } from '@sveltejs/kit';
import { ApiError, handleRouteError, proxy } from '$lib/server/api';

export const POST = async (event: RequestEvent) => {
  try {
    const id = event.params.id;
    const result = await proxy(event).post<{id: string; label: string; is_active: boolean}>(`/api/academic/semesters/${id}/activate`, {});
    return json(result);
  } catch (e) {
    return handleRouteError(e, 'semester activate');
  }
};
