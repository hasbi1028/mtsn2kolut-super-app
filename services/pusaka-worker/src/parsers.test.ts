import test from 'node:test';
import assert from 'node:assert/strict';

import { extractJam, parseTodayFromText } from './parsers.js';

test('extractJam reads inline and trailing time values', () => {
	assert.equal(extractJam('Jam Masuk 07:12 WITA', 'Jam Masuk'), '07:12 WITA');
	assert.equal(extractJam('Jam Pulang\n-\nCatatan', 'Jam Pulang'), '-');
});

test('parseTodayFromText returns the latest block for the requested day', () => {
	const todayLabel = 'Jumat, 01 Mei 2026';
	const text = [
		'Kamis, 30 April 2026',
		'Jam Masuk 07:10 WITA',
		'Jumat, 01 Mei 2026',
		'Jam Masuk',
		'07:15 WITA',
		'Jam Pulang',
		'-',
		'Jumat, 01 Mei 2026',
		'Jam Masuk 07:18 WITA',
		'Jam Pulang 16:03 WITA'
	].join('\n');

	const parsed = parseTodayFromText(text, todayLabel);
	assert.ok(parsed);
	assert.equal(parsed?.jam_masuk, '07:18 WITA');
	assert.equal(parsed?.jam_pulang, '16:03 WITA');
});
