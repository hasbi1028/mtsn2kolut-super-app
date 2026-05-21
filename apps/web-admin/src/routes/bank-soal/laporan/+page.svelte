<script lang="ts">
	import { onMount } from 'svelte';
	import { clientApiPathWithQuery, readClientApiData } from '$lib/client/api';

	type ReportKey = 'input' | 'progress' | 'revision' | 'reviewer' | 'readiness' | 'honor';
	type StatusOption = { value: string; label: string };
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
	let workflowStatuses = $state<string[]>([]);
	let statusDropdownOpen = $state(false);
	let includeSystem = $state(true);
	let loading = $state(false);
	let exporting = $state('');
	let compactMode = $state(false);
	let compactDensity = $state<'readable' | 'dense'>('readable');
	let error = $state('');
	let report = $state<ReportResult | null>(null);

	const workflowStatusOptions: StatusOption[] = [
		{ value: 'draft', label: 'Draft' },
		{ value: 'submitted', label: 'Submitted' },
		{ value: 'review', label: 'Review' },
		{ value: 'revision_needed', label: 'Perlu Revisi' },
		{ value: 'approved', label: 'Approved' },
		{ value: 'published', label: 'Published' },
		{ value: 'rejected', label: 'Rejected' },
		{ value: 'archived', label: 'Archived' }
	];

	let selectedStatusLabels = $derived(
		workflowStatuses
			.map((status) => workflowStatusOptions.find((option) => option.value === status)?.label ?? status)
			.filter(Boolean)
	);
	let statusSummary = $derived(
		selectedStatusLabels.length === 0
			? 'Semua status'
			: selectedStatusLabels.length === 1
				? selectedStatusLabels[0]
				: `${selectedStatusLabels.length} status dipilih`
	);

	function toggleWorkflowStatus(value: string) {
		if (workflowStatuses.includes(value)) {
			workflowStatuses = workflowStatuses.filter((status) => status !== value);
			return;
		}
		workflowStatuses = [...workflowStatuses, value];
	}

	function clearWorkflowStatuses() {
		workflowStatuses = [];
	}

	const subjectShortNames: Record<string, string> = {
		'Pendidikan Jasmani, Olahraga, dan Kesehatan': 'PJOK',
		'Pendidikan Jasmani Olahraga dan Kesehatan': 'PJOK',
		'Ilmu Pengetahuan Alam': 'IPA',
		'Ilmu Pengetahuan Sosial': 'IPS',
		'Bahasa Indonesia': 'B. Indonesia',
		'Bahasa Inggris': 'B. Inggris',
		'Bahasa Arab': 'B. Arab',
		'Pendidikan Pancasila dan Kewarganegaraan': 'PPKn',
		'Pendidikan Pancasila': 'PPKn',
		'Mulok Kewirausahaan': 'Mulok KWU',
		'Muatan Lokal Kewirausahaan': 'Mulok KWU',
		'Sejarah Kebudayaan Islam': 'SKI',
		"Al-Qur'an Hadis": 'Qurdis',
		"Qur'an Hadits": 'Qurdis',
		'Akidah Akhlak': 'Akidah',
		'Fikih': 'Fikih',
		'Prakarya': 'Prakarya',
		'Matematika': 'MTK',
		'Informatika': 'Informatika'
	};

	const statusShortNames: Record<string, string> = {
		draft: 'Draft',
		submitted: 'Sub',
		review: 'Review',
		revision_needed: 'Rev',
		approved: 'Appr',
		published: 'Pub',
		rejected: 'Reject',
		archived: 'Arsip'
	};

	function compactSubjectName(value?: string) {
		const name = value?.trim() ?? '';
		return subjectShortNames[name] ?? name;
	}

	function compactStatuses(value?: string) {
		const text = value?.trim() ?? '';
		if (!text) return '';
		return text
			.split(',')
			.map((part) => {
				const [rawStatus, rawCount] = part.trim().split(':');
				const label = statusShortNames[rawStatus?.trim() ?? ''] ?? rawStatus?.trim() ?? '';
				const count = rawCount?.trim();
				return count ? `${label} ${count}` : label;
			})
			.filter(Boolean)
			.join(' · ');
	}

	function compactDateTime(value?: string) {
		const text = value?.trim() ?? '';
		const match = /^(\d{4})-(\d{2})-(\d{2})(?:[ T](\d{2}:\d{2}))?/.exec(text);
		if (!match) return text;
		return `${match[3]}/${match[2]}${match[4] ? ` ${match[4]}` : ''}`;
	}

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
		for (const status of workflowStatuses) p.append('workflow_status', status);
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

<div
	class="page-shell"
	class:compact-mode={compactMode}
	class:compact-readable={compactMode && compactDensity === 'readable'}
	class:compact-dense={compactMode && compactDensity === 'dense'}
>
	<section class="hero">
		<div>
			<p class="eyebrow">Bank Soal</p>
			<h1>Report Center</h1>
			<p class="muted">Laporan input, progres mapel, revisi, reviewer, kesiapan paket, dan volume tugas/honor tanpa bantuan AI.</p>
		</div>
		<div class="actions">
			<button
				type="button"
				class="secondary"
				class:compact-active={compactMode}
				onclick={() => (compactMode = !compactMode)}
				aria-pressed={compactMode}
			>
				{compactMode ? 'Mode Normal' : 'Mode Ringkas'}
			</button>
			{#if compactMode}
				<button
					type="button"
					class="secondary density-toggle"
					onclick={() => (compactDensity = compactDensity === 'readable' ? 'dense' : 'readable')}
					aria-label="Ganti kepadatan mode ringkas"
				>
					{compactDensity === 'readable' ? 'Nyaman 24–26 baris' : 'Padat ±31 baris'}
				</button>
			{/if}
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
		<div class="status-filter">
			<span class="filter-label">Status</span>
			<button
				type="button"
				class="status-trigger"
				aria-haspopup="listbox"
				aria-expanded={statusDropdownOpen}
				onclick={() => (statusDropdownOpen = !statusDropdownOpen)}
			>
				<span>{statusSummary}</span>
				<small>{workflowStatuses.length === 0 ? 'Filter multi status' : selectedStatusLabels.join(', ')}</small>
			</button>
			{#if statusDropdownOpen}
				<div class="status-menu" role="listbox" aria-label="Pilih status laporan" aria-multiselectable="true">
					<label class="status-option status-option-all">
						<input type="checkbox" checked={workflowStatuses.length === 0} onchange={clearWorkflowStatuses} />
						<span>Semua status</span>
					</label>
					<div class="status-divider"></div>
					{#each workflowStatusOptions as option}
						<label class="status-option">
							<input
								type="checkbox"
								checked={workflowStatuses.includes(option.value)}
								onchange={() => toggleWorkflowStatus(option.value)}
							/>
							<span>{option.label}</span>
						</label>
					{/each}
					<div class="status-menu-actions">
						<button type="button" class="secondary compact" onclick={clearWorkflowStatuses}>Reset</button>
						<button type="button" class="primary compact" onclick={() => (statusDropdownOpen = false)}>Selesai</button>
					</div>
				</div>
			{/if}
			{#if workflowStatuses.length > 0}
				<div class="status-chips" aria-label="Status aktif">
					{#each workflowStatuses as status}
						<button type="button" class="status-chip" onclick={() => toggleWorkflowStatus(status)}>
							{workflowStatusOptions.find((option) => option.value === status)?.label ?? status}
							<span aria-hidden="true">×</span>
						</button>
					{/each}
				</div>
			{/if}
		</div>
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
			{#if compactMode}
				<div class="compact-banner" aria-label="Ringkasan laporan untuk screenshot">
					<strong>{report.summary.total} soal</strong>
					<span>{report.summary.authors} pembuat/grup</span>
					<span>{report.summary.subjects} mapel</span>
					<span>{report.rows.length} baris</span>
					<span>{compactDensity === 'readable' ? 'Nyaman 24–26 baris' : 'Padat ±31 baris'}</span>
				</div>
			{/if}
			<div class="table-head">
				<div><h2>{report.title}</h2><p>{report.period_label} · {report.access_note}</p></div>
				<span>{report.rows.length} baris</span>
			</div>
			<div class="table-wrap" aria-label="Tabel laporan lengkap">
				<table>
					<thead><tr><th>No</th><th>Utama</th><th>Mapel</th><th>Tingkat</th><th class="wide-only">Tugas</th><th>Total</th><th class="wide-only">PG</th><th class="wide-only">Essay</th><th class="wide-only">Lain</th><th>Status</th><th>Terakhir</th><th class="wide-only">Keterangan</th></tr></thead>
					<tbody>
						{#each report.rows as row}
							<tr class:system={row.system_row} class:warning={row.data_warning}>
								<td>{row.no}</td>
								<td><b>{row.primary}</b><br><small>{row.secondary}</small></td>
								<td>{compactMode ? compactSubjectName(row.subject_name) : row.subject_name}</td><td>{row.level_name}</td><td class="wide-only">{row.task}</td>
								<td class="num">{row.total}</td><td class="num wide-only">{row.pg}</td><td class="num wide-only">{row.essay}</td><td class="num wide-only">{row.other}</td>
								<td>{compactMode ? compactStatuses(row.statuses) : row.statuses}</td><td>{compactMode ? compactDateTime(row.last_input) : row.last_input}</td><td class="wide-only">{row.notes}</td>
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

	.compact-active {
		background: #ccfbf1;
		color: #0f766e;
		box-shadow: inset 0 0 0 2px #0f766e;
	}
	.compact-banner {
		display: flex;
		gap: 6px;
		flex-wrap: wrap;
		align-items: center;
		border: 1px solid #d1fae5;
		background: #ecfdf5;
		color: #065f46;
		border-radius: 12px;
		padding: 8px 9px;
		font-size: 13px;
	}
	.compact-banner span {
		border-left: 1px solid #a7f3d0;
		padding-left: 6px;
	}
	.compact-mode {
		gap: 7px;
		padding: 8px;
	}
	.compact-mode .hero {
		border-radius: 16px;
		padding: 12px 14px;
		align-items: center;
	}
	.compact-mode .hero .eyebrow,
	.compact-mode .hero .muted,
	.compact-mode .tabs,
	.compact-mode .filters,
	.compact-mode .summary,
	.compact-mode .wide-only {
		display: none;
	}
	.compact-mode h1 {
		font-size: 20px;
	}
	.compact-mode .actions {
		gap: 6px;
	}
	.compact-mode .actions button:not(.compact-active):not(.density-toggle) {
		display: none;
	}
	.compact-mode .density-toggle {
		background: #e0f2fe;
		color: #075985;
	}
	.compact-mode .table-card {
		padding: 6px;
		border-radius: 14px;
		box-shadow: 0 6px 18px #0f172a12;
	}
	.compact-mode .table-head {
		padding: 3px 3px 8px;
	}
	.compact-mode .table-head h2 {
		font-size: 16px;
	}
	.compact-mode .table-head p,
	.compact-mode .table-head span {
		font-size: 12px;
	}
	.compact-mode .table-wrap {
		display: block;
		border: 1px solid #e5e7eb;
		border-radius: 10px;
		overflow-x: auto;
	}
	.compact-mode table {
		min-width: 0;
		width: 100%;
		table-layout: auto;
	}
	.compact-mode th {
		position: static;
		background: #f8fafc;
		color: #475569;
		font-size: 11.5px;
		padding: 7px 3px;
		border-bottom: 1px solid #dbe3ea;
	}
	.compact-mode td {
		font-size: 12px;
		padding: 7px 3px;
		line-height: 1.24;
	}
	.compact-mode td:first-child,
	.compact-mode th:first-child {
		width: 20px;
		text-align: center;
		color: #64748b;
	}
	.compact-mode td:nth-child(2) {
		max-width: 170px;
	}
	.compact-mode td:nth-child(2) b {
		display: block;
		overflow: hidden;
		text-overflow: ellipsis;
		white-space: nowrap;
	}
	.compact-mode td:nth-child(2) small {
		display: none;
	}
	.compact-mode td:nth-child(3),
	.compact-mode td:nth-child(4),
	.compact-mode td:nth-child(10),
	.compact-mode td:nth-child(11) {
		white-space: nowrap;
	}
	.compact-mode tr:nth-child(even) td {
		background: #f8fafc;
	}

	.compact-mode td:nth-child(3) {
		max-width: 104px;
		overflow: hidden;
		text-overflow: ellipsis;
	}
	.compact-mode td:nth-child(4),
	.compact-mode th:nth-child(4) {
		width: 34px;
	}
	.compact-mode td:nth-child(6),
	.compact-mode th:nth-child(6) {
		width: 30px;
	}
	.compact-mode td:nth-child(10) {
		max-width: 132px;
		font-weight: 700;
		color: #334155;
	}
	.compact-mode td:nth-child(11),
	.compact-mode th:nth-child(11) {
		width: 56px;
		color: #475569;
		text-align: right;
	}

	.compact-mode.compact-dense .compact-banner {
		padding: 7px 8px;
		font-size: 12px;
	}
	.compact-mode.compact-dense .table-card {
		padding: 5px;
	}
	.compact-mode.compact-dense .table-head {
		padding: 1px 2px 4px;
	}
	.compact-mode.compact-dense .table-head h2 {
		font-size: 14px;
	}
	.compact-mode.compact-dense .table-head p,
	.compact-mode.compact-dense .table-head span {
		font-size: 10.5px;
		line-height: 1.12;
	}
	.compact-mode.compact-dense th {
		font-size: 10.5px;
		padding: 3px 2.5px;
		line-height: 1.08;
	}
	.compact-mode.compact-dense td {
		font-size: 10.8px;
		padding: 2px 2.5px;
		line-height: 1.08;
	}
	.compact-mode.compact-dense td:nth-child(2) {
		max-width: 150px;
	}
	.compact-mode.compact-dense td:nth-child(3) {
		max-width: 92px;
	}
	.compact-mode.compact-dense td:nth-child(10) {
		max-width: 118px;
	}


	.status-filter { position: relative; display: grid; gap: 6px; min-width: 220px; }
	.filter-label { font-size: 12px; font-weight: 800; color: #475569; }
	.status-trigger { border: 1px solid #cbd5e1; border-radius: 10px; padding: 8px 10px; min-height: 42px; background: white; display: grid; gap: 2px; text-align: left; cursor: pointer; min-width: 220px; }
	.status-trigger span { font-weight: 800; color: #0f172a; }
	.status-trigger small { color: #64748b; overflow: hidden; text-overflow: ellipsis; white-space: nowrap; max-width: 260px; }
	.status-menu { position: absolute; top: calc(100% + 6px); left: 0; z-index: 20; width: min(320px, calc(100vw - 32px)); background: white; border: 1px solid #cbd5e1; border-radius: 14px; padding: 10px; box-shadow: 0 16px 36px #0f172a22; display: grid; gap: 4px; }
	.status-option { display: flex; align-items: center; gap: 8px; padding: 8px; border-radius: 10px; font-size: 13px; color: #0f172a; cursor: pointer; }
	.status-option:hover { background: #f1f5f9; }
	.status-option input { width: auto; min-width: 0; padding: 0; }
	.status-option-all { font-weight: 900; }
	.status-divider { border-top: 1px solid #e5e7eb; margin: 4px 0; }
	.status-menu-actions { display: flex; gap: 8px; justify-content: flex-end; border-top: 1px solid #e5e7eb; padding-top: 8px; margin-top: 4px; }
	.compact { min-height: 34px; padding: 7px 10px; font-size: 12px; }
	.status-chips { display: flex; flex-wrap: wrap; gap: 6px; max-width: 360px; }
	.status-chip { border: 1px solid #99f6e4; background: #ccfbf1; color: #0f766e; border-radius: 999px; padding: 4px 8px; font-size: 11px; font-weight: 900; cursor: pointer; display: inline-flex; align-items: center; gap: 6px; }
	.status-chip span { font-size: 13px; line-height: 1; }

	@media (max-width: 760px) {
		.page-shell { padding: 12px; gap: 12px; }
		.hero { display: grid; border-radius: 18px; padding: 18px; }
		.hero .actions { width: 100%; display: grid; grid-template-columns: 1fr 1fr; }
		.hero .actions button { width: 100%; padding-inline: 10px; font-size: 12px; }
		.compact-mode .hero { display: flex; padding: 10px; }
		.compact-mode .hero .actions { width: auto; display: flex; margin-left: auto; }
		.compact-mode .hero .actions button { width: auto; padding-inline: 8px; }
		.compact-mode { padding: 6px; gap: 6px; }
		.compact-mode .hero { padding: 8px; }
		.compact-mode .table-card { padding: 5px; border-radius: 12px; }
		.compact-mode .table-head { padding: 2px 2px 6px; }
		.compact-mode.compact-dense .table-head { padding: 0 1px 3px; }
		.compact-mode th { padding-inline: 2.5px; }
		.compact-mode td { padding-inline: 2.5px; }
		.compact-mode.compact-dense th { padding: 2px 2px; }
		.compact-mode.compact-dense td { padding: 1.5px 2px; }
		.compact-mode .compact-banner { gap: 5px; padding-inline: 7px; }
		.compact-mode .compact-banner span { padding-left: 5px; }
		.filters { display: grid; grid-template-columns: 1fr; padding: 12px; }
		.filters label, .filters button, input, select, .status-filter, .status-trigger { width: 100%; min-width: 0; }
		.status-menu { position: static; width: 100%; box-shadow: 0 10px 24px #0f172a18; }
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
		.compact-mode .table-wrap { display: block; }
		.compact-mode .mobile-rows { display: none; }
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
