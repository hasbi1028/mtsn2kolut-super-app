import { confirmAction } from '$lib/confirm-dialog';

export async function confirmDiscardChanges(hasChanges: boolean): Promise<boolean> {
	if (!hasChanges) return true;
	return confirmAction({
		title: 'Buang Perubahan?',
		message: 'Ada perubahan yang belum disimpan. Perubahan akan hilang jika Anda melanjutkan.',
		confirmLabel: 'Buang Perubahan',
		cancelLabel: 'Tetap Edit',
		tone: 'warning',
	});
}

export function bindBeforeUnload(hasChanges: () => boolean): () => void {
	if (typeof window === 'undefined') return () => undefined;
	const handleBeforeUnload = (event: BeforeUnloadEvent) => {
		if (!hasChanges()) return;
		event.preventDefault();
		event.returnValue = '';
	};
	window.addEventListener('beforeunload', handleBeforeUnload);
	return () => window.removeEventListener('beforeunload', handleBeforeUnload);
}

