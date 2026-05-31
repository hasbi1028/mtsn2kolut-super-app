import { describe, expect, it } from 'vitest';
import {
	buildPackageReadinessCsv,
	filterPackagesByReadiness,
	packageReadinessBadge
} from './package-readiness';
import type { PackageReadinessListItem, PackageReadinessFilter } from './package-readiness';

const basePackage = {
	id: 'pkg-base',
	subject_id: 'subject-1',
	subject_name: 'IPA',
	title: 'Paket IPA',
	duration_minutes: 90,
	question_count: 25,
	session_count: 0,
	locked: false
};

function item(overrides: Partial<PackageReadinessListItem>): PackageReadinessListItem {
	return {
		...basePackage,
		...overrides,
		readiness: {
			status: 'siap',
			target_pg_count: 20,
			target_essay_count: 5,
			missing_pg_count: 0,
			missing_essay_count: 0,
			question_count: 25,
			pg_count: 20,
			essay_count: 5,
			total_points: 25,
			published_count: 25,
			unpublished_count: 0,
			metadata_gap_count: 0,
			session_count: 0,
			locked: false,
			ready: true,
			...(overrides.readiness ?? {})
		}
	};
}

describe('package readiness list helpers', () => {
	const packages = [
		item({ id: 'ready', title: 'Siap' }),
		item({
			id: 'kurang-pg',
			title: 'Kurang PG',
			readiness: { status: 'kurang', ready: false, missing_pg_count: 2, pg_count: 18 }
		}),
		item({
			id: 'kurang-essay',
			title: 'Kurang Essay',
			readiness: { status: 'kurang', ready: false, missing_essay_count: 1, essay_count: 4 }
		}),
		item({
			id: 'metadata-gap',
			title: 'Metadata Gap',
			readiness: { status: 'kurang', ready: false, metadata_gap_count: 3 }
		}),
		item({
			id: 'locked',
			title: 'Terkunci',
			locked: true,
			readiness: { status: 'locked', locked: true, ready: true }
		})
	];

	it.each<[PackageReadinessFilter, string[]]>([
		['ready', ['ready']],
		['kurang_pg', ['kurang-pg']],
		['kurang_essay', ['kurang-essay']],
		['metadata_gap', ['metadata-gap']],
		['locked', ['locked']]
	])('filters %s packages', (filter, expectedIds) => {
		expect(filterPackagesByReadiness(packages, filter).map((pkg) => pkg.id)).toEqual(expectedIds);
	});

	it('labels locked packages before ready state', () => {
		expect(packageReadinessBadge(packages[4]).label).toBe('Terkunci');
	});

	it('exports package readiness as operator-friendly CSV', () => {
		const csv = buildPackageReadinessCsv(packages.slice(0, 2));
		expect(csv.split('\n')[0]).toBe('No,Mapel,Judul Paket,Status,PG,Essay,Gap Metadata,Belum Terbit,Poin,Pemakaian');
		expect(csv).toContain('1,IPA,Siap,Siap dikunci,20/20,5/5,0,0,25,0 sesi');
		expect(csv).toContain('2,IPA,Kurang PG,Kurang PG,18/20,5/5,0,0,25,0 sesi');
	});
});
