class ExamLoginPayload {
  const ExamLoginPayload({
    required this.participantId,
    required this.student,
    required this.session,
    required this.room,
    required this.questions,
    required this.answeredCount,
    required this.totalQuestions,
    required this.timeRemainingSeconds,
  });

  final String participantId;
  final ExamStudent student;
  final ExamSession session;
  final ExamRoom? room;
  final List<ExamQuestion> questions;
  final int answeredCount;
  final int totalQuestions;
  final int timeRemainingSeconds;

  factory ExamLoginPayload.fromJson(Map<String, dynamic> json) {
    return ExamLoginPayload(
      participantId: json['participant_id'] as String? ?? '',
      student: ExamStudent.fromJson(
        json['student'] as Map<String, dynamic>? ?? const {},
      ),
      session: ExamSession.fromJson(
        json['session'] as Map<String, dynamic>? ?? const {},
      ),
      room: json['room'] is Map<String, dynamic>
          ? ExamRoom.fromJson(json['room'] as Map<String, dynamic>)
          : null,
      questions: ((json['questions'] as List<dynamic>?) ?? const [])
          .whereType<Map<String, dynamic>>()
          .map(ExamQuestion.fromJson)
          .toList(),
      answeredCount: json['answered_count'] as int? ?? 0,
      totalQuestions: json['total_questions'] as int? ?? 0,
      timeRemainingSeconds: json['time_remaining_seconds'] as int? ?? 0,
    );
  }
}

class ExamStudent {
  const ExamStudent({required this.nis, required this.nama});

  final String nis;
  final String nama;

  factory ExamStudent.fromJson(Map<String, dynamic> json) {
    return ExamStudent(
      nis: json['nis'] as String? ?? '-',
      nama: json['nama'] as String? ?? 'Siswa',
    );
  }
}

class ExamSession {
  const ExamSession({
    required this.id,
    required this.title,
    required this.scheduledStart,
    required this.scheduledEnd,
    required this.durationMinutes,
  });

  final String id;
  final String title;
  final DateTime? scheduledStart;
  final DateTime? scheduledEnd;
  final int durationMinutes;

  factory ExamSession.fromJson(Map<String, dynamic> json) {
    return ExamSession(
      id: json['id'] as String? ?? '',
      title: json['title'] as String? ?? 'Sesi Ujian',
      scheduledStart: _tryParseDateTime(json['scheduled_start']),
      scheduledEnd: _tryParseDateTime(json['scheduled_end']),
      durationMinutes: json['duration_minutes'] as int? ?? 0,
    );
  }
}

class ExamRoom {
  const ExamRoom({required this.roomName});

  final String roomName;

  factory ExamRoom.fromJson(Map<String, dynamic> json) {
    return ExamRoom(roomName: json['room_name'] as String? ?? '-');
  }
}

class ExamQuestion {
  const ExamQuestion({
    required this.id,
    this.questionType = 'multiple_choice',
    required this.questionText,
    required this.stemHtml,
    required this.stimulusHtml,
    required this.stemMediaUrl,
    required this.stimulusMediaUrl,
    required this.stemAudioUrl,
    required this.stimulusAudioUrl,
    required this.options,
  });

  final String id;
  final String questionType;
  final String questionText;
  final String stemHtml;
  final String stimulusHtml;
  final String stemMediaUrl;
  final String stimulusMediaUrl;
  final String stemAudioUrl;
  final String stimulusAudioUrl;
  final List<ExamOption> options;

  bool get isEssay =>
      questionType == 'essay' || (questionType.isEmpty && options.isEmpty);
  bool get isShortAnswer => questionType == 'short_answer';
  bool get isMultipleAnswer => questionType == 'multiple_answer';
  bool get isMatching => questionType == 'matching';
  bool get isTextAnswer => isEssay || isShortAnswer;
  bool get hasRichContent =>
      stemHtml.trim().isNotEmpty || stimulusHtml.trim().isNotEmpty;

  factory ExamQuestion.fromJson(Map<String, dynamic> json) {
    final options = ((json['options'] as List<dynamic>?) ?? const [])
        .whereType<Map<String, dynamic>>()
        .map(ExamOption.fromJson)
        .toList();
    final type =
        json['question_type'] as String? ??
        (options.isEmpty ? 'essay' : 'multiple_choice');
    return ExamQuestion(
      id: json['id'] as String? ?? '',
      questionType: type,
      questionText: json['question_text'] as String? ?? '',
      stemHtml: json['stem_html'] as String? ?? '',
      stimulusHtml: json['stimulus_html'] as String? ?? '',
      stemMediaUrl: json['stem_media_url'] as String? ?? '',
      stimulusMediaUrl: json['stimulus_media_url'] as String? ?? '',
      stemAudioUrl: json['stem_audio_url'] as String? ?? '',
      stimulusAudioUrl: json['stimulus_audio_url'] as String? ?? '',
      options: options,
    );
  }
}

class ExamOption {
  const ExamOption({
    required this.label,
    required this.text,
    this.matchLabel = '',
    this.matchText = '',
    this.isDistractor = false,
  });

  final String label;
  final String text;
  final String matchLabel;
  final String matchText;
  final bool isDistractor;

  factory ExamOption.fromJson(Map<String, dynamic> json) {
    return ExamOption(
      label: json['label'] as String? ?? '',
      text: json['text'] as String? ?? '',
      matchLabel: json['match_label'] as String? ?? '',
      matchText: json['match_text'] as String? ?? '',
      isDistractor: json['is_distractor'] as bool? ?? false,
    );
  }
}

class ExamStatusPayload {
  const ExamStatusPayload({
    required this.answeredCount,
    required this.totalQuestions,
    required this.timeRemainingSeconds,
    required this.isSubmitted,
  });

  final int answeredCount;
  final int totalQuestions;
  final int timeRemainingSeconds;
  final bool isSubmitted;

  factory ExamStatusPayload.fromJson(Map<String, dynamic> json) {
    return ExamStatusPayload(
      answeredCount: json['answered_count'] as int? ?? 0,
      totalQuestions: json['total_questions'] as int? ?? 0,
      timeRemainingSeconds: json['time_remaining_seconds'] as int? ?? 0,
      isSubmitted: json['is_submitted'] as bool? ?? false,
    );
  }
}

DateTime? _tryParseDateTime(Object? raw) {
  if (raw is! String || raw.trim().isEmpty) {
    return null;
  }
  return DateTime.tryParse(raw);
}
