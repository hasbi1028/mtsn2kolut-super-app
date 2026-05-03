import 'package:flutter_test/flutter_test.dart';
import 'package:mobile/src/exam_session_store.dart';

void main() {
  group('ExamSessionStore', () {
    late ExamSessionStore store;
    late MemorySnapshotValueStore metadataStore;
    late MemorySnapshotValueStore secureStore;

    setUp(() {
      metadataStore = MemorySnapshotValueStore();
      secureStore = MemorySnapshotValueStore();
      store = ExamSessionStore(
        metadataStore: metadataStore,
        secureStore: secureStore,
      );
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
      expect(metadataStore.values['exam_active_snapshot_secure'], isNull);
      expect(secureStore.values['exam_active_snapshot_secure'], isNotNull);
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
      metadataStore.values['exam_active_snapshot'] = '"bukan-object-json"';

      final restored = await store.loadSnapshot();

      expect(restored, isNull);
    });

    test('returns null for malformed snapshot json', () async {
      metadataStore.values['exam_active_snapshot'] = '{bukan-json-valid';

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

    test('returns null when secure snapshot payload is missing', () async {
      metadataStore.values['exam_active_snapshot'] =
          '{"base_url":"http://10.0.2.2:8080","student_name":"Siti"}';

      final restored = await store.loadSnapshot();

      expect(restored, isNull);
    });

    test('loads legacy plaintext snapshot for migration compatibility', () async {
      metadataStore.values['exam_active_snapshot'] =
          '{"base_url":"http://10.0.2.2:8080","exam_token":"token-1","device_fingerprint":"android:test","student_name":"Siti","student_nis":"24001","session_title":"Matematika","room_name":"Lab 1","scheduled_start_iso":"","scheduled_end_iso":"","duration_minutes":90,"current_question_index":1,"answers":{"q1":"A"},"pending_answers":{"q2":"B"},"played_audio_question_ids":["q3"],"last_server_contact_iso":"","last_sync_failure_iso":"","consecutive_sync_failures":0}';

      final restored = await store.loadSnapshot();

      expect(restored, isNotNull);
      expect(restored?.examToken, 'token-1');
      expect(restored?.answers, {'q1': 'A'});
      expect(restored?.pendingAnswers, {'q2': 'B'});
      expect(
        metadataStore.values['exam_active_snapshot'],
        isNot(contains('exam_token')),
      );
      expect(
        secureStore.values['exam_active_snapshot_secure'],
        contains('token-1'),
      );
    });
  });
}

class MemorySnapshotValueStore implements SnapshotValueStore {
  final Map<String, String> values = <String, String>{};

  @override
  Future<void> delete(String key) async {
    values.remove(key);
  }

  @override
  Future<String?> read(String key) async => values[key];

  @override
  Future<void> write(String key, String value) async {
    values[key] = value;
  }
}
