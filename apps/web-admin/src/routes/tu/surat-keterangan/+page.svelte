<script lang="ts">
	import PlusIcon from '@lucide/svelte/icons/plus';
	import PrinterIcon from '@lucide/svelte/icons/printer';
	import RefreshCcwIcon from '@lucide/svelte/icons/refresh-ccw';
	import XCircleIcon from '@lucide/svelte/icons/x-circle';
	import AsyncContent from '$lib/components/AsyncContent.svelte';
	import LoadingButton from '$lib/components/LoadingButton.svelte';
	import RecoveryPanel from '$lib/components/RecoveryPanel.svelte';
	import { Badge } from '$lib/components/ui/badge';
	import { Button } from '$lib/components/ui/button';
	import * as Card from '$lib/components/ui/card';
	import { Input } from '$lib/components/ui/input';
	import { Skeleton } from '$lib/components/ui/skeleton';
	import * as Table from '$lib/components/ui/table';
	import { Textarea } from '$lib/components/ui/textarea';
	import { toast } from '$lib/components/ui/sonner';
	import { onMount } from 'svelte';
	import { confirmAction } from '$lib/confirm-dialog';

	type CertificateTemplate = {
		id: string;
		code: string;
		name: string;
		description: string;
		default_purpose: string;
		body_template: string;
	};

	type StudentOption = {
		id: string;
		nis: string;
		nisn: string;
		nama: string;
		status: string;
		class_name: string;
		parent_name: string;
		tempat_lahir: string;
		tanggal_lahir: string | null;
	};

	type StudentCertificate = {
		id: string;
		template_code: string;
		template_name: string;
		student_name: string;
		student_nis: string;
		class_name: string;
		nomor_surat: string;
		classification_code: string;
		tanggal_surat: string;
		purpose: string;
		recipient: string;
		status: string;
		created_by_username: string;
	};

	type Overview = {
		templates: CertificateTemplate[];
		students: StudentOption[];
		certificates: StudentCertificate[];
	};

	type ApiEnvelope<T> = {
		data?: T;
		error?: string;
		message?: string;
	};

	type CertificateForm = {
		template_id: string;
		student_id: string;
		tanggal_surat: string;
		purpose: string;
		recipient: string;
		remarks: string;
	};

	const STUDENT_STATUSES = [
		['', 'Semua status'],
		['active', 'Aktif'],
		['alumni', 'Alumni'],
		['mutated', 'Pindah'],
		['prospective', 'Calon siswa']
	] as const;

	let overviewPromise = $state<Promise<Overview> | null>(null);
	let templates = $state<CertificateTemplate[]>([]);
	let students = $state<StudentOption[]>([]);
	let search = $state('');
	let status = $state('');
	let templateCode = $state('');
	let studentSearch = $state('');
	let studentStatus = $state('');
	let refreshBusy = $state(false);
	let filterBusy = $state(false);
	let createBusy = $state(false);
	let studentBusy = $state(false);
	let cancelBusy = $state<Record<string, boolean>>({});
	let form = $state<CertificateForm>({
		template_id: '',
		student_id: '',
		tanggal_surat: new Date().toISOString().slice(0, 10),
		purpose: '',
		recipient: 'Yang berkepentingan',
		remarks: ''
	});

	let selectedTemplate = $derived(templates.find((template) => template.id === form.template_id));
	let selectedStudent = $derived(students.find((student) => student.id === form.student_id));

	function isRecord(value: unknown): value is Record<string, unknown> {
		return typeof value === 'object' && value !== null;
	}

	function apiErrorMessage(payload: unknown) {
		if (!isRecord(payload)) return '';
		const error = payload.error;
		if (typeof error === 'string' && error.trim()) return error;
		const message = payload.message;
		if (typeof message === 'string' && message.trim()) return message;
		return '';
	}

	async function readApi<T>(response: Response, fallbackMessage: string): Promise<T> {
		const payload = (await response.json().catch(() => null)) as ApiEnvelope<T> | T | null;
		const message = apiErrorMessage(payload);
		if (!response.ok) throw new Error(message || fallbackMessage);
		if (isRecord(payload) && typeof payload.error === 'string' && payload.error.trim()) throw new Error(payload.error);
		if (isRecord(payload) && 'data' in payload) {
			const envelope = payload as ApiEnvelope<T>;
			if (envelope.data === undefined) throw new Error(fallbackMessage);
			return envelope.data;
		}
		if (payload === null) throw new Error(fallbackMessage);
		return payload as T;
	}

	function applyOverview(overview: Overview) {
		templates = overview.templates ?? [];
		students = overview.students ?? [];
		if (!form.template_id && templates[0]) {
			selectTemplate(templates[0].id);
		}
	}

	async function fetchOverview(): Promise<Overview> {
		const certificateParams = new URLSearchParams();
		if (search.trim()) certificateParams.set('search', search.trim());
		if (status) certificateParams.set('status', status);
		if (templateCode) certificateParams.set('template_code', templateCode);
		const studentParams = new URLSearchParams();
		if (studentSearch.trim()) studentParams.set('search', studentSearch.trim());
		if (studentStatus) studentParams.set('status', studentStatus);
		const [templatesRes, studentsRes, certificatesRes] = await Promise.all([
			fetch('/api/tu/surat-keterangan/templates'),
			fetch(`/api/tu/surat-keterangan/students?${studentParams.toString()}`),
			fetch(`/api/tu/surat-keterangan?${certificateParams.toString()}`)
		]);
		return {
			templates: await readApi<CertificateTemplate[]>(templatesRes, 'Gagal memuat template surat'),
			students: await readApi<StudentOption[]>(studentsRes, 'Gagal memuat data siswa'),
			certificates: await readApi<StudentCertificate[]>(certificatesRes, 'Gagal memuat arsip surat')
		};
	}

	function loadOverview() {
		overviewPromise = fetchOverview().then((overview) => {
			applyOverview(overview);
			return overview;
		});
		return overviewPromise;
	}

	async function refreshOverview() {
		const overview = await fetchOverview();
		applyOverview(overview);
		overviewPromise = Promise.resolve(overview);
	}

	async function refreshOverviewAction() {
		refreshBusy = true;
		try {
			await refreshOverview();
		} catch (error) {
			toast.error(overviewErrorMessage(error));
		} finally {
			refreshBusy = false;
		}
	}

	async function applyCertificateFilters() {
		filterBusy = true;
		try {
			await refreshOverview();
		} catch (error) {
			toast.error(overviewErrorMessage(error));
		} finally {
			filterBusy = false;
		}
	}

	function retryOverview(reset?: () => void) {
		reset?.();
		loadOverview();
	}

	function overviewErrorMessage(error: unknown) {
		if (error instanceof Error && error.message.trim()) return error.message;
		return 'Gagal memuat surat keterangan siswa';
	}

	function handleRenderError(error: unknown) {
		console.error('TU student certificate render failed', error);
	}

	function selectTemplate(templateId: string) {
		form.template_id = templateId;
		const template = templates.find((item) => item.id === templateId);
		if (template) form.purpose = template.default_purpose;
	}

	function resetForm() {
		form = {
			template_id: templates[0]?.id ?? '',
			student_id: '',
			tanggal_surat: new Date().toISOString().slice(0, 10),
			purpose: templates[0]?.default_purpose ?? '',
			recipient: 'Yang berkepentingan',
			remarks: ''
		};
	}

	async function loadStudents() {
		studentBusy = true;
		try {
			const params = new URLSearchParams();
			if (studentSearch.trim()) params.set('search', studentSearch.trim());
			if (studentStatus) params.set('status', studentStatus);
			const res = await fetch(`/api/tu/surat-keterangan/students?${params.toString()}`);
			students = await readApi<StudentOption[]>(res, 'Gagal memuat data siswa');
		} catch (error) {
			toast.error(error instanceof Error ? error.message : 'Gagal memuat data siswa');
		} finally {
			studentBusy = false;
		}
	}

	async function createCertificate() {
		if (!form.template_id) { toast.error('Pilih jenis surat'); return; }
		if (!form.student_id) { toast.error('Pilih siswa'); return; }
		if (!form.tanggal_surat) { toast.error('Tanggal surat wajib diisi'); return; }
		if (!form.purpose.trim()) { toast.error('Keperluan surat wajib diisi'); return; }
		createBusy = true;
		try {
			const res = await fetch('/api/tu/surat-keterangan', {
				method: 'POST',
				headers: { 'Content-Type': 'application/json' },
				body: JSON.stringify(form)
			});
			const created = await readApi<StudentCertificate>(res, 'Gagal membuat surat keterangan');
			toast.success(`Surat ${created.nomor_surat} berhasil diterbitkan`);
			resetForm();
			await refreshOverview();
		} catch (error) {
			toast.error(error instanceof Error ? error.message : 'Gagal membuat surat keterangan');
		} finally {
			createBusy = false;
		}
	}

	async function cancelCertificate(certificate: StudentCertificate) {
		const confirmed = await confirmAction({
			title: 'Batalkan Arsip Surat',
			message: `Batalkan arsip ${certificate.nomor_surat}?`,
			confirmLabel: 'Batalkan Arsip',
			tone: 'warning'
		});
		if (!confirmed) return;
		cancelBusy = { ...cancelBusy, [certificate.id]: true };
		try {
			const res = await fetch(`/api/tu/surat-keterangan/${certificate.id}/cancel`, {
				method: 'POST',
				headers: { 'Content-Type': 'application/json' },
				body: JSON.stringify({ remarks: 'Dibatalkan dari halaman Tata Usaha' })
			});
			await readApi<unknown>(res, 'Gagal membatalkan surat');
			toast.success('Surat keterangan dibatalkan');
			await refreshOverview();
		} catch (error) {
			toast.error(error instanceof Error ? error.message : 'Gagal membatalkan surat');
		} finally {
			cancelBusy = { ...cancelBusy, [certificate.id]: false };
		}
	}

	function formatDate(value: string | null | undefined) {
		if (!value) return '-';
		const date = new Date(value);
		if (Number.isNaN(date.getTime())) return value.slice(0, 10);
		return new Intl.DateTimeFormat('id-ID', { dateStyle: 'medium' }).format(date);
	}

	function statusLabel(value: string) {
		if (value === 'issued') return 'Terbit';
		if (value === 'canceled') return 'Batal';
		return value;
	}

	onMount(() => {
		loadOverview();
	});
</script>

<svelte:head>
	<title>Surat Keterangan Siswa | MTsN 2 Kolaka Utara</title>
</svelte:head>

<div class="space-y-6">
	<div class="flex flex-col gap-3 lg:flex-row lg:items-end lg:justify-between">
		<div class="space-y-1">
			<p class="text-sm font-medium text-emerald-700">Tata Usaha</p>
			<h1 class="text-2xl font-semibold tracking-normal text-slate-950">Surat Keterangan Siswa</h1>
			<p class="max-w-3xl text-sm leading-6 text-muted-foreground">
				Penerbitan surat keterangan siswa memakai nomor surat keluar pusat dengan klasifikasi PP.00.4.
			</p>
		</div>
		<LoadingButton variant="outline" loading={refreshBusy} loadingLabel="Memuat..." onclick={() => void refreshOverviewAction()}>
			<RefreshCcwIcon class="size-4" />
			Refresh
		</LoadingButton>
	</div>

	<AsyncContent promise={overviewPromise} onerror={handleRenderError}>
		{#snippet pending()}
			<div class="grid gap-4 lg:grid-cols-[minmax(320px,420px)_1fr]">
				<Skeleton class="h-[520px] rounded-lg" />
				<Skeleton class="h-[520px] rounded-lg" />
			</div>
		{/snippet}

		{#snippet failed(error, reset)}
			<RecoveryPanel
				title="Surat Keterangan Belum Bisa Dimuat"
				message={overviewErrorMessage(error)}
				onRetry={() => retryOverview(reset)}
			/>
		{/snippet}

		{#snippet children(value)}
			{@const overview = value as Overview}
			<div class="grid gap-4 lg:grid-cols-[minmax(320px,420px)_1fr]">
				<Card.Root>
					<Card.Header>
						<Card.Title>Terbitkan Surat</Card.Title>
						<Card.Description>Nomor surat dibuat otomatis dari agenda surat keluar.</Card.Description>
					</Card.Header>
					<Card.Content>
						<form class="space-y-4" onsubmit={(event) => { event.preventDefault(); void createCertificate(); }}>
							<div class="space-y-2">
								<label for="cert-template" class="text-sm font-medium">Jenis surat</label>
								<select
									id="cert-template"
									value={form.template_id}
									onchange={(event) => selectTemplate((event.currentTarget as HTMLSelectElement).value)}
									class="w-full rounded-md border border-input bg-background px-3 py-2 text-sm"
								>
									<option value="">Pilih template</option>
									{#each overview.templates as template (template.id)}
										<option value={template.id}>{template.name}</option>
									{/each}
								</select>
							</div>

							<div class="grid gap-2 sm:grid-cols-[1fr_auto]">
								<div class="space-y-2">
									<label for="student-search" class="text-sm font-medium">Cari siswa</label>
									<Input id="student-search" bind:value={studentSearch} placeholder="Nama, NIS, atau NISN" />
								</div>
								<div class="space-y-2">
									<label for="student-status" class="text-sm font-medium">Status</label>
									<select id="student-status" bind:value={studentStatus} class="w-full rounded-md border border-input bg-background px-3 py-2 text-sm sm:w-36">
										{#each STUDENT_STATUSES as [value, label] (value)}
											<option {value}>{label}</option>
										{/each}
									</select>
								</div>
							</div>
							<LoadingButton type="button" variant="outline" loading={studentBusy} loadingLabel="Memuat siswa..." onclick={() => void loadStudents()}>
								<RefreshCcwIcon class="size-4" />
								Cari Siswa
							</LoadingButton>

							<div class="space-y-2">
								<label for="cert-student" class="text-sm font-medium">Siswa</label>
								<select id="cert-student" bind:value={form.student_id} class="w-full rounded-md border border-input bg-background px-3 py-2 text-sm">
									<option value="">Pilih siswa</option>
									{#each students as student (student.id)}
										<option value={student.id}>{student.nama} · {student.nis}{student.class_name ? ` · ${student.class_name}` : ''}</option>
									{/each}
								</select>
							</div>

							{#if selectedStudent}
								<div class="rounded-lg border border-emerald-100 bg-emerald-50 px-3 py-2 text-sm text-emerald-950">
									<p class="font-medium">{selectedStudent.nama}</p>
									<p class="mt-1 text-emerald-800">
										NIS {selectedStudent.nis || '-'} · NISN {selectedStudent.nisn || '-'} · {selectedStudent.class_name || 'Tanpa kelas'}
									</p>
								</div>
							{/if}

							<div class="grid gap-3 sm:grid-cols-2">
								<div class="space-y-2">
									<label for="cert-date" class="text-sm font-medium">Tanggal surat</label>
									<Input id="cert-date" type="date" bind:value={form.tanggal_surat} />
								</div>
								<div class="space-y-2">
									<label for="cert-recipient" class="text-sm font-medium">Tujuan</label>
									<Input id="cert-recipient" bind:value={form.recipient} />
								</div>
							</div>

							<div class="space-y-2">
								<label for="cert-purpose" class="text-sm font-medium">Keperluan</label>
								<Textarea id="cert-purpose" bind:value={form.purpose} rows={3} />
							</div>

							<div class="space-y-2">
								<label for="cert-remarks" class="text-sm font-medium">Catatan TU</label>
								<Textarea id="cert-remarks" bind:value={form.remarks} rows={3} placeholder="Opsional" />
							</div>

							{#if selectedTemplate}
								<div class="rounded-lg border bg-muted/30 px-3 py-3 text-sm leading-6 text-muted-foreground">
									<p class="font-medium text-foreground">{selectedTemplate.name}</p>
									<p>{selectedTemplate.body_template}</p>
								</div>
							{/if}

							<div class="flex flex-wrap gap-2">
								<LoadingButton type="submit" loading={createBusy} loadingLabel="Menerbitkan...">
									<PlusIcon class="size-4" />
									Terbitkan
								</LoadingButton>
								<Button type="button" variant="outline" onclick={resetForm}>Reset</Button>
							</div>
						</form>
					</Card.Content>
				</Card.Root>

				<Card.Root>
					<Card.Header>
						<div class="flex flex-col gap-3 xl:flex-row xl:items-end xl:justify-between">
							<div>
								<Card.Title>Arsip Surat</Card.Title>
								<Card.Description>Riwayat surat keterangan siswa yang sudah bernomor.</Card.Description>
							</div>
							<div class="grid gap-2 sm:grid-cols-[minmax(180px,1fr)_150px_180px]">
								<Input aria-label="Cari arsip surat" bind:value={search} placeholder="Cari nomor, siswa, keperluan" />
								<select bind:value={status} aria-label="Filter status" class="rounded-md border border-input bg-background px-3 py-2 text-sm">
									<option value="">Semua status</option>
									<option value="issued">Terbit</option>
									<option value="canceled">Batal</option>
								</select>
								<select bind:value={templateCode} aria-label="Filter jenis surat" class="rounded-md border border-input bg-background px-3 py-2 text-sm">
									<option value="">Semua jenis</option>
									{#each overview.templates as template (template.id)}
										<option value={template.code}>{template.name}</option>
									{/each}
								</select>
							</div>
						</div>
					</Card.Header>
					<Card.Content class="space-y-4">
						<div class="flex flex-wrap gap-2">
							<LoadingButton variant="outline" loading={filterBusy} loadingLabel="Menerapkan..." onclick={() => void applyCertificateFilters()}>
								<RefreshCcwIcon class="size-4" />
								Terapkan Filter
							</LoadingButton>
						</div>
						<div class="overflow-x-auto rounded-lg border">
							<Table.Root>
								<Table.Header>
									<Table.Row>
										<Table.Head>Nomor</Table.Head>
										<Table.Head>Siswa</Table.Head>
										<Table.Head>Jenis</Table.Head>
										<Table.Head>Keperluan</Table.Head>
										<Table.Head>Status</Table.Head>
										<Table.Head class="text-right">Aksi</Table.Head>
									</Table.Row>
								</Table.Header>
								<Table.Body>
									{#if overview.certificates.length === 0}
										<Table.Row>
											<Table.Cell colspan={6} class="h-24 text-center text-muted-foreground">
												Belum ada surat keterangan.
											</Table.Cell>
										</Table.Row>
									{:else}
										{#each overview.certificates as certificate (certificate.id)}
											<Table.Row>
												<Table.Cell class="min-w-52">
													<p class="font-medium text-slate-950">{certificate.nomor_surat}</p>
													<p class="text-xs text-muted-foreground">{formatDate(certificate.tanggal_surat)} · {certificate.classification_code}</p>
												</Table.Cell>
												<Table.Cell class="min-w-48">
													<p class="font-medium">{certificate.student_name}</p>
													<p class="text-xs text-muted-foreground">{certificate.student_nis || '-'} · {certificate.class_name || 'Tanpa kelas'}</p>
												</Table.Cell>
												<Table.Cell class="min-w-44">{certificate.template_name}</Table.Cell>
												<Table.Cell class="min-w-64 text-sm text-muted-foreground">{certificate.purpose}</Table.Cell>
												<Table.Cell>
													<Badge variant={certificate.status === 'issued' ? 'default' : 'secondary'}>{statusLabel(certificate.status)}</Badge>
												</Table.Cell>
												<Table.Cell class="min-w-40 text-right">
													<div class="inline-flex gap-2">
														<Button
															variant="outline"
															size="icon-sm"
															href={`/tu/surat-keterangan/${certificate.id}/print`}
															aria-label={`Cetak ${certificate.nomor_surat}`}
														>
															<PrinterIcon class="size-4" />
														</Button>
														<LoadingButton
															variant="destructive"
															size="icon-sm"
															loading={cancelBusy[certificate.id] === true}
															disabled={certificate.status !== 'issued'}
															aria-label={`Batalkan ${certificate.nomor_surat}`}
															onclick={() => void cancelCertificate(certificate)}
														>
															<XCircleIcon class="size-4" />
														</LoadingButton>
													</div>
												</Table.Cell>
											</Table.Row>
										{/each}
									{/if}
								</Table.Body>
							</Table.Root>
						</div>
					</Card.Content>
				</Card.Root>
			</div>
		{/snippet}
	</AsyncContent>
</div>
