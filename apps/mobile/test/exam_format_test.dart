import 'package:flutter_test/flutter_test.dart';
import 'package:mobile/src/exam_format.dart';

void main() {
  group('formatExamSchedule', () {
    test('renders full schedule when start end and duration exist', () {
      final result = formatExamSchedule(
        scheduledStartIso: '2026-05-01T08:00:00+08:00',
        scheduledEndIso: '2026-05-01T09:30:00+08:00',
        durationMinutes: 90,
      );

      expect(result, '01/05/2026 08:00 • s.d. 09:30 • 90 menit');
    });

    test('renders partial schedule when only duration exists', () {
      final result = formatExamSchedule(
        scheduledStartIso: '',
        scheduledEndIso: '',
        durationMinutes: 60,
      );

      expect(result, '60 menit');
    });

    test('falls back when no schedule metadata exists', () {
      final result = formatExamSchedule(
        scheduledStartIso: '',
        scheduledEndIso: '',
        durationMinutes: 0,
      );

      expect(result, 'Jadwal belum tersedia');
    });
  });

  group('formatRestoreHealthLabel', () {
    test('returns warning label for repeated sync failures', () {
      final result = formatRestoreHealthLabel(
        lastServerContactIso: '2026-05-01T08:44:00+08:00',
        lastSyncFailureIso: '2026-05-01T08:46:00+08:00',
        consecutiveSyncFailures: 3,
      );

      expect(result, 'Perlu perhatian koneksi');
    });

    test('returns stable label when last contact exists without failures', () {
      final result = formatRestoreHealthLabel(
        lastServerContactIso: '2026-05-01T08:44:00+08:00',
        lastSyncFailureIso: '',
        consecutiveSyncFailures: 0,
      );

      expect(result, 'Terakhir stabil');
    });

    test('returns disturbed label when a failure exists', () {
      final result = formatRestoreHealthLabel(
        lastServerContactIso: '',
        lastSyncFailureIso: '2026-05-01T08:46:00+08:00',
        consecutiveSyncFailures: 0,
      );

      expect(result, 'Pernah terganggu');
    });

    test('returns empty-history label when no health metadata exists', () {
      final result = formatRestoreHealthLabel(
        lastServerContactIso: '',
        lastSyncFailureIso: '',
        consecutiveSyncFailures: 0,
      );

      expect(result, 'Belum ada riwayat koneksi');
    });
  });

  group('formatRestoreClock', () {
    test('formats valid iso clock', () {
      final result = formatRestoreClock('2026-05-01T08:44:00+08:00');

      expect(result, '08:44');
    });

    test('returns dash for empty value', () {
      final result = formatRestoreClock('');

      expect(result, '-');
    });

    test('returns dash for invalid value', () {
      final result = formatRestoreClock('bukan-tanggal');

      expect(result, '-');
    });
  });
}
