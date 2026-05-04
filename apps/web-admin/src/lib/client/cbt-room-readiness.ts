export type CbtRoomReadiness = {
	room_count: number;
	total_capacity: number;
	participant_count: number;
	assigned_participant_count: number;
	unassigned_participant_count: number;
	missing_seat_count: number;
	rooms_without_proctor: number;
	proctor_assignment_count: number;
};

export type CbtReadinessTone = 'success' | 'warning' | 'info';

export function roomReadinessTone(readiness: CbtRoomReadiness | null): CbtReadinessTone {
	if (!readiness) return 'info';
	if (
		readiness.unassigned_participant_count > 0
		|| readiness.missing_seat_count > 0
		|| readiness.rooms_without_proctor > 0
		|| readiness.total_capacity < readiness.participant_count
	) {
		return 'warning';
	}
	return 'success';
}

export function roomReadinessMessage(readiness: CbtRoomReadiness | null): string {
	if (!readiness) return 'Kesiapan ruangan belum dimuat.';
	return `${readiness.assigned_participant_count}/${readiness.participant_count} peserta sudah punya ruang, kapasitas total ${readiness.total_capacity}, ${readiness.rooms_without_proctor} ruang belum punya pengawas, ${readiness.missing_seat_count} peserta belum punya nomor meja.`;
}

export function cbtRoomSetupErrorMessage(error: unknown, fallback: string): string {
	if (error instanceof Error && error.message.trim() && !error.message.toLowerCase().includes('fetch')) {
		const message = error.message.trim();
		if (message.includes('403') || /forbidden|unauthori[sz]ed|akses|permission/i.test(message)) {
			return 'Akses pengaturan sesi hanya untuk admin/operator CBT.';
		}
		return message;
	}
	return fallback;
}
