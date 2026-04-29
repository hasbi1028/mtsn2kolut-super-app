export type QuestionVariantId = 'studio' | 'wizard' | 'grid' | 'document' | 'review' | 'package-fit';

export type QuestionVariantConfig = {
	id: QuestionVariantId;
	title: string;
	shortTitle: string;
	description: string;
	persona: string;
	highlights: string[];
	cautions: string[];
};

export type QuestionVariantEvaluation = {
	ease: number;
	speed: number;
	review: number;
	fit: number;
	note: string;
	updatedAt: string;
};

export const QUESTION_VARIANTS: QuestionVariantConfig[] = [
	{
		id: 'studio',
		title: 'Studio',
		shortTitle: 'Studio',
		description: 'Pendekatan baseline yang paling dekat dengan halaman bank soal sekarang: katalog, detail, dan form lengkap dalam satu workspace.',
		persona: 'Guru yang ingin semua alat ada di satu tempat.',
		highlights: ['Paling lengkap', 'Mudah dibandingkan dengan sistem saat ini', 'Cocok untuk admin kurikulum'],
		cautions: ['Bisa terasa padat untuk guru baru'],
	},
	{
		id: 'wizard',
		title: 'Wizard',
		shortTitle: 'Wizard',
		description: 'Alur bertahap yang memecah authoring menjadi langkah sederhana agar guru baru tidak langsung dibanjiri metadata.',
		persona: 'Guru beginner yang ingin dipandu langkah demi langkah.',
		highlights: ['Beban kognitif ringan', 'Beginner lebih terarah', 'Enak untuk membuat soal dari nol'],
		cautions: ['Lebih lambat untuk editing cepat'],
	},
	{
		id: 'grid',
		title: 'Grid',
		shortTitle: 'Grid',
		description: 'Pendekatan batch-friendly dengan katalog dan input yang terasa lebih kompak dan operasional.',
		persona: 'Guru yang ingin input banyak soal dengan ritme cepat.',
		highlights: ['Cepat untuk scanning katalog', 'Nuansa operasional', 'Bagus untuk kerja repetitif'],
		cautions: ['Kurang naratif untuk authoring panjang'],
	},
	{
		id: 'document',
		title: 'Document',
		shortTitle: 'Document',
		description: 'Editor-first. Fokus ke pengalaman menulis soal seperti dokumen, lalu memeriksa preview peserta di sampingnya.',
		persona: 'Guru yang berpikir dalam bentuk narasi, bacaan, dan stimulus.',
		highlights: ['Nyaman untuk soal bacaan', 'Preview terasa kuat', 'Cocok untuk rich text'],
		cautions: ['Daftar katalog terasa sekunder'],
	},
	{
		id: 'review',
		title: 'Review',
		shortTitle: 'Review',
		description: 'Workflow-first. Menonjolkan readiness, status review, dan aksi kurasi agar mudah dipakai reviewer atau admin mapel.',
		persona: 'Admin/guru senior yang fokus ke quality control soal.',
		highlights: ['Antrian review jelas', 'Aksi approve/publish dekat', 'Mudahkan kurasi'],
		cautions: ['Kurang santai untuk penulisan awal'],
	},
	{
		id: 'package-fit',
		title: 'Package Fit',
		shortTitle: 'Package Fit',
		description: 'Menulis soal sambil melihat kecocokan dengan paket, tingkat, dan blueprint ujian yang akan dibangun nanti.',
		persona: 'Guru yang menulis soal sambil membayangkan paket ujian target.',
		highlights: ['Konteks paket terasa', 'Bagus untuk perencanaan UTS/UAS', 'Mendorong metadata cukup'],
		cautions: ['Perlu disiplin isi tingkat dan topik'],
	},
];

export const QUESTION_VARIANT_MAP = Object.fromEntries(QUESTION_VARIANTS.map((variant) => [variant.id, variant])) as Record<QuestionVariantId, QuestionVariantConfig>;

export function evaluationStorageKey(variant: QuestionVariantId) {
	return `mtsn2-question-eval:${variant}`;
}
