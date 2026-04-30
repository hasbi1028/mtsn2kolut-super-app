import 'package:flutter/material.dart';

import '../exam_format.dart';
import '../exam_session_store.dart';

class ExamRestoreFailedScreen extends StatelessWidget {
  const ExamRestoreFailedScreen({
    super.key,
    required this.snapshot,
    required this.message,
  });

  final ExamSessionSnapshot snapshot;
  final String message;

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
