import 'dart:convert';
import 'dart:io';

import 'package:flutter_test/flutter_test.dart';
import 'package:mobile/src/exam_api.dart';

void main() {
  group('ExamApiClient', () {
    const deviceFingerprint = 'android:test';
    late HttpServer server;
    late String baseUrl;

    setUp(() async {
      server = await HttpServer.bind(InternetAddress.loopbackIPv4, 0);
      baseUrl = 'http://${server.address.address}:${server.port}';
    });

    tearDown(() async {
      await server.close(force: true);
    });

    test('examAssetHeaders carries exam token and device fingerprint', () {
      final client = ExamApiClient(
        baseUrl: 'http://127.0.0.1',
        deviceFingerprint: deviceFingerprint,
      );

      expect(client.examAssetHeaders(' token-1 '), <String, String>{
        'X-Exam-Token': 'token-1',
        'X-Device-Fingerprint': deviceFingerprint,
      });
      expect(client.examAssetHeaders(''), isEmpty);
    });

    test('normalizes and rejects unsafe operator base URLs', () {
      expect(
        ExamApiClient.normalizeBaseUrl(' http://127.0.0.1:8080/ '),
        'http://127.0.0.1:8080',
      );
      expect(
        () => ExamApiClient(baseUrl: 'ftp://127.0.0.1:8080'),
        throwsA(isA<ExamApiException>()),
      );
      expect(
        () => ExamApiClient(baseUrl: 'http://127.0.0.1:8080/api?x=1'),
        throwsA(isA<ExamApiException>()),
      );
      expect(
        () => ExamApiClient(baseUrl: 'https://user:pass@127.0.0.1:8080'),
        throwsA(isA<ExamApiException>()),
      );
    });

    test('asset credentials are limited to relative and same-origin URLs', () {
      final client = ExamApiClient(
        baseUrl: 'https://exam.example.test:8443',
        deviceFingerprint: deviceFingerprint,
      );

      expect(
        client.examAssetHeadersForUrl('token-1', '/assets/audio.mp3'),
        containsPair('X-Exam-Token', 'token-1'),
      );
      expect(
        client.examAssetHeadersForUrl('token-1', 'assets/audio.mp3'),
        containsPair('X-Exam-Token', 'token-1'),
      );
      expect(
        client.examAssetHeadersForUrl(
          'token-1',
          'https://exam.example.test:8443/assets/audio.mp3',
        ),
        containsPair('X-Device-Fingerprint', deviceFingerprint),
      );
      expect(
        client.examAssetHeadersForUrl(
          'token-1',
          '//exam.example.test:8443/assets/audio.mp3',
        ),
        isEmpty,
      );
      expect(
        client.examAssetHeadersForUrl(
          'token-1',
          '//cdn.example.test/assets/audio.mp3',
        ),
        isEmpty,
      );
      expect(
        client.examAssetHeadersForUrl(
          'token-1',
          'https://cdn.example.test/assets/audio.mp3',
        ),
        isEmpty,
      );
    });

    test('login unwraps data envelope into payload', () async {
      server.listen((request) async {
        expect(request.uri.path, '/api/exam/login');
        expect(request.method, 'POST');

        final rawBody = await utf8.decoder.bind(request).join();
        expect(jsonDecode(rawBody), <String, dynamic>{
          'token': 'token-1',
          'device_fingerprint': 'android:test',
        });

        request.response
          ..statusCode = 200
          ..headers.contentType = ContentType.json
          ..write(
            jsonEncode({
              'data': {
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
                'questions': const [],
                'answered_count': 0,
                'total_questions': 20,
                'time_remaining_seconds': 1800,
              },
            }),
          );
        await request.response.close();
      });

      final client = ExamApiClient(
        baseUrl: baseUrl,
        deviceFingerprint: deviceFingerprint,
      );
      final payload = await client.login(
        token: 'token-1',
        deviceFingerprint: 'android:test',
      );

      expect(payload.participantId, 'participant-1');
      expect(payload.session.title, 'Matematika Kelas VIII');
      expect(payload.room?.roomName, 'Lab 1');
    });

    test('sendHeartbeat sends exam token and empty payload contract', () async {
      server.listen((request) async {
        expect(request.uri.path, '/api/exam/heartbeat');
        expect(request.method, 'POST');
        expect(request.headers.value('X-Exam-Token'), 'token-1');
        expect(
          request.headers.value('X-Device-Fingerprint'),
          deviceFingerprint,
        );

        final rawBody = await utf8.decoder.bind(request).join();
        expect(jsonDecode(rawBody), <String, dynamic>{});

        request.response
          ..statusCode = 200
          ..headers.contentType = ContentType.json
          ..write(
            jsonEncode({
              'data': {'status': 'ok'},
            }),
          );
        await request.response.close();
      });

      final client = ExamApiClient(
        baseUrl: baseUrl,
        deviceFingerprint: deviceFingerprint,
      );
      await client.sendHeartbeat('token-1');
    });

    test('maps socket transport failures into controlled exception', () async {
      final client = ExamApiClient(baseUrl: 'http://127.0.0.1:1');

      await expectLater(
        () => client.getStatus('token-1'),
        throwsA(
          isA<ExamApiException>()
              .having((error) => error.statusCode, 'statusCode', isNull)
              .having(
                (error) => error.message,
                'message',
                'Tidak bisa terhubung ke server ujian.',
              ),
        ),
      );
    });

    test('maps request timeouts into controlled exception', () async {
      server.listen((request) async {
        await Future<void>.delayed(const Duration(milliseconds: 120));
        request.response
          ..statusCode = 200
          ..headers.contentType = ContentType.json
          ..write(jsonEncode({'data': <String, Object?>{}}));
        await request.response.close();
      });

      final client = ExamApiClient(
        baseUrl: baseUrl,
        requestTimeout: const Duration(milliseconds: 20),
      );

      await expectLater(
        () => client.getStatus('token-1'),
        throwsA(
          isA<ExamApiException>().having(
            (error) => error.message,
            'message',
            'Koneksi ke server ujian terlalu lama merespons.',
          ),
        ),
      );
    });

    test('getStatus unwraps submission status from data envelope', () async {
      server.listen((request) async {
        expect(request.uri.path, '/api/exam/status');
        expect(request.headers.value('X-Exam-Token'), 'token-1');
        expect(
          request.headers.value('X-Device-Fingerprint'),
          deviceFingerprint,
        );

        request.response
          ..statusCode = 200
          ..headers.contentType = ContentType.json
          ..write(
            jsonEncode({
              'data': {
                'answered_count': 10,
                'total_questions': 20,
                'time_remaining_seconds': 600,
                'is_submitted': true,
              },
            }),
          );
        await request.response.close();
      });

      final client = ExamApiClient(
        baseUrl: baseUrl,
        deviceFingerprint: deviceFingerprint,
      );
      final payload = await client.getStatus('token-1');

      expect(payload.answeredCount, 10);
      expect(payload.totalQuestions, 20);
      expect(payload.timeRemainingSeconds, 600);
      expect(payload.isSubmitted, isTrue);
    });

    test('saveAnswer sends exam token and answer payload contract', () async {
      server.listen((request) async {
        expect(request.uri.path, '/api/exam/answer');
        expect(request.method, 'POST');
        expect(request.headers.value('X-Exam-Token'), 'token-1');
        expect(
          request.headers.value('X-Device-Fingerprint'),
          deviceFingerprint,
        );

        final rawBody = await utf8.decoder.bind(request).join();
        expect(jsonDecode(rawBody), <String, dynamic>{
          'question_id': 'question-7',
          'answer': 'Pilihan B',
        });

        request.response
          ..statusCode = 200
          ..headers.contentType = ContentType.json
          ..write(
            jsonEncode({
              'data': {'status': 'recorded'},
            }),
          );
        await request.response.close();
      });

      final client = ExamApiClient(
        baseUrl: baseUrl,
        deviceFingerprint: deviceFingerprint,
      );
      await client.saveAnswer(
        token: 'token-1',
        questionId: 'question-7',
        answer: 'Pilihan B',
      );
    });

    test('sendEvent sends event type and data payload contract', () async {
      server.listen((request) async {
        expect(request.uri.path, '/api/exam/event');
        expect(request.method, 'POST');
        expect(request.headers.value('X-Exam-Token'), 'token-1');
        expect(
          request.headers.value('X-Device-Fingerprint'),
          deviceFingerprint,
        );

        final rawBody = await utf8.decoder.bind(request).join();
        expect(jsonDecode(rawBody), <String, dynamic>{
          'event_type': 'repeat_resume_attempt',
          'data': <String, dynamic>{'count': 2, 'source': 'resume_gate'},
        });

        request.response
          ..statusCode = 200
          ..headers.contentType = ContentType.json
          ..write(
            jsonEncode({
              'data': {'status': 'recorded'},
            }),
          );
        await request.response.close();
      });

      final client = ExamApiClient(
        baseUrl: baseUrl,
        deviceFingerprint: deviceFingerprint,
      );
      await client.sendEvent(
        token: 'token-1',
        eventType: 'repeat_resume_attempt',
        data: const <String, Object?>{'count': 2, 'source': 'resume_gate'},
      );
    });

    test('submit sends exam token and empty payload contract', () async {
      server.listen((request) async {
        expect(request.uri.path, '/api/exam/submit');
        expect(request.method, 'POST');
        expect(request.headers.value('X-Exam-Token'), 'token-1');
        expect(
          request.headers.value('X-Device-Fingerprint'),
          deviceFingerprint,
        );

        final rawBody = await utf8.decoder.bind(request).join();
        expect(jsonDecode(rawBody), <String, dynamic>{});

        request.response
          ..statusCode = 200
          ..headers.contentType = ContentType.json
          ..write(
            jsonEncode({
              'data': {'status': 'submitted'},
            }),
          );
        await request.response.close();
      });

      final client = ExamApiClient(
        baseUrl: baseUrl,
        deviceFingerprint: deviceFingerprint,
      );
      await client.submit('token-1');
    });

    test('throws controlled exception when data envelope is missing', () async {
      server.listen((request) async {
        request.response
          ..statusCode = 200
          ..headers.contentType = ContentType.json
          ..write(jsonEncode({'message': 'ok tanpa data'}));
        await request.response.close();
      });

      final client = ExamApiClient(
        baseUrl: baseUrl,
        deviceFingerprint: deviceFingerprint,
      );

      await expectLater(
        () => client.getStatus('token-1'),
        throwsA(
          isA<ExamApiException>().having(
            (error) => error.message,
            'message',
            'Respons server ujian tidak lengkap.',
          ),
        ),
      );
    });

    test(
      'throws controlled exception when success response json is malformed',
      () async {
        server.listen((request) async {
          request.response
            ..statusCode = 200
            ..headers.contentType = ContentType.json
            ..write('{bukan-json-valid');
          await request.response.close();
        });

        final client = ExamApiClient(
          baseUrl: baseUrl,
          deviceFingerprint: deviceFingerprint,
        );

        await expectLater(
          () => client.getStatus('token-1'),
          throwsA(
            isA<ExamApiException>()
                .having((error) => error.statusCode, 'statusCode', 200)
                .having(
                  (error) => error.message,
                  'message',
                  'Respons server ujian tidak valid.',
                ),
          ),
        );
      },
    );

    test('prefers backend message field on http error', () async {
      server.listen((request) async {
        request.response
          ..statusCode = 409
          ..headers.contentType = ContentType.json
          ..write(
            jsonEncode({
              'message': 'Token sudah terhubung dengan perangkat lain.',
            }),
          );
        await request.response.close();
      });

      final client = ExamApiClient(
        baseUrl: baseUrl,
        deviceFingerprint: deviceFingerprint,
      );

      await expectLater(
        () => client.login(token: 'token-1', deviceFingerprint: 'android:test'),
        throwsA(
          isA<ExamApiException>()
              .having((error) => error.statusCode, 'statusCode', 409)
              .having(
                (error) => error.message,
                'message',
                'Token sudah terhubung dengan perangkat lain.',
              ),
        ),
      );
    });

    test(
      'throws controlled exception when success response json is not an object',
      () async {
        server.listen((request) async {
          request.response
            ..statusCode = 200
            ..headers.contentType = ContentType.json
            ..write(jsonEncode(['bukan-object']));
          await request.response.close();
        });

        final client = ExamApiClient(
          baseUrl: baseUrl,
          deviceFingerprint: deviceFingerprint,
        );

        await expectLater(
          () => client.getStatus('token-1'),
          throwsA(
            isA<ExamApiException>()
                .having((error) => error.statusCode, 'statusCode', 200)
                .having(
                  (error) => error.message,
                  'message',
                  'Respons server ujian tidak valid.',
                ),
          ),
        );
      },
    );

    test(
      'throws controlled exception when error response json is malformed',
      () async {
        server.listen((request) async {
          request.response
            ..statusCode = 500
            ..headers.contentType = ContentType.json
            ..write('{bukan-json-valid');
          await request.response.close();
        });

        final client = ExamApiClient(
          baseUrl: baseUrl,
          deviceFingerprint: deviceFingerprint,
        );

        await expectLater(
          () => client.submit('token-1'),
          throwsA(
            isA<ExamApiException>()
                .having((error) => error.statusCode, 'statusCode', 500)
                .having(
                  (error) => error.message,
                  'message',
                  'Respons error server ujian tidak valid.',
                ),
          ),
        );
      },
    );

    test(
      'falls back to error field then generic message on http error',
      () async {
        var requestCount = 0;

        server.listen((request) async {
          requestCount += 1;

          request.response
            ..statusCode = 500
            ..headers.contentType = ContentType.json;

          if (requestCount == 1) {
            request.response.write(
              jsonEncode({'error': 'Server ujian sedang bermasalah.'}),
            );
          } else {
            request.response.write(
              jsonEncode({'detail': 'tanpa message/error'}),
            );
          }

          await request.response.close();
        });

        final client = ExamApiClient(
          baseUrl: baseUrl,
          deviceFingerprint: deviceFingerprint,
        );

        await expectLater(
          () => client.submit('token-1'),
          throwsA(
            isA<ExamApiException>().having(
              (error) => error.message,
              'message',
              'Server ujian sedang bermasalah.',
            ),
          ),
        );

        await expectLater(
          () => client.sendHeartbeat('token-1'),
          throwsA(
            isA<ExamApiException>().having(
              (error) => error.message,
              'message',
              'Permintaan ke server ujian gagal.',
            ),
          ),
        );
      },
    );

    test(
      'throws controlled exception when error response json is not an object',
      () async {
        server.listen((request) async {
          request.response
            ..statusCode = 500
            ..headers.contentType = ContentType.json
            ..write(jsonEncode(['bukan-object']));
          await request.response.close();
        });

        final client = ExamApiClient(
          baseUrl: baseUrl,
          deviceFingerprint: deviceFingerprint,
        );

        await expectLater(
          () => client.submit('token-1'),
          throwsA(
            isA<ExamApiException>()
                .having((error) => error.statusCode, 'statusCode', 500)
                .having(
                  (error) => error.message,
                  'message',
                  'Respons error server ujian tidak valid.',
                ),
          ),
        );
      },
    );
  });
}
