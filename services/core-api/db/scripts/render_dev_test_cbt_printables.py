from __future__ import annotations

import json
import os
import subprocess
from dataclasses import dataclass
from datetime import datetime
from pathlib import Path
from zoneinfo import ZoneInfo

from reportlab.lib import colors
from reportlab.lib.pagesizes import A4, landscape
from reportlab.lib.units import mm
from reportlab.lib.utils import simpleSplit
from reportlab.pdfgen import canvas

ROOT = Path('/home/servermtsn2kolut/mtsn2kolut-super-app')
SCRIPT_DIR = Path(__file__).resolve().parent
OUT_DIR = ROOT / 'docs/pdf/dev-test-cbt'
CARD_PDF = OUT_DIR / 'kartu-ujian-dev-test-cbt.pdf'
ATTENDANCE_PDF = OUT_DIR / 'lembar-hadir-dev-test-cbt.pdf'
TZ = ZoneInfo('Asia/Makassar')
DATABASE_URL = 'postgresql://pusaka:pusaka_dev@localhost:5432/pusaka'

TITLE = 'Kegiatan Dev Test CBT'
SESSION_TITLE = 'Sesi Dev Test CBT'
ROOM_NAME = 'Ruang Dev Test 01'


@dataclass
class Participant:
    nis: str
    nama: str
    token: str
    seat_no: int


@dataclass
class SeedData:
    event_title: str
    session_title: str
    subject_name: str
    class_name: str
    academic_year_name: str
    room_name: str
    room_token: str
    room_capacity: int
    scheduled_start: datetime
    scheduled_end: datetime
    proctor_name: str
    proctor_nip: str
    participants: list[Participant]


def run_node_query() -> dict:
    js = r"""
import pg from 'pg';
const { Pool } = pg;
const pool = new Pool({ connectionString: process.env.DATABASE_URL });

const queryOne = async (sql, params = []) => {
  const res = await pool.query(sql, params);
  return res.rows[0] ?? null;
};

try {
  const event = await queryOne(`
    SELECT e.id, e.title AS event_title, ay.name AS academic_year_name
    FROM cbt_exam_events e
    JOIN academic_years ay ON ay.id = e.academic_year_id
    WHERE e.title = $1
    ORDER BY e.created_at DESC
    LIMIT 1
  `, ['Kegiatan Dev Test CBT']);

  if (!event) throw new Error('event not found');

  const session = await queryOne(`
    SELECT s.id, s.title AS session_title, s.scheduled_start, s.scheduled_end,
           sc.name AS class_name,
           sub.name AS subject_name
    FROM cbt_exam_sessions s
    JOIN school_classes sc ON sc.id = s.class_id
    JOIN cbt_packages p ON p.id = s.package_id
    JOIN subjects sub ON sub.id = p.subject_id
    WHERE s.title = $1
    ORDER BY s.created_at DESC
    LIMIT 1
  `, ['Sesi Dev Test CBT']);

  if (!session) throw new Error('session not found');

  const room = await queryOne(`
    SELECT r.id, r.room_name, r.room_token, r.capacity
    FROM cbt_exam_rooms r
    WHERE r.session_id = $1 AND r.room_name = $2
    ORDER BY r.created_at DESC
    LIMIT 1
  `, [session.id, 'Ruang Dev Test 01']);

  if (!room) throw new Error('room not found');

  const proctor = await queryOne(`
    SELECT e.nip, e.nama
    FROM cbt_room_proctors rp
    JOIN employees e ON e.id = rp.employee_id
    WHERE rp.exam_room_id = $1
    ORDER BY rp.assigned_at ASC
    LIMIT 1
  `, [room.id]);

  if (!proctor) throw new Error('proctor not found');

  const participants = await pool.query(`
    SELECT st.nis, st.nama, p.token, p.seat_no
    FROM cbt_exam_participants p
    JOIN students st ON st.id = p.student_id
    WHERE p.session_id = $1
    ORDER BY p.seat_no ASC, st.nama ASC
  `, [session.id]);

  const payload = {
    event_title: event.event_title,
    session_title: session.session_title,
    subject_name: session.subject_name,
    class_name: session.class_name,
    academic_year_name: event.academic_year_name,
    room_name: room.room_name,
    room_token: room.room_token,
    room_capacity: room.capacity,
    scheduled_start: session.scheduled_start,
    scheduled_end: session.scheduled_end,
    proctor_name: proctor.nama,
    proctor_nip: proctor.nip,
    participants: participants.rows,
  };
  console.log(JSON.stringify(payload));
} finally {
  await pool.end();
}
"""
    proc = subprocess.run(
        ['node', '--input-type=module', '-e', js],
        cwd=SCRIPT_DIR,
        env={**os.environ, 'DATABASE_URL': DATABASE_URL},
        capture_output=True,
        text=True,
        check=True,
    )
    return json.loads(proc.stdout)


def to_seed_data(payload: dict) -> SeedData:
    return SeedData(
        event_title=payload['event_title'],
        session_title=payload['session_title'],
        subject_name=payload['subject_name'],
        class_name=payload['class_name'],
        academic_year_name=payload['academic_year_name'],
        room_name=payload['room_name'],
        room_token=payload['room_token'],
        room_capacity=int(payload['room_capacity']),
        scheduled_start=datetime.fromisoformat(payload['scheduled_start'].replace('Z', '+00:00')).astimezone(TZ),
        scheduled_end=datetime.fromisoformat(payload['scheduled_end'].replace('Z', '+00:00')).astimezone(TZ),
        proctor_name=payload['proctor_name'],
        proctor_nip=payload['proctor_nip'],
        participants=[
            Participant(
                nis=row['nis'],
                nama=row['nama'],
                token=row['token'],
                seat_no=int(row['seat_no']),
            )
            for row in payload['participants']
        ],
    )


def wrap_lines(text: str, font_name: str, font_size: int, max_width: float) -> list[str]:
    return simpleSplit(text, font_name, font_size, max_width)


def draw_header(c: canvas.Canvas, title: str, subtitle: str, page_label: str, pagesize=A4):
    width, height = pagesize
    c.setFillColor(colors.HexColor('#0f172a'))
    c.setFont('Helvetica-Bold', 12)
    c.drawString(16 * mm, height - 14 * mm, title)
    c.setFillColor(colors.HexColor('#475569'))
    c.setFont('Helvetica', 9)
    c.drawRightString(width - 16 * mm, height - 14 * mm, page_label)
    c.setStrokeColor(colors.HexColor('#cbd5e1'))
    c.line(16 * mm, height - 16 * mm, width - 16 * mm, height - 16 * mm)
    c.setFillColor(colors.HexColor('#334155'))
    c.setFont('Helvetica', 10)
    c.drawString(16 * mm, height - 22 * mm, subtitle)


MONTHS_ID = {
    1: 'Januari',
    2: 'Februari',
    3: 'Maret',
    4: 'April',
    5: 'Mei',
    6: 'Juni',
    7: 'Juli',
    8: 'Agustus',
    9: 'September',
    10: 'Oktober',
    11: 'November',
    12: 'Desember',
}


def fmt_wita(dt: datetime) -> str:
    return f"{dt.day:02d} {MONTHS_ID[dt.month]} {dt.year}, {dt:%H.%M} WITA"


def draw_card(c: canvas.Canvas, data: SeedData, participant: Participant, x: float, y: float, w: float, h: float):
    c.setStrokeColor(colors.HexColor('#0f766e'))
    c.setLineWidth(1.1)
    c.roundRect(x, y, w, h, 8, stroke=1, fill=0)

    c.setFillColor(colors.HexColor('#0f766e'))
    c.roundRect(x, y + h - 16 * mm, w, 16 * mm, 8, stroke=0, fill=1)
    c.setFillColor(colors.white)
    c.setFont('Helvetica-Bold', 12)
    c.drawString(x + 5 * mm, y + h - 9.5 * mm, 'KARTU UJIAN')
    c.setFont('Helvetica', 8.5)
    c.drawRightString(x + w - 5 * mm, y + h - 9.5 * mm, data.event_title)

    c.setFillColor(colors.HexColor('#0f172a'))
    c.setFont('Helvetica-Bold', 14)
    c.drawString(x + 5 * mm, y + h - 24 * mm, data.subject_name)
    c.setFont('Helvetica', 9.5)
    c.setFillColor(colors.HexColor('#475569'))
    c.drawString(x + 5 * mm, y + h - 30 * mm, f'Tahun ajaran: {data.academic_year_name}')
    c.drawString(x + 5 * mm, y + h - 35 * mm, f'Kelas: {data.class_name}')

    left_x = x + 5 * mm
    right_x = x + w / 2 + 3 * mm
    start_y = y + h - 44 * mm
    line_h = 6 * mm
    fields_left = [
        ('Nama', participant.nama),
        ('NIS', participant.nis),
        ('Ruang', data.room_name),
        ('Kursi', str(participant.seat_no)),
    ]
    fields_right = [
        ('Token', participant.token),
        ('Sesi', data.session_title),
        ('Mulai', fmt_wita(data.scheduled_start)),
        ('Selesai', fmt_wita(data.scheduled_end)),
    ]

    def draw_fields(base_x: float, fields: list[tuple[str, str]]):
        yy = start_y
        for label, value in fields:
            c.setFillColor(colors.HexColor('#0f172a'))
            c.setFont('Helvetica-Bold', 9)
            c.drawString(base_x, yy, f'{label}:')
            c.setFont('Helvetica', 9)
            c.setFillColor(colors.black)
            text = str(value)
            max_width = w / 2 - 12 * mm
            if label == 'Token':
                c.setFont('Courier-Bold', 11)
                c.setFillColor(colors.HexColor('#0f172a'))
                c.drawString(base_x, yy - 1.5 * mm, text)
                yy -= line_h
                continue
            wrapped = wrap_lines(text, 'Helvetica', 9, max_width)
            if not wrapped:
                wrapped = ['-']
            c.drawString(base_x + 17 * mm, yy, wrapped[0])
            extra_y = yy
            for extra in wrapped[1:]:
                extra_y -= 4.2 * mm
                c.drawString(base_x + 17 * mm, extra_y, extra)
            yy -= line_h * max(1, len(wrapped))

    draw_fields(left_x, fields_left)
    draw_fields(right_x, fields_right)

    c.setStrokeColor(colors.HexColor('#cbd5e1'))
    c.line(x + 5 * mm, y + 16 * mm, x + w - 5 * mm, y + 16 * mm)
    c.setFillColor(colors.HexColor('#475569'))
    c.setFont('Helvetica', 8)
    c.drawString(x + 5 * mm, y + 11 * mm, 'Tunjukkan kartu ini saat masuk ruang ujian. Token hanya untuk satu peserta.')
    c.drawRightString(x + w - 5 * mm, y + 11 * mm, f'Kapasitas ruang: {data.room_capacity}')


def build_cards_pdf(data: SeedData):
    c = canvas.Canvas(str(CARD_PDF), pagesize=A4)
    c.setTitle('Kartu Ujian Dev Test CBT')
    c.setAuthor('Hermes Agent')
    c.setSubject('Paket cetak kartu ujian untuk development test')
    width, height = A4
    card_w = width - 32 * mm
    card_h = 112 * mm
    gap = 8 * mm
    top_y = height - 22 * mm - card_h
    bottom_y = 18 * mm

    for index, participant in enumerate(data.participants):
        if index % 2 == 0:
            page_no = index // 2 + 1
            draw_header(c, 'Paket Cetak Kartu Ujian', f'{data.event_title} • {data.room_name}', f'Halaman {page_no}', pagesize=A4)
            draw_card(c, data, participant, 16 * mm, top_y, card_w, card_h)
        else:
            draw_card(c, data, participant, 16 * mm, bottom_y, card_w, card_h)
            c.showPage()
    if len(data.participants) % 2 == 1:
        c.showPage()
    c.save()


def build_attendance_pdf(data: SeedData):
    pagesize = landscape(A4)
    c = canvas.Canvas(str(ATTENDANCE_PDF), pagesize=pagesize)
    c.setTitle('Lembar Hadir Dev Test CBT')
    c.setAuthor('Hermes Agent')
    c.setSubject('Lembar hadir dan verifikasi untuk development test')
    width, height = pagesize
    draw_header(c, 'Lembar Hadir & Verifikasi', f'{data.event_title} • {data.room_name}', 'Halaman 1', pagesize=pagesize)

    c.setFillColor(colors.HexColor('#0f172a'))
    c.setFont('Helvetica-Bold', 15)
    c.drawString(16 * mm, height - 30 * mm, data.event_title)
    c.setFont('Helvetica', 10)
    c.setFillColor(colors.HexColor('#334155'))
    meta_lines = [
        f'Sesi: {data.session_title}',
        f'Pengawas: {data.proctor_name} ({data.proctor_nip})',
        f'Ruang: {data.room_name} | Token Ruang: {data.room_token} | Kapasitas: {data.room_capacity}',
        f'Jadwal: {fmt_wita(data.scheduled_start)} sampai {fmt_wita(data.scheduled_end)}',
    ]
    yy = height - 36 * mm
    for line in meta_lines:
        c.drawString(16 * mm, yy, line)
        yy -= 5 * mm

    table_x = 16 * mm
    table_y = 18 * mm
    table_w = width - 32 * mm
    table_h = height - 74 * mm
    row_count = len(data.participants) + 1
    row_h = table_h / row_count

    cols = [
        ('No', 12 * mm),
        ('Nama', 64 * mm),
        ('NIS', 28 * mm),
        ('Token', 35 * mm),
        ('Kursi', 18 * mm),
        ('Tanda tangan', table_w - (12 + 64 + 28 + 35 + 18) * mm),
    ]
    c.setStrokeColor(colors.HexColor('#94a3b8'))
    c.setFillColor(colors.HexColor('#e2e8f0'))
    c.rect(table_x, table_y + table_h - row_h, table_w, row_h, stroke=1, fill=1)

    # Header row
    cx = table_x
    c.setFillColor(colors.HexColor('#0f172a'))
    c.setFont('Helvetica-Bold', 9)
    for label, col_w in cols:
        c.rect(cx, table_y + table_h - row_h, col_w, row_h, stroke=1, fill=0)
        c.drawCentredString(cx + col_w / 2, table_y + table_h - row_h + 3.2 * mm, label)
        cx += col_w

    # Data rows
    for i, participant in enumerate(data.participants, start=1):
        y = table_y + table_h - row_h * (i + 1)
        c.setFillColor(colors.white)
        c.rect(table_x, y, table_w, row_h, stroke=1, fill=1)
        values = [str(i), participant.nama, participant.nis, participant.token, str(participant.seat_no), '']
        cx = table_x
        for idx, ((_, col_w), value) in enumerate(zip(cols, values)):
            c.rect(cx, y, col_w, row_h, stroke=1, fill=0)
            c.setFont('Helvetica', 8.5 if idx != 5 else 8)
            c.setFillColor(colors.black)
            if idx == 0 or idx == 4:
                c.drawCentredString(cx + col_w / 2, y + 2.8 * mm, value)
            elif idx == 5:
                pass
            else:
                trimmed = value
                if len(trimmed) > 32:
                    trimmed = trimmed[:29] + '...'
                c.drawString(cx + 2 * mm, y + 2.8 * mm, trimmed)
            cx += col_w

    c.setFillColor(colors.HexColor('#475569'))
    c.setFont('Helvetica', 8.5)
    c.drawString(16 * mm, 10 * mm, 'Catatan: tanda tangan peserta ditulis pada kolom terakhir. Jika perlu, cetak ulang satu lembar khusus untuk cadangan.')
    c.showPage()
    c.save()


def main():
    OUT_DIR.mkdir(parents=True, exist_ok=True)
    payload = run_node_query()
    data = to_seed_data(payload)
    build_cards_pdf(data)
    build_attendance_pdf(data)
    print(json.dumps({
        'cards_pdf': str(CARD_PDF),
        'attendance_pdf': str(ATTENDANCE_PDF),
        'participants': len(data.participants),
        'room': data.room_name,
    }, ensure_ascii=False))


if __name__ == '__main__':
    main()
