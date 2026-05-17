import { describe, expect, it, beforeEach } from 'vitest';
import { audioLabel, clearProctorAudioThrottle, shouldThrottleAudio } from './proctor-audio';

describe('proctor audio throttling', () => {
	beforeEach(() => clearProctorAudioThrottle());

	it('labels Indonesian alarm levels', () => {
		expect(audioLabel('critical')).toBe('Peringatan kritis');
		expect(audioLabel('technical')).toBe('Gangguan teknis');
	});

	it('throttles per alarm key and then allows after the level window', () => {
		expect(shouldThrottleAudio('participant-1:app_switch', 'warning', 1_000)).toBe(false);
		expect(shouldThrottleAudio('participant-1:app_switch', 'warning', 30_000)).toBe(true);
		expect(shouldThrottleAudio('participant-1:app_switch', 'warning', 122_000)).toBe(false);
	});

	it('caps room audio globally at five sounds per minute', () => {
		for (let i = 0; i < 5; i += 1) {
			expect(shouldThrottleAudio(`participant-${i}`, 'critical', 10_000 + i)).toBe(false);
		}
		expect(shouldThrottleAudio('participant-6', 'critical', 11_000)).toBe(true);
		expect(shouldThrottleAudio('participant-6', 'critical', 72_000)).toBe(false);
	});
});
