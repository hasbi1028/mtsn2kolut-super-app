import { error, type RequestEvent } from '@sveltejs/kit';
import { createReadStream } from 'node:fs';
import { readdir, readFile, stat } from 'node:fs/promises';
import path from 'node:path';
import { Readable } from 'node:stream';

const CBT_RELEASE_DIR = path.resolve(process.cwd(), '../../releases/mobile/cbt');
const NO_STORE = 'no-store, max-age=0';

const CBT_ARTIFACTS = [
	{
		fileName: 'cbt-flutter-universal-release.apk',
		label: 'Arsip APK Universal Release (nonaktif)',
		description: 'Artifact Flutter tahap lanjutan/nonaktif; cocok untuk arm64, armeabi-v7a, dan x86_64; bukan jalur resmi ujian tahun ini.',
		variant: 'release',
		abi: 'universal',
		recommendedFor: 'Arsip/verifikasi atau uji teknis internal; siswa diarahkan ke Portal Ujian Web.'
	},
	{
		fileName: 'cbt-flutter-arm64-v8a-release.apk',
		label: 'Arsip APK arm64-v8a Release (nonaktif)',
		description: 'Artifact Flutter tahap lanjutan/nonaktif untuk mayoritas HP Android 64-bit modern; bukan jalur resmi ujian tahun ini.',
		variant: 'release',
		abi: 'arm64-v8a',
		recommendedFor: 'Arsip/verifikasi atau uji teknis internal pada HP Android 64-bit.'
	},
	{
		fileName: 'cbt-flutter-armeabi-v7a-release.apk',
		label: 'Arsip APK armeabi-v7a Release (nonaktif)',
		description: 'Artifact Flutter tahap lanjutan/nonaktif untuk sebagian HP lama 32-bit yang tidak bisa memasang arm64.',
		variant: 'release',
		abi: 'armeabi-v7a',
		recommendedFor: 'Arsip/verifikasi atau uji teknis internal pada HP lama 32-bit.'
	},
	{
		fileName: 'cbt-flutter-demo-debug.apk',
		label: 'Arsip APK DEMO Debug (nonaktif)',
		description: 'Build khusus uji manual cepat dengan tombol Masuk Mode DEMO; tidak untuk ujian resmi.',
		variant: 'demo-debug',
		abi: 'universal',
		recommendedFor: 'Demo teknis tahap lanjutan; operasi tahun ini memakai Portal Ujian Web.'
	},
	{
		fileName: 'SHA256SUMS.txt',
		label: 'SHA256SUMS',
		description: 'Daftar checksum untuk memverifikasi keutuhan file APK.',
		variant: 'checksum',
		abi: 'text',
		recommendedFor: 'Verifikasi file setelah upload/distribusi.'
	}
] as const;

type CbtArtifactDefinition = (typeof CBT_ARTIFACTS)[number];

function artifactPath(fileName: string) {
	if (!CBT_ARTIFACTS.some((artifact) => artifact.fileName === fileName)) {
		throw error(404, 'Artifact CBT tidak dikenal');
	}
	return path.join(CBT_RELEASE_DIR, fileName);
}

function contentType(fileName: string) {
	if (fileName.endsWith('.apk')) return 'application/vnd.android.package-archive';
	if (fileName.endsWith('.txt')) return 'text/plain; charset=utf-8';
	return 'application/octet-stream';
}

async function readChecksums() {
	const checksums = new Map<string, string>();
	try {
		const text = await readFile(path.join(CBT_RELEASE_DIR, 'SHA256SUMS.txt'), 'utf8');
		for (const line of text.split('\n')) {
			const trimmed = line.trim();
			if (!trimmed) continue;
			const [sha256, filePath] = trimmed.split(/\s+/, 2);
			const fileName = path.basename(filePath ?? '');
			if (sha256 && fileName) checksums.set(fileName, sha256);
		}
	} catch {
		// Missing checksum file should not hide downloadable artifacts.
	}
	return checksums;
}

export async function listCbtMobileArtifacts(event: RequestEvent) {
	let dirEntries: string[] = [];
	try {
		dirEntries = await readdir(CBT_RELEASE_DIR);
	} catch {
		throw error(404, 'Artifact APK CBT belum tersedia di server');
	}
	const available = new Set(dirEntries);
	const checksums = await readChecksums();
	const baseUrl = event.url.origin;
	const artifacts = [];

	for (const definition of CBT_ARTIFACTS) {
		if (!available.has(definition.fileName)) continue;
		const fullPath = artifactPath(definition.fileName);
		const fileStat = await stat(fullPath);
		if (!fileStat.isFile()) continue;
		artifacts.push({
			...definition,
			sizeBytes: fileStat.size,
			updatedAt: fileStat.mtime.toISOString(),
			sha256: checksums.get(definition.fileName) ?? '',
			downloadUrl: `/releases/mobile/cbt/${encodeURIComponent(definition.fileName)}`,
			absoluteDownloadUrl: `${baseUrl}/releases/mobile/cbt/${encodeURIComponent(definition.fileName)}`
		});
	}

	return {
		title: 'Arsip Download APK CBT (nonaktif)',
		primaryClient: 'Portal Ujian Web /ujian',
		fallbackClient: 'Flutter APK arsip/tahap lanjutan',
		demoWebUrl: '/ujian?demo=1',
		artifactDir: CBT_RELEASE_DIR,
		generatedAt: new Date().toISOString(),
		artifacts
	};
}

export async function serveCbtMobileArtifact(fileName: string) {
	const definition = CBT_ARTIFACTS.find((artifact) => artifact.fileName === fileName) as CbtArtifactDefinition | undefined;
	if (!definition) throw error(404, 'Artifact CBT tidak dikenal');
	const fullPath = artifactPath(fileName);
	let fileStat;
	try {
		fileStat = await stat(fullPath);
	} catch {
		throw error(404, 'Artifact APK CBT belum tersedia');
	}
	if (!fileStat.isFile()) throw error(404, 'Artifact APK CBT belum tersedia');

	const headers = new Headers({
		'Content-Type': contentType(fileName),
		'Content-Length': String(fileStat.size),
		'Cache-Control': NO_STORE,
		'X-Content-Type-Options': 'nosniff',
		'Content-Disposition': `attachment; filename="${definition.fileName}"`
	});
	const stream = Readable.toWeb(createReadStream(fullPath)) as ReadableStream<Uint8Array>;
	return new Response(stream, { headers });
}
