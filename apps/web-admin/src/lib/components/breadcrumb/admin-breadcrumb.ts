import { findActiveSidebarHref } from '$lib/components/sidebar/sidebar-active';
import { filterSidebarNavGroupsByAccess } from '$lib/components/sidebar/sidebar-access';
import {
	dashboardNavItem,
	sidebarNavGroups,
	type SidebarFlatItem,
} from '$lib/components/sidebar/sidebar-config';
import { flattenSidebarNavGroups, numberSidebarNavGroups } from '$lib/components/sidebar/sidebar-tree';

export type AdminBreadcrumbCrumb = {
	label: string;
	href?: string;
	section?: string;
};

type BreadcrumbPattern = {
	match: RegExp;
	crumbs: AdminBreadcrumbCrumb[];
};

const segmentLabelMap: Record<string, string> = {
	archive: 'Arsip',
	cbt: 'CBT',
	cetak: 'Cetak',
	copies: 'Eksemplar',
	edit: 'Ubah',
	'exam-cards': 'Kartu Peserta',
	members: 'Peserta',
	minutes: 'Berita Acara',
	matrix: 'Matriks',
	new: 'Tambah',
	'pengawas-cards': 'Kartu Pengawas',
	print: 'Cetak',
	'print-pack': 'Paket Cetak',
	proctoring: 'Pengawasan',
	release: 'Rilis',
	report: 'Laporan',
	rooms: 'Ruang',
	scan: 'Pindai',
	soal: 'Soal'
};

const fallbackPatterns: BreadcrumbPattern[] = [
	{
		match: /^\/asesmen\/kegiatan\/[^/]+\/exam-cards\/?$/,
		crumbs: [
			{ label: 'Asesmen Ujian', section: '7' },
			{ label: 'Persiapan', href: '/asesmen/persiapan', section: '7.1.2' },
			{ label: 'Kegiatan', href: '/asesmen/kegiatan', section: '7.1.2.1' },
			{ label: 'Kartu Peserta', section: '7.1.2.1.1' }
		]
	},
	{
		match: /^\/asesmen\/kegiatan\/[^/]+\/(cetak|archive|members)\/?$/,
		crumbs: [
			{ label: 'Asesmen Ujian', section: '7' },
			{ label: 'Persiapan', href: '/asesmen/persiapan', section: '7.1.2' },
			{ label: 'Kegiatan', href: '/asesmen/kegiatan', section: '7.1.2.1' },
			{ label: 'Dokumen & Arsip', section: '7.1.2.1.2' }
		]
	},
	{
		match: /^\/asesmen\/kegiatan\/[^/]+\/?$/,
		crumbs: [
			{ label: 'Asesmen Ujian', section: '7' },
			{ label: 'Persiapan', href: '/asesmen/persiapan', section: '7.1.2' },
			{ label: 'Kegiatan', href: '/asesmen/kegiatan', section: '7.1.2.1' },
			{ label: 'Detail Kegiatan', section: '7.1.2.1.3' }
		]
	},
	{
		match: /^\/asesmen\/kegiatan(\/new)?\/?$/,
		crumbs: [
			{ label: 'Asesmen Ujian', section: '7' },
			{ label: 'Persiapan', href: '/asesmen/persiapan', section: '7.1.2' },
			{ label: 'Kegiatan', section: '7.1.2.1' }
		]
	},
	{
		match: /^\/asesmen\/paket\/[^/]+\/?$/,
		crumbs: [
			{ label: 'Asesmen Ujian', section: '7' },
			{ label: 'Persiapan', href: '/asesmen/persiapan', section: '7.1.2' },
			{ label: 'Paket Asesmen', href: '/asesmen/paket', section: '7.1.2.2' },
			{ label: 'Detail Paket', section: '7.1.2.2.1' }
		]
	},
	{
		match: /^\/asesmen\/paket(\/new)?\/?$/,
		crumbs: [
			{ label: 'Asesmen Ujian', section: '7' },
			{ label: 'Persiapan', href: '/asesmen/persiapan', section: '7.1.2' },
			{ label: 'Paket Asesmen', section: '7.1.2.2' }
		]
	},
	{
		match: /^\/asesmen\/sesi\/[^/]+\/rooms\/[^/]+\/proctoring(\/report)?\/?$/,
		crumbs: [
			{ label: 'Asesmen Ujian', section: '7' },
			{ label: 'Pelaksanaan', href: '/asesmen/pelaksanaan', section: '7.1.3' },
			{ label: 'Ruang Saya', href: '/asesmen/ruang-saya', section: '7.1.3.1' },
			{ label: 'Pengawasan Ruang', section: '7.1.3.1.1' }
		]
	},
	{
		match: /^\/asesmen\/sesi\/[^/]+\/(proctoring|minutes)(\/report)?\/?$/,
		crumbs: [
			{ label: 'Asesmen Ujian', section: '7' },
			{ label: 'Pelaksanaan', href: '/asesmen/pelaksanaan', section: '7.1.3' },
			{ label: 'Sesi', href: '/asesmen/sesi', section: '7.1.2.3' },
			{ label: 'Operasional Sesi', section: '7.1.3.2' }
		]
	},
	{
		match: /^\/asesmen\/sesi\/[^/]+\/?$/,
		crumbs: [
			{ label: 'Asesmen Ujian', section: '7' },
			{ label: 'Persiapan', href: '/asesmen/persiapan', section: '7.1.2' },
			{ label: 'Sesi & Ruang', href: '/asesmen/sesi', section: '7.1.2.3' },
			{ label: 'Detail Sesi', section: '7.1.2.3.1' }
		]
	},
	{
		match: /^\/asesmen\/sesi(\/new)?\/?$/,
		crumbs: [
			{ label: 'Asesmen Ujian', section: '7' },
			{ label: 'Persiapan', href: '/asesmen/persiapan', section: '7.1.2' },
			{ label: 'Sesi & Ruang', section: '7.1.2.3' }
		]
	},
	{
		match: /^\/asesmen\/ruang-saya\/?$/,
		crumbs: [
			{ label: 'Asesmen Ujian', section: '7' },
			{ label: 'Pelaksanaan', href: '/asesmen/pelaksanaan', section: '7.1.3' },
			{ label: 'Ruang Saya', section: '7.1.3.1' }
		]
	},
	{
		match: /^\/asesmen\/aplikasi-siswa(\/matrix|\/release)?\/?$/,
		crumbs: [
			{ label: 'Asesmen Ujian', section: '7' },
			{ label: 'Pelaksanaan', href: '/asesmen/pelaksanaan', section: '7.1.3' },
			{ label: 'Perangkat Siswa', section: '7.1.3.3' }
		]
	},
	{
		match: /^\/asesmen\/non-tes\/?$/,
		crumbs: [
			{ label: 'Asesmen Ujian', section: '7' },
			{ label: 'Penilaian Non-Tes', section: '7.2' }
		]
	},
	{
		match: /^\/bank-soal\/(analisis-butir|cetak|laporan|mapel-kd|pengaturan|alat|penerbitan|daftar)\/?$/,
		crumbs: [
			{ label: 'Bank Soal', section: '6' },
			{ label: 'Kelola Soal', href: '/bank-soal', section: '6.1' },
			{ label: 'Fitur Bank Soal', section: '6.1.5' }
		]
	}
];

const numberedSidebarNavGroups = numberSidebarNavGroups(sidebarNavGroups);
const numberedDashboardNavItem = {
	...dashboardNavItem,
	section: '0',
	numberedLabel: '0 Dashboard'
};

export function visibleBreadcrumbNavItems(roles: string[], permissions: string[]) {
	return [
		{ ...numberedDashboardNavItem, group: 'Akses Cepat', groupSection: '0', ancestors: [], ancestorSections: [], breadcrumb: ['Akses Cepat', dashboardNavItem.label] },
		...flattenSidebarNavGroups(filterSidebarNavGroupsByAccess(numberedSidebarNavGroups, roles, permissions))
	];
}

export function buildAdminBreadcrumbs(pathname: string, roles: string[], permissions: string[]): AdminBreadcrumbCrumb[] {
	const fallback = fallbackPatterns.find((pattern) => pattern.match.test(pathname));
	if (fallback) return normalizeCrumbs(fallback.crumbs);

	const items = visibleBreadcrumbNavItems(roles, permissions);
	const activeHref = findActiveSidebarHref(pathname, items);
	const activeItem = items.find((item) => item.href === activeHref);
	if (activeItem) return sidebarItemCrumbs(activeItem, pathname);

	const globalItems = allBreadcrumbNavItems();
	const globalActiveHref = findActiveSidebarHref(pathname, globalItems);
	const globalActiveItem = globalItems.find((item) => item.href === globalActiveHref);
	if (globalActiveItem) return sidebarItemCrumbs(globalActiveItem, pathname);

	return normalizeCrumbs(genericCrumbs(pathname));
}

export function compactAdminBreadcrumbs(crumbs: AdminBreadcrumbCrumb[]) {
	if (crumbs.length <= 3) return crumbs;
	return [crumbs[0], crumbs[crumbs.length - 2], crumbs[crumbs.length - 1]];
}

function sidebarItemCrumbs(item: SidebarFlatItem, pathname: string): AdminBreadcrumbCrumb[] {
	const labels = [item.group, ...(item.ancestors ?? []), item.label].filter(Boolean);
	const sections = [item.groupSection, ...(item.ancestorSections ?? []), item.section].filter(
		(section): section is string => typeof section === 'string' && section.length > 0
	);
	const crumbs: AdminBreadcrumbCrumb[] = labels.map((label, index) => ({ label, section: sections[index] }));

	const extraSegments = extraPathSegments(pathname, item.href);
	if (extraSegments.length > 0 && crumbs.length > 0) {
		let parentSection = crumbs[crumbs.length - 1]?.section;
		crumbs[crumbs.length - 1] = { label: item.label, href: item.href, section: parentSection };

		for (const segment of extraSegments) {
			parentSection = parentSection ? `${parentSection}.1` : undefined;
			crumbs.push({ label: titleizeSegment(segment), section: parentSection });
		}
	}

	return normalizeCrumbs(crumbs);
}

function allBreadcrumbNavItems() {
	return [
		{ ...numberedDashboardNavItem, group: 'Akses Cepat', groupSection: '0', ancestors: [], ancestorSections: [], breadcrumb: ['Akses Cepat', dashboardNavItem.label] },
		...flattenSidebarNavGroups(numberedSidebarNavGroups)
	];
}

function extraPathSegments(pathname: string, baseHref: string) {
	if (pathname === baseHref) return [];
	const suffix = pathname.slice(baseHref.length).replace(/^\/+|\/+$/g, '');
	return suffix.split('/').filter(Boolean);
}

function genericCrumbs(pathname: string): AdminBreadcrumbCrumb[] {
	const segments = pathname.split('/').filter(Boolean);
	if (segments.length === 0) return [{ label: 'Dashboard', section: '0' }];
	return [
		{ label: 'Dashboard', href: '/', section: '0' },
		...segments.map((segment, index) => ({ label: titleizeSegment(segment), section: `0.${index + 1}` }))
	];
}

function titleizeSegment(segment: string) {
	if (!segment) return 'Halaman';
	if (isLikelyId(segment)) return 'Detail';
	const mapped = segmentLabelMap[segment.toLowerCase()];
	if (mapped) return mapped;
	return segment
		.split('-')
		.filter(Boolean)
		.map((part) => part.charAt(0).toUpperCase() + part.slice(1))
		.join(' ');
}

function isLikelyId(segment: string) {
	return /^[0-9a-f]{8,}-/i.test(segment) || /^[0-9a-f]{24,}$/i.test(segment) || /^\d+$/.test(segment);
}

function normalizeCrumbs(crumbs: AdminBreadcrumbCrumb[]) {
	return crumbs
		.filter((crumb) => crumb.label.trim().length > 0)
		.map((crumb, index, all) => ({
			label: crumb.label.trim(),
			href: index === all.length - 1 ? undefined : crumb.href,
			section: crumb.section
		}));
}
