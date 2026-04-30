String formatExamSchedule({
  required String scheduledStartIso,
  required String scheduledEndIso,
  required int durationMinutes,
}) {
  final start = _tryParse(scheduledStartIso);
  final end = _tryParse(scheduledEndIso);

  final parts = <String>[];
  if (start != null) {
    parts.add(_formatDateTime(start));
  }
  if (end != null) {
    parts.add('s.d. ${_formatTime(end)}');
  }
  if (durationMinutes > 0) {
    parts.add('$durationMinutes menit');
  }

  return parts.isEmpty ? 'Jadwal belum tersedia' : parts.join(' • ');
}

String formatRestoreHealthLabel({
  required String lastServerContactIso,
  required String lastSyncFailureIso,
  required int consecutiveSyncFailures,
}) {
  if (consecutiveSyncFailures >= 3) {
    return 'Perlu perhatian koneksi';
  }
  if (lastServerContactIso.trim().isNotEmpty &&
      lastSyncFailureIso.trim().isEmpty) {
    return 'Terakhir stabil';
  }
  if (lastSyncFailureIso.trim().isNotEmpty) {
    return 'Pernah terganggu';
  }
  return 'Belum ada riwayat koneksi';
}

String formatRestoreClock(String rawIso) {
  final parsed = _tryParse(rawIso);
  if (parsed == null) {
    return '-';
  }
  return _formatTime(parsed);
}

DateTime? _tryParse(String raw) {
  if (raw.trim().isEmpty) {
    return null;
  }
  return DateTime.tryParse(raw)?.toLocal();
}

String _formatDateTime(DateTime value) {
  final date =
      '${value.day.toString().padLeft(2, '0')}/${value.month.toString().padLeft(2, '0')}/${value.year}';
  return '$date ${_formatTime(value)}';
}

String _formatTime(DateTime value) {
  final hour = value.hour.toString().padLeft(2, '0');
  final minute = value.minute.toString().padLeft(2, '0');
  return '$hour:$minute';
}
