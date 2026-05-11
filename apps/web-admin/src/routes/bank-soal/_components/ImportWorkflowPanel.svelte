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
			? 'CSV disimpan sebagai soal khusus kegiatan terpilih. Pakai mode ini hanya untuk stok kegiatan yang tidak boleh masuk bank soal pakai ulang.'
			: 'CSV disimpan sebagai konsep Bank Soal pakai ulang tanpa kegiatan khusus, sehingga bisa dipakai ulang lintas paket.'
	);
	let importScopeCopy = $derived(
		specialEventQuestionMode
			? 'Khusus kegiatan: soal dikaitkan ke kegiatan yang sedang dipilih.'
			: 'Pakai ulang: konteks kegiatan hanya membantu cek kebutuhan, bukan tujuan impor.'
	);
	let readyImportCount = $derived(importResult?.would_import ?? importResult?.valid ?? importResult?.imported ?? 0);
	let hasImportErrors = $derived((importResult?.errors.length ?? 0) > 0);
	let canConfirmImport = $derived(Boolean(importSubjectId && hasImportFile && importDryRunDone && readyImportCount > 0 && !hasImportErrors));
	let importFileSizeLabel = $derived(formatFileSize(importFileSize));
	let importFileStateLabel = $derived(importDryRunDone ? (hasImportErrors ? 'Perlu perbaikan' : 'Pratinjau bersih') : 'Belum pratinjau');

	let importSteps = $derived([
		{ label: '1. Format Isian', desc: 'Pakai struktur CSV resmi agar kolom tipe, kunci, dan opsi konsisten.', ready: true },
		{ label: '2. Mapel & File', desc: importSubjectId && hasImportFile ? `${importFileName || 'CSV dipilih'} · ${importFileSizeLabel}` : 'Pilih mata pelajaran dan unggah CSV.', ready: Boolean(importSubjectId && hasImportFile) },
		{ label: '3. Cek Data', desc: importDryRunDone ? (hasImportErrors ? 'Ada masalah yang perlu diperbaiki.' : `${readyImportCount} soal siap diimpor.`) : 'Wajib sebelum impor final.', ready: importDryRunDone && !hasImportErrors },
		{ label: '4. Konfirmasi', desc: canConfirmImport ? 'Impor final sudah aman dijalankan.' : 'Menunggu pratinjau bersih.', ready: canConfirmImport },
	]);

	function handleSubjectChange(event: Event) {
		onSubjectChange((event.currentTarget as HTMLSelectElement).value);
	}

	function formatFileSize(bytes: number): string {
		if (bytes <= 0) return '0 KB';
		if (bytes < 1024 * 1024) return `${Math.max(1, Math.round(bytes / 1024))} KB`;
		return `${(bytes / (1024 * 1024)).toFixed(1)} MB`;
	}
</script>

<section class="overflow-hidden rounded-2xl border border-primary/20 bg-card shadow-sm" aria-labelledby="legacy-import-title">
	<div class="border-b border-primary/20 bg-gradient-to-r from-primary/10 via-card to-warning/10 p-4 md:p-5">
		<div class="flex flex-col gap-4 lg:flex-row lg:items-start lg:justify-between">
			<div class="min-w-0">
				<p class="text-[10px] font-black uppercase tracking-[0.28em] text-primary">Studio Impor Bank Soal</p>
				<h2 id="legacy-import-title" class="mt-1 text-xl font-black uppercase italic tracking-tight text-foreground">Masukkan banyak soal sekaligus</h2>
				<p class="mt-2 max-w-3xl text-sm leading-6 text-muted-foreground">{importIntroCopy}</p>
				<div class="mt-3 flex flex-wrap gap-2">
					<span class="rounded-full border border-primary/20 bg-card px-2.5 py-1 text-[10px] font-bold uppercase tracking-wide text-primary">{selectedImportContext}</span>
					<span class="rounded-full border border-warning/30 bg-card px-2.5 py-1 text-[10px] font-bold uppercase tracking-wide text-warning">{importScopeCopy}</span>
				</div>
			</div>
			<LoadingButton variant="outline" size="sm" onclick={onTemplate} loading={templateBusy} loadingLabel="Mengunduh..." class="h-9 shrink-0 bg-card text-xs">
				Unduh Format Isian CSV
			</LoadingButton>
		</div>
	</div>
	<div class="space-y-4 p-4 md:p-5">
		<div class="grid gap-3 md:grid-cols-2 xl:grid-cols-4">
			{#each importSteps as step (step.label)}
				<div class="rounded-xl border p-3 {step.ready ? 'border-primary/20 bg-primary/10 text-primary' : 'border-border bg-muted/50 text-foreground'}">
					<div class="flex items-start justify-between gap-2">
						<p class="text-[10px] font-black uppercase tracking-[0.18em]">{step.label}</p>
						<span class="rounded-full bg-card/80 px-2 py-0.5 text-[10px] font-bold">{step.ready ? 'OK' : 'Menunggu'}</span>
					</div>
					<p class="mt-2 text-xs leading-5 opacity-80">{step.desc}</p>
				</div>
			{/each}
		</div>
		<div class="rounded-lg border border-border bg-card px-3 py-2 text-xs text-muted-foreground">
			<span class="font-semibold text-foreground">Format tipe:</span> boleh kosong untuk PG lama, atau diisi: pg_kompleks, benar_salah, setuju_tidak_setuju, isian, essay, menjodohkan.
		</div>
		<div class="grid gap-3 md:grid-cols-2">
			<div>
				<label for="legacy-import-subject" class="mb-1 block text-xs font-medium text-muted-foreground">
					Mata Pelajaran <span class="text-destructive">*</span>
				</label>
				<select
					id="legacy-import-subject"
					value={importSubjectId}
					onchange={handleSubjectChange}
					class="w-full rounded-md border border-border bg-card px-2.5 py-2 text-sm text-foreground focus:outline-none focus:ring-2 focus:ring-ring"
				>
					<option value="">-- Pilih Mapel --</option>
					{#each subjects as subject (subject.id)}
						<option value={subject.id}>{subject.name}</option>
					{/each}
				</select>
			</div>
			<div>
				<label for="legacy-import-file" class="mb-1 block text-xs font-medium text-muted-foreground">
					File CSV <span class="text-destructive">*</span>
				</label>
					<input
						id="legacy-import-file"
						type="file"
						accept=".csv,text/csv"
						onchange={onFileChange}
						class="block w-full rounded-md border border-border bg-card px-2.5 py-2 text-sm text-foreground file:mr-3 file:rounded-md file:border-0 file:bg-success/10 file:px-3 file:py-1.5 file:text-sm file:font-medium file:text-success"
					/>
					{#if hasImportFile}
						<div class="mt-2 flex flex-wrap items-center gap-2 rounded-md border border-success/20 bg-success/10 px-2.5 py-1.5 text-xs text-success">
							<span class="max-w-56 truncate font-semibold">{importFileName || 'File CSV dipilih'}</span>
							<span class="text-success">{importFileSizeLabel}</span>
							<span class="rounded-full bg-card px-2 py-0.5 font-semibold {importDryRunDone && !hasImportErrors ? 'text-success' : 'text-warning'}">
								{importFileStateLabel}
							</span>
						</div>
					{/if}
				</div>
			</div>
			{#if importResult}
				<div class="rounded-md border border-border bg-muted/50 p-3 text-sm text-foreground" aria-live="polite">
					<div class="grid grid-cols-2 gap-2 text-center md:grid-cols-4">
						<div>
							<div class="text-lg font-bold">{importDryRunDone ? (importResult.would_import ?? importResult.valid ?? 0) : importResult.imported}</div>
							<div class="text-[10px] uppercase text-success">{importDryRunDone ? 'Akan Masuk' : 'Masuk'}</div>
					</div>
					<div>
						<div class="text-lg font-bold">{importResult.skipped}</div>
						<div class="text-[10px] uppercase text-success">Lewat</div>
					</div>
						<div>
							<div class="text-lg font-bold">{importResult.total_rows}</div>
							<div class="text-[10px] uppercase text-success">Baris</div>
						</div>
						<div>
							<div class="text-lg font-bold {hasImportErrors ? 'text-warning' : 'text-foreground'}">{importResult.errors.length}</div>
							<div class="text-[10px] uppercase {hasImportErrors ? 'text-warning' : 'text-success'}">Error</div>
						</div>
					</div>
					{#if importDryRunDone}
						<p class="mt-3 border-t border-success/20 pt-2 text-xs font-semibold {hasImportErrors ? 'text-warning' : 'text-success'}">
							{hasImportErrors
								? 'Pratinjau menemukan masalah. Perbaiki file CSV lalu cek data ulang sebelum impor.'
								: 'Pratinjau cek data bersih. Konfirmasi impor sudah aman dijalankan.'}
						</p>
					{/if}
					{#if importResult.errors.length > 0}
						<p class="mt-3 border-t border-success/20 pt-2 text-xs font-semibold text-warning">Error import ditampilkan agar kolom wajib, format tipe, dan encoding bisa diperbaiki sebelum upload ulang.</p>
					<ul class="mt-3 space-y-1 border-t border-success/20 pt-2 text-xs text-warning">
						{#each importResult.errors.slice(0, 6) as error (`legacy-import-error-${error}`)}
							<li>{error}</li>
						{/each}
					</ul>
				{/if}
				{#if importResult.duplicate_codes.length > 0}
					<div class="mt-3 border-t border-success/20 pt-2 text-xs text-warning">
						<p class="font-semibold">Kode duplikat dilewati:</p>
						<p class="mt-1 break-words font-mono text-[11px]">{importResult.duplicate_codes.slice(0, 24).join(', ')}{importResult.duplicate_codes.length > 24 ? `, +${importResult.duplicate_codes.length - 24} lagi` : ''}</p>
					</div>
				{/if}
			</div>
		{/if}
		<div class="flex flex-wrap justify-end gap-2 border-t border-border pt-4">
			<Button variant="outline" onclick={onBack}>Kembali ke Daftar</Button>
				<LoadingButton onclick={onDryRun} loading={importBusy} loadingLabel="Pratinjau..." disabled={importBusy || !importSubjectId || !hasImportFile} variant="outline" class="disabled:opacity-50">
					Pratinjau Dry-run
				</LoadingButton>
				<LoadingButton onclick={onConfirmImport} loading={importBusy} loadingLabel="Import..." disabled={importBusy || !canConfirmImport} class="bg-success text-background hover:bg-success disabled:opacity-50">
					Konfirmasi Import
				</LoadingButton>
		</div>
	</div>
</section>
