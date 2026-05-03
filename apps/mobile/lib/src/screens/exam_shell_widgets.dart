import 'package:flutter/material.dart';

import '../exam_error_messages.dart';
import 'exam_shell_connection.dart';

class StatTile extends StatelessWidget {
  const StatTile({
    super.key,
    required this.label,
    required this.value,
    this.accent = false,
  });

  final String label;
  final String value;
  final bool accent;

  @override
  Widget build(BuildContext context) {
    final theme = Theme.of(context);
    return Container(
      width: double.infinity,
      padding: const EdgeInsets.all(16),
      decoration: BoxDecoration(
        color: accent ? Colors.red.shade50 : const Color(0xFFF6F8F3),
        borderRadius: BorderRadius.circular(18),
      ),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          Text(label, style: theme.textTheme.bodySmall),
          const SizedBox(height: 4),
          Text(
            value,
            style: theme.textTheme.titleLarge?.copyWith(
              fontWeight: FontWeight.w800,
              color: accent ? Colors.red.shade700 : null,
            ),
          ),
        ],
      ),
    );
  }
}

class InlineMessage extends StatelessWidget {
  const InlineMessage({super.key, required this.tone, required this.message});

  final BannerTone tone;
  final String message;

  @override
  Widget build(BuildContext context) {
    final theme = Theme.of(context);
    final color = switch (tone) {
      BannerTone.error => Colors.red.shade700,
      BannerTone.success => const Color(0xFF0B7A3B),
      BannerTone.info => theme.colorScheme.primary,
    };

    return Container(
      width: double.infinity,
      padding: const EdgeInsets.all(14),
      decoration: BoxDecoration(
        color: color.withValues(alpha: 0.08),
        borderRadius: BorderRadius.circular(16),
        border: Border.all(color: color.withValues(alpha: 0.18)),
      ),
      child: Text(
        message,
        style: theme.textTheme.bodyMedium?.copyWith(
          color: color,
          fontWeight: FontWeight.w600,
        ),
      ),
    );
  }
}

class SyncStatusChip extends StatelessWidget {
  const SyncStatusChip({super.key, required this.label, required this.tone});

  final String label;
  final SyncTone tone;

  @override
  Widget build(BuildContext context) {
    final theme = Theme.of(context);
    final (background, foreground) = switch (tone) {
      SyncTone.success => (const Color(0xFFE8F5EC), const Color(0xFF0B7A3B)),
      SyncTone.warning => (const Color(0xFFFFF3D8), const Color(0xFF9A6700)),
      SyncTone.danger => (const Color(0xFFFDE7E9), Colors.red.shade700),
    };

    return Container(
      padding: const EdgeInsets.symmetric(horizontal: 12, vertical: 7),
      decoration: BoxDecoration(
        color: background,
        borderRadius: BorderRadius.circular(999),
      ),
      child: Text(
        label,
        style: theme.textTheme.labelMedium?.copyWith(
          color: foreground,
          fontWeight: FontWeight.w800,
        ),
      ),
    );
  }
}

class QuestionMediaCard extends StatelessWidget {
  const QuestionMediaCard({
    super.key,
    required this.url,
    this.headers = const <String, String>{},
  });

  final String url;
  final Map<String, String> headers;

  @override
  Widget build(BuildContext context) {
    final theme = Theme.of(context);
    return Container(
      width: double.infinity,
      padding: const EdgeInsets.all(12),
      decoration: BoxDecoration(
        color: const Color(0xFFF6F8F3),
        borderRadius: BorderRadius.circular(18),
        border: Border.all(color: theme.colorScheme.outlineVariant),
      ),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          Text(
            'Media soal',
            style: theme.textTheme.labelLarge?.copyWith(
              color: theme.colorScheme.primary,
              fontWeight: FontWeight.w800,
            ),
          ),
          const SizedBox(height: 10),
          ClipRRect(
            borderRadius: BorderRadius.circular(14),
            child: Image.network(
              url,
              headers: headers.isEmpty ? null : headers,
              fit: BoxFit.contain,
              errorBuilder: (context, _, _) {
                return Container(
                  width: double.infinity,
                  padding: const EdgeInsets.all(18),
                  color: Colors.white,
                  child: Text(
                    'Media tidak dapat dimuat.\n$url',
                    style: theme.textTheme.bodySmall,
                  ),
                );
              },
              loadingBuilder: (context, child, progress) {
                if (progress == null) {
                  return child;
                }
                return Container(
                  width: double.infinity,
                  padding: const EdgeInsets.all(24),
                  color: Colors.white,
                  child: Column(
                    mainAxisSize: MainAxisSize.min,
                    children: [
                      const CircularProgressIndicator(strokeWidth: 2),
                      const SizedBox(height: 10),
                      Text('Memuat media...', style: theme.textTheme.bodySmall),
                    ],
                  ),
                );
              },
            ),
          ),
        ],
      ),
    );
  }
}

class ConnectionWarningCard extends StatelessWidget {
  const ConnectionWarningCard({
    super.key,
    required this.failureCount,
    required this.lastFailureAt,
    required this.onRetry,
  });

  final int failureCount;
  final String lastFailureAt;
  final VoidCallback? onRetry;

  @override
  Widget build(BuildContext context) {
    final theme = Theme.of(context);
    return Container(
      width: double.infinity,
      padding: const EdgeInsets.all(16),
      decoration: BoxDecoration(
        color: const Color(0xFFFFF3D8),
        borderRadius: BorderRadius.circular(18),
        border: Border.all(color: const Color(0xFFE5C172)),
      ),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          Text(
            'Koneksi perlu diperhatikan',
            style: theme.textTheme.titleMedium?.copyWith(
              color: const Color(0xFF9A6700),
              fontWeight: FontWeight.w800,
            ),
          ),
          const SizedBox(height: 8),
          Text(
            'Perangkat mengalami $failureCount gangguan sinkron berturut-turut. Terakhir tercatat pukul $lastFailureAt.',
            style: theme.textTheme.bodyMedium?.copyWith(height: 1.5),
          ),
          const SizedBox(height: 12),
          OutlinedButton.icon(
            onPressed: onRetry,
            icon: const Icon(Icons.sync),
            label: const Text('Coba Sinkron Ulang'),
          ),
        ],
      ),
    );
  }
}

class ConnectionHealthCard extends StatelessWidget {
  const ConnectionHealthCard({
    super.key,
    required this.label,
    required this.description,
    required this.tone,
  });

  final String label;
  final String description;
  final HealthTone tone;

  @override
  Widget build(BuildContext context) {
    final theme = Theme.of(context);
    final (background, border, foreground, icon) = switch (tone) {
      HealthTone.good => (
        const Color(0xFFE8F5EC),
        const Color(0xFFA9D4B8),
        const Color(0xFF0B7A3B),
        Icons.cloud_done,
      ),
      HealthTone.warning => (
        const Color(0xFFFFF3D8),
        const Color(0xFFE5C172),
        const Color(0xFF9A6700),
        Icons.save_outlined,
      ),
      HealthTone.danger => (
        const Color(0xFFFDE7E9),
        const Color(0xFFE8A5AB),
        const Color(0xFFC03645),
        Icons.sync_problem,
      ),
    };

    return Container(
      width: double.infinity,
      padding: const EdgeInsets.all(16),
      decoration: BoxDecoration(
        color: background,
        borderRadius: BorderRadius.circular(18),
        border: Border.all(color: border),
      ),
      child: Row(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          Icon(icon, color: foreground),
          const SizedBox(width: 12),
          Expanded(
            child: Column(
              crossAxisAlignment: CrossAxisAlignment.start,
              children: [
                Text(
                  label,
                  style: theme.textTheme.titleMedium?.copyWith(
                    color: foreground,
                    fontWeight: FontWeight.w800,
                  ),
                ),
                const SizedBox(height: 6),
                Text(
                  description,
                  style: theme.textTheme.bodySmall?.copyWith(height: 1.5),
                ),
              ],
            ),
          ),
        ],
      ),
    );
  }
}

class DegradedModeCard extends StatelessWidget {
  const DegradedModeCard({
    super.key,
    required this.pendingCount,
    required this.failureCount,
    required this.onRetry,
  });

  final int pendingCount;
  final int failureCount;
  final VoidCallback? onRetry;

  @override
  Widget build(BuildContext context) {
    final theme = Theme.of(context);
    return Container(
      width: double.infinity,
      padding: const EdgeInsets.all(16),
      decoration: BoxDecoration(
        color: const Color(0xFFFDE7E9),
        borderRadius: BorderRadius.circular(18),
        border: Border.all(color: const Color(0xFFE8A5AB)),
      ),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          Text(
            'Mode koneksi menurun aktif',
            style: theme.textTheme.titleMedium?.copyWith(
              color: Colors.red.shade700,
              fontWeight: FontWeight.w800,
            ),
          ),
          const SizedBox(height: 8),
          Text(
            pendingCount > 0
                ? 'Sinkron gagal $failureCount kali berturut-turut dan masih ada $pendingCount jawaban lokal. Kirim ujian ditahan sampai koneksi minimal pulih.'
                : 'Sinkron gagal $failureCount kali berturut-turut. Perbarui status dulu sebelum mengirim ujian.',
            style: theme.textTheme.bodyMedium?.copyWith(height: 1.5),
          ),
          const SizedBox(height: 12),
          FilledButton.icon(
            onPressed: onRetry,
            icon: const Icon(Icons.sync_problem),
            label: const Text('Pulihkan Sinkron'),
          ),
        ],
      ),
    );
  }
}

class SupervisorAttentionCard extends StatelessWidget {
  const SupervisorAttentionCard({
    super.key,
    required this.lastContactAt,
    required this.staleDuration,
    required this.escalationThreshold,
    required this.escalated,
    required this.onRetry,
  });

  final String lastContactAt;
  final String staleDuration;
  final String escalationThreshold;
  final bool escalated;
  final VoidCallback? onRetry;

  @override
  Widget build(BuildContext context) {
    final theme = Theme.of(context);
    final titleColor = escalated
        ? Colors.red.shade700
        : const Color(0xFF9A6700);
    final background = escalated
        ? const Color(0xFFFDE7E9)
        : const Color(0xFFFFF3D8);
    final border = escalated
        ? const Color(0xFFE8A5AB)
        : const Color(0xFFE5C172);
    final action = escalated
        ? FilledButton.icon(
            onPressed: onRetry,
            icon: const Icon(Icons.priority_high),
            label: const Text('Intervensi dan Sinkron Ulang'),
          )
        : OutlinedButton.icon(
            onPressed: onRetry,
            icon: const Icon(Icons.support_agent),
            label: const Text('Periksa dan Sinkron Ulang'),
          );

    return Container(
      width: double.infinity,
      padding: const EdgeInsets.all(16),
      decoration: BoxDecoration(
        color: background,
        borderRadius: BorderRadius.circular(18),
        border: Border.all(color: border),
      ),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          Text(
            escalated
                ? 'Pengawas harus segera intervensi'
                : 'Perlu intervensi pengawas',
            style: theme.textTheme.titleMedium?.copyWith(
              color: titleColor,
              fontWeight: FontWeight.w800,
            ),
          ),
          const SizedBox(height: 8),
          Text(
            escalated
                ? 'Status koneksi bertahan di level waspada selama $staleDuration sejak kontak server terakhir pukul $lastContactAt. Pengawas sebaiknya segera memeriksa perangkat, jaringan, dan memastikan sinkron ulang berhasil sebelum peserta melanjutkan tanpa pengawasan.'
                : 'Status koneksi berada di level waspada selama $staleDuration sejak kontak server terakhir pukul $lastContactAt. Minta pengawas memeriksa jaringan perangkat lalu lakukan sinkron ulang.',
            style: theme.textTheme.bodyMedium?.copyWith(height: 1.5),
          ),
          const SizedBox(height: 10),
          Container(
            width: double.infinity,
            padding: const EdgeInsets.all(12),
            decoration: BoxDecoration(
              color: Colors.white.withValues(alpha: 0.7),
              borderRadius: BorderRadius.circular(14),
            ),
            child: Text(
              escalated
                  ? 'Ambang eskalasi keras sudah terlewati setelah $escalationThreshold tanpa kontak server baru.'
                  : 'Jika kondisi ini bertahan sampai $escalationThreshold tanpa kontak server baru, panel ini akan naik ke mode intervensi keras.',
              style: theme.textTheme.bodySmall?.copyWith(
                height: 1.45,
                fontWeight: FontWeight.w600,
              ),
            ),
          ),
          const SizedBox(height: 12),
          action,
        ],
      ),
    );
  }
}

class QuestionAudioStatusChip extends StatelessWidget {
  const QuestionAudioStatusChip({super.key, required this.hasPlayed});

  final bool hasPlayed;

  @override
  Widget build(BuildContext context) {
    final theme = Theme.of(context);
    final background = hasPlayed
        ? theme.colorScheme.primary.withValues(alpha: 0.12)
        : const Color(0xFFFFF3D8);
    final foreground = hasPlayed
        ? theme.colorScheme.primary
        : const Color(0xFF9A6700);

    return Container(
      padding: const EdgeInsets.symmetric(horizontal: 12, vertical: 8),
      decoration: BoxDecoration(
        color: background,
        borderRadius: BorderRadius.circular(999),
      ),
      child: Row(
        mainAxisSize: MainAxisSize.min,
        children: [
          Icon(
            hasPlayed ? Icons.headset_mic : Icons.hearing_outlined,
            size: 16,
            color: foreground,
          ),
          const SizedBox(width: 8),
          Text(
            hasPlayed ? 'Audio soal sudah diputar' : 'Audio soal belum diputar',
            style: theme.textTheme.labelMedium?.copyWith(
              color: foreground,
              fontWeight: FontWeight.w700,
            ),
          ),
        ],
      ),
    );
  }
}

class ExamGuidanceCard extends StatelessWidget {
  const ExamGuidanceCard({super.key, required this.notice});

  final ExamGuidanceNotice notice;

  @override
  Widget build(BuildContext context) {
    final theme = Theme.of(context);
    final (background, border, foreground, icon) = switch (notice.tone) {
      ExamGuidanceTone.info => (
        const Color(0xFFE8F5EC),
        const Color(0xFFA9D4B8),
        const Color(0xFF0B7A3B),
        Icons.task_alt,
      ),
      ExamGuidanceTone.warning => (
        const Color(0xFFFFF3D8),
        const Color(0xFFE5C172),
        const Color(0xFF9A6700),
        Icons.warning_amber_rounded,
      ),
      ExamGuidanceTone.danger => (
        const Color(0xFFFDE7E9),
        const Color(0xFFE8A5AB),
        const Color(0xFFC03645),
        Icons.sync_problem,
      ),
    };

    return Container(
      width: double.infinity,
      padding: const EdgeInsets.all(16),
      decoration: BoxDecoration(
        color: background,
        borderRadius: BorderRadius.circular(18),
        border: Border.all(color: border),
      ),
      child: Row(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          Icon(icon, color: foreground),
          const SizedBox(width: 12),
          Expanded(
            child: Column(
              crossAxisAlignment: CrossAxisAlignment.start,
              children: [
                Text(
                  notice.title,
                  style: theme.textTheme.titleSmall?.copyWith(
                    color: foreground,
                    fontWeight: FontWeight.w800,
                  ),
                ),
                const SizedBox(height: 6),
                Text(
                  notice.message,
                  style: theme.textTheme.bodySmall?.copyWith(height: 1.5),
                ),
              ],
            ),
          ),
        ],
      ),
    );
  }
}
