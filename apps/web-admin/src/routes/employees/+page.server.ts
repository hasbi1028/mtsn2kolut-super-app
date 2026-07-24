import { error, fail } from '@sveltejs/kit';
import type { PageServerLoad, Actions } from './$types.js';

type Employee = {
  id: string;
  pegawai_uid: string;
  nip: string;
  nama: string;
  unit_kerja: string;
  employment_type: string;
  tanggal_lahir: string;
  jenis_kelamin: string;
  tempat_lahir: string;
  pusaka_eligible: boolean;
  has_pusaka_account: boolean;
  pusaka_is_enabled: boolean;
  is_active: boolean;
};

type EmployeesPayload = {
  items?: Employee[];
  data?: { items?: Employee[] };
  error?: string;
  message?: string;
};

function extractEmployees(payload: EmployeesPayload | Employee[]): Employee[] {
  if (Array.isArray(payload)) return payload;
  return payload.items ?? payload.data?.items ?? [];
}

function getCookieHeaders(request: Request, url: URL): Record<string, string> {
  const cookie = request.headers.get('cookie') || url.searchParams.toString();
  return cookie ? { cookie } : {};
}

export const load: PageServerLoad = async ({ fetch, url, request }) => {
  const res = await fetch(`${url.origin}/api/employees`, {
    headers: getCookieHeaders(request, url),
  });

  if (!res.ok) {
    throw error(res.status, 'Gagal memuat data pegawai');
  }

  const payload: EmployeesPayload | Employee[] = await res.json();
  const employees = extractEmployees(payload);

  return { employees };
};

export const actions: Actions = {
  tambah: async ({ request, fetch, url }) => {
    const data = await request.formData();

    const nama = String(data.get('nama') ?? '').trim();
    const employmentType = String(data.get('employment_type') ?? '').trim();

    if (!nama || !employmentType) {
      return fail(400, { tambahError: 'Nama dan status kepegawaian wajib diisi.' });
    }

    const pusakaUsername = String(data.get('pusaka_username') ?? '').trim();
    const pusakaPassword = String(data.get('pusaka_password') ?? '').trim();

    if ((pusakaUsername && !pusakaPassword) || (!pusakaUsername && pusakaPassword)) {
      return fail(400, { tambahError: 'Username dan password PUSAKA harus diisi berpasangan.' });
    }

    const isPusakaEligible = employmentType === 'pns' || employmentType === 'pppk';
    if (!isPusakaEligible && (pusakaUsername || pusakaPassword)) {
      return fail(400, { tambahError: 'Hanya pegawai PNS atau PPPK yang boleh memiliki akun PUSAKA.' });
    }

    const body: Record<string, string> = {
      nama,
      employment_type: employmentType,
    };
    for (const field of ['nip', 'unit_kerja', 'tanggal_lahir', 'jenis_kelamin', 'tempat_lahir']) {
      const val = String(data.get(field) ?? '').trim();
      if (val) body[field] = val;
    }
    if (pusakaUsername) body.pusaka_username = pusakaUsername;
    if (pusakaPassword) body.pusaka_password = pusakaPassword;

    const res = await fetch(`${url.origin}/api/employees`, {
      method: 'POST',
      headers: { 'content-type': 'application/json', ...getCookieHeaders(request, url) },
      body: JSON.stringify(body),
    });

    if (!res.ok) {
      let errMsg = 'Gagal menyimpan pegawai.';
      try {
        const errPayload = await res.json();
        errMsg = errPayload.error ?? errPayload.message ?? errMsg;
      } catch { /* use default */ }
      return fail(res.status, { tambahError: errMsg });
    }

    return { tambahSuccess: 'Pegawai baru berhasil ditambahkan.' };
  },

  hapus: async ({ request, fetch, url }) => {
    const data = await request.formData();
    const id = String(data.get('id') ?? '').trim();

    if (!id) {
      return fail(400, { hapusError: 'ID pegawai tidak valid.' });
    }

    const res = await fetch(`${url.origin}/api/employees/${id}`, {
      method: 'DELETE',
      headers: getCookieHeaders(request, url),
    });

    if (!res.ok) {
      let errMsg = 'Gagal menghapus pegawai.';
      try {
        const contentType = res.headers.get('content-type') || '';
        if (contentType.includes('json')) {
          const e = await res.json();
          errMsg = e.error || e.message || errMsg;
        } else {
          const text = await res.text();
          if (text && text.length < 200) errMsg = text;
        }
      } catch {}
      return fail(res.status, { hapusError: errMsg });
    }

    return { hapusSuccess: 'Pegawai berhasil dihapus.' };
  },

  nonaktifkan: async ({ request, fetch, url }) => {
    const data = await request.formData();
    const id = String(data.get('id') ?? '').trim();

    if (!id) {
      return fail(400, { nonaktifError: 'ID pegawai tidak valid.' });
    }

    const isActive = data.get('is_active') === 'true';

    const res = await fetch(`${url.origin}/api/employees/${id}/status`, {
      method: 'PATCH',
      headers: { 'content-type': 'application/json', ...getCookieHeaders(request, url) },
      body: JSON.stringify({ is_active: !isActive }),
    });

    if (!res.ok) {
      let errMsg = 'Gagal mengubah status pegawai.';
      try {
        const contentType = res.headers.get('content-type') || '';
        if (contentType.includes('json')) {
          const e = await res.json();
          errMsg = e.error || e.message || errMsg;
        } else {
          const text = await res.text();
          if (text && text.length < 200) errMsg = text;
        }
      } catch {}
      return fail(res.status, { nonaktifError: errMsg });
    }

    return { nonaktifSuccess: `Pegawai ${isActive ? 'dinonaktifkan' : 'diaktifkan'}.` };
  },
};
