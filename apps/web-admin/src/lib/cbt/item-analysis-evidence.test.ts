import { describe, expect, it } from 'vitest';

import {
	CRONBACH_ALPHA_STATUS,
	formatItemAnalysisAccuracyLabel,
	itemAnalysisEmptyStateCopy,
	itemAnalysisMetricInventory,
	summarizeItemAnalysisEvidence
} from './item-analysis-evidence';

describe('CBT item-analysis evidence helpers', () => {
	it('summarizes unanswered count and per-type accuracy without fake psychometrics', () => {
		const summary = summarizeItemAnalysisEvidence([
			{
				question_type: 'multiple_choice',
				submitted_count: 10,
				answered_count: 9,
				blank_count: 1,
				correct_count: 6,
				incorrect_count: 3,
				unscored_count: 0,
				difficulty_index: 0.6,
				discrimination_index: 0.25
			},
			{
				question_type: 'essay',
				submitted_count: 10,
				answered_count: 8,
				blank_count: 2,
				correct_count: 0,
				incorrect_count: 0,
				unscored_count: 3,
				difficulty_index: 0.72,
				discrimination_index: 0.1
			}
		]);

		expect(summary.totalItems).toBe(2);
		expect(summary.unansweredCount).toBe(3);
		expect(summary.unscoredCount).toBe(3);
		expect(summary.perType).toEqual([
			{
				questionType: 'essay',
				itemCount: 1,
				submittedCount: 10,
				answeredCount: 8,
				unansweredCount: 2,
				correctCount: 0,
				accuracy: 0,
				averageDifficulty: 0.72,
				averageDiscrimination: 0.1
			},
			{
				questionType: 'multiple_choice',
				itemCount: 1,
				submittedCount: 10,
				answeredCount: 9,
				unansweredCount: 1,
				correctCount: 6,
				accuracy: 6 / 9,
				averageDifficulty: 0.6,
				averageDiscrimination: 0.25
			}
		]);
	});

	it('keeps Cronbach alpha documented as deferred unless a real implementation exists', () => {
		expect(CRONBACH_ALPHA_STATUS).toBe('documented_deferred_until_formula_and_dataset_are_tested');
		expect(itemAnalysisMetricInventory()).toContainEqual({
			metric: 'Cronbach alpha',
			status: 'deferred',
			guard: 'No value is emitted until a formula and fixed validation dataset are implemented and tested.'
		});
	});

	it('provides C3 operator-friendly accuracy labels and empty-state copy', () => {
		expect(
			formatItemAnalysisAccuracyLabel({ questionType: 'multiple_choice', accuracy: 0.625, answeredCount: 8, correctCount: 5 })
		).toBe('Pilihan ganda: 62.5% benar (5/8 terjawab benar)');
		expect(
			formatItemAnalysisAccuracyLabel({ questionType: 'essay', accuracy: 0, answeredCount: 0, correctCount: 0 })
		).toBe('Uraian: belum cukup respons untuk akurasi per tipe');
		expect(itemAnalysisEmptyStateCopy).toContain('Belum cukup respons');
		expect(itemAnalysisEmptyStateCopy).toContain('tidak membuat analitik palsu');
	});
});
