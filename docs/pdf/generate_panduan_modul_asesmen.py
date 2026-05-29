from pathlib import Path
from reportlab.lib.pagesizes import A4
from reportlab.lib.units import mm
from reportlab.lib import colors
from reportlab.pdfgen import canvas
from reportlab.lib.utils import ImageReader
from reportlab.pdfbase.pdfmetrics import stringWidth

ROOT = Path('/home/servermtsn2kolut/mtsn2kolut-super-app')
OUT = ROOT / 'docs/pdf/panduan-teknis-modul-asesmen.pdf'
ASSET = ROOT / 'docs/pdf/assets-asesmen-panduan-current'

W, H = A4
M = 16 * mm
TITLE = 'Panduan Teknis Modul Asesmen'
SUB = 'mtsn2kolut super app'
FONT = 'Helvetica'
FONT_BOLD = 'Helvetica-Bold'


def wrap_text(text, font, size, max_width):
    words = text.split()
    lines = []
    current = ''
    for word in words:
        trial = f'{current} {word}'.strip()
        if stringWidth(trial, font, size) <= max_width or not current:
            current = trial
        else:
            lines.append(current)
            current = word
    if current:
        lines.append(current)
    return lines


def draw_wrapped(c, text, x, y, font=FONT, size=11, leading=14, max_width=None, color=colors.black):
    if max_width is None:
        max_width = W - 2 * M
    c.setFillColor(color)
    c.setFont(font, size)
    yy = y
    for line in wrap_text(text, font, size, max_width):
        c.drawString(x, yy, line)
        yy -= leading
    return yy


def draw_page_header(c, page_num):
    c.setFillColor(colors.HexColor('#0f172a'))
    c.setFont(FONT_BOLD, 11)
    c.drawString(M, H - M + 2, TITLE)
    c.setFont(FONT, 9)
    c.setFillColor(colors.HexColor('#475569'))
    c.drawRightString(W - M, H - M + 2, f'Halaman {page_num}')
    c.setStrokeColor(colors.HexColor('#dbeafe'))
    c.line(M, H - M - 4, W - M, H - M - 4)


def add_cover(c):
    draw_page_header(c, 1)
    y = H - 32 * mm
    c.setFillColor(colors.HexColor('#14532d'))
    c.setFont(FONT_BOLD, 20)
    c.drawString(M, y, TITLE)
    y -= 10 * mm
    c.setFont(FONT, 11)
    c.setFillColor(colors.HexColor('#334155'))
    for line in [
        'Edisi revisi menggunakan screenshot dari mtsn2kolut super app yang aktif.',
        'Fokus dokumentasi: halaman setelah autentikasi di area Asesmen/CBT.',
        'Halaman public/login tidak ditampilkan di panduan ini.',
    ]:
        c.drawString(M, y, line)
        y -= 6 * mm

    y -= 4 * mm
    c.setFillColor(colors.HexColor('#0f766e'))
    c.roundRect(M, y - 34 * mm, W - 2 * M, 30 * mm, 6 * mm, stroke=0, fill=1)
    c.setFillColor(colors.white)
    c.setFont(FONT_BOLD, 12)
    c.drawString(M + 5 * mm, y - 10 * mm, 'Ringkasan alur pakai')
    c.setFont(FONT, 10)
    bullets = [
        '1. Buka area Asesmen setelah akun sudah login.',
        '2. Peserta masuk lewat Portal Ujian Peserta.',
        '3. Cek identitas sebelum lanjut ke ujian.',
        '4. Pengawas membuka Portal Pengawasan Ruang.',
        '5. Pantau status ruang, peringatan, dan peserta.',
    ]
    by = y - 15 * mm
    for b in bullets:
        c.drawString(M + 6 * mm, by, b)
        by -= 4.7 * mm

    y2 = 95 * mm
    c.setFillColor(colors.HexColor('#fee2e2'))
    c.roundRect(M, y2, W - 2 * M, 18 * mm, 5 * mm, stroke=0, fill=1)
    c.setFillColor(colors.HexColor('#7f1d1d'))
    c.setFont(FONT_BOLD, 10)
    c.drawString(M + 5 * mm, y2 + 11 * mm, 'Catatan')
    c.setFont(FONT, 9.5)
    c.drawString(M + 5 * mm, y2 + 6.5 * mm, 'Garis tunjuk merah menandai area yang perlu dibaca lebih dulu.')
    c.drawString(M + 5 * mm, y2 + 2.2 * mm, 'Beberapa data pada screenshot disamarkan bila sensitif.')

    c.showPage()


def add_image_page(c, page_num, img_path, title, caption):
    draw_page_header(c, page_num)
    top = H - 28 * mm
    c.setFillColor(colors.HexColor('#0f172a'))
    c.setFont(FONT_BOLD, 16)
    c.drawString(M, top, title)
    c.setFont(FONT, 10)
    c.setFillColor(colors.HexColor('#475569'))
    draw_wrapped(c, caption, M, top - 6 * mm, font=FONT, size=10, leading=12, max_width=W - 2 * M)

    img = ImageReader(str(img_path))
    iw, ih = img.getSize()
    avail_w = W - 2 * M
    avail_h = H - 66 * mm
    scale = min(avail_w / iw, avail_h / ih)
    dw = iw * scale
    dh = ih * scale
    x = (W - dw) / 2
    y = 18 * mm
    c.setStrokeColor(colors.HexColor('#cbd5e1'))
    c.roundRect(x - 2, y - 2, dw + 4, dh + 4, 6, stroke=1, fill=0)
    c.drawImage(str(img_path), x, y, width=dw, height=dh, preserveAspectRatio=True, mask='auto')

    c.setFillColor(colors.HexColor('#64748b'))
    c.setFont(FONT, 8.5)
    c.drawString(M, 8 * mm, 'Panduan ini khusus halaman setelah login; tidak menampilkan layar public/login.')
    c.showPage()


def main():
    c = canvas.Canvas(str(OUT), pagesize=A4)
    c.setTitle(TITLE)
    c.setAuthor('Hermes Agent')
    c.setSubject('Panduan teknis modul asesmen')
    c.setCreator('ReportLab')

    add_cover(c)
    add_image_page(
        c, 2,
        ASSET / '02-ujian-form-annotated.png',
        'Portal Ujian Peserta',
        'Form ini dipakai setelah peserta masuk ke area internal. Fokus baca: kode kartu/QR token, PIN, lalu tombol Lanjutkan.'
    )
    add_image_page(
        c, 3,
        ASSET / '03-ujian-konfirmasi-annotated.png',
        'Konfirmasi Identitas Peserta',
        'Halaman ini muncul setelah QR/PIN diterima. Pastikan nama, kelas/NIS, ruang/meja, dan sesi ujian sudah benar sebelum lanjut.'
    )
    add_image_page(
        c, 4,
        ASSET / '04-pengawas-annotated.png',
        'Portal Pengawasan Ruang',
        'Dashboard pengawas untuk memantau ruang aktif, tab Ruang/Peringatan/Peserta, dan tombol aksi operasional.'
    )

    c.save()
    print(OUT)


if __name__ == '__main__':
    main()
