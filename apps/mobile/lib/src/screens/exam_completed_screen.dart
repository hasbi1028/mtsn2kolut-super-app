import 'package:flutter/material.dart';

class ExamCompletedScreen extends StatelessWidget {
  const ExamCompletedScreen({
    super.key,
    required this.studentName,
    required this.studentNis,
    required this.sessionTitle,
    required this.roomName,
    required this.answeredCount,
    required this.totalQuestions,
    required this.wasAutoSubmitted,
  });

  final String studentName;
  final String studentNis;
  final String sessionTitle;
  final String roomName;
  final int answeredCount;
  final int totalQuestions;
  final bool wasAutoSubmitted;

  @override
  Widget build(BuildContext context) {
    final theme = Theme.of(context);

    return PopScope(
      canPop: false,
      child: Scaffold(
        body: SafeArea(
          child: Center(
            child: ConstrainedBox(
              constraints: const BoxConstraints(maxWidth: 720),
              child: Padding(
                padding: const EdgeInsets.all(24),
                child: Card(
                  child: Padding(
                    padding: const EdgeInsets.all(28),
                    child: Column(
                      mainAxisSize: MainAxisSize.min,
                      crossAxisAlignment: CrossAxisAlignment.start,
                      children: [
                        Container(
                          padding: const EdgeInsets.symmetric(
                            horizontal: 14,
                            vertical: 8,
                          ),
                          decoration: BoxDecoration(
                            color: theme.colorScheme.primary.withValues(
                              alpha: 0.12,
                            ),
                            borderRadius: BorderRadius.circular(999),
                          ),
                          child: Text(
                            'CBT MTsN 2 Kolaka Utara',
                            style: theme.textTheme.labelLarge?.copyWith(
                              color: theme.colorScheme.primary,
                              fontWeight: FontWeight.w800,
                            ),
                          ),
                        ),
                        const SizedBox(height: 18),
                        Text(
                          wasAutoSubmitted
                              ? 'Ujian ditutup otomatis.'
                              : 'Ujian berhasil dikirim.',
                          style: theme.textTheme.headlineMedium?.copyWith(
                            fontWeight: FontWeight.w800,
                            color: const Color(0xFF14361D),
                          ),
                        ),
                        const SizedBox(height: 12),
                        Text(
                          wasAutoSubmitted
                              ? 'Waktu ujian telah habis dan jawaban Anda sudah dikirim ke server.'
                              : 'Jawaban Anda sudah diterima server. Silakan menunggu arahan pengawas.',
                          style: theme.textTheme.bodyLarge?.copyWith(
                            height: 1.55,
                          ),
                        ),
                        const SizedBox(height: 24),
                        Wrap(
                          spacing: 12,
                          runSpacing: 12,
                          children: [
                            _SummaryTile(label: 'Nama', value: studentName),
                            _SummaryTile(label: 'NIS', value: studentNis),
                            _SummaryTile(label: 'Ruang', value: roomName),
                            _SummaryTile(
                              label: 'Terjawab',
                              value: '$answeredCount / $totalQuestions soal',
                            ),
                          ],
                        ),
                        const SizedBox(height: 24),
                        Container(
                          width: double.infinity,
                          padding: const EdgeInsets.all(18),
                          decoration: BoxDecoration(
                            color: const Color(0xFFF6F8F3),
                            borderRadius: BorderRadius.circular(20),
                            border: Border.all(
                              color: theme.colorScheme.outlineVariant,
                            ),
                          ),
                          child: Column(
                            crossAxisAlignment: CrossAxisAlignment.start,
                            children: [
                              Text(
                                sessionTitle,
                                style: theme.textTheme.titleMedium?.copyWith(
                                  fontWeight: FontWeight.w800,
                                ),
                              ),
                              const SizedBox(height: 8),
                              Text(
                                'Tampilan ini menandakan sesi pada perangkat ini sudah selesai. '
                                'Jangan menutup atau meminjamkan perangkat sebelum pengawas memastikan proses ujian benar-benar selesai.',
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
            ),
          ),
        ),
      ),
    );
  }
}

class _SummaryTile extends StatelessWidget {
  const _SummaryTile({required this.label, required this.value});

  final String label;
  final String value;

  @override
  Widget build(BuildContext context) {
    final theme = Theme.of(context);

    return Container(
      width: 150,
      padding: const EdgeInsets.all(16),
      decoration: BoxDecoration(
        color: Colors.white,
        borderRadius: BorderRadius.circular(18),
        border: Border.all(color: theme.colorScheme.outlineVariant),
      ),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          Text(label, style: theme.textTheme.bodySmall),
          const SizedBox(height: 6),
          Text(
            value,
            style: theme.textTheme.titleMedium?.copyWith(
              fontWeight: FontWeight.w800,
            ),
          ),
        ],
      ),
    );
  }
}
