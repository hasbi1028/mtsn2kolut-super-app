<script lang="ts">
	import { onMount } from 'svelte';
	import { page } from '$app/state';
	import { resolve } from '$app/paths';
	import * as Table from '$lib/components/ui/table';
	import { Button } from '$lib/components/ui/button';
	import { Input } from '$lib/components/ui/input';
	import { Skeleton } from '$lib/components/ui/skeleton';
	import AsyncContent from '$lib/components/AsyncContent.svelte';
	import LoadingButton from '$lib/components/LoadingButton.svelte';
	import RecoveryPanel from '$lib/components/RecoveryPanel.svelte';
	import { toast } from '$lib/components/ui/sonner';
	import { confirmAction } from '$lib/confirm-dialog';
	import { clientApiPath, readClientApiData, readClientJson } from '$lib/client/api';
	import { displayName } from '$lib/utils/display-name';

	type Subject = { id: string; name: string; code?: string };
	type CbtEvent = { id: string; title: string; status?: string; exam_type?: string; academic_year_name?: string };
	type UserOption = { id: string; username: string; roles?: string[]; employee_id?: string | null; profile_nama?: string | null };
	type EventMemberRole = 'panitia' | 'pembuat_soal' | 'reviewer' | 'proktor' | 'pengawas' | 'korektor';
	type SubjectScopedRole = 'pembuat_soal' | 'reviewer' | 'korektor';
	type EventMember = {
		id: string;
		user_id: string;
		employee_id?: string | null;
		subject_id?: string | null;
		role: EventMemberRole;
		username?: string;
		employee_nama?: string;
		employee_name?: string;
		subject_name?: string;
		subject_code?: string;
	};
	type AcademicPayload = { subjects?: Subject[] };
	type UsersPayload = UserOption[] | { items?: UserOption[]; users?: UserOption[] };
	type PagePayload = { info: CbtEvent; subjects: Subject[]; users: UserOption[]; members: EventMember[] };

	const eventId = page.params.id ?? '';
	const roles: Array<{ value: EventMemberRole; label: string; desc: string }> = [
		{ value: 'panitia', label: 'Panitia', desc: 'Koordinasi kegiatan' },
		{ value: 'pembuat_soal', label: 'Pembuat Soal', desc: 'Menyusun bank soal' },
		{ value: 'reviewer', label: 'Pemeriksa Soal', desc: 'Menelaah mutu soal' },
		{ value: 'proktor', label: 'Pengawas Teknis', desc: 'Membantu kelancaran sesi ujian' },
		{ value: 'pengawas', label: 'Pengawas', desc: 'Pengawasan ruang' },
		{ value: 'korektor', label: 'Korektor', desc: 'Koreksi uraian' }
	];
	const subjectScopedRoles: SubjectScopedRole[] = ['pembuat_soal', 'reviewer', 'korektor'];

	let info = $state<CbtEvent | null>(null);
	let subjects = $state<Subject[]>([]);
	let users = $state<UserOption[]>([]);
	let members = $state<EventMember[]>([]);
	let payloadPromise = $state<Promise<PagePayload> | null>(null);
	let formUserId = $state('');
	let formRole = $state<EventMemberRole>('pembuat_soal');
	let formSubjectId = $state('');
	let editingId = $state('');
	let busy = $state(false);
	let deletingId = $state('');
	let search = $state('');

	let filteredMembers = $derived.by(() => {
		const q = search.trim().toLowerCase();
		if (!q) return members;
		return members.filter((member) => [memberDisplayName(member), roleLabel(member.role), member.subject_name, member.subject_code]
			.filter(Boolean)
			.some((value) => String(value).toLowerCase().includes(q)));
	});
	let subjectSelectorAvailable = $derived(isSubjectScopedRole(formRole));
	let subjectSelectorHelper = $derived(subjectSelectorAvailable
		? 'Mapel dipakai untuk membatasi kerja pembuat soal, reviewer, atau korektor.'
		: 'Peran panitia, proktor, dan pengawas bersifat operasional, jadi mapel tidak dikirim untuk penugasan ini.');

	function isSubjectScopedRole(role: EventMemberRole): role is SubjectScopedRole {
		return subjectScopedRoles.includes(role as SubjectScopedRole);
	}

	function normalizeUsers(payload: UsersPayload): UserOption[] {
		if (Array.isArray(payload)) return payload;
		return payload.items ?? payload.users ?? [];
	}

	function userDisplayName(user: UserOption): string {
		return displayName({ display_name: user.profile_nama, username: user.username, id: user.id }, 'Pengguna');
	}

	function memberDisplayName(member: EventMember): string {
		return displayName({ display_name: member.employee_nama ?? member.employee_name, username: member.username, id: member.user_id }, 'Pengguna');
	}

	function memberSecondaryLabel(member: EventMember): string {
		const primary = memberDisplayName(member);
		const username = (member.username ?? '').trim();
		return username && username !== primary ? username : '';
	}

	function roleLabel(role: string): string {
		return roles.find((item) => item.value === role)?.label ?? role;
	}

	function subjectLabel(member: EventMember): string {
		if (!isSubjectScopedRole(member.role)) return 'Tidak berlaku';
		return member.subject_name ?? member.subject_code ?? 'Semua mapel';
	}

	function handleRoleChange(event: Event) {
		const target = event.currentTarget as HTMLSelectElement;
		formRole = target.value as EventMemberRole;
		if (!isSubjectScopedRole(formRole)) formSubjectId = '';
	}

	function errorMessage(error: unknown, fallback: string) {
		return error instanceof Error && error.message.trim() ? error.message : fallback;
	}

	async function fetchPayload(): Promise<PagePayload> {
		const [nextInfo, academic, nextUsers, nextMembers] = await Promise.all([
			fetch(clientApiPath`/api/asesmen/events/${eventId}`).then((response) => readClientApiData<CbtEvent>(response, 'Gagal memuat kegiatan ujian')),
			fetch('/api/bank-soal/soal-support/subjects').then((response) => readClientApiData<AcademicPayload>(response, 'Gagal memuat mapel')),
			fetch('/api/users').then((response) => readClientApiData<UsersPayload>(response, 'Gagal memuat pengguna')),
			fetch(clientApiPath`/api/asesmen/events/${eventId}/members`).then((response) => readClientApiData<EventMember[]>(response, 'Gagal memuat penugasan'))
		]);
		return { info: nextInfo, subjects: academic.subjects ?? [], users: normalizeUsers(nextUsers), members: Array.isArray(nextMembers) ? nextMembers : [] };
	}

	function applyPayload(payload: PagePayload) {
		info = payload.info;
		subjects = payload.subjects;
		users = payload.users;
		members = payload.members;
	}

	function load() {
		payloadPromise = fetchPayload().then((payload) => {
			applyPayload(payload);
			return payload;
		});
	}

	function resetForm() {
		editingId = '';
		formUserId = '';
		formRole = 'pembuat_soal';
		formSubjectId = '';
	}

	function editMember(member: EventMember) {
		editingId = member.id;
		formUserId = member.user_id;
		formRole = member.role;
		formSubjectId = isSubjectScopedRole(member.role) ? (member.subject_id ?? '') : '';
	}

	async function saveMember() {
		if (!formUserId) {
			toast.warning('Pilih pengguna terlebih dahulu.');
			return;
		}
		busy = true;
		try {
			const selectedUser = users.find((user) => user.id === formUserId);
			const payload = {
				user_id: formUserId,
				employee_id: selectedUser?.employee_id || undefined,
				role: formRole,
				subject_id: subjectSelectorAvailable && formSubjectId ? formSubjectId : undefined
			};
			const url = editingId ? clientApiPath`/api/asesmen/events/${eventId}/members/${editingId}` : clientApiPath`/api/asesmen/events/${eventId}/members`;
			await fetch(url, {
				method: editingId ? 'PUT' : 'POST',
				headers: { 'Content-Type': 'application/json' },
				body: JSON.stringify(payload)
			}).then((response) => readClientJson<unknown>(response));
			toast.success(editingId ? 'Penugasan diperbarui' : 'Penugasan ditambahkan');
			resetForm();
			load();
		} catch (error) {
			toast.error(errorMessage(error, 'Gagal menyimpan penugasan'));
		} finally {
			busy = false;
		}
	}

	async function deleteMember(member: EventMember) {
		if (!(await confirmAction({ title: 'Hapus Penugasan', message: `Hapus ${memberDisplayName(member)} dari kegiatan ini?`, confirmLabel: 'Hapus', tone: 'danger' }))) return;
		deletingId = member.id;
		try {
			await fetch(clientApiPath`/api/asesmen/events/${eventId}/members/${member.id}`, { method: 'DELETE' }).then((response) => readClientJson<unknown>(response));
			toast.success('Penugasan dihapus');
			if (editingId === member.id) resetForm();
			load();
		} catch (error) {
			toast.error(errorMessage(error, 'Gagal menghapus penugasan'));
		} finally {
			deletingId = '';
		}
	}

	onMount(load);
</script>

<svelte:head><title>Penugasan CBT — {info?.title ?? 'Kegiatan'}</title></svelte:head>

<div class="space-y-6 p-6">
	<div class="flex items-center gap-2 text-sm text-muted-foreground">
		<a href={resolve('/asesmen/kegiatan')} class="hover:text-foreground">Kegiatan Ujian</a>
		<span>/</span>
		<a href={resolve(`/asesmen/kegiatan/${eventId}`)} class="hover:text-foreground">{info?.title ?? 'Detail'}</a>
		<span>/</span>
		<span class="font-medium text-foreground">Penugasan</span>
	</div>

	<AsyncContent promise={payloadPromise}>
		{#snippet pending()}
			<div class="space-y-4">
				<Skeleton class="h-8 w-72" />
				<Skeleton class="h-32 w-full" />
				<Skeleton class="h-64 w-full" />
			</div>
		{/snippet}
		{#snippet failed(error, reset)}
			<RecoveryPanel title="Penugasan Belum Tersaji" message={errorMessage(error, 'Gagal memuat penugasan kegiatan.')} onRetry={() => { reset?.(); load(); }} />
		{/snippet}
		{#snippet children(value)}
			{@const current = value as PagePayload}
			<div class="flex flex-col gap-3 lg:flex-row lg:items-start lg:justify-between">
				<div>
					<p class="text-xs font-bold uppercase tracking-[0.18em] text-success">Panitia & Penugasan Soal</p>
					<h1 class="mt-1 text-2xl font-semibold text-foreground">{current.info.title}</h1>
					<p class="mt-1 text-sm text-muted-foreground">Kelola pembuat soal, reviewer, proktor, pengawas, dan korektor untuk kegiatan CBT ini.</p>
				</div>
				<a href={resolve(`/bank-soal/tambah?event_id=${eventId}`)} class="inline-flex rounded-md border border-success/20 bg-success/10 px-3 py-2 text-sm font-semibold text-success hover:bg-success/15">Buka Komposer Bank Soal</a>
			</div>

			<section class="rounded-xl border border-success/20 bg-card p-4 shadow-sm">
				<div class="grid gap-3 lg:grid-cols-[minmax(0,1.4fr)_11rem_14rem_auto] lg:items-end">
					<div>
						<label for="member-user" class="mb-1 block text-xs font-medium text-muted-foreground">Pengguna</label>
						<select id="member-user" bind:value={formUserId} class="h-9 w-full rounded-md border border-border bg-card px-2.5 text-sm text-foreground focus:outline-none focus:ring-2 focus:ring-ring">
							<option value="">-- Pilih pengguna --</option>
							{#each users as user (user.id)}
								<option value={user.id}>{userDisplayName(user)}</option>
							{/each}
						</select>
					</div>
					<div>
						<label for="member-role" class="mb-1 block text-xs font-medium text-muted-foreground">Peran</label>
						<select id="member-role" value={formRole} onchange={handleRoleChange} class="h-9 w-full rounded-md border border-border bg-card px-2.5 text-sm text-foreground focus:outline-none focus:ring-2 focus:ring-ring">
							{#each roles as role (role.value)}
								<option value={role.value}>{role.label}</option>
							{/each}
						</select>
					</div>
					<div>
						<label for="member-subject" class="mb-1 block text-xs font-medium text-muted-foreground">Mapel Penugasan</label>
						{#if subjectSelectorAvailable}
							<select id="member-subject" bind:value={formSubjectId} class="h-9 w-full rounded-md border border-border bg-card px-2.5 text-sm text-foreground focus:outline-none focus:ring-2 focus:ring-ring">
								<option value="">Semua mapel</option>
								{#each subjects as subject (subject.id)}
									<option value={subject.id}>{subject.name}</option>
								{/each}
							</select>
						{:else}
							<select id="member-subject" value="" disabled class="h-9 w-full rounded-md border border-border bg-muted/50 px-2.5 text-sm text-muted-foreground">
								<option value="">Tidak berlaku untuk peran ini</option>
							</select>
						{/if}
						<p class="mt-1 text-[11px] text-muted-foreground">{subjectSelectorHelper}</p>
					</div>
					<div class="flex gap-2">
						{#if editingId}<Button variant="outline" onclick={resetForm} disabled={busy}>Batal</Button>{/if}
						<LoadingButton onclick={() => void saveMember()} loading={busy} loadingLabel="Menyimpan..." disabled={busy || !formUserId} class="bg-success text-background hover:bg-success disabled:opacity-50">{editingId ? 'Simpan' : 'Tambah'}</LoadingButton>
					</div>
				</div>
			</section>

			<section class="rounded-xl border border-border bg-card shadow-sm">
				<div class="flex flex-col gap-3 border-b border-border p-4 md:flex-row md:items-center md:justify-between">
					<div>
						<h2 class="text-base font-semibold text-foreground">Daftar Penugasan</h2>
						<p class="text-sm text-muted-foreground">{members.length} orang terhubung dengan kegiatan ini.</p>
					</div>
					<Input placeholder="Cari nama/peran/mapel..." bind:value={search} class="h-9 md:w-72" />
				</div>
				<div class="overflow-x-auto">
					<Table.Root>
						<Table.Header>
							<Table.Row class="bg-muted/50">
								<Table.Head>Nama</Table.Head>
								<Table.Head>Peran</Table.Head>
								<Table.Head>Mapel</Table.Head>
								<Table.Head class="text-right">Aksi</Table.Head>
							</Table.Row>
						</Table.Header>
						<Table.Body>
							{#each filteredMembers as member (member.id)}
								<Table.Row>
									<Table.Cell class="font-medium text-foreground">
										{memberDisplayName(member)}
										{@const secondary = memberSecondaryLabel(member)}
										{#if secondary}<div class="text-xs font-normal text-muted-foreground">{secondary}</div>{/if}
									</Table.Cell>
									<Table.Cell><span class="rounded bg-success/10 px-2 py-1 text-xs font-semibold text-success">{roleLabel(member.role)}</span></Table.Cell>
									<Table.Cell class="text-sm text-muted-foreground">{subjectLabel(member)}</Table.Cell>
									<Table.Cell class="text-right">
										<Button variant="outline" size="sm" class="mr-2 h-8" onclick={() => editMember(member)}>Edit</Button>
										<LoadingButton variant="outline" size="sm" class="h-8 border-destructive/30 text-destructive hover:bg-destructive/10" onclick={() => void deleteMember(member)} loading={deletingId === member.id} loadingLabel="Hapus..." disabled={deletingId !== ''}>Hapus</LoadingButton>
									</Table.Cell>
								</Table.Row>
							{:else}
								<Table.Row><Table.Cell colspan={4} class="py-10 text-center text-sm text-muted-foreground">Belum ada penugasan.</Table.Cell></Table.Row>
							{/each}
						</Table.Body>
					</Table.Root>
				</div>
			</section>
		{/snippet}
	</AsyncContent>
</div>
