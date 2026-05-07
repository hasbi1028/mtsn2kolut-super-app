import {
	isAuthExpiredSyncStatus,
	type BankSoalQuestionSyncInput,
	type BankSoalQuestionSyncIntent,
} from './bank-soal-offline';

export type BuildBankSoalQuestionSyncOptions<TPayload> = {
	draftKey: string;
	editingId: string | null;
	intent: BankSoalQuestionSyncIntent;
	payload: TPayload;
};

export type QueueQuestionSaveDecision = {
	isOnline: boolean;
	responseReceived: boolean;
	responseStatus?: number;
};

export function buildBankSoalQuestionSyncInput<TPayload>({
	draftKey,
	editingId,
	intent,
	payload,
}: BuildBankSoalQuestionSyncOptions<TPayload>): BankSoalQuestionSyncInput<TPayload> {
	return {
		draftKey,
		intent,
		method: editingId ? 'PUT' : 'POST',
		endpoint: editingId ? `/api/bank-soal/questions/${encodeURIComponent(editingId)}` : '/api/bank-soal/questions',
		questionId: editingId ?? undefined,
		payload,
	};
}

export function shouldQueueBankSoalQuestionSave(decision: QueueQuestionSaveDecision): boolean {
	if (!decision.isOnline) return true;
	if (!decision.responseReceived) return true;
	return typeof decision.responseStatus === 'number' && isAuthExpiredSyncStatus(decision.responseStatus);
}

export function bankSoalQueueAuthRequiredMessage(count: number): string {
	return count > 1
		? `${count} perubahan Bank Soal tersimpan lokal. Masuk ulang lalu sinkronkan.`
		: 'Perubahan Bank Soal tersimpan lokal. Masuk ulang lalu sinkronkan.';
}
