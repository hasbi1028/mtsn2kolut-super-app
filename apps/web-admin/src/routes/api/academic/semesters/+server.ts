import { json } from '@sveltejs/kit';
import type { RequestEvent } from '@sveltejs/kit';
import { ApiError, handleRouteError, proxy } from '$lib/server/api';

export interface Semester {
  id: string;
  academic_year_id: string;
  academic_year_name: string;
  name: string;
  label: string;
  start_date: string;
  end_date: string;
  is_active: boolean;
  created_at: string;
  updated_at: string;
}

export const GET = async (event: RequestEvent) => {
  try {
    const items = await proxy(event).get<Semester[]>('/api/academic/semesters');
    return json({ items });
  } catch (e) {
    return handleRouteError(e, 'semesters GET');
  }
};

export const POST = async (event: RequestEvent) => {
  try {
    const body = await event.request.json();
    const result = await proxy(event).post<Semester>('/api/academic/semesters', body);
    return json(result, { status: 201 });
  } catch (e) {
    return handleRouteError(e, 'semesters POST');
  }
};
