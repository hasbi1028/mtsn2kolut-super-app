<script lang="ts">
	import { page } from '$app/state';
	import { AlertTriangle, Wrench } from 'lucide-svelte';
	import {
		canBypassMaintenance,
		maintenanceModuleLabel,
		routeAffectedByMaintenance,
		type MaintenanceStatus
	} from '$lib/maintenance/modules';
	import type { AuthUser } from '$lib/server/auth';

	type Props = {
		status?: MaintenanceStatus | null;
		user?: AuthUser | null;
	};

	let { status = null, user = null }: Props = $props();

	const windowState = $derived(status?.window);
	const affected = $derived(Boolean(status?.active && routeAffectedByMaintenance(windowState, page.url.pathname)));
	const bypass = $derived(canBypassMaintenance(user, windowState));
	const moduleText = $derived((windowState?.affected_modules ?? []).map(maintenanceModuleLabel).join(', '));
	const toneClass = $derived(windowState?.severity === 'critical'
		? 'border-red-200 bg-red-50 text-red-900'
		: windowState?.severity === 'warning'
			? 'border-amber-200 bg-amber-50 text-amber-900'
			: 'border-emerald-200 bg-emerald-50 text-emerald-900');
</script>

{#if affected && windowState}
	<section class={`mb-4 rounded-lg border px-4 py-3 text-sm ${toneClass}`}>
		<div class="flex gap-3">
			{#if bypass}
				<Wrench class="mt-0.5 h-5 w-5 shrink-0" aria-hidden="true" />
			{:else}
				<AlertTriangle class="mt-0.5 h-5 w-5 shrink-0" aria-hidden="true" />
			{/if}
			<div class="min-w-0 space-y-1">
				<p class="font-semibold">{windowState.title}</p>
				<p>{windowState.message}</p>
				<p class="text-xs">
					Mode {windowState.mode === 'read_only' ? 'baca saja' : windowState.mode}
					{#if moduleText}
						· Modul: {moduleText}
					{/if}
					{#if bypass}
						· Bypass admin aktif untuk sesi Anda.
					{/if}
				</p>
			</div>
		</div>
	</section>
{/if}
