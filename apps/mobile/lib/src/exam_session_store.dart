import 'dart:convert';

import 'package:shared_preferences/shared_preferences.dart';

class ExamSessionSnapshot {
  const ExamSessionSnapshot({
    required this.baseUrl,
    required this.examToken,
    required this.deviceFingerprint,
    required this.currentQuestionIndex,
    required this.answers,
  });

  final String baseUrl;
  final String examToken;
  final String deviceFingerprint;
  final int currentQuestionIndex;
  final Map<String, String> answers;

  Map<String, Object?> toJson() {
    return <String, Object?>{
      'base_url': baseUrl,
      'exam_token': examToken,
      'device_fingerprint': deviceFingerprint,
      'current_question_index': currentQuestionIndex,
      'answers': answers,
    };
  }

  factory ExamSessionSnapshot.fromJson(Map<String, dynamic> json) {
    return ExamSessionSnapshot(
      baseUrl: json['base_url'] as String? ?? '',
      examToken: json['exam_token'] as String? ?? '',
      deviceFingerprint: json['device_fingerprint'] as String? ?? '',
      currentQuestionIndex: json['current_question_index'] as int? ?? 0,
      answers: ((json['answers'] as Map<String, dynamic>?) ?? const {}).map(
        (key, value) => MapEntry(key, value.toString()),
      ),
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
