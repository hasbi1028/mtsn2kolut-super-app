import 'package:flutter/material.dart';

class ExamStatusGuideScreen extends StatelessWidget {
  const ExamStatusGuideScreen({super.key});

  @override
  Widget build(BuildContext context) {
    final theme = Theme.of(context);

    return Scaffold(
      appBar: AppBar(title: const Text('Panduan Status Ujian')),
      body: SafeArea(
        child: Center(
          child: ConstrainedBox(
            constraints: const BoxConstraints(maxWidth: 860),
            child: ListView(
              padding: const EdgeInsets.all(24),
              children: [
                Text(
                  'Arti status koneksi di aplikasi peserta',
                  style: theme.textTheme.headlineMedium?.copyWith(
                    fontWeight: FontWeight.w800,
                    color: const Color(0xFF14361D),
                  ),
                ),
                const SizedBox(height: 12),
                Text(
                  'Panduan ini membantu siswa dan pengawas membaca kondisi sinkron perangkat selama ujian BYOD berlangsung.',
                  style: theme.textTheme.bodyLarge?.copyWith(height: 1.55),
                ),
                const SizedBox(height: 24),
                const _GuideCard(
                  title: 'Aman',
                  subtitle: 'Kondisi stabil',
                  description:
                      'Perangkat terakhir berhasil terhubung ke server dan tidak ada jawaban yang menunggu dikirim.',
                  tone: _GuideTone.good,
                  icon: Icons.cloud_done,
                ),
                const SizedBox(height: 14),
                const _GuideCard(
                  title: 'Perlu sinkron',
                  subtitle: 'Jawaban aman di perangkat',
                  description:
                      'Sebagian jawaban masih tersimpan di perangkat dan perlu dikirim ulang ke server. Siswa tetap berada di layar ujian.',
                  tone: _GuideTone.warning,
                  icon: Icons.save_outlined,
                ),
                const SizedBox(height: 14),
                const _GuideCard(
                  title: 'Perlu pengawas',
                  subtitle: 'Koneksi atau sesi perlu dicek',
                  description:
                      'Aplikasi mendeteksi sesi perlu dicek ulang, kontak server sudah stale, atau koneksi berulang kali gagal.',
                  tone: _GuideTone.danger,
                  icon: Icons.cloud_off,
                ),
                const SizedBox(height: 14),
                const _GuideCard(
                  title: 'Tidak didukung aplikasi siswa',
                  subtitle: 'Butuh prosedur ruang',
                  description:
                      'Tipe soal tertentu seperti hotspot atau unggah berkas belum dijawab langsung di aplikasi siswa. Pengawas perlu mencatat tindak lanjut sesuai prosedur.',
                  tone: _GuideTone.danger,
                  icon: Icons.support_agent,
                ),
                const SizedBox(height: 24),
                Container(
                  padding: const EdgeInsets.all(18),
                  decoration: BoxDecoration(
                    color: const Color(0xFFF6F8F3),
                    borderRadius: BorderRadius.circular(20),
                    border: Border.all(color: theme.colorScheme.outlineVariant),
                  ),
                  child: Column(
                    crossAxisAlignment: CrossAxisAlignment.start,
                    children: [
                      Text(
                        'Catatan untuk pengawas',
                        style: theme.textTheme.titleMedium?.copyWith(
                          fontWeight: FontWeight.w800,
                        ),
                      ),
                      const SizedBox(height: 10),
                      Text(
                        'Jika status berubah ke Perlu pengawas, minta siswa tetap berada di layar ujian, tekan tombol sinkron ulang, dan tunggu sampai status kembali Aman atau Perlu sinkron selesai sebelum mengizinkan kirim jawaban akhir.',
                        style: theme.textTheme.bodyMedium?.copyWith(
                          height: 1.55,
                        ),
                      ),
                    ],
                  ),
                ),
              ],
            ),
          ),
        ),
      ),
    );
  }
}

enum _GuideTone { good, warning, danger }

class _GuideCard extends StatelessWidget {
  const _GuideCard({
    required this.title,
    required this.subtitle,
    required this.description,
    required this.tone,
    required this.icon,
  });

  final String title;
  final String subtitle;
  final String description;
  final _GuideTone tone;
  final IconData icon;

  @override
  Widget build(BuildContext context) {
    final theme = Theme.of(context);
    final (background, border, foreground) = switch (tone) {
      _GuideTone.good => (
        const Color(0xFFE8F5EC),
        const Color(0xFFA9D4B8),
        const Color(0xFF0B7A3B),
      ),
      _GuideTone.warning => (
        const Color(0xFFFFF3D8),
        const Color(0xFFE5C172),
        const Color(0xFF9A6700),
      ),
      _GuideTone.danger => (
        const Color(0xFFFDE7E9),
        const Color(0xFFE8A5AB),
        const Color(0xFFC03645),
      ),
    };

    return Container(
      padding: const EdgeInsets.all(18),
      decoration: BoxDecoration(
        color: background,
        borderRadius: BorderRadius.circular(20),
        border: Border.all(color: border),
      ),
      child: Row(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          Icon(icon, color: foreground, size: 24),
          const SizedBox(width: 14),
          Expanded(
            child: Column(
              crossAxisAlignment: CrossAxisAlignment.start,
              children: [
                Text(
                  title,
                  style: theme.textTheme.titleMedium?.copyWith(
                    color: foreground,
                    fontWeight: FontWeight.w800,
                  ),
                ),
                const SizedBox(height: 4),
                Text(
                  subtitle,
                  style: theme.textTheme.labelLarge?.copyWith(
                    color: foreground,
                    fontWeight: FontWeight.w700,
                  ),
                ),
                const SizedBox(height: 8),
                Text(
                  description,
                  style: theme.textTheme.bodyMedium?.copyWith(height: 1.5),
                ),
              ],
            ),
          ),
        ],
      ),
    );
  }
}
