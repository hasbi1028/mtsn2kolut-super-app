/**
 * Validates the Authorization: Bearer <WORKER_TOKEN> header.
 * If WORKER_TOKEN env is not set, all requests are allowed (dev mode).
 */
export function checkWorkerAuth(request: Request): boolean {
	const token = process.env.WORKER_TOKEN ?? '';
	if (!token) return true;
	return request.headers.get('authorization') === `Bearer ${token}`;
}
