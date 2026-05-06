<script lang="ts">
	import { Button } from '$lib/components/ui/button';

	let { onClear }: { onClear: () => void } = $props();
</script>

<section class="rounded-lg border border-amber-200 bg-amber-50/70 p-3 text-sm text-amber-950" aria-label="Peringatan draft lokal komposer">
	<div class="flex flex-col gap-3 md:flex-row md:items-center md:justify-between">
		<p>
			<span class="font-semibold">Perangkat bersama:</span> komposer menyimpan draft lokal di browser ini. Logout akan menghapus draft CBT lokal; gunakan tombol hapus jika selesai memakai komputer bersama.
		</p>
		<Button variant="outline" class="border-amber-300 bg-white text-amber-900 hover:bg-amber-100" onclick={onClear}>
			Hapus Draft Lokal
		</Button>
	</div>
</section>
