<script lang="ts">
	import QRCode from 'qrcode';

	type CardSide = 'front' | 'back' | 'both';
	type CardData = {
		card_no?: string;
		qr_url?: string;
		qr_token?: string;
		nama?: string;
		nis?: string;
		nisn?: string;
		class_name?: string;
		class_code?: string;
		photo_url?: string;
		status?: string;
		valid_until?: string;
	};

	let { card, side = 'both', print = false, onQrReady }: { card: CardData; side?: CardSide; print?: boolean; onQrReady?: (ready: boolean) => void } = $props();

	const qrValue = $derived(card.qr_url || card.qr_token || '');
	const hasSecureQr = $derived(Boolean(qrValue));
	const classLabel = $derived(card.class_name || card.class_code || '-');
	let qrDataUrl = $state('');

	$effect(() => {
		const value = qrValue;
		qrDataUrl = '';
		onQrReady?.(false);
		if (!value) return;
		QRCode.toDataURL(value, { errorCorrectionLevel: 'M', margin: 1, width: 192, color: { dark: '#052e16', light: '#ffffff' } })
			.then((url) => {
				if (qrValue === value) {
					qrDataUrl = url;
					onQrReady?.(true);
				}
			})
			.catch(() => {
				qrDataUrl = '';
				onQrReady?.(false);
			});
	});

	function initials(name?: string) {
		return (name || 'Siswa')
			.split(/\s+/)
			.filter(Boolean)
			.slice(0, 2)
			.map((part) => part[0]?.toUpperCase())
			.join('') || 'S';
	}

</script>

<div class:print-sheet={print} class="student-id-card-wrap side-{side}">
	{#if side === 'front' || side === 'both'}
		<section class="id-card front" aria-label="Kartu siswa depan">
			<div class="gold-ribbon"></div>
			<header>
				<div class="logo">MTs</div>
				<div>
					<p class="ministry">KEMENTERIAN AGAMA</p>
					<h2>MTsN 2 KOLAKA UTARA</h2>
					<p class="subtitle">KARTU IDENTITAS SISWA TERPADU</p>
				</div>
			</header>
			<div class="front-body">
				<div class="photo">
					{#if card.photo_url}
						<img src={card.photo_url} alt="Foto siswa" />
					{:else}
						<span>{initials(card.nama)}</span>
					{/if}
				</div>
				<div class="identity">
					<p class="label">Nama Siswa</p>
					<h1>{card.nama || 'NAMA SISWA'}</h1>
					<div class="fields">
						<div><span>NIS</span><strong>{card.nis || '-'}</strong></div>
						<div><span>Kelas</span><strong>{classLabel}</strong></div>
						<div><span>ID Kartu</span><strong>{card.card_no || '-'}</strong></div>
						<div><span>Status</span><strong>{card.status || 'active'}</strong></div>
					</div>
				</div>
			</div>
			<footer>
				<span>Berlaku selama siswa aktif</span>
				<span>Official Student ID</span>
			</footer>
		</section>
	{/if}

	{#if side === 'back' || side === 'both'}
		<section class="id-card back" aria-label="Kartu siswa belakang">
			<div class="back-title">
				<h2>SCAN UNTUK VERIFIKASI</h2>
				<p>QR bukan password. Login tetap memerlukan PIN/token sesuai layanan.</p>
			</div>
			<div class="qr-box" aria-label="QR verifikasi kartu">
				<div class:qr={hasSecureQr && qrDataUrl} class:qr-missing={!hasSecureQr || !qrDataUrl} aria-label="QR verifikasi kartu">
					{#if qrDataUrl && hasSecureQr}
						<img src={qrDataUrl} alt="QR verifikasi kartu siswa" />
					{:else}
						<span>QR belum tersedia — jangan cetak</span>
					{/if}
				</div>
			</div>

			<div class="back-info">
				<p>ID Kartu</p>
				<strong>{card.card_no || '-'}</strong>
				<small>{hasSecureQr ? 'QR aman aktif — token tidak dicetak sebagai teks' : 'QR belum tersedia — jangan cetak kartu ini'}</small>
			</div>
			<div class="rules">
				<p>Kartu ini milik madrasah. Jika ditemukan, mohon kembalikan ke MTsN 2 Kolaka Utara.</p>
				<p>Digunakan untuk pengenal, portal siswa, presensi, perpustakaan, layanan, dan validasi CBT.</p>
			</div>
		</section>
	{/if}
</div>

<style>
	.student-id-card-wrap { display: flex; flex-wrap: wrap; gap: 18px; align-items: flex-start; }
	.id-card { position: relative; overflow: hidden; width: 340px; height: 214px; border-radius: 18px; border: 1px solid #d6c28d; background: #fffdf4; color: #122117; box-shadow: 0 18px 40px rgba(15, 23, 42, .14); font-family: Inter, ui-sans-serif, system-ui, sans-serif; }
	.front { background: linear-gradient(135deg, #fffdf4 0%, #f6ecd1 58%, #e7cf82 100%); }
	.front:before { content: ''; position: absolute; inset: -55px auto auto -45px; width: 170px; height: 170px; border-radius: 999px; background: #0f5b35; opacity: .96; }
	.gold-ribbon { position: absolute; right: -55px; top: -35px; width: 170px; height: 120px; rotate: 25deg; background: linear-gradient(135deg, #d9b24c, #f6df94); opacity: .9; }
	header { position: relative; z-index: 1; display: flex; gap: 10px; align-items: center; padding: 18px 20px 10px; }
	.logo { display: grid; place-items: center; width: 42px; height: 42px; border-radius: 12px; background: #fff; color: #0f5b35; font-weight: 900; font-size: 13px; box-shadow: 0 8px 18px rgba(0,0,0,.14); }
	.ministry { margin: 0; font-size: 8px; letter-spacing: .12em; font-weight: 800; color: #56705d; }
	header h2 { margin: 0; font-size: 17px; line-height: 1; font-weight: 900; letter-spacing: .02em; }
	.subtitle { margin: 4px 0 0; font-size: 8px; font-weight: 800; color: #0f5b35; letter-spacing: .08em; }
	.front-body { position: relative; z-index: 1; display: grid; grid-template-columns: 92px 1fr; gap: 14px; padding: 8px 20px; }
	.photo { display: grid; place-items: center; width: 92px; height: 110px; border-radius: 16px; border: 4px solid #fff; background: linear-gradient(135deg, #dbe7df, #97baa5); color: #0f5b35; font-size: 26px; font-weight: 900; box-shadow: 0 10px 22px rgba(15, 91, 53, .22); overflow: hidden; }
	.photo img { width: 100%; height: 100%; object-fit: cover; }
	.identity .label { margin: 4px 0 2px; color: #5d6b5e; font-size: 8px; text-transform: uppercase; letter-spacing: .1em; font-weight: 800; }
	.identity h1 { margin: 0 0 8px; font-size: 20px; line-height: 1.08; color: #0b2e1d; text-transform: uppercase; }
	.fields { display: grid; grid-template-columns: 1fr 1fr; gap: 5px 10px; font-size: 9px; }
	.fields span { display: block; color: #66735f; font-weight: 700; }
	.fields strong { display: block; color: #17251b; font-size: 10px; }
	footer { position: absolute; left: 20px; right: 20px; bottom: 12px; display: flex; justify-content: space-between; color: #31513c; font-size: 8px; font-weight: 800; }
	.back { background: linear-gradient(135deg, #0f5b35, #173d2b 60%, #0b271a); color: #fff8df; padding: 18px; }
	.back:after { content: ''; position: absolute; right: -60px; bottom: -70px; width: 180px; height: 180px; border-radius: 999px; background: rgba(217,178,76,.32); }
	.back-title h2 { margin: 0; font-size: 18px; letter-spacing: .05em; }
	.back-title p { margin: 4px 0 12px; max-width: 245px; color: #e7d9ab; font-size: 9px; line-height: 1.35; }
	.qr-box { position: absolute; left: 20px; top: 78px; width: 104px; height: 104px; border-radius: 14px; background: #fff; padding: 9px; box-shadow: 0 12px 28px rgba(0,0,0,.24); }
	.qr { display: grid; grid-template-columns: repeat(17, 1fr); gap: 1px; width: 86px; height: 86px; }
	.qr img { width: 86px; height: 86px; display: block; }
	.qr-missing { display: grid; place-items: center; width: 86px; height: 86px; border: 2px dashed #b91c1c; color: #b91c1c; background: #fff7ed; text-align: center; font-size: 8px; font-weight: 800; line-height: 1.2; padding: 6px; }
	.back-info { position: absolute; left: 142px; right: 20px; top: 84px; }
	.back-info p { margin: 0; font-size: 8px; color: #e8d79e; text-transform: uppercase; letter-spacing: .12em; font-weight: 800; }
	.back-info strong { display: block; margin-top: 4px; font-size: 18px; color: #ffd966; }
	.back-info small { display: block; margin-top: 7px; max-height: 30px; overflow: hidden; color: #d9eadd; font-size: 7px; line-height: 1.25; word-break: break-all; }
	.rules { position: absolute; left: 142px; right: 20px; bottom: 20px; color: #e4ecd9; font-size: 8px; line-height: 1.35; }
	.rules p { margin: 0 0 4px; }
	@media print {
		:global(body) { background: #fff !important; }
		.id-card { box-shadow: none; print-color-adjust: exact; -webkit-print-color-adjust: exact; }
		.print-sheet { gap: 8mm; }
	}
</style>
