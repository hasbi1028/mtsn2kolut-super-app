<script lang="ts">
	import { resolve } from '$app/paths';
	import BankSoalHealthDashboard from '../_components/BankSoalHealthDashboard.svelte';

	type PageData = {
		user?: {
			role?: string;
			roles?: string[];
			permissions?: string[];
		};
	};

	let { data }: { data: PageData } = $props();

	type ToolHref =
		| '/bank-soal/cetak'
		| '/bank-soal/penerbitan'
		| '/bank-soal/mapel-kd'
		| '/bank-soal/analisis-butir'
		| '/bank-soal/laporan'
		| '/bank-soal/pengaturan';

	type ToolLink = {
		href: ToolHref;
		label: string;
		description: string;
		permissions: string[];
	};

	const tools: ToolLink[] = [
		{ href: '/bank-soal/cetak', label: 'Cetak Soal', description: 'Format cetak dan kebutuhan arsip.', permissions: ['bank_soal.read'] },
		{ href: '/bank-soal/penerbitan', label: 'Penerbitan', description: 'Publikasi soal yang sudah disahkan.', permissions: ['bank_soal.publish'] },
		{ href: '/bank-soal/mapel-kd', label: 'Mapel & KD', description: 'Referensi kurikulum untuk penulis soal.', permissions: ['bank_soal.read', 'bank_soal.create'] },
		{ href: '/bank-soal/analisis-butir', label: 'Analisis Butir', description: 'Pantau mutu dan revisi berbasis pemakaian.', permissions: ['bank_soal.analytics'] },
		{ href: '/bank-soal/laporan', label: 'Laporan', description: 'Rekap kualitas dan distribusi Bank Soal.', permissions: ['bank_soal.read', 'bank_soal.analytics', 'bank_soal.review'] },
		{ href: '/bank-soal/pengaturan', label: 'Pengaturan', description: 'Standar, reviewer, dan konfigurasi admin.', permissions: ['bank_soal.settings'] },
	];

	function hasToolAccess(tool: ToolLink): boolean {
		const roles = data.user?.roles ?? (data.user?.role ? [data.user.role] : []);
		if (roles.includes('admin')) return true;
		const permissions = data.user?.permissions ?? [];
		return tool.permissions.some((permission) => permissions.includes(permission));
	}

	let visibleTools = $derived(tools.filter(hasToolAccess));
</script>

<div class="space-y-5">
	<section class="rounded-lg border bg-card p-4 shadow-sm">
		<div class="mb-3">
			<p class="text-sm font-semibold text-foreground">Alat teknis</p>
			<p class="text-sm text-muted-foreground">Fitur lanjutan dikumpulkan di sini agar sidebar tetap ringkas.</p>
		</div>
		<div class="grid gap-3 md:grid-cols-2 xl:grid-cols-3">
			{#each visibleTools as tool (tool.href)}
				<a
					class="rounded-lg border bg-background p-3 text-sm transition hover:border-primary/40 hover:bg-primary/5 focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring"
					href={resolve(tool.href)}
				>
					<span class="block font-medium text-foreground">{tool.label}</span>
					<span class="mt-1 block text-muted-foreground">{tool.description}</span>
				</a>
			{/each}
		</div>
	</section>

	<BankSoalHealthDashboard {data} />
</div>
