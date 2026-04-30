import 'package:flutter_test/flutter_test.dart';
import 'package:mobile/src/exam_session_store.dart';
import 'package:shared_preferences/shared_preferences.dart';

void main() {
  group('ExamSessionStore', () {
    late ExamSessionStore store;

    setUp(() {
      SharedPreferences.setMockInitialValues(const {});
      store = ExamSessionStore();
    });

    test('saves and loads base url', () async {
      await store.saveBaseUrl('http://10.0.2.2:8080');

      final baseUrl = await store.loadBaseUrl();

      expect(baseUrl, 'http://10.0.2.2:8080');
    });

    test('saves and loads snapshot with byod metadata intact', () async {
      const snapshot = ExamSessionSnapshot(
        baseUrl: 'http://10.0.2.2:8080',
        examToken: 'token-1',
        deviceFingerprint: 'android:test',
        studentName: 'Siti Aminah',
        studentNis: '24001',
        sessionTitle: 'Matematika Kelas VIII',
        roomName: 'Lab 1',
        scheduledStartIso: '2026-05-01T08:00:00+08:00',
        scheduledEndIso: '2026-05-01T09:30:00+08:00',
        durationMinutes: 90,
        currentQuestionIndex: 3,
        answers: {'question-1': 'B'},
        pendingAnswers: {'essay-2': 'Jawaban lokal'},
        playedAudioQuestionIds: ['audio-1'],
        lastServerContactIso: '2026-05-01T08:44:00+08:00',
        lastSyncFailureIso: '2026-05-01T08:46:00+08:00',
        consecutiveSyncFailures: 2,
      );

      await store.saveSnapshot(snapshot);

      final restored = await store.loadSnapshot();
      final rememberedBaseUrl = await store.loadBaseUrl();

      expect(restored, isNotNull);
      expect(restored?.baseUrl, snapshot.baseUrl);
      expect(restored?.examToken, snapshot.examToken);
      expect(restored?.studentName, snapshot.studentName);
      expect(restored?.sessionTitle, snapshot.sessionTitle);
      expect(restored?.currentQuestionIndex, snapshot.currentQuestionIndex);
      expect(restored?.answers, snapshot.answers);
      expect(restored?.pendingAnswers, snapshot.pendingAnswers);
      expect(restored?.playedAudioQuestionIds, snapshot.playedAudioQuestionIds);
      expect(restored?.lastServerContactIso, snapshot.lastServerContactIso);
      expect(restored?.lastSyncFailureIso, snapshot.lastSyncFailureIso);
      expect(
        restored?.consecutiveSyncFailures,
        snapshot.consecutiveSyncFailures,
      );
      expect(rememberedBaseUrl, snapshot.baseUrl);
    });

    test('returns null for missing snapshot', () async {
      final restored = await store.loadSnapshot();

      expect(restored, isNull);
    });

    test('returns null for invalid snapshot payload', () async {
      SharedPreferences.setMockInitialValues({
        'exam_active_snapshot': '"bukan-object-json"',
      });
      store = ExamSessionStore();

      final restored = await store.loadSnapshot();

      expect(restored, isNull);
    });

    test('clears snapshot without removing remembered base url', () async {
      const snapshot = ExamSessionSnapshot(
        baseUrl: 'http://10.0.2.2:8080',
        examToken: 'token-1',
        deviceFingerprint: 'android:test',
        studentName: 'Siti Aminah',
        studentNis: '24001',
        sessionTitle: 'Matematika Kelas VIII',
        roomName: 'Lab 1',
        scheduledStartIso: '2026-05-01T08:00:00+08:00',
        scheduledEndIso: '2026-05-01T09:30:00+08:00',
        durationMinutes: 90,
        currentQuestionIndex: 0,
        answers: {},
        pendingAnswers: {},
        playedAudioQuestionIds: [],
        lastServerContactIso: '',
        lastSyncFailureIso: '',
        consecutiveSyncFailures: 0,
      );

      await store.saveSnapshot(snapshot);
      await store.clearSnapshot();

      final restored = await store.loadSnapshot();
      final rememberedBaseUrl = await store.loadBaseUrl();

      expect(restored, isNull);
      expect(rememberedBaseUrl, snapshot.baseUrl);
    });
  });
}
