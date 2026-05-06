import { existsSync, readdirSync, readFileSync, statSync } from 'node:fs';
import path from 'node:path';
import { describe, expect, it } from 'vitest';

const SRC_ROOT = path.resolve(process.cwd(), 'src');

function toPosix(value: string): string {
	return value.split(path.sep).join('/');
}

function isSourceFile(filePath: string): boolean {
	return filePath.endsWith('.svelte') || filePath.endsWith('.ts');
}

function listSourceFiles(dir: string): string[] {
	const entries = readdirSync(dir);
	return entries.flatMap((entry) => {
		const filePath = path.join(dir, entry);
		const stat = statSync(filePath);
		if (stat.isDirectory()) return listSourceFiles(filePath);
		return isSourceFile(filePath) ? [filePath] : [];
	});
}

describe('active frontend API aliases', () => {
	it('keeps source fetches off removed /api/cbt BFF routes', () => {
		const offenders = listSourceFiles(SRC_ROOT)
			.filter((filePath) => !filePath.endsWith('.test.ts'))
			.flatMap((filePath) => {
				const rel = toPosix(path.relative(SRC_ROOT, filePath));
				return readFileSync(filePath, 'utf8')
					.split('\n')
					.flatMap((line, index) => (line.includes('/api/cbt') ? [`${rel}:${index + 1}`] : []));
			});

		expect(offenders).toEqual([]);
	});

	it('keeps removed route trees absent from web-admin source', () => {
		const removedRoutes = [
			'routes/cbt',
			'routes/api/cbt',
			'routes/bank-soal/komposer',
			'routes/bank-soal/import',
			'routes/bank-soal/review'
		].map((routePath) => path.join(SRC_ROOT, routePath));

		expect(removedRoutes.filter((routePath) => existsSync(routePath)).map((routePath) => toPosix(path.relative(SRC_ROOT, routePath)))).toEqual([]);
	});
});
