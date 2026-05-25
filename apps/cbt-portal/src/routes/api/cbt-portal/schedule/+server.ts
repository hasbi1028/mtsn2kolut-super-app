import type { RequestHandler } from './$types';
import { proxyCore } from '$lib/api';

export const GET: RequestHandler = async (event) => proxyCore(event, '/api/cbt-portal/schedule');
