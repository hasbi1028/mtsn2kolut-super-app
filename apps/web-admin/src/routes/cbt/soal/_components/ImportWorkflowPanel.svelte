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
			importFileName = '',
			importFileSize = 0,
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
			importFileName: string;
			importFileSize: number;
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
			? 'CSV disimpan sebagai soal khusus kegiatan terpilih. Pakai mode ini hanya untuk stok event yang tidak boleh masuk repositori reusable.'
			: 'CSV disimpan sebagai draft Bank Soal reusable tanpa event_id, sehingga bisa dipakai ulang lintas paket.'
	);
	let importScopeCopy = $derived(
		specialEventQuestionMode
			? 'Khusus event: soal dikaitkan ke kegiatan yang sedang dipilih.'
			: 'Reusable: konteks kegiatan hanya membantu cek kebutuhan, bukan tujuan import.'
	);
	let readyImportCount = $derived(importResult?.would_import ?? importResult?.valid ?? importResult?.imported ?? 0);
	let hasImportErrors = $derived((importResult?.errors.length ?? 0) > 0);
	let canConfirmImport = $derived(Boolean(importSubjectId && hasImportFile && importDryRunDone && readyImportCount > 0 && !hasImportErrors));
	let importFileSizeLabel = $derived(formatFileSize(importFileSize));
	let importFileStateLabel = $derived(importDryRunDone ? (hasImportErrors ? 'Perlu perbaikan' : 'Preview bersih') : 'Belum preview');

	function handleSubjectChange(event: Event) {
		onSubjectChange((event.currentTarget as HTMLSelectElement).value);
	}

	function formatFileSize(bytes: number): string {
		if (bytes <= 0) return '0 KB';
		if (bytes < 1024 * 1024) return `${Math.max(1, Math.round(bytes / 1024))} KB`;
		return `${(bytes / (1024 * 1024)).toFixed(1)} MB`;
	}
</script>

<section class="rounded-xl border border-slate-200 bg-white shadow-sm" aria-labelledby="legacy-import-title">
	<div class="space-y-4 p-4 md:p-5">
		<div>
			<p class="text-xs font-bold uppercase tracking-wider text-green-700">Upload CSV</p>
			<h2 id="legacy-import-title" class="mt-1 text-base font-semibold text-slate-800">Masukkan banyak soal sekaligus</h2>
			<p class="mt-1 text-xs text-slate-500">{importIntroCopy}</p>
			<div class="mt-3 rounded-lg border border-slate-200 bg-slate-50 px-3 py-2 text-xs text-slate-700">
				<p><span class="font-semibold text-slate-900">Tujuan import:</span> {selectedImportContext}</p>
				<p class="mt-1 text-slate-500">{importScopeCopy}</p>
			</div>
		</div>
		<div class="flex flex-wrap items-center gap-2 rounded-lg border border-slate-200 bg-white px-3 py-2 text-xs text-slate-600">
			<LoadingButton variant="outline" size="sm" onclick={onTemplate} loading={templateBusy} loadingLabel="Mengunduh..." class="h-8 bg-white text-xs">
				Download Template
			</LoadingButton>
			<span>Kolom tipe boleh kosong untuk PG lama, atau diisi: pg_kompleks, benar_salah, setuju_tidak_setuju, isian, essay, menjodohkan.</span>
		</div>
		<div class="grid gap-3 md:grid-cols-2">
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
					{#if hasImportFile}
						<div class="mt-2 flex flex-wrap items-center gap-2 rounded-md border border-green-100 bg-green-50 px-2.5 py-1.5 text-xs text-green-900">
							<span class="max-w-56 truncate font-semibold">{importFileName || 'File CSV dipilih'}</span>
							<span class="text-green-700">{importFileSizeLabel}</span>
							<span class="rounded-full bg-white px-2 py-0.5 font-semibold {importDryRunDone && !hasImportErrors ? 'text-green-700' : 'text-amber-700'}">
								{importFileStateLabel}
							</span>
						</div>
					{/if}
				</div>
			</div>
			{#if importResult}
				<div class="rounded-md border border-slate-200 bg-slate-50 p-3 text-sm text-slate-800" aria-live="polite">
					<div class="grid grid-cols-2 gap-2 text-center md:grid-cols-4">
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
						<div>
							<div class="text-lg font-bold {hasImportErrors ? 'text-amber-700' : 'text-slate-900'}">{importResult.errors.length}</div>
							<div class="text-[10px] uppercase {hasImportErrors ? 'text-amber-700' : 'text-green-700'}">Error</div>
						</div>
					</div>
					{#if importDryRunDone}
						<p class="mt-3 border-t border-green-200 pt-2 text-xs font-semibold {hasImportErrors ? 'text-amber-900' : 'text-green-900'}">
							{hasImportErrors
								? 'Preview menemukan error. Perbaiki file CSV lalu jalankan dry-run ulang sebelum import.'
								: 'Preview dry-run bersih. Konfirmasi Import sudah aman dijalankan.'}
						</p>
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
				<LoadingButton onclick={onConfirmImport} loading={importBusy} loadingLabel="Import..." disabled={importBusy || !canConfirmImport} class="bg-green-700 text-white hover:bg-green-800 disabled:opacity-50">
					Konfirmasi Import
				</LoadingButton>
		</div>
	</div>
</section>
