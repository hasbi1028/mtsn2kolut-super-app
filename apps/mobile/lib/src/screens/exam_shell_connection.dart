class ExamShellConnectionViewModel {
  const ExamShellConnectionViewModel({
    required this.isSubmitted,
    required this.pendingAnswerCount,
    required this.consecutiveSyncFailures,
    required this.hasError,
    required this.lastServerContactAt,
    required this.lastSyncFailureAt,
    required this.isSavingAnswer,
    required this.isSyncingStatus,
    required this.isResumingExam,
    required this.resumeCheckRequired,
  });

  static const int degradedFailureThreshold = 3;
  static const int staleAttentionThresholdSeconds = 120;
  static const int staleEscalationThresholdSeconds = 240;
  static const int staleSyncLabelThresholdSeconds = 90;

  final bool isSubmitted;
  final int pendingAnswerCount;
  final int consecutiveSyncFailures;
  final bool hasError;
  final DateTime? lastServerContactAt;
  final DateTime? lastSyncFailureAt;
  final bool isSavingAnswer;
  final bool isSyncingStatus;
  final bool isResumingExam;
  final bool resumeCheckRequired;

  bool get isDegradedMode =>
      !isSubmitted && consecutiveSyncFailures >= degradedFailureThreshold;

  bool get lastContactIsStale {
    final last = lastServerContactAt;
    if (last == null) {
      return false;
    }
    return DateTime.now().difference(last).inSeconds >=
        staleSyncLabelThresholdSeconds;
  }

  bool get needsSupervisorAttention {
    final last = lastServerContactAt;
    if (last == null || isSubmitted || isDegradedMode) {
      return false;
    }
    return DateTime.now().difference(last).inSeconds >=
        staleAttentionThresholdSeconds;
  }

  bool get needsEscalatedSupervisorAttention {
    final last = lastServerContactAt;
    if (last == null || isSubmitted || isDegradedMode) {
      return false;
    }
    return DateTime.now().difference(last).inSeconds >=
        staleEscalationThresholdSeconds;
  }

  String get staleAttentionDurationLabel {
    final last = lastServerContactAt;
    if (last == null) {
      return '-';
    }
    final seconds = DateTime.now().difference(last).inSeconds;
    if (seconds < 60) {
      return '$seconds detik';
    }
    final minutes = seconds ~/ 60;
    final remainingSeconds = seconds % 60;
    if (minutes < 60) {
      return '$minutes menit ${remainingSeconds.toString().padLeft(2, '0')} detik';
    }
    final hours = minutes ~/ 60;
    final remainingMinutes = minutes % 60;
    return '$hours jam ${remainingMinutes.toString().padLeft(2, '0')} menit';
  }

  String get staleEscalationThresholdLabel {
    final minutes = staleEscalationThresholdSeconds ~/ 60;
    if (minutes < 60) {
      return '$minutes menit';
    }
    final hours = minutes ~/ 60;
    final remainingMinutes = minutes % 60;
    return '$hours jam ${remainingMinutes.toString().padLeft(2, '0')} menit';
  }

  String get syncStatusLabel {
    if (isResumingExam || resumeCheckRequired) {
      return 'Cek Ulang';
    }
    if (isDegradedMode) {
      return 'Menurun';
    }
    if (lastContactIsStale) {
      return 'Waspada';
    }
    if (isSavingAnswer || isSyncingStatus) {
      return 'Sinkron';
    }
    if (pendingAnswerCount > 0) {
      return 'Lokal';
    }
    if (hasError) {
      return 'Gangguan';
    }
    return 'Tersambung';
  }

  SyncTone get syncStatusTone {
    if (isDegradedMode || hasError) {
      return SyncTone.danger;
    }
    if (isResumingExam ||
        resumeCheckRequired ||
        pendingAnswerCount > 0 ||
        lastContactIsStale) {
      return SyncTone.warning;
    }
    return SyncTone.success;
  }

  String get connectionHealthTitle {
    if (isDegradedMode) {
      return 'Menurun';
    }
    if (lastContactIsStale) {
      return 'Waspada';
    }
    if (hasError) {
      return 'Gangguan';
    }
    if (pendingAnswerCount > 0) {
      return 'Lokal';
    }
    return 'Tersambung';
  }

  String get connectionHealthDescription {
    if (isDegradedMode) {
      return 'Sinkron berulang kali gagal. Submit manual ditahan sampai koneksi membaik.';
    }
    if (lastContactIsStale) {
      return 'Perangkat sudah cukup lama tidak menyentuh server. Perbarui status agar pengawas tahu koneksi masih sehat.';
    }
    if (hasError) {
      return 'Server belum merespons stabil. Pantau jaringan dan coba sinkron ulang.';
    }
    if (pendingAnswerCount > 0) {
      return '$pendingAnswerCount jawaban masih aman di perangkat dan menunggu sinkron.';
    }
    return 'Perangkat terakhir berhasil terhubung ke server tanpa jawaban lokal tertahan.';
  }

  HealthTone get connectionHealthTone {
    if (isDegradedMode || hasError) {
      return HealthTone.danger;
    }
    if (lastContactIsStale || pendingAnswerCount > 0) {
      return HealthTone.warning;
    }
    return HealthTone.good;
  }

  String formatClock(DateTime? value) {
    if (value == null) {
      return '-';
    }
    final local = value.toLocal();
    final hour = local.hour.toString().padLeft(2, '0');
    final minute = local.minute.toString().padLeft(2, '0');
    final second = local.second.toString().padLeft(2, '0');
    return '$hour:$minute:$second';
  }

  static String formatDuration(int seconds) {
    final duration = Duration(seconds: seconds.clamp(0, 999999));
    final hours = duration.inHours.toString().padLeft(2, '0');
    final minutes = (duration.inMinutes % 60).toString().padLeft(2, '0');
    final secs = (duration.inSeconds % 60).toString().padLeft(2, '0');
    return '$hours:$minutes:$secs';
  }
}

enum SyncTone { success, warning, danger }

enum HealthTone { good, warning, danger }

enum BannerTone { error, success, info }
