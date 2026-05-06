<script lang="ts">
	import { onMount } from 'svelte';
	import { resolve } from '$app/paths';
	import AsyncContent from '$lib/components/AsyncContent.svelte';
	import RecoveryPanel from '$lib/components/RecoveryPanel.svelte';
	import { Skeleton } from '$lib/components/ui/skeleton';
	import { readClientApiData } from '$lib/client/api';

	type SummaryResponse = {
		counts?: Partial<Record<'total' | 'draft' | 'review' | 'revision' | 'approved' | 'published' | 'package_usage', number>>;
		by_subject?: Array<{ subject_name?: string; subject_code?: string; total?: number }>;
		by_cognitive_level?: Array<{ cognitive_level?: string; total?: number }>;
	};

	type SubjectPayload = {
		subjects?: Array<{ id: string; name: string; code?: string }>;
	};

	type Payload = {
		summary: SummaryResponse;
		subjects: SubjectPayload['subjects'];
	};

	type SopItem = {
		title: string;
		desc: string;
		owner: string;
		status: string;
	};

	const workflowSteps = [
		{ label: 'Draft', desc: 'Guru menyusun metadata, naskah, opsi/kunci, dan pembahasan sebelum diajukan.', tone: 'slate' },
		{ label: 'Review', desc: 'Reviewer memeriksa substansi, konstruksi, bahasa, kunci/rubrik, dan kesesuaian KD/CP/TP.', tone: 'amber' },
		{ label: 'Revisi', desc: 'Soal dikembalikan jika perlu perbaikan. Catatan reviewer wajib jelas dan bisa ditindaklanjuti.', tone: 'rose' },
		{ label: 'Approved', desc: 'Soal lolos review dan siap dipakai untuk paket asesmen internal.', tone: 'emerald' },
		{ label: 'Published', desc: 'Soal tersedia untuk pemakaian paket dan menjadi bagian repositori reusable.', tone: 'green' }
	];

	const qualityRules = [
		'Isi metadata mapel, kelas/fase, KD/CP/TP, materi, level kognitif, dan kesulitan sebelum review.',
		'Naskah soal wajib jelas, bebas ambigu, dan tidak bergantung pada informasi di luar stimulus.',
		'Soal pilihan wajib memiliki kunci benar; essay/isian wajib memiliki rubrik atau jawaban acuan.',
		'Pembahasan dianjurkan untuk semua tipe soal agar bank soal bisa dipakai ulang untuk remedial/pengayaan.',
		'Gunakan flag HOTS hanya jika soal menuntut analisis, evaluasi, atau kreasi; bukan sekadar narasi panjang.',
		'Soal yang sudah dipakai paket/jawaban tidak diedit sembarang; lakukan duplikasi/revisi versi bila perlu perubahan besar.'
	];

	const sopItems: SopItem[] = [
		{ title: 'Import massal', desc: 'Gunakan template resmi, jalankan preview dry-run, validasi error, baru lakukan import final.', owner: 'Admin/Guru', status: 'Aktif' },
		{ title: 'Review berkala', desc: 'Prioritaskan antrean review dan soal revisi sebelum periode asesmen aktif.', owner: 'Reviewer/Admin', status: 'Aktif' },
		{ title: 'Coverage mapel', desc: 'Pantau Mapel & KD untuk memastikan soal tersebar merata dan metadata kurikulum lengkap.', owner: 'Admin Kurikulum', status: 'Monitoring' },
		{ title: 'Analisis butir', desc: 'Gunakan halaman analisis untuk melihat pemakaian, HOTS, status review, dan tindak lanjut.', owner: 'Admin/Guru', status: 'Monitoring' }
	];

	const integrations = [
		{ name: 'Daftar Soal', path: resolve('/bank-soal/daftar'), desc: 'Sumber data utama untuk pencarian, filter, pagination, dan aksi per soal.' },
		{ name: 'Komposer', path: resolve('/bank-soal/tambah'), desc: 'Pembuatan/edit soal dengan autosave, shortcut, validasi, dan preview siswa.' },
		{ name: 'Review', path: resolve('/bank-soal/verifikasi'), desc: 'Antrean verifikasi, catatan reviewer, timeline, approve/revisi.' },
		{ name: 'Import', path: resolve('/bank-soal/impor'), desc: 'Preview dry-run dan import final dari Word/Excel/template.' },
		{ name: 'Asesmen Paket', path: resolve('/asesmen/paket'), desc: 'Pemakaian soal terbit ke paket asesmen.' }
	];

	let promise = $state<Promise<Payload> | null>(null);
	let summary = $state<SummaryResponse>({});
	let subjects = $state<Payload['subjects']>([]);

	async function fetchPayload(): Promise<Payload> {
		const [summaryPayload, subjectPayload] = await Promise.all([
			fetch('/api/bank-soal/summary').then((response) => readClientApiData<SummaryResponse>(response, 'Gagal memuat ringkasan Bank Soal')),
			fetch('/api/bank-soal/soal-support/subjects').then((response) => readClientApiData<SubjectPayload>(response, 'Gagal memuat mapel'))
		]);
		return { summary: summaryPayload ?? {}, subjects: subjectPayload.subjects ?? [] };
	}

	function load() {
		promise = fetchPayload().then((payload) => {
			summary = payload.summary;
			subjects = payload.subjects;
			return payload;
		});
	}

	let totalQuestions = $derived(summary.counts?.total ?? 0);
	let readyQuestions = $derived((summary.counts?.approved ?? 0) + (summary.counts?.published ?? 0));
	let pendingReview = $derived(summary.counts?.review ?? 0);
	let completionRate = $derived(totalQuestions > 0 ? Math.round((readyQuestions / totalQuestions) * 100) : 0);
	let subjectCoverage = $derived(summary.by_subject?.length ?? subjects?.length ?? 0);
	let cognitiveCoverage = $derived(summary.by_cognitive_level?.filter((item) => (item.total ?? 0) > 0).length ?? 0);
	let operationalStatus = $derived([
		{ label: 'Kesiapan Bank Soal', value: `${completionRate}%`, desc: `${readyQuestions} dari ${totalQuestions} soal approved/published` },
		{ label: 'Antrean Review', value: pendingReview, desc: 'Soal menunggu keputusan reviewer' },
		{ label: 'Coverage Mapel', value: subjectCoverage, desc: 'Mapel muncul pada ringkasan/sumber akademik' },
		{ label: 'Level Kognitif', value: cognitiveCoverage, desc: 'Kategori Bloom/C-level berisi soal' }
	]);

	onMount(load);
</script>

<svelte:head><title>Pengaturan Bank Soal</title></svelte:head>

<div class="space-y-5 p-4 md:p-6">
	<section class="overflow-hidden rounded-2xl border border-emerald-100 bg-white shadow-sm">
		<div class="bg-gradient-to-r from-emerald-50 via-white to-slate-50 p-4 md:p-5">
			<div class="flex flex-col gap-4 lg:flex-row lg:items-start lg:justify-between">
				<div>
					<p class="text-[10px] font-black uppercase tracking-[0.28em] text-emerald-700">Governance Bank Soal</p>
					<h1 class="mt-1 text-2xl font-black uppercase italic tracking-tight text-slate-950">Pengaturan & SOP</h1>
					<p class="mt-2 max-w-3xl text-sm leading-6 text-slate-600">Pusat panduan operasional Bank Soal: workflow, standar kualitas, SOP import/review, dan integrasi dengan modul Asesmen.</p>
				</div>
				<div class="flex flex-wrap gap-2">
					<a href={resolve('/bank-soal')} class="rounded-md border border-slate-200 bg-white px-3 py-2 text-sm font-semibold text-slate-700 hover:bg-slate-50">Dashboard</a>
					<a href={resolve('/bank-soal/analisis-butir')} class="rounded-md border border-emerald-200 bg-emerald-50 px-3 py-2 text-sm font-semibold text-emerald-800 hover:bg-emerald-100">Analisis</a>
				</div>
			</div>
		</div>
	</section>

	<AsyncContent {promise}>
		{#snippet pending()}
			<Skeleton class="h-80 w-full" />
		{/snippet}
		{#snippet failed(error, reset)}
			<RecoveryPanel title="Pengaturan belum bisa dimuat" message={error instanceof Error ? error.message : 'Gagal memuat data pengaturan.'} onRetry={() => { reset?.(); load(); }} />
		{/snippet}
		{#snippet children()}
			<section class="grid gap-3 md:grid-cols-2 xl:grid-cols-4">
				{#each operationalStatus as item (item.label)}
					<article class="rounded-xl border border-slate-200 bg-white p-4 shadow-sm">
						<p class="text-[10px] font-bold uppercase tracking-[0.18em] text-slate-500">{item.label}</p>
						<p class="mt-2 text-3xl font-black text-slate-950">{item.value}</p>
						<p class="mt-1 text-xs text-slate-500">{item.desc}</p>
					</article>
				{/each}
			</section>

			<section class="grid gap-4 xl:grid-cols-[minmax(0,1fr)_24rem]">
				<div class="space-y-4">
					<div class="rounded-xl border border-slate-200 bg-white p-4 shadow-sm">
						<h2 class="text-base font-bold text-slate-900">Workflow Standar</h2>
						<p class="mt-1 text-xs text-slate-500">Status ini menjadi acuan operasional saat soal bergerak dari draft sampai siap dipakai.</p>
						<div class="mt-4 grid gap-3 md:grid-cols-5">
							{#each workflowSteps as step, index (step.label)}
								<div class="rounded-xl border border-slate-200 bg-slate-50 p-3">
									<div class="flex items-center gap-2">
										<span class="flex size-7 items-center justify-center rounded-full bg-white text-xs font-black text-slate-700">{index + 1}</span>
										<p class="text-sm font-bold text-slate-900">{step.label}</p>
									</div>
									<p class="mt-2 text-xs leading-5 text-slate-600">{step.desc}</p>
								</div>
							{/each}
						</div>
					</div>

					<div class="rounded-xl border border-slate-200 bg-white p-4 shadow-sm">
						<h2 class="text-base font-bold text-slate-900">Standar Kualitas Minimum</h2>
						<div class="mt-4 grid gap-2 md:grid-cols-2">
							{#each qualityRules as rule, index (rule)}
								<div class="rounded-lg border border-emerald-100 bg-emerald-50/60 p-3 text-sm leading-5 text-emerald-950">
									<span class="mr-2 inline-flex size-5 items-center justify-center rounded-full bg-emerald-700 text-[10px] font-black text-white">{index + 1}</span>{rule}
								</div>
							{/each}
						</div>
					</div>

					<div class="rounded-xl border border-slate-200 bg-white p-4 shadow-sm">
						<h2 class="text-base font-bold text-slate-900">Integrasi Modul</h2>
						<div class="mt-4 grid gap-3 md:grid-cols-2">
							{#each integrations as item (item.path)}
								<a href={item.path} class="rounded-lg border border-slate-200 bg-slate-50 p-3 transition hover:border-emerald-200 hover:bg-emerald-50">
									<p class="text-sm font-bold text-slate-900">{item.name}</p>
									<p class="mt-1 text-xs leading-5 text-slate-500">{item.desc}</p>
									<p class="mt-2 font-mono text-[10px] text-emerald-700">{item.path}</p>
								</a>
							{/each}
						</div>
					</div>
				</div>

				<aside class="space-y-4">
					<div class="rounded-xl border border-slate-200 bg-white p-4 shadow-sm">
						<h2 class="text-base font-bold text-slate-900">SOP Operasional</h2>
						<div class="mt-3 space-y-3">
							{#each sopItems as item (item.title)}
								<div class="rounded-lg border border-slate-100 bg-slate-50 p-3">
									<div class="flex items-start justify-between gap-3">
										<div><p class="text-sm font-bold text-slate-900">{item.title}</p><p class="mt-1 text-xs leading-5 text-slate-500">{item.desc}</p></div>
										<span class="rounded-full border border-emerald-200 bg-white px-2 py-1 text-[10px] font-bold uppercase text-emerald-700">{item.status}</span>
									</div>
									<p class="mt-2 text-[10px] font-bold uppercase tracking-wide text-slate-400">PIC: {item.owner}</p>
								</div>
							{/each}
						</div>
					</div>

					<div class="rounded-xl border border-amber-200 bg-amber-50 p-4 text-amber-950">
						<p class="text-sm font-bold">Catatan konfigurasi teknis</p>
						<p class="mt-2 text-xs leading-5 text-amber-800">Halaman ini masih berupa pusat kontrol operasional berbasis data live. Perubahan aturan sistem yang bersifat mutasi tetap perlu endpoint khusus agar aman, ter-audit, dan tidak mengganggu periode asesmen aktif.</p>
					</div>

					<div class="rounded-xl border border-slate-200 bg-white p-4 shadow-sm">
						<h2 class="text-base font-bold text-slate-900">Rute Aktif</h2>
						<div class="mt-3 space-y-2 text-xs text-slate-600">
							<p><span class="font-mono text-emerald-700">/bank-soal/*</span> untuk UI aktif.</p>
							<p><span class="font-mono text-emerald-700">/api/bank-soal/*</span> untuk BFF aktif.</p>
							<p>Namespace lama tetap compatibility/deprecated dan tidak dipakai untuk fitur baru.</p>
						</div>
					</div>
				</aside>
			</section>
		{/snippet}
	</AsyncContent>
</div>
