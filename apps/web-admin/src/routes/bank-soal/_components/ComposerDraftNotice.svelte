<script lang="ts">
	import { Button } from '$lib/components/ui/button';

	let { onClear }: { onClear: () => void } = $props();
</script>

<section class="rounded-lg border border-warning/30 bg-warning/10 p-3 text-sm text-warning" aria-label="Peringatan konsep tersimpan komposer">
	<div class="flex flex-col gap-3 md:flex-row md:items-center md:justify-between">
		<p>
			<span class="font-semibold">Perangkat bersama:</span> komposer menyimpan konsep tersimpan di browser ini. Logout akan menghapus konsep dan antrian Bank Soal lokal; gunakan tombol hapus jika selesai memakai komputer bersama.
		</p>
		<Button variant="outline" class="border-warning/30 bg-card text-warning hover:bg-warning/15" onclick={onClear}>
			Hapus Konsep Tersimpan
		</Button>
	</div>
</section>
