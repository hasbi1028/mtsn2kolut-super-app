import 'package:flutter/material.dart';

import '../exam_error_messages.dart';
import '../exam_format.dart';
import '../exam_session_store.dart';

class ExamRestoreFailedScreen extends StatelessWidget {
  const ExamRestoreFailedScreen({
    super.key,
    required this.snapshot,
    required this.message,
    this.notice,
  });

  final ExamSessionSnapshot snapshot;
  final String message;
  final ExamGuidanceNotice? notice;

  @override
  Widget build(BuildContext context) {
    final theme = Theme.of(context);

    return Scaffold(
      body: SafeArea(
        child: Center(
          child: ConstrainedBox(
            constraints: const BoxConstraints(maxWidth: 760),
            child: Padding(
              padding: const EdgeInsets.all(24),
              child: Card(
                child: Padding(
                  padding: const EdgeInsets.all(28),
                  child: Column(
                    mainAxisSize: MainAxisSize.min,
                    crossAxisAlignment: CrossAxisAlignment.start,
                    children: [
                      Text(
                        'Sesi lama tidak bisa dipulihkan',
                        style: theme.textTheme.headlineMedium?.copyWith(
                          fontWeight: FontWeight.w800,
                          color: const Color(0xFF14361D),
                        ),
                      ),
                      const SizedBox(height: 12),
                      Text(
                        message,
                        style: theme.textTheme.bodyLarge?.copyWith(
                          height: 1.55,
                        ),
                      ),
                      if (notice != null) ...[
                        const SizedBox(height: 18),
                        _GuidancePanel(notice: notice!),
                      ],
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
                              snapshot.sessionTitle,
                              style: theme.textTheme.titleMedium?.copyWith(
                                fontWeight: FontWeight.w800,
                              ),
                            ),
                            const SizedBox(height: 10),
                            Text(
                              '${snapshot.studentName} • ${snapshot.studentNis}\nRuang ${snapshot.roomName}',
                              style: theme.textTheme.bodyMedium?.copyWith(
                                height: 1.5,
                              ),
                            ),
                            const SizedBox(height: 10),
                            Text(
                              formatExamSchedule(
                                scheduledStartIso: snapshot.scheduledStartIso,
                                scheduledEndIso: snapshot.scheduledEndIso,
                                durationMinutes: snapshot.durationMinutes,
                              ),
                              style: theme.textTheme.bodySmall?.copyWith(
                                color: theme.colorScheme.primary,
                                fontWeight: FontWeight.w700,
                              ),
                            ),
                            const SizedBox(height: 12),
                            Wrap(
                              spacing: 8,
                              runSpacing: 8,
                              children: [
                                _RestoreStatusChip(
                                  label: formatRestoreHealthLabel(
                                    lastServerContactIso:
                                        snapshot.lastServerContactIso,
                                    lastSyncFailureIso:
                                        snapshot.lastSyncFailureIso,
                                    consecutiveSyncFailures:
                                        snapshot.consecutiveSyncFailures,
                                  ),
                                  warning:
                                      snapshot.consecutiveSyncFailures >= 3,
                                ),
                                _RestoreStatusChip(
                                  label:
                                      '${snapshot.answers.length} jawaban lokal',
                                ),
                                _RestoreStatusChip(
                                  label:
                                      '${snapshot.pendingAnswers.length} belum tersinkron',
                                  warning: snapshot.pendingAnswers.isNotEmpty,
                                ),
                                if (snapshot.lastServerContactIso
                                    .trim()
                                    .isNotEmpty)
                                  _RestoreStatusChip(
                                    label:
                                        'Kontak server ${formatRestoreClock(snapshot.lastServerContactIso)}',
                                  ),
                                if (snapshot.lastSyncFailureIso
                                    .trim()
                                    .isNotEmpty)
                                  _RestoreStatusChip(
                                    label:
                                        'Gangguan ${formatRestoreClock(snapshot.lastSyncFailureIso)}',
                                    warning: true,
                                  ),
                              ],
                            ),
                          ],
                        ),
                      ),
                      const SizedBox(height: 24),
                      SizedBox(
                        width: double.infinity,
                        child: FilledButton(
                          onPressed: () => Navigator.of(context).pop(),
                          child: const Text('Kembali ke Login Token'),
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
    );
  }
}

class _GuidancePanel extends StatelessWidget {
  const _GuidancePanel({required this.notice});

  final ExamGuidanceNotice notice;

  @override
  Widget build(BuildContext context) {
    final theme = Theme.of(context);
    final (background, border, foreground, icon) = switch (notice.tone) {
      ExamGuidanceTone.info => (
        const Color(0xFFEAF4EB),
        const Color(0xFF7FB08A),
        const Color(0xFF1E5B2F),
        Icons.info_outline,
      ),
      ExamGuidanceTone.warning => (
        const Color(0xFFFFF3D8),
        const Color(0xFFF2C46D),
        const Color(0xFF9A6700),
        Icons.warning_amber_rounded,
      ),
      ExamGuidanceTone.danger => (
        const Color(0xFFFDE8E8),
        const Color(0xFFE8A4A4),
        const Color(0xFF9F2F2F),
        Icons.gpp_bad_outlined,
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
                  style: theme.textTheme.bodyMedium?.copyWith(
                    color: foreground,
                    height: 1.45,
                  ),
                ),
              ],
            ),
          ),
        ],
      ),
    );
  }
}

class _RestoreStatusChip extends StatelessWidget {
  const _RestoreStatusChip({required this.label, this.warning = false});

  final String label;
  final bool warning;

  @override
  Widget build(BuildContext context) {
    final theme = Theme.of(context);
    final background = warning
        ? const Color(0xFFFFF3D8)
        : theme.colorScheme.primary.withValues(alpha: 0.12);
    final foreground = warning
        ? const Color(0xFF9A6700)
        : theme.colorScheme.primary;

    return Container(
      padding: const EdgeInsets.symmetric(horizontal: 10, vertical: 7),
      decoration: BoxDecoration(
        color: background,
        borderRadius: BorderRadius.circular(999),
      ),
      child: Text(
        label,
        style: theme.textTheme.labelSmall?.copyWith(
          color: foreground,
          fontWeight: FontWeight.w700,
        ),
      ),
    );
  }
}
