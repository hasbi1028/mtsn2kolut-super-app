<script lang="ts">
	import { onMount } from 'svelte';
	import * as Card from '$lib/components/ui/card';
	import * as Table from '$lib/components/ui/table';
	import { Button } from '$lib/components/ui/button';
	import { Input } from '$lib/components/ui/input';
	import { Badge } from '$lib/components/ui/badge';
	import { toast } from '$lib/components/ui/sonner';
	import { Skeleton } from '$lib/components/ui/skeleton';
	import LoadingButton from '$lib/components/LoadingButton.svelte';
	import EmptyStatePanel from '$lib/components/EmptyStatePanel.svelte';
	import RecoveryPanel from '$lib/components/RecoveryPanel.svelte';

	type User = {
		id: string; username: string; roles: string[];
		employee_id: string | null; student_id: string | null; parent_id: string | null;
		profile_nama: string | null;
		is_active: boolean;
		created_at: string;
	};
	type Employee = { id: string; nama: string; nip: string; };
	type Student = { id: string; nama: string; nis: string; };
	type Parent = { id: string; nama: string; phone: string; };

	let users = $state<User[]>([]);
	let employees = $state<Employee[]>([]);
	let students = $state<Student[]>([]);
	let parents = $state<Parent[]>([]);
	let loading = $state(true);
	let error = $state('');
	let showForm = $state(false);

	let fUsername = $state('');
	let fPassword = $state('');
	let fRoles = $state<string[]>(['guru']);
	let fEmpId = $state('');
	let fStuId = $state('');
	let fParId = $state('');
	let fBusy = $state(false);
	let formHint = $derived.by(() => {
		if (fRoles.includes('siswa')) return 'Akun siswa wajib ditautkan ke satu profil siswa.';
		if (fRoles.includes('ortu')) return 'Akun orang tua wajib ditautkan ke satu profil orang tua.';
		if (fRoles.includes('guru') || fRoles.includes('staf')) return 'Akun guru/staf wajib ditautkan ke satu profil pegawai.';
		return 'Akun admin murni boleh tanpa tautan profil.';
	});

	const availableRoles = [
		{ value: 'admin', label: 'Administrator' },
		{ value: 'guru', label: 'Guru' },
		{ value: 'staf', label: 'Staf' },
		{ value: 'siswa', label: 'Siswa' },
		{ value: 'ortu', label: 'Orang Tua' },
	];

	async function load() {
		try {
			const [uRes, eRes, sRes, pRes] = await Promise.all([
				fetch('/api/users'),
				fetch('/api/employees'),
				fetch('/api/students'),
				fetch('/api/parents'),
			]);
			users = await uRes.json();
			const eJson = await eRes.json();
			employees = eJson.data ?? eJson ?? [];
			students = await sRes.json();
			parents = await pRes.json();
		} catch {
			error = 'Gagal memuat data';
		} finally {
			loading = false;
		}
	}

	async function createUser() {
		if (!fUsername || !fPassword || fRoles.length === 0) return;
		fBusy = true;
		try {
			const res = await fetch('/api/users', {
				method: 'POST',
				headers: { 'Content-Type': 'application/json' },
				body: JSON.stringify({
					username: fUsername, password: fPassword,
					roles: fRoles, 
					employee_id: fEmpId || null,
					student_id: fStuId || null,
					parent_id: fParId || null,
				}),
			});
			if (!res.ok) { const j = await res.json(); toast.error(j.error ?? 'Gagal'); return; }
			fUsername = ''; fPassword = ''; fRoles = ['guru']; fEmpId = ''; fStuId = ''; fParId = '';
			showForm = false;
			toast.success('Pengguna berhasil dibuat');
			await load();
		} finally { fBusy = false; }
	}

	async function deleteUser(id: string, name: string) {
		if (name === 'admin') return alert('User admin utama tidak bisa dihapus');
		if (!confirm(`Hapus pengguna "${name}"?`)) return;
		await fetch(`/api/users/${id}`, { method: 'DELETE' });
		await load();
	}

	async function toggleUserStatus(user: User) {
		if (user.username === 'admin' && user.is_active) {
			toast.error('Akun admin utama tidak bisa dinonaktifkan');
			return;
		}
		const next = !user.is_active;
		const actionLabel = next ? 'mengaktifkan' : 'menonaktifkan';
		if (!confirm(`${actionLabel} akun "${user.username}"?`)) return;
		const res = await fetch(`/api/users/${user.id}/status`, {
			method: 'PATCH',
			headers: { 'Content-Type': 'application/json' },
			body: JSON.stringify({ is_active: next }),
		});
		if (!res.ok) {
			const payload = await res.json().catch(() => ({}));
			toast.error(payload.error ?? 'Gagal memperbarui status akun');
			return;
		}
		toast.success(next ? 'Akun diaktifkan' : 'Akun dinonaktifkan');
		await load();
	}

	function toggleRole(role: string) {
		if (fRoles.includes(role)) {
			fRoles = fRoles.filter(r => r !== role);
		} else {
			fRoles = [...fRoles, role];
		}
	}

	onMount(load);
</script>

<svelte:head><title>Manajemen Pengguna — MTSN 2 Kolut</title></svelte:head>

<div class="space-y-6 p-6">
	<div class="flex flex-wrap items-start justify-between gap-4">
		<div>
			<h1 class="text-2xl font-semibold text-slate-800">Manajemen Pengguna</h1>
			<p class="text-sm text-slate-500 mt-1">Kelola akun akses sistem dengan RBAC terpadu</p>
		</div>
		<Button onclick={() => (showForm = !showForm)}>
			{showForm ? 'Batal' : '+ Tambah Pengguna'}
		</Button>
	</div>

	<div class="grid gap-3 md:grid-cols-4">
		<div class="rounded-2xl border border-emerald-100 bg-emerald-50 px-4 py-4">
			<p class="text-[11px] font-semibold uppercase tracking-[0.22em] text-emerald-700">Total Akun</p>
			<p class="mt-2 text-2xl font-semibold text-slate-900">{users.length}</p>
			<p class="text-sm text-slate-600">akun yang sudah dapat masuk ke sistem</p>
		</div>
		<div class="rounded-2xl border border-sky-100 bg-sky-50 px-4 py-4">
			<p class="text-[11px] font-semibold uppercase tracking-[0.22em] text-sky-700">Akun Aktif</p>
			<p class="mt-2 text-2xl font-semibold text-slate-900">{users.filter((item) => item.is_active).length}</p>
			<p class="text-sm text-slate-600">akun yang saat ini masih aktif digunakan</p>
		</div>
		<div class="rounded-2xl border border-amber-100 bg-amber-50 px-4 py-4">
			<p class="text-[11px] font-semibold uppercase tracking-[0.22em] text-amber-700">Multi-Role</p>
			<p class="mt-2 text-2xl font-semibold text-slate-900">{users.filter((item) => (item.roles ?? []).length > 1).length}</p>
			<p class="text-sm text-slate-600">akun yang memegang lebih dari satu role</p>
		</div>
		<div class="rounded-2xl border border-violet-100 bg-violet-50 px-4 py-4">
			<p class="text-[11px] font-semibold uppercase tracking-[0.22em] text-violet-700">Terhubung Profil</p>
			<p class="mt-2 text-2xl font-semibold text-slate-900">{users.filter((item) => item.profile_nama).length}</p>
			<p class="text-sm text-slate-600">akun yang sudah terkait dengan entitas sekolah</p>
		</div>
	</div>

	{#if error}
		<RecoveryPanel title="Data Pengguna Belum Tersaji" message={error} onRetry={load} />
	{/if}

	{#if showForm}
		<Card.Root>
			<Card.Header class="pb-2"><Card.Title class="text-base">Tambah Akun Baru</Card.Title></Card.Header>
			<Card.Content class="space-y-4">
				<div class="grid gap-4 sm:grid-cols-2">
					<div class="space-y-3">
						<div>
							<label for="u-name" class="text-xs text-slate-500 mb-1 block">Username</label>
							<Input id="u-name" bind:value={fUsername} placeholder="Gunakan nama akun yang mudah dikenali" />
						</div>
						<div>
							<label for="u-pass" class="text-xs text-slate-500 mb-1 block">Password</label>
							<Input id="u-pass" type="password" bind:value={fPassword} placeholder="Minimal 8 karakter" />
						</div>
						<div>
							<p class="mb-2 block text-xs text-slate-500">Peran Akses (boleh pilih lebih dari satu)</p>
							<div class="flex flex-wrap gap-2">
								{#each availableRoles as r (r.value)}
									<button 
										class={`px-3 py-1 text-xs rounded-full border transition-colors ${fRoles.includes(r.value) ? 'bg-green-700 text-white border-green-700' : 'bg-white text-slate-600 border-slate-200'}`}
										onclick={() => toggleRole(r.value)}
									>
										{r.label}
									</button>
								{/each}
							</div>
							<p class="mt-2 text-xs text-slate-500">{formHint}</p>
						</div>
					</div>
					
					<div class="space-y-3">
						{#if fRoles.includes('guru') || fRoles.includes('staf')}
							<div>
								<label for="u-emp" class="text-xs text-slate-500 mb-1 block">Hubungkan ke Pegawai</label>
								<select id="u-emp" class="w-full rounded-md border border-input bg-background px-3 py-2 text-sm" bind:value={fEmpId}>
									<option value="">-- Pilih Pegawai --</option>
									{#each employees as e (e.id)}
										<option value={e.id}>{e.nama} ({e.nip})</option>
									{/each}
								</select>
							</div>
						{/if}

						{#if fRoles.includes('siswa')}
							<div>
								<label for="u-stu" class="text-xs text-slate-500 mb-1 block">Hubungkan ke Siswa</label>
								<select id="u-stu" class="w-full rounded-md border border-input bg-background px-3 py-2 text-sm" bind:value={fStuId}>
									<option value="">-- Pilih Siswa --</option>
									{#each students as s (s.id)}
										<option value={s.id}>{s.nama} ({s.nis})</option>
									{/each}
								</select>
							</div>
						{/if}

						{#if fRoles.includes('ortu')}
							<div>
								<label for="u-par" class="text-xs text-slate-500 mb-1 block">Hubungkan ke Orang Tua</label>
								<select id="u-par" class="w-full rounded-md border border-input bg-background px-3 py-2 text-sm" bind:value={fParId}>
									<option value="">-- Pilih Orang Tua --</option>
									{#each parents as p (p.id)}
										<option value={p.id}>{p.nama} ({p.phone})</option>
									{/each}
								</select>
							</div>
						{/if}
					</div>
				</div>
				<div class="flex gap-2">
					<LoadingButton onclick={createUser} loading={fBusy} disabled={fBusy || !fUsername || !fPassword || fRoles.length === 0}>
						Simpan Pengguna
					</LoadingButton>
					<Button variant="outline" onclick={() => (showForm = false)}>Batal</Button>
				</div>
			</Card.Content>
		</Card.Root>
	{/if}

	<Card.Root class="overflow-hidden border-slate-200 shadow-sm">
		<Card.Content class="p-0">
			{#if loading && users.length === 0}
				<div class="space-y-3 p-6">
					{#each Array.from({ length: 5 }) as _, index (`user-skeleton-${index}`)}
						<div class="grid gap-3 md:grid-cols-[1fr_1.1fr_1fr_0.7fr_0.8fr_auto] md:items-center">
							<Skeleton class="h-5 w-28" />
							<Skeleton class="h-5 w-36" />
							<Skeleton class="h-5 w-40" />
							<Skeleton class="h-6 w-20" />
							<Skeleton class="h-5 w-24" />
							<Skeleton class="h-9 w-40 justify-self-end" />
						</div>
					{/each}
				</div>
			{:else}
				<div class="hidden overflow-x-auto lg:block">
				<Table.Root>
					<Table.Header>
						<Table.Row class="bg-slate-50">
							<Table.Head>Username</Table.Head>
							<Table.Head>Roles</Table.Head>
							<Table.Head>Profil Terhubung</Table.Head>
							<Table.Head>Status</Table.Head>
							<Table.Head>Dibuat</Table.Head>
							<Table.Head></Table.Head>
						</Table.Row>
					</Table.Header>
					<Table.Body>
						{#each users as u (u.id)}
							<Table.Row>
								<Table.Cell class="font-medium">{u.username}</Table.Cell>
								<Table.Cell>
									<div class="flex flex-wrap gap-1">
										{#each u.roles || [] as r (r)}
											<Badge variant={r === 'admin' ? 'default' : 'secondary'} class="text-[10px] uppercase">{r}</Badge>
										{/each}
									</div>
								</Table.Cell>
								<Table.Cell class="text-sm text-slate-600">{u.profile_nama || '—'}</Table.Cell>
								<Table.Cell>
									<Badge variant={u.is_active ? 'outline' : 'destructive'}>
										{u.is_active ? 'Aktif' : 'Nonaktif'}
									</Badge>
								</Table.Cell>
								<Table.Cell class="text-xs text-slate-400">{new Date(u.created_at).toLocaleDateString()}</Table.Cell>
								<Table.Cell class="text-right">
									<div class="flex justify-end gap-2">
										<Button
											variant="outline"
											size="sm"
											onclick={() => toggleUserStatus(u)}
										>
											{u.is_active ? 'Nonaktifkan' : 'Aktifkan'}
										</Button>
										<Button variant="ghost" size="sm" onclick={() => deleteUser(u.id, u.username)}
											class="text-red-600 hover:text-red-700 hover:bg-red-50">Hapus</Button>
									</div>
								</Table.Cell>
							</Table.Row>
						{:else}
							<Table.Row>
								<Table.Cell colspan={6} class="p-4">
									<EmptyStatePanel
										compact
										title="Belum ada data pengguna"
										description="Tambahkan akun pertama untuk mulai menghubungkan pegawai, siswa, atau orang tua ke akses sistem."
									/>
								</Table.Cell>
							</Table.Row>
						{/each}
					</Table.Body>
				</Table.Root>
				</div>

				<div class="grid gap-3 p-4 lg:hidden">
					{#each users as u (u.id)}
						<div class="rounded-2xl border border-slate-200 bg-white p-4 shadow-sm">
							<div class="flex items-start justify-between gap-3">
								<div class="min-w-0">
									<p class="text-sm font-semibold text-slate-900">{u.username}</p>
									<div class="mt-1 flex flex-wrap gap-1">
										{#each u.roles || [] as r (r)}
											<Badge variant={r === 'admin' ? 'default' : 'secondary'} class="text-[10px] capitalize">{r}</Badge>
										{/each}
									</div>
									<p class="mt-2 text-xs text-slate-500">{u.profile_nama || 'Tidak terhubung profil'}</p>
								</div>
								<Badge variant={u.is_active ? 'outline' : 'destructive'}>{u.is_active ? 'Aktif' : 'Nonaktif'}</Badge>
							</div>
							<div class="mt-4 flex gap-2">
								<Button
									variant="outline"
									size="sm"
									class="flex-1 justify-center"
									onclick={() => toggleUserStatus(u)}
								>
									{u.is_active ? 'Nonaktifkan' : 'Aktifkan'}
								</Button>
								<Button variant="ghost" size="sm" onclick={() => deleteUser(u.id, u.username)}
									class="flex-1 justify-center text-red-600 hover:text-red-700 hover:bg-red-50">Hapus</Button>
							</div>
						</div>
					{:else}
						<EmptyStatePanel
							title="Belum ada data pengguna"
							description="Tambahkan akun pertama agar role sekolah dan akses portal bisa mulai dikelola dari panel ini."
						/>
					{/each}
				</div>
			{/if}
		</Card.Content>
	</Card.Root>
</div>
