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

function canSeeInventoryAttention(roles: string[]) {
	return roles.includes('admin') || roles.includes('staf');
}

function canSeePusakaAttention(roles: string[]) {
	return roles.includes('admin');
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
				.then((res) => (res.ok ? res.json() : null))
				.then((payload) => {
					nextAttention.inventory = Number(payload?.data?.perlu_restok ?? payload?.perlu_restok ?? 0);
				})
				.catch(() => {
					nextAttention.inventory = 0;
				})
		);
		requests.push(
			fetchImpl('/api/library/stats')
				.then((res) => (res.ok ? res.json() : null))
				.then((payload) => {
					const overdue = Number(payload?.data?.terlambat ?? payload?.terlambat ?? 0);
					const unpaid = Number(payload?.data?.denda_belum_lunas ?? payload?.denda_belum_lunas ?? 0);
					nextAttention.library = overdue + unpaid;
				})
				.catch(() => {
					nextAttention.library = 0;
				})
		);
	}

	if (canSeePusakaAttention(roles)) {
		requests.push(
			fetchImpl('/api/queue/stats')
				.then((res) => (res.ok ? res.json() : null))
				.then((payload) => {
					nextAttention.pusaka = Number(payload?.failed ?? payload?.data?.failed ?? 0);
				})
				.catch(() => {
					nextAttention.pusaka = 0;
				})
		);
	}

	await Promise.allSettled(requests);
	return nextAttention;
}
