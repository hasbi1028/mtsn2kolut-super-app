<script lang="ts">
	import { Button } from '$lib/components/ui/button';
	import LoadingButton from '$lib/components/LoadingButton.svelte';

	type Subject = { id: string; name: string; code: string };
	type LegacyImportResult = {
		total_rows: number;
		valid?: number;
		would_import?: number;
		imported: number;
		skipped: number;
		errors: string[];
		duplicate_codes: string[];
	};

	let {
		subjects,
		selectedImportContext,
		specialEventQuestionMode,
		importSubjectId = $bindable(''),
		importBusy,
		hasImportFile,
		importDryRunDone,
		importResult,
		templateBusy,
		onTemplate,
		onFileChange,
		onSubjectChange,
		onDryRun,
		onConfirmImport,
		onBack
	}: {
		subjects: Subject[];
		selectedImportContext: string;
		specialEventQuestionMode: boolean;
		importSubjectId: string;
		importBusy: boolean;
		hasImportFile: boolean;
		importDryRunDone: boolean;
		importResult: LegacyImportResult | null;
		templateBusy: boolean;
		onTemplate: () => void;
		onFileChange: (event: Event) => void;
		onSubjectChange: (subjectId: string) => void;
		onDryRun: () => void;
		onConfirmImport: () => void;
		onBack: () => void;
	} = $props();

	let importIntroCopy = $derived(
		specialEventQuestionMode
			? 'Hasil upload CSV akan disimpan sebagai soal khusus kegiatan yang sedang dipilih. Kolom tipe boleh kosong untuk PG lama, atau diisi: pg_kompleks, benar_salah, setuju_tidak_setuju, isian, essay, menjodohkan.'
			: 'Hasil upload disimpan sebagai draft di Bank Soal reusable. Kolom tipe boleh kosong untuk PG lama, atau diisi: pg_kompleks, benar_salah, setuju_tidak_setuju, isian, essay, menjodohkan.'
	);
	let importScopeCopy = $derived(
		specialEventQuestionMode
			? 'Mode khusus kegiatan aktif: CSV dikirim dengan event_id kegiatan terpilih, sehingga soal menjadi stok khusus event tersebut.'
			: 'Konteks kegiatan hanya membantu filter dan cek kebutuhan; import tetap menambah stok reusable lintas paket.'
	);
	let stepTwoCopy = $derived(
		specialEventQuestionMode
			? 'Mapel wajib dipilih; CSV akan dikaitkan ke kegiatan terpilih sebagai soal khusus event.'
			: 'Mapel wajib dipilih; import default masuk Bank Soal reusable tanpa event_id.'
	);

	function handleSubjectChange(event: Event) {
		onSubjectChange((event.currentTarget as HTMLSelectElement).value);
	}
</script>

<section class="rounded-xl border border-slate-200 bg-white shadow-sm" aria-labelledby="legacy-import-title">
	<div class="space-y-4 p-4 md:p-5">
		<div>
			<p class="text-xs font-bold uppercase tracking-wider text-green-700">Upload CSV</p>
			<h2 id="legacy-import-title" class="mt-1 text-base font-semibold text-slate-800">Masukkan banyak soal sekaligus</h2>
			<p class="mt-1 text-xs text-slate-500">{importIntroCopy}</p>
			<p class="mt-2 rounded-md border border-slate-200 bg-slate-50 px-2 py-1.5 text-xs text-slate-700">Tujuan import: <span class="font-semibold text-emerald-800">{selectedImportContext}</span></p>
			<p class="mt-1 text-xs text-slate-500">{importScopeCopy}</p>
		</div>
		<div class="grid gap-2 text-xs sm:grid-cols-4">
			<div class="rounded-md border border-emerald-200 bg-white p-3 text-emerald-900 shadow-sm">
				<p class="font-semibold">1. Download Template</p>
				<LoadingButton variant="outline" size="sm" onclick={onTemplate} loading={templateBusy} loadingLabel="Mengunduh..." class="mt-2 h-7 bg-white text-[11px]">
					Template
				</LoadingButton>
			</div>
			<div class="rounded-md border border-slate-200 bg-white p-3 text-slate-700">
				<p class="font-semibold">2. Pilih Mapel</p>
				<p class="mt-1 text-[11px]">{stepTwoCopy}</p>
			</div>
			<div class="rounded-md border border-slate-200 bg-white p-3 text-slate-700">
				<p class="font-semibold">3. Upload File</p>
				<p class="mt-1 text-[11px]">Gunakan CSV UTF-8 agar huruf dan simbol aman.</p>
			</div>
			<div class="rounded-md border border-slate-200 bg-white p-3 text-slate-700">
				<p class="font-semibold">4. Lihat Hasil Import</p>
				<p class="mt-1 text-[11px]">Jumlah masuk, dilewati, dan error muncul setelah upload.</p>
			</div>
		</div>
		<div class="space-y-3">
			<div>
				<label for="legacy-import-subject" class="mb-1 block text-xs font-medium text-slate-600">
					Mata Pelajaran <span class="text-red-500">*</span>
				</label>
				<select
					id="legacy-import-subject"
					value={importSubjectId}
					onchange={handleSubjectChange}
					class="w-full rounded-md border border-slate-200 bg-white px-2.5 py-2 text-sm text-slate-700 focus:outline-none focus:ring-2 focus:ring-green-500"
				>
					<option value="">-- Pilih Mapel --</option>
					{#each subjects as subject (subject.id)}
						<option value={subject.id}>{subject.name}</option>
					{/each}
				</select>
			</div>
			<div>
				<label for="legacy-import-file" class="mb-1 block text-xs font-medium text-slate-600">
					File CSV <span class="text-red-500">*</span>
				</label>
				<input
					id="legacy-import-file"
					type="file"
					accept=".csv,text/csv"
					onchange={onFileChange}
					class="block w-full rounded-md border border-slate-200 bg-white px-2.5 py-2 text-sm text-slate-700 file:mr-3 file:rounded-md file:border-0 file:bg-green-50 file:px-3 file:py-1.5 file:text-sm file:font-medium file:text-green-800"
				/>
			</div>
		</div>
		{#if importResult}
			<div class="rounded-md border border-slate-200 bg-slate-50 p-3 text-sm text-slate-800" aria-live="polite">
				<div class="grid grid-cols-3 gap-2 text-center">
					<div>
						<div class="text-lg font-bold">{importDryRunDone ? (importResult.would_import ?? importResult.valid ?? 0) : importResult.imported}</div>
						<div class="text-[10px] uppercase text-green-700">{importDryRunDone ? 'Akan Masuk' : 'Masuk'}</div>
					</div>
					<div>
						<div class="text-lg font-bold">{importResult.skipped}</div>
						<div class="text-[10px] uppercase text-green-700">Lewat</div>
					</div>
					<div>
						<div class="text-lg font-bold">{importResult.total_rows}</div>
						<div class="text-[10px] uppercase text-green-700">Baris</div>
					</div>
				</div>
				{#if importDryRunDone}
					<p class="mt-3 border-t border-green-200 pt-2 text-xs font-semibold text-green-900">Preview dry-run selesai. Periksa error dan kode duplikat sebelum menekan Konfirmasi Import.</p>
				{/if}
				{#if importResult.errors.length > 0}
					<p class="mt-3 border-t border-green-200 pt-2 text-xs font-semibold text-amber-900">Error import ditampilkan agar kolom wajib, format tipe, dan encoding bisa diperbaiki sebelum upload ulang.</p>
					<ul class="mt-3 space-y-1 border-t border-green-200 pt-2 text-xs text-amber-800">
						{#each importResult.errors.slice(0, 6) as error (`legacy-import-error-${error}`)}
							<li>{error}</li>
						{/each}
					</ul>
				{/if}
				{#if importResult.duplicate_codes.length > 0}
					<div class="mt-3 border-t border-green-200 pt-2 text-xs text-amber-900">
						<p class="font-semibold">Kode duplikat dilewati:</p>
						<p class="mt-1 break-words font-mono text-[11px]">{importResult.duplicate_codes.slice(0, 24).join(', ')}{importResult.duplicate_codes.length > 24 ? `, +${importResult.duplicate_codes.length - 24} lagi` : ''}</p>
					</div>
				{/if}
			</div>
		{/if}
		<div class="flex flex-wrap justify-end gap-2 border-t border-slate-100 pt-4">
			<Button variant="outline" onclick={onBack}>Kembali ke Daftar</Button>
			<LoadingButton onclick={onDryRun} loading={importBusy} loadingLabel="Preview..." disabled={importBusy || !importSubjectId || !hasImportFile} variant="outline" class="disabled:opacity-50">
				Preview Dry-run
			</LoadingButton>
			<LoadingButton onclick={onConfirmImport} loading={importBusy} loadingLabel="Import..." disabled={importBusy || !importSubjectId || !hasImportFile || !importDryRunDone} class="bg-green-700 text-white hover:bg-green-800 disabled:opacity-50">
				Konfirmasi Import
			</LoadingButton>
		</div>
	</div>
</section>
