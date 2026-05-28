import { describe, expect, it } from 'vitest';
import {
	buildRoomAssignmentPayload,
	canSaveRoomAssignment,
	roomAssignmentModeLabel,
	roomAssignmentOptions,
	type RoomAssignmentPreview
} from './room-assignment';

describe('room assignment helpers', () => {
	it('builds guarded payloads for each mix policy', () => {
		expect(buildRoomAssignmentPayload('same_class')).toEqual({
			mix_policy: 'same_class',
			assignment_mode: 'random_balanced',
			allow_cross_grade: false,
			is_special_event: false
		});
		expect(buildRoomAssignmentPayload('same_grade')).toMatchObject({
			mix_policy: 'same_grade',
			allow_cross_grade: false,
			is_special_event: false
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

	it('labels backend mix policies in field-facing Indonesian copy', () => {
		expect(roomAssignmentModeLabel('same_class')).toBe('Per Kelas');
		expect(roomAssignmentModeLabel('same_grade')).toBe('Campur Satu Tingkat');
		expect(roomAssignmentModeLabel('mixed_scope')).toBe('Campur Lintas Tingkat');
		expect(roomAssignmentModeLabel('unexpected')).toBe('Campur Lintas Tingkat');
	});

	it('only allows save when every participant is placed', () => {
		const readyPreview: RoomAssignmentPreview = {
			summary: {
				participant_count: 32,
				room_count: 2,
				capacity_total: 40,
				assigned_count: 32,
				unassigned_count: 0,
				mix_policy: 'mixed_scope'
			},
			rooms: []
		};
		expect(canSaveRoomAssignment(null)).toBe(false);
		expect(canSaveRoomAssignment({ ...readyPreview, summary: { ...readyPreview.summary, unassigned_count: 1 } })).toBe(false);
		expect(canSaveRoomAssignment(readyPreview)).toBe(true);
	});
});
