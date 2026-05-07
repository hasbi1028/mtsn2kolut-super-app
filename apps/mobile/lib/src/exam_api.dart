import 'dart:async';
import 'dart:convert';
import 'dart:io';

import 'package:flutter/foundation.dart';

import 'models.dart';

class ExamApiException implements Exception {
  const ExamApiException(this.message, {this.statusCode});

  final String message;
  final int? statusCode;

  @override
  String toString() => 'ExamApiException($statusCode): $message';
}

class ExamApiClient {
  ExamApiClient({
    required String baseUrl,
    this.deviceFingerprint,
    HttpClient? httpClient,
    this.requestTimeout = const Duration(seconds: 20),
    bool allowDebugPlainHttp = _defaultAllowDebugPlainHttp,
  }) : baseUrl = normalizeBaseUrl(
         baseUrl,
         allowDebugPlainHttp: allowDebugPlainHttp,
       ),
       _httpClient = httpClient ?? HttpClient() {
    _httpClient.connectionTimeout = const Duration(seconds: 10);
    _httpClient.userAgent = userAgent;
  }

  static const String userAgent =
      'MTsN2Kolut-CBT/1.0 (Flutter; Android; BYOD)';

  final String baseUrl;
  final String? deviceFingerprint;
  final Duration requestTimeout;
  final HttpClient _httpClient;

  static const bool _defaultAllowDebugPlainHttp = bool.fromEnvironment(
    'ALLOW_PLAINTEXT_EXAM_HTTP',
    defaultValue: !kReleaseMode,
  );

  static String normalizeBaseUrl(
    String rawBaseUrl, {
    bool allowDebugPlainHttp = _defaultAllowDebugPlainHttp,
  }) {
    final uri = Uri.tryParse(rawBaseUrl.trim());
    if (uri == null ||
        (uri.scheme != 'http' && uri.scheme != 'https') ||
        uri.host.trim().isEmpty ||
        uri.userInfo.isNotEmpty ||
        uri.hasQuery ||
        uri.hasFragment ||
        (uri.path.isNotEmpty && uri.path != '/')) {
      throw const ExamApiException(
        'Alamat server ujian tidak valid. Gunakan alamat HTTPS tanpa path atau query.',
      );
    }
    if (uri.scheme == 'http') {
      if (!allowDebugPlainHttp || !_isLocalTrialHttpHost(uri.host)) {
        throw const ExamApiException(
          'Alamat server ujian produksi harus memakai HTTPS. HTTP hanya diizinkan untuk uji lokal/operator pada localhost atau jaringan privat.',
        );
      }
    }
    return uri.replace(path: '', query: null, fragment: null).toString();
  }

  Map<String, String> examAssetHeaders(String token) {
    final trimmedToken = token.trim();
    if (trimmedToken.isEmpty) {
      return const <String, String>{};
    }
    final headers = <String, String>{'X-Exam-Token': trimmedToken};
    final boundDevice = deviceFingerprint?.trim() ?? '';
    if (boundDevice.isNotEmpty) {
      headers['X-Device-Fingerprint'] = boundDevice;
    }
    return headers;
  }

  Map<String, String> examAssetHeadersForUrl(String token, String assetUrl) {
    final uri = Uri.tryParse(resolveAssetUrl(assetUrl));
    if (uri == null) {
      return const <String, String>{};
    }
    if (uri.hasScheme || uri.hasAuthority) {
      if (!uri.hasScheme) {
        return const <String, String>{};
      }
      final baseUri = Uri.parse(baseUrl);
      if (uri.scheme != baseUri.scheme ||
          uri.host != baseUri.host ||
          uri.port != baseUri.port) {
        return const <String, String>{};
      }
    }
    return examAssetHeaders(token);
  }

  String resolveAssetUrl(String assetUrl) {
    final trimmed = assetUrl.trim();
    if (trimmed.isEmpty) {
      return trimmed;
    }
    final uri = Uri.tryParse(trimmed);
    if (uri == null) {
      return trimmed;
    }
    if (uri.hasScheme) {
      return uri.toString();
    }
    if (uri.hasAuthority) {
      return trimmed;
    }
    return Uri.parse(baseUrl).resolveUri(uri).toString();
  }

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
    final payload = await _sendJson(
      'POST',
      '/api/exam/heartbeat',
      examToken: token,
      body: const <String, Object?>{},
    );
    _expectDataStatus(payload, 'ok');
  }

  Future<void> sendEvent({
    required String token,
    required String eventType,
    Map<String, Object?> data = const <String, Object?>{},
  }) async {
    final payload = await _sendJson(
      'POST',
      '/api/exam/event',
      examToken: token,
      body: <String, Object?>{'event_type': eventType, 'data': data},
    );
    _expectDataStatus(payload, 'recorded');
  }

  Future<void> saveAnswer({
    required String token,
    required String questionId,
    required String answer,
  }) async {
    final payload = await _sendJson(
      'POST',
      '/api/exam/answer',
      examToken: token,
      body: <String, Object?>{'question_id': questionId, 'answer': answer},
    );
    _expectDataStatus(payload, 'recorded');
  }

  Future<void> submit(String token) async {
    final payload = await _sendJson(
      'POST',
      '/api/exam/submit',
      examToken: token,
      body: const <String, Object?>{},
    );
    _expectDataStatus(payload, 'submitted');
  }

  Future<Map<String, dynamic>> _sendJson(
    String method,
    String path, {
    String? examToken,
    Map<String, Object?>? body,
  }) async {
    try {
      final uri = Uri.parse('$baseUrl$path');
      final request = await _httpClient
          .openUrl(method, uri)
          .timeout(requestTimeout);
      request.headers.contentType = ContentType.json;
      request.headers.set(HttpHeaders.acceptHeader, 'application/json');
      for (final entry in examAssetHeaders(examToken ?? '').entries) {
        request.headers.set(entry.key, entry.value);
      }
      if (body != null) {
        request.write(jsonEncode(body));
      }

      final response = await request.close().timeout(requestTimeout);
      final raw = await utf8.decoder
          .bind(response)
          .join()
          .timeout(requestTimeout);
      final parsed = _parseJsonResponse(raw, statusCode: response.statusCode);

      if (response.statusCode >= 400) {
        throw ExamApiException(
          _extractMessage(parsed) ?? 'Permintaan ke server ujian gagal.',
          statusCode: response.statusCode,
        );
      }

      return parsed;
    } on SocketException {
      throw const ExamApiException('Tidak bisa terhubung ke server ujian.');
    } on HttpException {
      throw const ExamApiException('Koneksi ke server ujian tidak valid.');
    } on TimeoutException {
      throw const ExamApiException(
        'Koneksi ke server ujian terlalu lama merespons.',
      );
    } on FormatException {
      throw const ExamApiException(
        'Alamat server ujian tidak valid. Gunakan alamat http/https tanpa path atau query.',
      );
    }
  }

  Map<String, dynamic> _unwrapData(Map<String, dynamic> payload) {
    final data = payload['data'];
    if (data is Map<String, dynamic>) {
      return data;
    }
    throw const ExamApiException('Respons server ujian tidak lengkap.');
  }

  void _expectDataStatus(Map<String, dynamic> payload, String expectedStatus) {
    final data = _unwrapData(payload);
    if (data['status'] == expectedStatus) {
      return;
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

  Map<String, dynamic> _parseJsonResponse(
    String raw, {
    required int statusCode,
  }) {
    if (raw.isEmpty) {
      return <String, dynamic>{};
    }

    try {
      final decoded = jsonDecode(raw);
      if (decoded is Map<String, dynamic>) {
        return decoded;
      }
    } on FormatException {
      throw ExamApiException(
        statusCode >= 400
            ? 'Respons error server ujian tidak valid.'
            : 'Respons server ujian tidak valid.',
        statusCode: statusCode,
      );
    }

    throw ExamApiException(
      statusCode >= 400
          ? 'Respons error server ujian tidak valid.'
          : 'Respons server ujian tidak valid.',
      statusCode: statusCode,
    );
  }
}

bool _isLocalTrialHttpHost(String host) {
  final normalized = host.toLowerCase().trim();
  if (normalized == 'localhost' || normalized == '10.0.2.2') {
    return true;
  }

  final address = InternetAddress.tryParse(normalized);
  if (address == null) {
    return false;
  }
  if (address.isLoopback) {
    return true;
  }
  if (address.type != InternetAddressType.IPv4) {
    return false;
  }

  final octets = normalized.split('.').map(int.tryParse).toList();
  if (octets.length != 4 || octets.any((octet) => octet == null)) {
    return false;
  }
  final first = octets[0]!;
  final second = octets[1]!;
  return first == 10 ||
      (first == 172 && second >= 16 && second <= 31) ||
      (first == 192 && second == 168);
}
