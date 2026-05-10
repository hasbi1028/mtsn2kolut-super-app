import 'dart:convert';

import 'package:flutter_secure_storage/flutter_secure_storage.dart';
import 'package:shared_preferences/shared_preferences.dart';

class ExamSessionSnapshot {
  const ExamSessionSnapshot({
    required this.baseUrl,
    required this.examToken,
    this.roomToken = '',
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
  final String roomToken;
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
    return <String, Object?>{...toMetadataJson(), ...toSensitiveJson()};
  }

  Map<String, Object?> toMetadataJson() {
    return <String, Object?>{
      'base_url': baseUrl,
      'student_name': studentName,
      'student_nis': studentNis,
      'session_title': sessionTitle,
      'room_name': roomName,
      'scheduled_start_iso': scheduledStartIso,
      'scheduled_end_iso': scheduledEndIso,
      'duration_minutes': durationMinutes,
      'current_question_index': currentQuestionIndex,
      'played_audio_question_ids': playedAudioQuestionIds,
      'last_server_contact_iso': lastServerContactIso,
      'last_sync_failure_iso': lastSyncFailureIso,
      'consecutive_sync_failures': consecutiveSyncFailures,
    };
  }

  Map<String, Object?> toSensitiveJson() {
    return <String, Object?>{
      'exam_token': examToken,
      'room_token': roomToken,
      'device_fingerprint': deviceFingerprint,
      'answers': answers,
      'pending_answers': pendingAnswers,
    };
  }

  factory ExamSessionSnapshot.fromJson(Map<String, dynamic> json) {
    return ExamSessionSnapshot(
      baseUrl: json['base_url'] as String? ?? '',
      examToken: json['exam_token'] as String? ?? '',
      roomToken: json['room_token'] as String? ?? '',
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

abstract class SnapshotValueStore {
  Future<void> write(String key, String value);
  Future<String?> read(String key);
  Future<void> delete(String key);
}

class SharedPreferencesValueStore implements SnapshotValueStore {
  const SharedPreferencesValueStore();

  @override
  Future<void> delete(String key) async {
    final prefs = await SharedPreferences.getInstance();
    await prefs.remove(key);
  }

  @override
  Future<String?> read(String key) async {
    final prefs = await SharedPreferences.getInstance();
    return prefs.getString(key);
  }

  @override
  Future<void> write(String key, String value) async {
    final prefs = await SharedPreferences.getInstance();
    await prefs.setString(key, value);
  }
}

class SecureSnapshotValueStore implements SnapshotValueStore {
  const SecureSnapshotValueStore([
    this._storage = const FlutterSecureStorage(),
  ]);

  final FlutterSecureStorage _storage;

  @override
  Future<void> delete(String key) => _storage.delete(key: key);

  @override
  Future<String?> read(String key) => _storage.read(key: key);

  @override
  Future<void> write(String key, String value) =>
      _storage.write(key: key, value: value);
}

class ExamSessionStore {
  static const _baseUrlKey = 'exam_last_base_url';
  static const _snapshotMetadataKey = 'exam_active_snapshot';
  static const _snapshotSensitiveKey = 'exam_active_snapshot_secure';

  ExamSessionStore({
    SnapshotValueStore? metadataStore,
    SnapshotValueStore? secureStore,
  }) : _metadataStore = metadataStore ?? const SharedPreferencesValueStore(),
       _secureStore = secureStore ?? const SecureSnapshotValueStore();

  final SnapshotValueStore _metadataStore;
  final SnapshotValueStore _secureStore;

  Future<void> saveBaseUrl(String baseUrl) async {
    await _metadataStore.write(_baseUrlKey, baseUrl);
  }

  Future<String?> loadBaseUrl() async {
    return _metadataStore.read(_baseUrlKey);
  }

  Future<void> saveSnapshot(ExamSessionSnapshot snapshot) async {
    await _metadataStore.write(
      _snapshotMetadataKey,
      jsonEncode(snapshot.toMetadataJson()),
    );
    await _secureStore.write(
      _snapshotSensitiveKey,
      jsonEncode(snapshot.toSensitiveJson()),
    );
    await _metadataStore.write(_baseUrlKey, snapshot.baseUrl);
  }

  Future<ExamSessionSnapshot?> loadSnapshot() async {
    final rawMetadata = await _metadataStore.read(_snapshotMetadataKey);
    if (rawMetadata == null || rawMetadata.trim().isEmpty) {
      return null;
    }
    final decodedMetadata = _tryDecodeSnapshot(rawMetadata);
    if (decodedMetadata is! Map<String, dynamic>) {
      return null;
    }

    if (_isLegacySnapshot(decodedMetadata)) {
      final snapshot = ExamSessionSnapshot.fromJson(decodedMetadata);
      await saveSnapshot(snapshot);
      return snapshot;
    }

    final rawSensitive = await _secureStore.read(_snapshotSensitiveKey);
    if (rawSensitive == null || rawSensitive.trim().isEmpty) {
      return null;
    }
    final decodedSensitive = _tryDecodeSnapshot(rawSensitive);
    if (decodedSensitive is! Map<String, dynamic>) {
      return null;
    }

    return ExamSessionSnapshot.fromJson(<String, dynamic>{
      ...decodedMetadata,
      ...decodedSensitive,
    });
  }

  Future<void> clearSnapshot() async {
    await _metadataStore.delete(_snapshotMetadataKey);
    await _secureStore.delete(_snapshotSensitiveKey);
  }
}

bool _isLegacySnapshot(Map<String, dynamic> json) {
  return json.containsKey('exam_token') ||
      json.containsKey('answers') ||
      json.containsKey('pending_answers') ||
      json.containsKey('device_fingerprint');
}

Object? _tryDecodeSnapshot(String raw) {
  try {
    return jsonDecode(raw);
  } on FormatException {
    return null;
  }
}
