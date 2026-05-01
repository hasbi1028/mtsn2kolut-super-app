import { readClientJson } from '$lib/client/api';

export type SidebarAttention = {
	inventory: number;
	library: number;
	pusaka: number;
};

const zeroAttention: SidebarAttention = {
	inventory: 0,
	library: 0,
	pusaka: 0
};

type JsonRecord = Record<string, unknown>;

function canSeeInventoryAttention(roles: string[]) {
	return roles.includes('admin') || roles.includes('staf');
}

function canSeePusakaAttention(roles: string[]) {
	return roles.includes('admin');
}

function isRecord(value: unknown): value is JsonRecord {
	return typeof value === 'object' && value !== null;
}

async function readStatsPayload(response: Response): Promise<JsonRecord | null> {
	if (!response.ok) return null;
	const payload = await readClientJson<unknown>(response).catch(() => null);
	return isRecord(payload) ? payload : null;
}

function numericStat(payload: JsonRecord | null, key: string): number {
	if (!payload) return 0;
	const data = isRecord(payload.data) ? payload.data : {};
	const value = data[key] ?? payload[key] ?? 0;
	const parsed = Number(value);
	return Number.isFinite(parsed) ? parsed : 0;
}

export async function fetchSidebarAttention(
	fetchImpl: typeof fetch,
	roles: string[]
): Promise<SidebarAttention> {
	const nextAttention: SidebarAttention = { ...zeroAttention };
	const requests: Promise<void>[] = [];

	if (canSeeInventoryAttention(roles)) {
		requests.push(
			fetchImpl('/api/inventory/stats')
				.then(readStatsPayload)
				.then((payload) => {
					nextAttention.inventory = numericStat(payload, 'perlu_restok');
				})
				.catch(() => {
					nextAttention.inventory = 0;
				})
		);
		requests.push(
			fetchImpl('/api/library/stats')
				.then(readStatsPayload)
				.then((payload) => {
					const overdue = numericStat(payload, 'terlambat');
					const unpaid = numericStat(payload, 'denda_belum_lunas');
					nextAttention.library = overdue + unpaid;
				})
				.catch(() => {
					nextAttention.library = 0;
				})
		);
	}

	if (canSeePusakaAttention(roles)) {
		requests.push(
			fetchImpl('/api/pusaka/jobs/stats')
				.then(readStatsPayload)
				.then((payload) => {
					nextAttention.pusaka = numericStat(payload, 'failed');
				})
				.catch(() => {
					nextAttention.pusaka = 0;
				})
		);
	}

	await Promise.allSettled(requests);
	return nextAttention;
}
