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
    required this.questionText,
    required this.stemHtml,
    required this.stimulusHtml,
    required this.stemMediaUrl,
    required this.stimulusMediaUrl,
    required this.options,
  });

  final String id;
  final String questionText;
  final String stemHtml;
  final String stimulusHtml;
  final String stemMediaUrl;
  final String stimulusMediaUrl;
  final List<ExamOption> options;

  bool get isEssay => options.isEmpty;
  bool get hasRichContent =>
      stemHtml.trim().isNotEmpty || stimulusHtml.trim().isNotEmpty;

  factory ExamQuestion.fromJson(Map<String, dynamic> json) {
    return ExamQuestion(
      id: json['id'] as String? ?? '',
      questionText: json['question_text'] as String? ?? '',
      stemHtml: json['stem_html'] as String? ?? '',
      stimulusHtml: json['stimulus_html'] as String? ?? '',
      stemMediaUrl: json['stem_media_url'] as String? ?? '',
      stimulusMediaUrl: json['stimulus_media_url'] as String? ?? '',
      options: ((json['options'] as List<dynamic>?) ?? const [])
          .whereType<Map<String, dynamic>>()
          .map(ExamOption.fromJson)
          .toList(),
    );
  }
}

class ExamOption {
  const ExamOption({required this.label, required this.text});

  final String label;
  final String text;

  factory ExamOption.fromJson(Map<String, dynamic> json) {
    return ExamOption(
      label: json['label'] as String? ?? '',
      text: json['text'] as String? ?? '',
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
