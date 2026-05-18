<script lang="ts">
	import { onMount } from 'svelte';
	import { clientApiPathWithQuery, readClientApiData } from '$lib/client/api';

	type ReportKey = 'input' | 'progress' | 'revision' | 'reviewer' | 'readiness' | 'honor';
	type ReportRow = {
		no: number;
		primary: string;
		secondary?: string;
		subject_name?: string;
		level_name?: string;
		task?: string;
		total: number;
		pg: number;
		essay: number;
		other: number;
		draft: number;
		submitted: number;
		review: number;
		revision_needed: number;
		approved: number;
		published: number;
		rejected: number;
		shortage?: number;
		volume?: number;
		unit?: string;
		statuses?: string;
		first_input?: string;
		last_input?: string;
		notes?: string;
		system_row?: boolean;
		data_warning?: boolean;
	};
	type ReportResult = {
		title: string;
		period_label: string;
		generated_at: string;
		summary: {
			total: number;
			pg: number;
			essay: number;
			other: number;
			authors: number;
			subjects: number;
			missing_level: number;
			system_rows: number;
			shortage: number;
		};
		rows: ReportRow[];
		access_note?: string;
	};

	const tabs: Array<{ key: ReportKey; label: string; desc: string }> = [
		{ key: 'input', label: 'Input Soal', desc: 'Rekap pembuat, mapel, tingkat, jenis soal, dan status workflow.' },
		{ key: 'progress', label: 'Progres Mapel', desc: 'Kelengkapan target PG/Essay per mapel dan tingkat.' },
		{ key: 'revision', label: 'Perlu Revisi', desc: 'Soal yang perlu ditindaklanjuti pembuat soal.' },
		{ key: 'reviewer', label: 'Reviewer', desc: 'Beban verifikasi, soal menunggu review, dan hasil review.' },
		{ key: 'readiness', label: 'Siap Paket', desc: 'Soal approved/published yang siap dipakai paket.' },
		{ key: 'honor', label: 'Honor/Tugas', desc: 'Volume tugas sebagai bahan lampiran SK/honor.' }
	];

	let activeReport = $state<ReportKey>('input');
	let periodPreset = $state('this_month');
	let startDate = $state('');
	let endDate = $state('');
	let subjectId = $state('');
	let targetLevel = $state('');
	let authorUsername = $state('');
	let workflowStatus = $state('');
	let includeSystem = $state(true);
	let loading = $state(false);
	let exporting = $state('');
	let error = $state('');
	let report = $state<ReportResult | null>(null);

	function params(format?: string) {
		const p = new URLSearchParams();
		p.set('report', activeReport);
		p.set('period_preset', periodPreset);
		if (periodPreset === 'custom') {
			if (startDate) p.set('start_date', startDate);
			if (endDate) p.set('end_date', endDate);
		}
		if (subjectId.trim()) p.set('subject_id', subjectId.trim());
		if (targetLevel) p.set('target_level', targetLevel);
		if (authorUsername.trim()) p.set('author_username', authorUsername.trim());
		if (workflowStatus) p.set('workflow_status', workflowStatus);
		p.set('include_system', includeSystem ? 'true' : 'false');
		if (format) p.set('format', format);
		return p;
	}

	async function loadReport() {
		loading = true;
		error = '';
		try {
			report = await readClientApiData<ReportResult>(await fetch(clientApiPathWithQuery('/api/bank-soal/reports', params())));
		} catch (e) {
			error = e instanceof Error ? e.message : 'Gagal memuat laporan';
		} finally {
			loading = false;
		}
	}

	async function exportReport(format: 'png' | 'csv' | 'html' | 'pdf') {
		exporting = format;
		error = '';
		try {
			const response = await fetch(clientApiPathWithQuery('/api/bank-soal/reports/export', params(format)));
			if (!response.ok) throw new Error(`Export gagal (HTTP ${response.status})`);
			const blob = await response.blob();
			const disposition = response.headers.get('content-disposition') ?? '';
			const match = /filename="?([^";]+)"?/i.exec(disposition);
			const filename = match?.[1] ?? `laporan-bank-soal-${activeReport}.${format === 'csv' ? 'csv' : format}`;
			const url = URL.createObjectURL(blob);
			const a = document.createElement('a');
			a.href = url;
			a.download = filename;
			document.body.appendChild(a);
			a.click();
			a.remove();
			URL.revokeObjectURL(url);
		} catch (e) {
			error = e instanceof Error ? e.message : 'Export gagal';
		} finally {
			exporting = '';
		}
	}

	onMount(() => {
		void loadReport();
	});
</script>

<svelte:head><title>Laporan Bank Soal</title></svelte:head>

<div class="page-shell">
	<section class="hero">
		<div>
			<p class="eyebrow">Bank Soal</p>
			<h1>Report Center</h1>
			<p class="muted">Laporan input, progres mapel, revisi, reviewer, kesiapan paket, dan volume tugas/honor tanpa bantuan AI.</p>
		</div>
		<div class="actions">
			<button class="secondary" onclick={() => exportReport('csv')} disabled={loading || !!exporting}>{exporting === 'csv' ? 'Menyiapkan…' : 'Download Excel/CSV'}</button>
			<button class="secondary" onclick={() => exportReport('html')} disabled={loading || !!exporting}>Download HTML</button>
			<button class="secondary" onclick={() => exportReport('pdf')} disabled={loading || !!exporting}>{exporting === 'pdf' ? 'Render PDF…' : 'Download PDF'}</button>
			<button class="primary" onclick={() => exportReport('png')} disabled={loading || !!exporting}>{exporting === 'png' ? 'Render gambar…' : 'Download Gambar'}</button>
		</div>
	</section>

	<section class="filters">
		<label>Periode
			<select bind:value={periodPreset}>
				<option value="today">Hari ini</option>
				<option value="this_week">Minggu ini</option>
				<option value="this_month">Bulan ini</option>
				<option value="custom">Custom</option>
			</select>
		</label>
		{#if periodPreset === 'custom'}
			<label>Mulai <input type="date" bind:value={startDate} /></label>
			<label>Sampai <input type="date" bind:value={endDate} /></label>
		{/if}
		<label>Tingkat
			<select bind:value={targetLevel}>
				<option value="">Semua</option>
				<option value="VII">VII</option>
				<option value="VIII">VIII</option>
				<option value="IX">IX</option>
			</select>
		</label>
		<label>Status
			<select bind:value={workflowStatus}>
				<option value="">Semua</option>
				<option value="draft">Draft</option>
				<option value="submitted">Submitted</option>
				<option value="review">Review</option>
				<option value="revision_needed">Perlu Revisi</option>
				<option value="approved">Approved</option>
				<option value="published">Published</option>
				<option value="rejected">Rejected</option>
			</select>
		</label>
		<label>Pembuat <input placeholder="username" bind:value={authorUsername} /></label>
		<label>Subject ID <input placeholder="opsional UUID mapel" bind:value={subjectId} /></label>
		<label class="check"><input type="checkbox" bind:checked={includeSystem} /> Tampilkan system/seed</label>
		<button class="primary" onclick={loadReport} disabled={loading}>{loading ? 'Memuat…' : 'Terapkan'}</button>
	</section>

	<nav class="tabs" aria-label="Tab laporan">
		{#each tabs as tab}
			<button class:active={activeReport === tab.key} onclick={() => { activeReport = tab.key; void loadReport(); }}>
				<strong>{tab.label}</strong><span>{tab.desc}</span>
			</button>
		{/each}
	</nav>

	{#if error}<div class="alert">{error}</div>{/if}

	{#if report}
		<section class="summary">
			<div><span>Total</span><strong>{report.summary.total}</strong></div>
			<div><span>PG</span><strong>{report.summary.pg}</strong></div>
			<div><span>Essay</span><strong>{report.summary.essay}</strong></div>
			<div><span>Lain</span><strong>{report.summary.other}</strong></div>
			<div><span>Pembuat/Grup</span><strong>{report.summary.authors}</strong></div>
			<div><span>Mapel</span><strong>{report.summary.subjects}</strong></div>
			<div><span>Kurang</span><strong>{report.summary.shortage}</strong></div>
			<div><span>Data warning</span><strong>{report.summary.missing_level}</strong></div>
		</section>

		<section class="table-card">
			<div class="table-head">
				<div><h2>{report.title}</h2><p>{report.period_label} · {report.access_note}</p></div>
				<span>{report.rows.length} baris</span>
			</div>
			<div class="table-wrap" aria-label="Tabel laporan lengkap">
				<table>
					<thead><tr><th>No</th><th>Utama</th><th>Mapel</th><th>Tingkat</th><th>Tugas</th><th>Total</th><th>PG</th><th>Essay</th><th>Lain</th><th>Status</th><th>Terakhir</th><th>Keterangan</th></tr></thead>
					<tbody>
						{#each report.rows as row}
							<tr class:system={row.system_row} class:warning={row.data_warning}>
								<td>{row.no}</td>
								<td><b>{row.primary}</b><br><small>{row.secondary}</small></td>
								<td>{row.subject_name}</td><td>{row.level_name}</td><td>{row.task}</td>
								<td class="num">{row.total}</td><td class="num">{row.pg}</td><td class="num">{row.essay}</td><td class="num">{row.other}</td>
								<td>{row.statuses}</td><td>{row.last_input}</td><td>{row.notes}</td>
							</tr>
						{/each}
					</tbody>
				</table>
			</div>

			<div class="mobile-rows" aria-label="Daftar laporan ringkas mobile">
				{#each report.rows as row}
					<article class:system={row.system_row} class:warning={row.data_warning}>
						<div class="mobile-row-head">
							<div><strong>{row.primary}</strong><small>{row.secondary}</small></div>
							<span>#{row.no}</span>
						</div>
						<div class="mobile-meta"><span>{row.subject_name}</span><span>{row.level_name}</span><span>{row.task}</span></div>
						<div class="mobile-counts"><b>{row.total}</b><span>Total</span><b>{row.pg}</b><span>PG</span><b>{row.essay}</b><span>Essay</span><b>{row.other}</b><span>Lain</span></div>
						<p>{row.statuses}</p>
						<small>Terakhir: {row.last_input} · {row.notes}</small>
					</article>
				{/each}
			</div>
		</section>
	{/if}
</div>

<style>
	.page-shell {
		padding: 24px;
		display: grid;
		gap: 18px;
	}
	.hero {
		display: flex;
		justify-content: space-between;
		gap: 16px;
		align-items: flex-start;
		background: linear-gradient(135deg, #0f766e, #111827);
		color: white;
		border-radius: 24px;
		padding: 26px;
	}
	.eyebrow { text-transform: uppercase; letter-spacing: .14em; font-weight: 800; opacity: .8; margin: 0 0 8px; }
	h1 { margin: 0; font-size: clamp(26px, 7vw, 34px); }
	.muted { opacity: .82; max-width: 760px; }
	.actions, .filters { display: flex; gap: 10px; flex-wrap: wrap; align-items: end; }
	.primary, .secondary { border: 0; border-radius: 12px; padding: 10px 14px; font-weight: 800; cursor: pointer; min-height: 42px; }
	.primary { background: #0f766e; color: white; }
	.secondary { background: white; color: #0f172a; }
	.filters, .table-card { background: white; border: 1px solid #e5e7eb; border-radius: 20px; padding: 16px; box-shadow: 0 10px 30px #0f172a10; }
	label { display: grid; gap: 6px; font-size: 12px; font-weight: 800; color: #475569; }
	input, select { border: 1px solid #cbd5e1; border-radius: 10px; padding: 9px; min-width: 130px; }
	.check { display: flex; align-items: center; gap: 8px; }
	.tabs { display: grid; grid-template-columns: repeat(auto-fit, minmax(180px, 1fr)); gap: 10px; }
	.tabs button { text-align: left; border: 1px solid #e5e7eb; background: white; border-radius: 16px; padding: 14px; display: grid; gap: 4px; cursor: pointer; }
	.tabs button.active { border-color: #0f766e; box-shadow: 0 0 0 3px #0f766e22; }
	.tabs span { font-size: 12px; color: #64748b; }
	.alert { background: #fee2e2; color: #991b1b; border-radius: 12px; padding: 12px; }
	.summary { display: grid; grid-template-columns: repeat(auto-fit, minmax(130px, 1fr)); gap: 12px; }
	.summary div { background: white; border: 1px solid #e5e7eb; border-radius: 16px; padding: 14px; }
	.summary span { display: block; color: #64748b; font-size: 12px; font-weight: 800; }
	.summary strong { font-size: 30px; }
	.table-head { display: flex; justify-content: space-between; gap: 12px; align-items: center; }
	.table-head h2 { margin: 0; }
	.table-head p { margin: 4px 0 0; color: #64748b; }
	.table-wrap { overflow: auto; -webkit-overflow-scrolling: touch; border-radius: 14px; }
	table { width: 100%; border-collapse: collapse; min-width: 1100px; }
	th { background: #0f172a; color: white; text-align: left; padding: 10px; font-size: 12px; position: sticky; top: 0; }
	td { border-bottom: 1px solid #e5e7eb; padding: 10px; font-size: 13px; vertical-align: top; }
	tr.system td, article.system { background: #eff6ff; }
	tr.warning td, article.warning { background: #fff7ed; }
	.num { text-align: right; font-variant-numeric: tabular-nums; }
	small { color: #64748b; }
	.mobile-rows { display: none; }

	@media (max-width: 760px) {
		.page-shell { padding: 12px; gap: 12px; }
		.hero { display: grid; border-radius: 18px; padding: 18px; }
		.hero .actions { width: 100%; display: grid; grid-template-columns: 1fr 1fr; }
		.hero .actions button { width: 100%; padding-inline: 10px; font-size: 12px; }
		.filters { display: grid; grid-template-columns: 1fr; padding: 12px; }
		.filters label, .filters button, input, select { width: 100%; min-width: 0; }
		.check { justify-content: flex-start; }
		.tabs { display: flex; overflow-x: auto; padding-bottom: 4px; scroll-snap-type: x mandatory; }
		.tabs button { min-width: 168px; scroll-snap-align: start; }
		.summary { grid-template-columns: repeat(2, minmax(0, 1fr)); }
		.summary div { padding: 12px; }
		.summary strong { font-size: 24px; }
		.table-card { padding: 12px; border-radius: 16px; }
		.table-head { display: grid; align-items: start; }
		.table-head span { justify-self: start; }
		.table-wrap { display: none; }
		.mobile-rows { display: grid; gap: 10px; }
		.mobile-rows article { border: 1px solid #e5e7eb; border-radius: 16px; padding: 12px; background: white; }
		.mobile-row-head { display: flex; justify-content: space-between; gap: 10px; align-items: flex-start; }
		.mobile-row-head div { display: grid; gap: 2px; }
		.mobile-row-head span { color: #64748b; font-size: 12px; font-weight: 800; }
		.mobile-meta { display: flex; flex-wrap: wrap; gap: 6px; margin: 10px 0; }
		.mobile-meta span { background: #f1f5f9; border-radius: 999px; padding: 4px 8px; color: #334155; font-size: 12px; }
		.mobile-counts { display: grid; grid-template-columns: repeat(4, auto 1fr); gap: 4px 6px; align-items: baseline; font-variant-numeric: tabular-nums; }
		.mobile-counts b { color: #0f766e; }
		.mobile-counts span { color: #64748b; font-size: 11px; }
		.mobile-rows p { margin: 10px 0 4px; font-size: 13px; color: #334155; }
	}

	@media (max-width: 420px) {
		.hero .actions { grid-template-columns: 1fr; }
		.summary { grid-template-columns: 1fr; }
		.mobile-counts { grid-template-columns: repeat(2, auto 1fr); }
	}
</style>
