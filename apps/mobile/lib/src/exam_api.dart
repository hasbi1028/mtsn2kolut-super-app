import 'dart:convert';
import 'dart:io';

import 'models.dart';

class ExamApiException implements Exception {
  const ExamApiException(this.message, {this.statusCode});

  final String message;
  final int? statusCode;

  @override
  String toString() => 'ExamApiException($statusCode): $message';
}

class ExamApiClient {
  ExamApiClient({required this.baseUrl, HttpClient? httpClient})
    : _httpClient = httpClient ?? HttpClient();

  final String baseUrl;
  final HttpClient _httpClient;

  Future<ExamLoginPayload> login({
    required String token,
    required String deviceFingerprint,
  }) async {
    final payload = await _sendJson(
      'POST',
      '/api/exam/login',
      body: <String, Object?>{
        'token': token,
        'device_fingerprint': deviceFingerprint,
      },
    );
    return ExamLoginPayload.fromJson(_unwrapData(payload));
  }

  Future<ExamStatusPayload> getStatus(String token) async {
    final payload = await _sendJson(
      'GET',
      '/api/exam/status',
      examToken: token,
    );
    return ExamStatusPayload.fromJson(_unwrapData(payload));
  }

  Future<void> sendHeartbeat(String token) async {
    await _sendJson(
      'POST',
      '/api/exam/heartbeat',
      examToken: token,
      body: const <String, Object?>{},
    );
  }

  Future<void> sendEvent({
    required String token,
    required String eventType,
    Map<String, Object?> data = const <String, Object?>{},
  }) async {
    await _sendJson(
      'POST',
      '/api/exam/event',
      examToken: token,
      body: <String, Object?>{'event_type': eventType, 'data': data},
    );
  }

  Future<void> saveAnswer({
    required String token,
    required String questionId,
    required String answer,
  }) async {
    await _sendJson(
      'POST',
      '/api/exam/answer',
      examToken: token,
      body: <String, Object?>{'question_id': questionId, 'answer': answer},
    );
  }

  Future<void> submit(String token) async {
    await _sendJson(
      'POST',
      '/api/exam/submit',
      examToken: token,
      body: const <String, Object?>{},
    );
  }

  Future<Map<String, dynamic>> _sendJson(
    String method,
    String path, {
    String? examToken,
    Map<String, Object?>? body,
  }) async {
    final uri = Uri.parse('$baseUrl$path');
    final request = await _httpClient.openUrl(method, uri);
    request.headers.contentType = ContentType.json;
    request.headers.set(HttpHeaders.acceptHeader, 'application/json');
    if (examToken != null && examToken.isNotEmpty) {
      request.headers.set('X-Exam-Token', examToken);
    }
    if (body != null) {
      request.write(jsonEncode(body));
    }

    final response = await request.close();
    final raw = await utf8.decoder.bind(response).join();
    final parsed = raw.isEmpty
        ? <String, dynamic>{}
        : jsonDecode(raw) as Map<String, dynamic>;

    if (response.statusCode >= 400) {
      throw ExamApiException(
        _extractMessage(parsed) ?? 'Permintaan ke server ujian gagal.',
        statusCode: response.statusCode,
      );
    }

    return parsed;
  }

  Map<String, dynamic> _unwrapData(Map<String, dynamic> payload) {
    final data = payload['data'];
    if (data is Map<String, dynamic>) {
      return data;
    }
    throw const ExamApiException('Respons server ujian tidak lengkap.');
  }

  String? _extractMessage(Map<String, dynamic> payload) {
    final directMessage = payload['message'];
    if (directMessage is String && directMessage.trim().isNotEmpty) {
      return directMessage;
    }

    final error = payload['error'];
    if (error is String && error.trim().isNotEmpty) {
      return error;
    }

    return null;
  }
}
