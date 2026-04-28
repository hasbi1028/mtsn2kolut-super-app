<script lang="ts">
	import { onMount } from 'svelte';
	import * as Card from '$lib/components/ui/card';
	import * as Table from '$lib/components/ui/table';
	import { Button } from '$lib/components/ui/button';
	import { Input } from '$lib/components/ui/input';
	import { Badge } from '$lib/components/ui/badge';

	type User = {
		id: string; username: string; role: 'admin' | 'guru';
		employee_id: string | null; employee_nama: string | null;
		created_at: string;
	};
	type Employee = { id: string; nama: string; nip: string; };

	let users = $state<User[]>([]);
	let employees = $state<Employee[]>([]);
	let loading = $state(true);
	let error = $state('');
	let toast = $state('');
	let showForm = $state(false);

	let fUsername = $state('');
	let fPassword = $state('');
	let fRole = $state<'admin' | 'guru'>('guru');
	let fEmpId = $state('');
	let fBusy = $state(false);

	async function load() {
		try {
			const [uRes, eRes] = await Promise.all([
				fetch('/api/users'),
				fetch('/api/employees'),
			]);
			users = await uRes.json();
			const eJson = await eRes.json();
			employees = eJson.data ?? eJson ?? [];
		} catch {
			error = 'Gagal memuat data pengguna';
		} finally {
			loading = false;
		}
	}

	async function createUser() {
		if (!fUsername || !fPassword) return;
		fBusy = true;
		try {
			const res = await fetch('/api/users', {
				method: 'POST',
				headers: { 'Content-Type': 'application/json' },
				body: JSON.stringify({
					username: fUsername, password: fPassword,
					role: fRole, employee_id: fEmpId || null,
				}),
			});
			if (!res.ok) { const j = await res.json(); toast = j.error ?? 'Gagal'; return; }
			fUsername = ''; fPassword = ''; fRole = 'guru'; fEmpId = '';
			showForm = false;
			toast = 'Pengguna berhasil dibuat';
			setTimeout(() => (toast = ''), 3000);
			await load();
		} finally { fBusy = false; }
	}

	async function deleteUser(id: string, name: string) {
		if (name === 'admin') return alert('User admin utama tidak bisa dihapus');
		if (!confirm(`Hapus pengguna "${name}"?`)) return;
		await fetch(`/api/users/${id}`, { method: 'DELETE' });
		await load();
	}

	onMount(load);
</script>

<svelte:head><title>Manajemen Pengguna — MTSN 2 Kolut</title></svelte:head>

<div class="space-y-6 p-6">
	<div class="flex items-center justify-between">
		<div>
			<h1 class="text-2xl font-semibold text-slate-800">Manajemen Pengguna</h1>
			<p class="text-sm text-slate-500 mt-1">Kelola akun akses sistem untuk admin dan guru</p>
		</div>
		<Button onclick={() => (showForm = !showForm)}>
			{showForm ? 'Batal' : '+ Tambah Pengguna'}
		</Button>
	</div>

	{#if toast}
		<div class="rounded-md bg-emerald-50 border border-emerald-200 p-3 text-sm text-emerald-800">{toast}</div>
	{/if}

	{#if showForm}
		<Card.Root>
			<Card.Header class="pb-2"><Card.Title class="text-base">Tambah Akun Baru</Card.Title></Card.Header>
			<Card.Content class="space-y-4">
				<div class="grid gap-3 sm:grid-cols-2">
					<div>
						<label for="u-name" class="text-xs text-slate-500 mb-1 block">Username</label>
						<Input id="u-name" bind:value={fUsername} placeholder="nama_pengguna" />
					</div>
					<div>
						<label for="u-pass" class="text-xs text-slate-500 mb-1 block">Password</label>
						<Input id="u-pass" type="password" bind:value={fPassword} placeholder="********" />
					</div>
					<div>
						<label for="u-role" class="text-xs text-slate-500 mb-1 block">Role</label>
						<select id="u-role" class="w-full rounded-md border border-input bg-background px-3 py-2 text-sm" bind:value={fRole}>
							<option value="guru">Guru</option>
							<option value="admin">Administrator</option>
						</select>
					</div>
					<div>
						<label for="u-emp" class="text-xs text-slate-500 mb-1 block">Hubungkan ke Pegawai (khusus Guru)</label>
						<select id="u-emp" class="w-full rounded-md border border-input bg-background px-3 py-2 text-sm" bind:value={fEmpId}>
							<option value="">-- Tidak Terhubung --</option>
							{#each employees as e}
								<option value={e.id}>{e.nama} ({e.nip})</option>
							{/each}
						</select>
					</div>
				</div>
				<div class="flex gap-2">
					<Button onclick={createUser} disabled={fBusy || !fUsername || !fPassword}>
						{fBusy ? 'Menyimpan...' : 'Simpan Pengguna'}
					</Button>
					<Button variant="outline" onclick={() => (showForm = false)}>Batal</Button>
				</div>
			</Card.Content>
		</Card.Root>
	{/if}

	<Card.Root>
		<Card.Content class="p-0 overflow-x-auto">
			<Table.Root>
				<Table.Header>
					<Table.Row class="bg-slate-50">
						<Table.Head>Username</Table.Head>
						<Table.Head>Role</Table.Head>
						<Table.Head>Nama Pegawai</Table.Head>
						<Table.Head>Dibuat</Table.Head>
						<Table.Head></Table.Head>
					</Table.Row>
				</Table.Header>
				<Table.Body>
					{#each users as u}
						<Table.Row>
							<Table.Cell class="font-medium">{u.username}</Table.Cell>
							<Table.Cell>
								<Badge variant={u.role === 'admin' ? 'default' : 'secondary'} class="capitalize">{u.role}</Badge>
							</Table.Cell>
							<Table.Cell class="text-sm text-slate-600">{u.employee_nama || '—'}</Table.Cell>
							<Table.Cell class="text-xs text-slate-400">{new Date(u.created_at).toLocaleDateString()}</Table.Cell>
							<Table.Cell class="text-right">
								<Button variant="ghost" size="sm" onclick={() => deleteUser(u.id, u.username)}
									class="text-red-600 hover:text-red-700 hover:bg-red-50">Hapus</Button>
							</Table.Cell>
						</Table.Row>
					{/each}
				</Table.Body>
			</Table.Root>
		</Card.Content>
	</Card.Root>
</div>
