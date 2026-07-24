import { json } from '@sveltejs/kit';
import type { RequestEvent } from '@sveltejs/kit';
import { ApiError, handleRouteError, proxy } from '$lib/server/api';

export const DELETE = async (event: RequestEvent) => {
  try {
    const id = event.params.id;
    await proxy(event).del(`/api/academic/semesters/${id}`);
    return new Response(null, { status: 204 });
  } catch (e) {
    return handleRouteError(e, 'semester DELETE');
  }
};
