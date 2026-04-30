<script lang="ts">
	let { active = false }: { active?: boolean } = $props();
</script>

<div
	class={`pointer-events-none fixed inset-x-0 top-0 z-[70] h-1 overflow-hidden transition-opacity duration-200 ${active ? 'opacity-100' : 'opacity-0'}`}
	aria-hidden="true"
>
	<div class="h-full w-full bg-emerald-100/70">
		<div
			class={`h-full bg-[oklch(0.38_0.13_145)] shadow-[0_0_18px_rgba(22,101,52,0.35)] ${active ? 'animate-[route-progress_1.25s_ease-in-out_infinite]' : ''}`}
		></div>
	</div>
</div>
