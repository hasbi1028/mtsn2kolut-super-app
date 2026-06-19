import { json } from '@sveltejs/kit';
import type { RequestHandler } from './$types';

export const GET: RequestHandler = async ({ fetch, request }) => {
  const res = await fetch('http://localhost:8080/health', {
    headers: { 'Content-Type': 'application/json' }
  });
  const data = await res.json();
  return json(data, { status: res.status });
};
