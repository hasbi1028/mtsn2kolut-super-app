#!/usr/bin/env node
import { readdirSync, readFileSync, statSync } from 'node:fs';
import { join, relative } from 'node:path';

const root = process.cwd();
const scanRoots = [
	'apps/web-admin/src/routes',
	'apps/web-admin/src/lib/components',
	'apps/web-admin/src/lib/client'
];

const patterns = [
	{ name: 'created_by', regex: /\bcreated_by\b/ },
	{ name: 'updated_by', regex: /\bupdated_by\b/ },
	{ name: 'deleted_by', regex: /\bdeleted_by\b/ },
	{ name: 'user_id', regex: /\buser_id\b/ },
	{ name: 'actor_id', regex: /\bactor_id\b/ },
	{ name: 'owner_id', regex: /\bowner_id\b/ },
	{ name: 'assignee_id', regex: /\bassignee_id\b/ },
	{ name: 'author_username', regex: /\bauthor_username\b/ },
	{ name: 'reviewer_username', regex: /\breviewer_username\b/ },
	{ name: 'approver_username', regex: /\bapprover_username\b/ },
	{ name: 'actor_username', regex: /\bactor_username\b/ },
	{ name: 'raw id interpolation', regex: /\{[^}\n]*(?:\.id|_id)[^}\n]*\}/ }
];

const ignoreFileParts = [
	'.svelte-kit/',
	'node_modules/',
	'.git/',
	'/ui/'
];

const allowedContext = [
	'ID internal',
	'Copy ID',
	'copy id',
	'debug',
	'Debug',
	'aria-describedby',
	'for=',
	'id=',
	'bind:this'
];

function walk(dir, files = []) {
	let entries = [];
	try {
		entries = readdirSync(dir);
	} catch {
		return files;
	}
	for (const entry of entries) {
		const path = join(dir, entry);
		const rel = relative(root, path);
		if (ignoreFileParts.some((part) => rel.includes(part))) continue;
		const st = statSync(path);
		if (st.isDirectory()) walk(path, files);
		else if (/\.(svelte|ts)$/.test(entry)) files.push(path);
	}
	return files;
}

const findings = [];
for (const scanRoot of scanRoots) {
	for (const file of walk(join(root, scanRoot))) {
		const rel = relative(root, file);
		const lines = readFileSync(file, 'utf8').split('\n');
		lines.forEach((line, index) => {
			if (allowedContext.some((token) => line.includes(token))) return;
			for (const pattern of patterns) {
				if (pattern.regex.test(line)) {
					findings.push({ file: rel, line: index + 1, pattern: pattern.name, text: line.trim().slice(0, 220) });
				}
			}
		});
	}
}

if (findings.length === 0) {
	console.log('PASS: tidak ada kandidat tampilan ID/username mentah di UI scan roots.');
	process.exit(0);
}

console.log(`Found ${findings.length} kandidat tampilan ID/username mentah. Triage: UI utama harus pakai displayName(), ID hanya boleh di debug/detail teknis.`);
for (const item of findings.slice(0, 250)) {
	console.log(`${item.file}:${item.line} [${item.pattern}] ${item.text}`);
}
if (findings.length > 250) {
	console.log(`... ${findings.length - 250} temuan lain disembunyikan. Persempit scan atau patch bertahap.`);
}

process.exit(0);
