import 'package:flutter_test/flutter_test.dart';
import 'package:mobile/src/models.dart';

void main() {
  group('ExamLoginPayload.fromJson', () {
    test('parses rich content and media fields', () {
      final payload = ExamLoginPayload.fromJson({
        'participant_id': 'participant-1',
        'student': {'nis': '24001', 'nama': 'Siti Aminah'},
        'session': {
          'id': 'session-1',
          'title': 'Matematika Kelas VIII',
          'scheduled_start': '2026-05-01T08:00:00+08:00',
          'scheduled_end': '2026-05-01T09:30:00+08:00',
          'duration_minutes': 90,
        },
        'room': {'room_name': 'Lab 1'},
        'questions': [
          {
            'id': 'question-1',
            'question_text': '',
            'stem_html': '<p>Soal utama</p>',
            'stimulus_html': '<p>Stimulus</p>',
            'stem_media_url': 'https://cdn.example.com/stem.png',
            'stimulus_media_url': 'https://cdn.example.com/stimulus.png',
            'stem_audio_url': 'https://cdn.example.com/stem.mp3',
            'stimulus_audio_url': 'https://cdn.example.com/stimulus.mp3',
            'options': [
              {'label': 'A', 'text': 'Pilihan A'},
            ],
          },
        ],
        'answered_count': 1,
        'total_questions': 10,
        'time_remaining_seconds': 1200,
      });

      expect(payload.participantId, 'participant-1');
      expect(payload.student.nis, '24001');
      expect(payload.session.title, 'Matematika Kelas VIII');
      expect(payload.room?.roomName, 'Lab 1');
      expect(payload.questions, hasLength(1));
      expect(payload.questions.first.stemHtml, '<p>Soal utama</p>');
      expect(payload.questions.first.stimulusHtml, '<p>Stimulus</p>');
      expect(
        payload.questions.first.stemMediaUrl,
        'https://cdn.example.com/stem.png',
      );
      expect(
        payload.questions.first.stimulusMediaUrl,
        'https://cdn.example.com/stimulus.png',
      );
      expect(
        payload.questions.first.stemAudioUrl,
        'https://cdn.example.com/stem.mp3',
      );
      expect(
        payload.questions.first.stimulusAudioUrl,
        'https://cdn.example.com/stimulus.mp3',
      );
    });

    test('uses safe defaults for missing nested payload fields', () {
      final payload = ExamLoginPayload.fromJson(const {});

      expect(payload.participantId, '');
      expect(payload.student.nis, '-');
      expect(payload.student.nama, 'Siswa');
      expect(payload.session.id, '');
      expect(payload.session.title, 'Sesi Ujian');
      expect(payload.room, isNull);
      expect(payload.questions, isEmpty);
      expect(payload.answeredCount, 0);
      expect(payload.totalQuestions, 0);
      expect(payload.timeRemainingSeconds, 0);
    });
  });

  group('ExamQuestion', () {
    test('detects essay and rich content states correctly', () {
      final richEssay = ExamQuestion.fromJson({
        'id': 'essay-1',
        'question_text': '',
        'stem_html': '<p>Uraikan jawaban Anda.</p>',
        'stimulus_html': '',
        'options': const [],
      });

      final plainMcq = ExamQuestion.fromJson({
        'id': 'mcq-1',
        'question_text': '2 + 2 = ...',
        'options': [
          {'label': 'A', 'text': '4'},
        ],
      });

      expect(richEssay.isEssay, isTrue);
      expect(richEssay.hasRichContent, isTrue);
      expect(plainMcq.isEssay, isFalse);
      expect(plainMcq.hasRichContent, isFalse);
    });
  });

  group('ExamStatusPayload.fromJson', () {
    test('parses submission status from backend payload', () {
      final payload = ExamStatusPayload.fromJson({
        'answered_count': 8,
        'total_questions': 20,
        'time_remaining_seconds': 900,
        'is_submitted': true,
      });

      expect(payload.answeredCount, 8);
      expect(payload.totalQuestions, 20);
      expect(payload.timeRemainingSeconds, 900);
      expect(payload.isSubmitted, isTrue);
    });

    test('defaults submission status to false when omitted', () {
      final payload = ExamStatusPayload.fromJson(const {});

      expect(payload.answeredCount, 0);
      expect(payload.totalQuestions, 0);
      expect(payload.timeRemainingSeconds, 0);
      expect(payload.isSubmitted, isFalse);
    });
  });
}
