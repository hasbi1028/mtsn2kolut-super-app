export const DEFAULT_PAGE_SIZE_OPTIONS = [12, 24, 48, 96] as const;

export type PaginationReason = 'page' | 'limit' | 'first' | 'previous' | 'next' | 'last';

export type PaginationChange = {
	page: number;
	limit: number;
	offset: number;
	reason: PaginationReason;
};

export type PaginationPageItem = number | 'ellipsis';

export function normalizePage(value: string | number | null | undefined, defaultPage = 1): number {
	const parsed = typeof value === 'number' ? value : Number.parseInt(value ?? '', 10);
	if (!Number.isFinite(parsed) || parsed < 1) return defaultPage;
	return Math.floor(parsed);
}

export function normalizePageSize(
	value: string | number | null | undefined,
	allowed: readonly number[] = DEFAULT_PAGE_SIZE_OPTIONS,
	defaultLimit = allowed[0] ?? 12
): number {
	const parsed = typeof value === 'number' ? value : Number.parseInt(value ?? '', 10);
	if (!Number.isFinite(parsed)) return defaultLimit;
	const size = Math.floor(parsed);
	return allowed.includes(size) ? size : defaultLimit;
}

export function clampPage(page: number, total: number, limit: number): number {
	const safeLimit = Math.max(1, limit);
	const pageCount = Math.max(1, Math.ceil(Math.max(0, total) / safeLimit));
	return Math.min(Math.max(1, Math.floor(page)), pageCount);
}

export function offsetForPage(page: number, limit: number): number {
	return Math.max(0, (Math.max(1, Math.floor(page)) - 1) * Math.max(1, Math.floor(limit)));
}

export function calculatePaginationRange(total: number, page: number, limit: number) {
	const safeTotal = Math.max(0, total);
	// If limit is 0, treat as "show all" - single page with all items
	const pageSize = limit === 0 ? safeTotal || 1 : Math.max(1, limit);
	const pageCount = Math.max(1, Math.ceil(safeTotal / pageSize));
	const safePage = Math.min(Math.max(1, Math.floor(page)), pageCount);
	const offset = offsetForPage(safePage, pageSize);
	return {
		page: safePage,
		limit: pageSize,
		offset,
		pageCount,
		start: safeTotal === 0 ? 0 : offset + 1,
		end: Math.min(safeTotal, safePage * pageSize),
		hasPrevious: safePage > 1,
		hasNext: safePage < pageCount
	};
}

export function paginateItems<T>(items: readonly T[], page: number, limit: number): T[] {
	const range = calculatePaginationRange(items.length, page, limit);
	return items.slice(range.offset, range.offset + range.limit);
}

export function buildPageItems(page: number, pageCount: number, siblingCount = 1): PaginationPageItem[] {
	const totalPages = Math.max(1, Math.floor(pageCount));
	const current = Math.min(Math.max(1, Math.floor(page)), totalPages);
	const siblings = Math.max(0, Math.floor(siblingCount));

	if (totalPages <= 7 + siblings * 2) {
		return Array.from({ length: totalPages }, (_, index) => index + 1);
	}

	const pages = new Set<number>();
	pages.add(1);
	pages.add(totalPages);
	for (let candidate = current - siblings; candidate <= current + siblings; candidate += 1) {
		if (candidate > 1 && candidate < totalPages) pages.add(candidate);
	}
	if (current <= 3 + siblings) {
		for (let candidate = 2; candidate <= Math.min(totalPages - 1, 4 + siblings * 2); candidate += 1) {
			pages.add(candidate);
		}
	}
	if (current >= totalPages - (2 + siblings)) {
		for (let candidate = Math.max(2, totalPages - (3 + siblings * 2)); candidate < totalPages; candidate += 1) {
			pages.add(candidate);
		}
	}

	const sorted = [...pages].sort((a, b) => a - b);
	const items: PaginationPageItem[] = [];
	let previous = 0;
	for (const pageNumber of sorted) {
		if (previous && pageNumber - previous > 1) items.push('ellipsis');
		items.push(pageNumber);
		previous = pageNumber;
	}
	return items;
}
