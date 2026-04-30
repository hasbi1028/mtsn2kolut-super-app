import 'package:flutter_test/flutter_test.dart';
import 'package:mobile/src/exam_session_store.dart';

void main() {
  group('ExamSessionSnapshot', () {
    test('roundtrips through toJson and fromJson', () {
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
        currentQuestionIndex: 4,
        answers: {'question-1': 'B'},
        pendingAnswers: {'essay-1': 'Jawaban lokal'},
        playedAudioQuestionIds: ['audio-1'],
        lastServerContactIso: '2026-05-01T08:44:00+08:00',
        lastSyncFailureIso: '2026-05-01T08:46:00+08:00',
        consecutiveSyncFailures: 2,
      );

      final restored = ExamSessionSnapshot.fromJson(snapshot.toJson());

      expect(restored.baseUrl, snapshot.baseUrl);
      expect(restored.examToken, snapshot.examToken);
      expect(restored.deviceFingerprint, snapshot.deviceFingerprint);
      expect(restored.studentName, snapshot.studentName);
      expect(restored.studentNis, snapshot.studentNis);
      expect(restored.sessionTitle, snapshot.sessionTitle);
      expect(restored.roomName, snapshot.roomName);
      expect(restored.scheduledStartIso, snapshot.scheduledStartIso);
      expect(restored.scheduledEndIso, snapshot.scheduledEndIso);
      expect(restored.durationMinutes, snapshot.durationMinutes);
      expect(restored.currentQuestionIndex, snapshot.currentQuestionIndex);
      expect(restored.answers, snapshot.answers);
      expect(restored.pendingAnswers, snapshot.pendingAnswers);
      expect(restored.playedAudioQuestionIds, snapshot.playedAudioQuestionIds);
      expect(restored.lastServerContactIso, snapshot.lastServerContactIso);
      expect(restored.lastSyncFailureIso, snapshot.lastSyncFailureIso);
      expect(
        restored.consecutiveSyncFailures,
        snapshot.consecutiveSyncFailures,
      );
    });

    test('coerces dynamic answer and audio values into strings', () {
      final restored = ExamSessionSnapshot.fromJson({
        'answers': {'question-1': 2},
        'pending_answers': {'essay-1': true},
        'played_audio_question_ids': [1, 'audio-2', false],
      });

      expect(restored.answers, {'question-1': '2'});
      expect(restored.pendingAnswers, {'essay-1': 'true'});
      expect(restored.playedAudioQuestionIds, ['1', 'audio-2', 'false']);
    });

    test('uses safe defaults for missing fields', () {
      final restored = ExamSessionSnapshot.fromJson(const {});

      expect(restored.baseUrl, '');
      expect(restored.examToken, '');
      expect(restored.deviceFingerprint, '');
      expect(restored.studentName, 'Siswa');
      expect(restored.studentNis, '-');
      expect(restored.sessionTitle, 'Sesi Ujian');
      expect(restored.roomName, '-');
      expect(restored.scheduledStartIso, '');
      expect(restored.scheduledEndIso, '');
      expect(restored.durationMinutes, 0);
      expect(restored.currentQuestionIndex, 0);
      expect(restored.answers, isEmpty);
      expect(restored.pendingAnswers, isEmpty);
      expect(restored.playedAudioQuestionIds, isEmpty);
      expect(restored.lastServerContactIso, '');
      expect(restored.lastSyncFailureIso, '');
      expect(restored.consecutiveSyncFailures, 0);
    });

    test('splits metadata and sensitive payloads predictably', () {
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
        currentQuestionIndex: 4,
        answers: {'question-1': 'B'},
        pendingAnswers: {'essay-1': 'Jawaban lokal'},
        playedAudioQuestionIds: ['audio-1'],
        lastServerContactIso: '2026-05-01T08:44:00+08:00',
        lastSyncFailureIso: '2026-05-01T08:46:00+08:00',
        consecutiveSyncFailures: 2,
      );

      final metadata = snapshot.toMetadataJson();
      final sensitive = snapshot.toSensitiveJson();

      expect(metadata.containsKey('exam_token'), isFalse);
      expect(metadata.containsKey('answers'), isFalse);
      expect(metadata.containsKey('pending_answers'), isFalse);
      expect(sensitive['exam_token'], 'token-1');
      expect(sensitive['answers'], {'question-1': 'B'});
      expect(sensitive['pending_answers'], {'essay-1': 'Jawaban lokal'});
    });
  });
}
