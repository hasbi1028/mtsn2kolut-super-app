export type Fetcher = (input: RequestInfo | URL, init?: RequestInit) => Promise<Response>;

export type SchoolProfile = {
	name: string;
	nsm: string;
	npsn: string;
	ministry_line: string;
	office_line: string;
	address: string;
	village: string;
	district: string;
	regency: string;
	province: string;
	postal_code: string;
	phone: string;
	email: string;
	website: string;
	head_name: string;
	head_nip: string;
};

type ApiEnvelope<T> = {
	data?: T;
	error?: string;
	message?: string;
};

export const defaultSchoolProfile: SchoolProfile = {
	name: 'MTs Negeri 2 Kolaka Utara',
	nsm: '',
	npsn: '',
	ministry_line: 'Kementerian Agama Republik Indonesia',
	office_line: 'Kantor Kementerian Agama Kabupaten Kolaka Utara',
	address: '',
	village: '',
	district: '',
	regency: 'Kolaka Utara',
	province: 'Sulawesi Tenggara',
	postal_code: '',
	phone: '',
	email: '',
	website: '',
	head_name: '',
	head_nip: '',
};

function isRecord(value: unknown): value is Record<string, unknown> {
	return typeof value === 'object' && value !== null;
}

function apiErrorMessage(payload: unknown) {
	if (!isRecord(payload)) return '';
	const error = payload.error;
	if (typeof error === 'string' && error.trim()) return error;
	const message = payload.message;
	if (typeof message === 'string' && message.trim()) return message;
	return '';
}

async function readApi<T>(response: Response, fallbackMessage: string): Promise<T> {
	const payload = (await response.json().catch(() => null)) as ApiEnvelope<T> | T | null;
	const message = apiErrorMessage(payload);
	if (!response.ok || message) throw new Error(message || fallbackMessage);
	if (isRecord(payload) && 'data' in payload) {
		const envelope = payload as ApiEnvelope<T>;
		if (envelope.data === undefined) throw new Error(fallbackMessage);
		return envelope.data;
	}
	if (payload === null) throw new Error(fallbackMessage);
	return payload as T;
}

export async function fetchSchoolProfile(fetcher: Fetcher = fetch) {
	const profile = await readApi<Partial<SchoolProfile>>(
		await fetcher('/api/school-profile'),
		'Gagal memuat profil madrasah.'
	);
	return { ...defaultSchoolProfile, ...profile };
}

export function schoolAddressLine(profile: SchoolProfile) {
	return [profile.address, profile.village, profile.district, profile.regency, profile.province, profile.postal_code]
		.map((item) => item.trim())
		.filter(Boolean)
		.join(', ');
}
