<script lang="ts">
	let stage = $state<'credential' | 'identity' | 'waiting' | 'exam' | 'review' | 'done'>('credential');
	let pin = $state('');
	let current = $state(0);
	let doubtful = $state(false);
	let selectedAnswer = $state('');
	let saveState = $state('aman lokal');

	const student = {
		name: 'Siswa Simulasi',
		className: 'VIII A',
		exam: 'Gladi CBT Madrasah',
		room: 'R01',
		seat: '12'
	};

	const questions = [
		{
			text: 'Contoh soal pertama untuk memastikan tampilan satu soal per layar mudah dibaca di HP.',
			options: ['Jawaban A', 'Jawaban B', 'Jawaban C', 'Jawaban D']
		},
		{
			text: 'Jika koneksi kurang stabil, apa yang sebaiknya dilakukan peserta?',
			options: ['Menutup browser', 'Panggil pengawas', 'Ganti perangkat tanpa izin', 'Muat ulang berkali-kali']
		}
	];

	function verifyCredential() {
		if (!pin.trim()) return;
		stage = 'identity';
	}

	function chooseAnswer(value: string) {
		selectedAnswer = value;
		saveState = 'tersimpan';
	}

	function nextQuestion() {
		if (current < questions.length - 1) {
			current += 1;
			doubtful = false;
			selectedAnswer = '';
			saveState = 'aman lokal';
			return;
		}
		stage = 'review';
	}
</script>

<svelte:head>
	<title>Portal Ujian Peserta</title>
	<meta name="description" content="Portal CBT peserta berbasis mobile-web sederhana dengan QR/PIN, ruang tunggu, dan satu soal per layar." />
</svelte:head>

<div class="min-h-screen bg-slate-50 text-slate-950" style="color-scheme: light;">
	<div class="mx-auto flex min-h-screen w-full max-w-md flex-col px-4 py-4">
		<header class="rounded-3xl border border-emerald-200 bg-white p-4 shadow-sm">
			<p class="text-xs font-black uppercase tracking-[0.2em] text-emerald-700">Portal Ujian Peserta</p>
			<h1 class="mt-1 text-2xl font-black">CBT Web</h1>
			<p class="mt-1 text-sm leading-6 text-slate-600">Masuk dengan QR/PIN kartu peserta. Jika ada kendala, panggil pengawas ruang.</p>
		</header>

		<main class="mt-4 flex-1">
			{#if stage === 'credential'}
				<section class="rounded-3xl border border-slate-200 bg-white p-4 shadow-sm">
					<h2 class="text-lg font-black">1. Masukkan QR/PIN</h2>
					<p class="mt-1 text-sm leading-6 text-slate-600">Demo ini tidak mengubah data ujian. Mode produksi akan membaca kartu peserta dari backend.</p>
					<label class="mt-4 block space-y-1.5 text-sm font-bold text-slate-700">
						<span>PIN kartu peserta</span>
						<input class="w-full rounded-2xl border border-slate-300 px-4 py-3 text-center text-xl font-black tracking-[0.3em] outline-none focus:border-emerald-500" inputmode="numeric" maxlength="8" bind:value={pin} placeholder="••••" />
					</label>
					<button class="mt-4 w-full rounded-2xl bg-emerald-700 px-4 py-3 text-sm font-black text-white disabled:bg-slate-300" disabled={!pin.trim()} onclick={verifyCredential}>Lanjut Verifikasi</button>
				</section>
			{:else if stage === 'identity'}
				<section class="rounded-3xl border border-slate-200 bg-white p-4 shadow-sm">
					<h2 class="text-lg font-black">2. Benarkah ini identitas Anda?</h2>
					<div class="mt-4 divide-y divide-slate-200 overflow-hidden rounded-2xl border border-slate-200 text-sm">
						<p class="flex justify-between gap-3 p-3"><span class="font-bold text-slate-500">Nama</span><strong>{student.name}</strong></p>
						<p class="flex justify-between gap-3 p-3"><span class="font-bold text-slate-500">Kelas</span><strong>{student.className}</strong></p>
						<p class="flex justify-between gap-3 p-3"><span class="font-bold text-slate-500">Ruang/Kursi</span><strong>{student.room}/{student.seat}</strong></p>
					</div>
					<button class="mt-4 w-full rounded-2xl bg-emerald-700 px-4 py-3 text-sm font-black text-white" onclick={() => (stage = 'waiting')}>Ya, Benar</button>
					<button class="mt-2 w-full rounded-2xl border border-slate-300 px-4 py-3 text-sm font-black text-slate-700" onclick={() => (stage = 'credential')}>Bukan Saya</button>
				</section>
			{:else if stage === 'waiting'}
				<section class="rounded-3xl border border-slate-200 bg-white p-4 shadow-sm">
					<h2 class="text-lg font-black">3. Ruang Tunggu</h2>
					<p class="mt-1 text-sm leading-6 text-slate-600">Tetap di halaman ini. Pengawas akan memberi arahan sebelum ujian dimulai.</p>
					<div class="mt-4 rounded-2xl bg-emerald-50 p-4 text-sm font-bold text-emerald-900">{student.exam} · {student.room}</div>
					<button class="mt-4 w-full rounded-2xl bg-emerald-700 px-4 py-3 text-sm font-black text-white" onclick={() => (stage = 'exam')}>Masuk Ujian</button>
				</section>
			{:else if stage === 'exam'}
				<section class="rounded-3xl border border-slate-200 bg-white p-4 shadow-sm">
					<div class="flex items-center justify-between gap-3 border-b border-slate-200 pb-3">
						<div>
							<p class="text-xs font-black uppercase tracking-[0.16em] text-emerald-700">Soal {current + 1} / {questions.length}</p>
							<h2 class="text-lg font-black">{student.exam}</h2>
						</div>
						<span class="rounded-full bg-slate-100 px-3 py-1 text-xs font-black text-slate-700">{saveState}</span>
					</div>
					<p class="mt-4 text-base font-bold leading-7">{questions[current].text}</p>
					<div class="mt-4 grid gap-2">
						{#each questions[current].options as option, index}
							<button class={`rounded-2xl border px-4 py-3 text-left text-sm font-black ${selectedAnswer === option ? 'border-emerald-500 bg-emerald-50 text-emerald-900' : 'border-slate-200 bg-white text-slate-800'}`} onclick={() => chooseAnswer(option)}>{String.fromCharCode(65 + index)}. {option}</button>
						{/each}
					</div>
				</section>
			{:else if stage === 'review'}
				<section class="rounded-3xl border border-slate-200 bg-white p-4 shadow-sm">
					<h2 class="text-lg font-black">Review sebelum selesai</h2>
					<p class="mt-1 text-sm leading-6 text-slate-600">Periksa jawaban. Jika sudah yakin, minta konfirmasi pengawas.</p>
					<div class="mt-4 grid grid-cols-3 gap-2 text-center text-xs font-black">
						<div class="rounded-2xl bg-emerald-50 p-3 text-emerald-900">Dijawab<br /><span class="text-xl">1</span></div>
						<div class="rounded-2xl bg-amber-50 p-3 text-amber-900">Ragu<br /><span class="text-xl">{doubtful ? 1 : 0}</span></div>
						<div class="rounded-2xl bg-slate-100 p-3 text-slate-700">Kosong<br /><span class="text-xl">1</span></div>
					</div>
					<button class="mt-4 w-full rounded-2xl bg-emerald-700 px-4 py-3 text-sm font-black text-white" onclick={() => (stage = 'done')}>Minta Konfirmasi Pengawas</button>
				</section>
			{:else}
				<section class="rounded-3xl border border-emerald-200 bg-emerald-50 p-4 text-emerald-950 shadow-sm">
					<h2 class="text-lg font-black">Permintaan selesai dikirim</h2>
					<p class="mt-1 text-sm leading-6">Tunggu pengawas mengonfirmasi. Jangan menutup halaman sebelum diarahkan.</p>
				</section>
			{/if}
		</main>

		{#if stage === 'exam'}
			<nav class="sticky bottom-0 mt-4 grid grid-cols-3 gap-2 rounded-3xl border border-slate-200 bg-white p-2 shadow-lg">
				<button class="rounded-2xl border border-slate-300 px-3 py-3 text-xs font-black text-slate-700 disabled:opacity-40" disabled={current === 0} onclick={() => (current -= 1)}>Sebelumnya</button>
				<button class={`rounded-2xl border px-3 py-3 text-xs font-black ${doubtful ? 'border-amber-300 bg-amber-50 text-amber-900' : 'border-slate-300 text-slate-700'}`} onclick={() => (doubtful = !doubtful)}>Ragu-ragu</button>
				<button class="rounded-2xl bg-emerald-700 px-3 py-3 text-xs font-black text-white" onclick={nextQuestion}>{current === questions.length - 1 ? 'Kumpulkan' : 'Berikutnya'}</button>
			</nav>
		{/if}
	</div>
</div>
