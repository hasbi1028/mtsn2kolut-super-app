import { env } from '$env/dynamic/private';
import { error } from '@sveltejs/kit';
import { createReadStream } from 'node:fs';
import { stat } from 'node:fs/promises';
import path from 'node:path';
import { Readable } from 'node:stream';

const DEFAULT_RELEASE_DIR = '/home/servermtsn2kolut/releases/mtsn2kolut-mobile';
const NO_STORE = 'no-store, max-age=0';

export type MobileReleaseFileOptions = {
	contentType: string;
	downloadName?: string;
	inline?: boolean;
};

function releaseRoot() {
	return path.resolve(env.MOBILE_RELEASE_DIR?.trim() || DEFAULT_RELEASE_DIR);
}

function safeReleasePath(relativePath: string) {
	if (relativePath.startsWith('/') || relativePath.includes('\0')) {
		throw error(400, 'Invalid release path');
	}
	const root = releaseRoot();
	const fullPath = path.resolve(root, relativePath);
	if (fullPath !== root && !fullPath.startsWith(`${root}${path.sep}`)) {
		throw error(400, 'Invalid release path');
	}
	return fullPath;
}

export async function serveMobileReleaseFile(relativePath: string, options: MobileReleaseFileOptions) {
	const fullPath = safeReleasePath(relativePath);
	let fileStat;
	try {
		fileStat = await stat(fullPath);
	} catch {
		throw error(404, 'Mobile release artifact belum tersedia');
	}
	if (!fileStat.isFile()) {
		throw error(404, 'Mobile release artifact belum tersedia');
	}

	const headers = new Headers({
		'Content-Type': options.contentType,
		'Content-Length': String(fileStat.size),
		'Cache-Control': NO_STORE,
		'X-Content-Type-Options': 'nosniff'
	});

	if (options.downloadName) {
		const disposition = options.inline ? 'inline' : 'attachment';
		headers.set('Content-Disposition', `${disposition}; filename="${options.downloadName}"`);
	}

	const stream = Readable.toWeb(createReadStream(fullPath)) as ReadableStream<Uint8Array>;
	return new Response(stream, { headers });
}
