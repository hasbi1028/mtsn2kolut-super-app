export const CRONBACH_ALPHA_STATUS = 'documented_deferred_until_formula_and_dataset_are_tested';

export type ItemAnalysisEvidenceRow = {
	question_type: string;
	submitted_count: number;
	answered_count: number;
	blank_count: number;
	correct_count: number;
	incorrect_count: number;
	unscored_count: number;
	difficulty_index: number;
	discrimination_index: number;
};

export type ItemAnalysisPerTypeSummary = {
	questionType: string;
	itemCount: number;
	submittedCount: number;
	answeredCount: number;
	unansweredCount: number;
	correctCount: number;
	accuracy: number;
	averageDifficulty: number;
	averageDiscrimination: number;
};

export type ItemAnalysisEvidenceSummary = {
	totalItems: number;
	unansweredCount: number;
	unscoredCount: number;
	perType: ItemAnalysisPerTypeSummary[];
};

export function summarizeItemAnalysisEvidence(rows: readonly ItemAnalysisEvidenceRow[]): ItemAnalysisEvidenceSummary {
	const buckets = new Map<string, ItemAnalysisPerTypeSummary & { difficultyTotal: number; discriminationTotal: number }>();
	let unansweredCount = 0;
	let unscoredCount = 0;

	for (const row of rows) {
		const questionType = row.question_type || 'unknown';
		const unanswered = Math.max(row.blank_count, row.submitted_count - row.answered_count, 0);
		unansweredCount += unanswered;
		unscoredCount += row.unscored_count;

		const current = buckets.get(questionType) ?? {
			questionType,
			itemCount: 0,
			submittedCount: 0,
			answeredCount: 0,
			unansweredCount: 0,
			correctCount: 0,
			accuracy: 0,
			averageDifficulty: 0,
			averageDiscrimination: 0,
			difficultyTotal: 0,
			discriminationTotal: 0
		};
		current.itemCount += 1;
		current.submittedCount += row.submitted_count;
		current.answeredCount += row.answered_count;
		current.unansweredCount += unanswered;
		current.correctCount += row.correct_count;
		current.difficultyTotal += row.difficulty_index;
		current.discriminationTotal += row.discrimination_index;
		buckets.set(questionType, current);
	}

	const perType = [...buckets.values()]
		.sort((a, b) => a.questionType.localeCompare(b.questionType, 'id-ID'))
		.map(({ difficultyTotal, discriminationTotal, ...bucket }) => ({
			...bucket,
			accuracy: bucket.answeredCount > 0 ? bucket.correctCount / bucket.answeredCount : 0,
			averageDifficulty: bucket.itemCount > 0 ? difficultyTotal / bucket.itemCount : 0,
			averageDiscrimination: bucket.itemCount > 0 ? discriminationTotal / bucket.itemCount : 0
		}));

	return {
		totalItems: rows.length,
		unansweredCount,
		unscoredCount,
		perType
	};
}

export function itemAnalysisMetricInventory() {
	return [
		{
			metric: 'Difficulty index',
			status: 'implemented',
			guard: 'Uses submitted participants; objective items use correct/submitted and essay uses average manual score.'
		},
		{
			metric: 'Discrimination index',
			status: 'implemented_with_sample_guard',
			guard: 'Uses top/bottom tercile groups when submitted samples exist; low sample interpretation remains operational guidance.'
		},
		{
			metric: 'Distractor / answer distribution',
			status: 'implemented',
			guard: 'Reports observed answer counts without inferring untested psychometric quality.'
		},
		{
			metric: 'Unanswered count',
			status: 'implemented',
			guard: 'Uses blank count and submitted-vs-answered gap.'
		},
		{
			metric: 'Per-type accuracy',
			status: 'implemented_helper',
			guard: 'Aggregates existing item-analysis rows by question type.'
		},
		{
			metric: 'Cronbach alpha',
			status: 'deferred',
			guard: 'No value is emitted until a formula and fixed validation dataset are implemented and tested.'
		}
	] as const;
}
