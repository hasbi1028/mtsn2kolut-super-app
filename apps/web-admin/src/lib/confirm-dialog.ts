export type ConfirmTone = 'default' | 'danger' | 'warning' | 'success';

export type ConfirmOptions = {
	title?: string;
	message: string;
	confirmLabel?: string;
	cancelLabel?: string;
	tone?: ConfirmTone;
	challenge?: string;
};

type ConfirmHandler = (options: Required<Omit<ConfirmOptions, 'challenge'>> & { challenge?: string }) => Promise<boolean>;

let handler: ConfirmHandler | null = null;

function normalizeConfirmOptions(options: ConfirmOptions) {
	return {
		title: options.title ?? 'Konfirmasi Tindakan',
		message: options.message,
		confirmLabel: options.confirmLabel ?? 'Lanjutkan',
		cancelLabel: options.cancelLabel ?? 'Batal',
		tone: options.tone ?? 'default',
		challenge: options.challenge,
	};
}

export function setConfirmHandler(nextHandler: ConfirmHandler | null) {
	handler = nextHandler;
}

export async function confirmAction(options: ConfirmOptions | string): Promise<boolean> {
	const normalized = normalizeConfirmOptions(
		typeof options === 'string' ? { message: options } : options
	);

	if (handler) {
		return handler(normalized);
	}

	if (typeof window === 'undefined') return false;
	return window.confirm(normalized.challenge ? `${normalized.message}\n\nKetik ${normalized.challenge} untuk melanjutkan.` : normalized.message);
}

export function confirmChallenge(options: Omit<ConfirmOptions, 'challenge'> & { challenge: string }) {
	return confirmAction({
		...options,
		confirmLabel: options.confirmLabel ?? 'Konfirmasi',
		tone: options.tone ?? 'danger',
	});
}
