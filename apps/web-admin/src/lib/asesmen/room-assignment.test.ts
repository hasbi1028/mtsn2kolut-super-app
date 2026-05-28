import { describe, expect, it } from 'vitest';
import {
	buildRoomAssignmentPayload,
	canSaveRoomAssignment,
	roomAssignmentModeLabel,
	roomAssignmentModeOptions,
	roomAssignmentOptions,
	type RoomAssignmentPreview
} from './room-assignment';

describe('room assignment helpers', () => {
	it('builds guarded payloads for automatic and manual modes', () => {
		expect(buildRoomAssignmentPayload('same_class')).toEqual({
			mix_policy: 'same_class',
			assignment_mode: 'random_balanced',
			allow_cross_grade: false,
			is_special_event: false,
			assignments: undefined
		});
		expect(buildRoomAssignmentPayload('same_grade', 'manual', [
			{ participant_id: 'p1', room_id: 'r1', seat_no: 1 }
		])).toMatchObject({
			mix_policy: 'same_grade',
			assignment_mode: 'manual',
			allow_cross_grade: false,
			is_special_event: false,
			assignments: [{ participant_id: 'p1', room_id: 'r1', seat_no: 1 }]
		});
		expect(buildRoomAssignmentPayload('mixed_scope')).toMatchObject({
			mix_policy: 'mixed_scope',
			allow_cross_grade: true,
			is_special_event: true
		});
	});

	it('keeps the three operator-facing policy options stable', () => {
		expect(roomAssignmentOptions.map((option) => option.value)).toEqual(['same_class', 'same_grade', 'mixed_scope']);
		expect(roomAssignmentOptions.every((option) => option.title.length > 0 && option.desc.length > 0)).toBe(true);
	});

	it('exposes automatic and manual modes with readable labels', () => {
		expect(roomAssignmentModeOptions.map((option) => option.value)).toEqual(['random_balanced', 'manual']);
		expect(roomAssignmentModeLabel('manual')).toBe('Manual per peserta');
		expect(roomAssignmentModeLabel('random_balanced')).toBe('Campur per Siswa Lintas Tingkat');
	});

	it('only allows save when every participant is placed', () => {
		const readyPreview: RoomAssignmentPreview = {
			summary: {
				participant_count: 32,
				room_count: 2,
				capacity_total: 40,
				assigned_count: 32,
				unassigned_count: 0,
				mix_policy: 'mixed_scope',
				assignment_mode: 'random_balanced'
			},
			rooms: []
		};
		expect(canSaveRoomAssignment(null)).toBe(false);
		expect(canSaveRoomAssignment({ ...readyPreview, summary: { ...readyPreview.summary, unassigned_count: 1 } })).toBe(false);
		expect(canSaveRoomAssignment(readyPreview)).toBe(true);
	});
});
