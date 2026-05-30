<script lang="ts">
	type Action = { label: string; href: string; tone?: string };
	type Feature = { title: string; detail: string; status: string; actions?: Action[] };
	type Step = { label: string; detail: string; badge: string };

	let { title, eyebrow, description, actions = [], stats = [], steps = [], features = [], tableTitle = 'Fitur operasional', tableDescription = 'Fitur di bawah mengikuti pola kerja CBT lama tetapi tetap memakai data/akses Super App.' } = $props<{
		title: string;
		eyebrow: string;
		description: string;
		actions?: Action[];
		stats?: Array<{ label: string; value: string; detail: string }>;
		steps?: Step[];
		features?: Feature[];
		tableTitle?: string;
		tableDescription?: string;
	}>();

	function actionClass(tone: Action['tone'] = 'soft') {
		if (tone === 'primary') return 'bg-emerald-700 text-white hover:bg-emerald-800 border-emerald-700';
		if (tone === 'danger') return 'bg-amber-50 text-amber-900 hover:bg-amber-100 border-amber-300';
		return 'bg-white text-slate-700 hover:border-emerald-300 hover:text-emerald-800 border-slate-300';
	}
</script>

<div class="space-y-4 pb-10">
	<header class="rounded-3xl border border-slate-200 bg-white p-5 shadow-sm">
		<p class="text-xs font-black uppercase tracking-[0.2em] text-emerald-700">{eyebrow}</p>
		<div class="mt-2 flex flex-col gap-3 lg:flex-row lg:items-end lg:justify-between">
			<div>
				<h1 class="text-3xl font-black tracking-tight md:text-4xl">{title}</h1>
				<p class="mt-2 max-w-3xl text-sm leading-6 text-slate-600">{description}</p>
			</div>
			<div class="flex flex-wrap gap-2">
				{#each actions as action (action.href + action.label)}
					<a href={action.href} class={`rounded-2xl border px-4 py-2 text-sm font-black shadow-sm ${actionClass(action.tone)}`}>{action.label}</a>
				{/each}
			</div>
		</div>
	</header>

	{#if stats.length}
		<section class="grid gap-3 sm:grid-cols-2 xl:grid-cols-4">
			{#each stats as stat (stat.label)}
				<div class="rounded-2xl border border-slate-200 bg-white p-4 shadow-sm">
					<p class="text-[10px] font-black uppercase tracking-[0.16em] text-slate-400">{stat.label}</p>
					<p class="mt-2 text-3xl font-black text-slate-950">{stat.value}</p>
					<p class="mt-1 text-xs font-semibold text-slate-500">{stat.detail}</p>
				</div>
			{/each}
		</section>
	{/if}

	{#if steps.length}
		<section class="rounded-3xl border border-slate-200 bg-white p-4 shadow-sm">
			<div class="mb-3 flex items-center justify-between gap-3">
				<div>
					<h2 class="text-lg font-black text-slate-950">Alur kerja</h2>
					<p class="text-sm font-semibold text-slate-500">Urutan sederhana untuk operator/panitia.</p>
				</div>
				<span class="rounded-full bg-emerald-50 px-3 py-1 text-xs font-black text-emerald-800">CBT Familiar</span>
			</div>
			<div class="grid gap-3 lg:grid-cols-4">
				{#each steps as step, index (step.label)}
					<div class="rounded-2xl border border-slate-200 bg-slate-50 p-4">
						<div class="flex items-center justify-between gap-2">
							<span class="flex h-8 w-8 items-center justify-center rounded-full bg-emerald-700 text-sm font-black text-white">{index + 1}</span>
							<span class="rounded-full bg-white px-2 py-1 text-[11px] font-black text-slate-600">{step.badge}</span>
						</div>
						<h3 class="mt-3 font-black text-slate-950">{step.label}</h3>
						<p class="mt-1 text-sm leading-6 text-slate-600">{step.detail}</p>
					</div>
				{/each}
			</div>
		</section>
	{/if}

	<section class="rounded-3xl border border-slate-200 bg-white shadow-sm">
		<div class="border-b border-slate-200 p-4">
			<h2 class="text-lg font-black text-slate-950">{tableTitle}</h2>
			<p class="text-sm font-semibold text-slate-500">{tableDescription}</p>
		</div>
		<div class="divide-y divide-slate-200">
			{#each features as feature (feature.title)}
				<div class="grid gap-3 p-4 text-sm lg:grid-cols-[1fr_auto] lg:items-center">
					<div class="min-w-0">
						<div class="flex flex-wrap items-center gap-2">
							<p class="font-black text-slate-950">{feature.title}</p>
							<span class="rounded-full bg-emerald-50 px-2 py-0.5 text-[11px] font-black uppercase text-emerald-800">{feature.status}</span>
						</div>
						<p class="mt-1 leading-6 text-slate-600">{feature.detail}</p>
					</div>
					{#if feature.actions?.length}
						<div class="flex flex-wrap gap-2 lg:justify-end">
							{#each feature.actions as action (action.href + action.label)}
								<a href={action.href} class={`rounded-xl border px-3 py-2 text-xs font-black shadow-sm ${actionClass(action.tone)}`}>{action.label}</a>
							{/each}
						</div>
					{/if}
				</div>
			{/each}
		</div>
	</section>
</div>
