import { describe, expect, it } from 'vitest';
import {
	countRunningSessions,
	countUnassignedParticipants,
	latestEventDocumentHubHref,
	summarizeWorkflowReadiness,
	workflowReadinessClass
} from './workflow-hub';

describe('asesmen workflow hub helpers', () => {
	it('uses the latest event document center when a session has event context', () => {
		expect(latestEventDocumentHubHref([{ event_id: '' }, { event_id: 'event-123' }])).toBe('/asesmen/kegiatan/event-123/cetak');
	});

	it('falls back to the kegiatan list when no event context exists', () => {
		expect(latestEventDocumentHubHref([{ status: 'scheduled' }])).toBe('/asesmen/kegiatan');
	});

	it('summarizes workflow readiness in operator priority order', () => {
		expect(summarizeWorkflowReadiness([])).toEqual({ label: 'Belum ada sesi', tone: 'neutral' });
		expect(summarizeWorkflowReadiness([{ unassigned_participant_count: 2, status: 'active' }])).toEqual({
			label: 'Perlu penempatan peserta',
			tone: 'warning'
		});
		expect(summarizeWorkflowReadiness([{ status: 'active' }])).toEqual({ label: 'Sedang berjalan', tone: 'success' });
		expect(summarizeWorkflowReadiness([{ scheduled_start: '2026-05-28T00:30:00Z' }], new Date('2026-05-28T02:00:00Z'))).toEqual({
			label: 'Siap hari ini',
			tone: 'neutral'
		});
	});

	it('counts session conditions used by the compact ringkasan UI', () => {
		const sessions = [
			{ status: 'active', unassigned_participant_count: 1 },
			{ session_status: 'active', unassigned_participant_count: 3 },
			{ status: 'scheduled' }
		];
		expect(countRunningSessions(sessions)).toBe(2);
		expect(countUnassignedParticipants(sessions)).toBe(4);
	});

	it('maps readiness tones to the existing compact badge classes', () => {
		expect(workflowReadinessClass('warning')).toContain('warning');
		expect(workflowReadinessClass('success')).toContain('success');
		expect(workflowReadinessClass('neutral')).toContain('muted');
	});
});
