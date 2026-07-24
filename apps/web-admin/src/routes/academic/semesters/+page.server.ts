import type { PageServerLoad } from './$types.js';

type Semester = {
  id: string;
  academic_year_id: string;
  academic_year_name: string;
  name: string;
  label: string;
  start_date: string;
  end_date: string;
  is_active: boolean;
};

export const load: PageServerLoad = async ({ fetch, url }) => {
  const res = await fetch(`${url.origin}/api/academic/semesters`);
  if (!res.ok) {
    const items: Semester[] = [];
    return { items };
  }
  const payload = await res.json();
  return { items: (payload.items ?? []) as Semester[] };
};
