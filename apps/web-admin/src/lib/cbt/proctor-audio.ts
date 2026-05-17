export type ProctorAudioLevel = 'warning' | 'medium' | 'critical' | 'technical';

const throttleByKey = new Map<string, number>();
const globalSoundTimes: number[] = [];

const LEVEL_INTERVAL_MS: Record<ProctorAudioLevel, number> = {
	warning: 120_000,
	medium: 60_000,
	critical: 30_000,
	technical: 120_000,
};

const TONE_PLAN: Record<ProctorAudioLevel, Array<[number, number]>> = {
	warning: [[740, 0.16]],
	medium: [[680, 0.14], [880, 0.18]],
	critical: [[920, 0.12], [720, 0.12], [920, 0.18]],
	technical: [[520, 0.18], [440, 0.2]],
};

export function canPlayProctorAudio(): boolean {
	if (typeof window === 'undefined') return false;
	return Boolean(window.AudioContext || (window as unknown as { webkitAudioContext?: typeof AudioContext }).webkitAudioContext);
}

export function audioLabel(level: ProctorAudioLevel | string): string {
	const labels: Record<string, string> = {
		warning: 'Peringatan ringan',
		medium: 'Peringatan sedang',
		critical: 'Peringatan kritis',
		technical: 'Gangguan teknis',
	};
	return labels[level] ?? 'Peringatan pengawas';
}

export function shouldThrottleAudio(key: string, level: ProctorAudioLevel | string, now = Date.now()): boolean {
	const normalizedLevel = normalizeAudioLevel(level);
	const cleanKey = `${normalizedLevel}:${key.trim() || 'global'}`;
	const lastPlayed = throttleByKey.get(cleanKey);
	if (lastPlayed !== undefined && now - lastPlayed < LEVEL_INTERVAL_MS[normalizedLevel]) return true;

	const recentGlobal = globalSoundTimes.filter((timestamp) => now - timestamp < 60_000);
	globalSoundTimes.length = 0;
	globalSoundTimes.push(...recentGlobal);
	if (recentGlobal.length >= 5) return true;

	throttleByKey.set(cleanKey, now);
	globalSoundTimes.push(now);
	return false;
}

export async function playProctorTone(level: ProctorAudioLevel | string): Promise<void> {
	if (!canPlayProctorAudio()) {
		throw new Error('AudioContext tidak tersedia');
	}
	const Ctx = window.AudioContext || (window as unknown as { webkitAudioContext?: typeof AudioContext }).webkitAudioContext;
	if (!Ctx) throw new Error('AudioContext tidak tersedia');
	const ctx = new Ctx();
	if (ctx.state === 'suspended') await ctx.resume();
	let cursor = ctx.currentTime;
	for (const [frequency, duration] of TONE_PLAN[normalizeAudioLevel(level)]) {
		const oscillator = ctx.createOscillator();
		const gain = ctx.createGain();
		oscillator.type = 'sine';
		oscillator.frequency.setValueAtTime(frequency, cursor);
		gain.gain.setValueAtTime(0.0001, cursor);
		gain.gain.exponentialRampToValueAtTime(0.16, cursor + 0.02);
		gain.gain.exponentialRampToValueAtTime(0.0001, cursor + duration);
		oscillator.connect(gain).connect(ctx.destination);
		oscillator.start(cursor);
		oscillator.stop(cursor + duration + 0.02);
		cursor += duration + 0.06;
	}
	setTimeout(() => void ctx.close().catch(() => undefined), Math.ceil((cursor - ctx.currentTime + 0.2) * 1000));
}

export function clearProctorAudioThrottle(): void {
	throttleByKey.clear();
	globalSoundTimes.length = 0;
}

function normalizeAudioLevel(level: ProctorAudioLevel | string): ProctorAudioLevel {
	if (level === 'critical' || level === 'medium' || level === 'technical' || level === 'warning') return level;
	return 'warning';
}
