import 'dart:convert';

import 'package:shared_preferences/shared_preferences.dart';

class ExamSessionSnapshot {
  const ExamSessionSnapshot({
    required this.baseUrl,
    required this.examToken,
    required this.deviceFingerprint,
    required this.studentName,
    required this.studentNis,
    required this.sessionTitle,
    required this.roomName,
    required this.scheduledStartIso,
    required this.scheduledEndIso,
    required this.durationMinutes,
    required this.currentQuestionIndex,
    required this.answers,
    required this.pendingAnswers,
    required this.playedAudioQuestionIds,
    required this.lastServerContactIso,
    required this.lastSyncFailureIso,
    required this.consecutiveSyncFailures,
  });

  final String baseUrl;
  final String examToken;
  final String deviceFingerprint;
  final String studentName;
  final String studentNis;
  final String sessionTitle;
  final String roomName;
  final String scheduledStartIso;
  final String scheduledEndIso;
  final int durationMinutes;
  final int currentQuestionIndex;
  final Map<String, String> answers;
  final Map<String, String> pendingAnswers;
  final List<String> playedAudioQuestionIds;
  final String lastServerContactIso;
  final String lastSyncFailureIso;
  final int consecutiveSyncFailures;

  Map<String, Object?> toJson() {
    return <String, Object?>{
      'base_url': baseUrl,
      'exam_token': examToken,
      'device_fingerprint': deviceFingerprint,
      'student_name': studentName,
      'student_nis': studentNis,
      'session_title': sessionTitle,
      'room_name': roomName,
      'scheduled_start_iso': scheduledStartIso,
      'scheduled_end_iso': scheduledEndIso,
      'duration_minutes': durationMinutes,
      'current_question_index': currentQuestionIndex,
      'answers': answers,
      'pending_answers': pendingAnswers,
      'played_audio_question_ids': playedAudioQuestionIds,
      'last_server_contact_iso': lastServerContactIso,
      'last_sync_failure_iso': lastSyncFailureIso,
      'consecutive_sync_failures': consecutiveSyncFailures,
    };
  }

  factory ExamSessionSnapshot.fromJson(Map<String, dynamic> json) {
    return ExamSessionSnapshot(
      baseUrl: json['base_url'] as String? ?? '',
      examToken: json['exam_token'] as String? ?? '',
      deviceFingerprint: json['device_fingerprint'] as String? ?? '',
      studentName: json['student_name'] as String? ?? 'Siswa',
      studentNis: json['student_nis'] as String? ?? '-',
      sessionTitle: json['session_title'] as String? ?? 'Sesi Ujian',
      roomName: json['room_name'] as String? ?? '-',
      scheduledStartIso: json['scheduled_start_iso'] as String? ?? '',
      scheduledEndIso: json['scheduled_end_iso'] as String? ?? '',
      durationMinutes: json['duration_minutes'] as int? ?? 0,
      currentQuestionIndex: json['current_question_index'] as int? ?? 0,
      answers: ((json['answers'] as Map<String, dynamic>?) ?? const {}).map(
        (key, value) => MapEntry(key, value.toString()),
      ),
      pendingAnswers:
          ((json['pending_answers'] as Map<String, dynamic>?) ?? const {}).map(
            (key, value) => MapEntry(key, value.toString()),
          ),
      playedAudioQuestionIds:
          ((json['played_audio_question_ids'] as List<dynamic>?) ?? const [])
              .map((value) => value.toString())
              .toList(),
      lastServerContactIso: json['last_server_contact_iso'] as String? ?? '',
      lastSyncFailureIso: json['last_sync_failure_iso'] as String? ?? '',
      consecutiveSyncFailures: json['consecutive_sync_failures'] as int? ?? 0,
    );
  }
}

class ExamSessionStore {
  static const _baseUrlKey = 'exam_last_base_url';
  static const _snapshotKey = 'exam_active_snapshot';

  Future<void> saveBaseUrl(String baseUrl) async {
    final prefs = await SharedPreferences.getInstance();
    await prefs.setString(_baseUrlKey, baseUrl);
  }

  Future<String?> loadBaseUrl() async {
    final prefs = await SharedPreferences.getInstance();
    return prefs.getString(_baseUrlKey);
  }

  Future<void> saveSnapshot(ExamSessionSnapshot snapshot) async {
    final prefs = await SharedPreferences.getInstance();
    await prefs.setString(_snapshotKey, jsonEncode(snapshot.toJson()));
    await prefs.setString(_baseUrlKey, snapshot.baseUrl);
  }

  Future<ExamSessionSnapshot?> loadSnapshot() async {
    final prefs = await SharedPreferences.getInstance();
    final raw = prefs.getString(_snapshotKey);
    if (raw == null || raw.trim().isEmpty) {
      return null;
    }
    final decoded = jsonDecode(raw);
    if (decoded is! Map<String, dynamic>) {
      return null;
    }
    return ExamSessionSnapshot.fromJson(decoded);
  }

  Future<void> clearSnapshot() async {
    final prefs = await SharedPreferences.getInstance();
    await prefs.remove(_snapshotKey);
  }
}
