import 'dart:convert';
import 'dart:io';

import 'package:flutter_test/flutter_test.dart';
import 'package:mobile/src/exam_api.dart';

void main() {
  group('ExamApiClient', () {
    late HttpServer server;
    late String baseUrl;

    setUp(() async {
      server = await HttpServer.bind(InternetAddress.loopbackIPv4, 0);
      baseUrl = 'http://${server.address.address}:${server.port}';
    });

    tearDown(() async {
      await server.close(force: true);
    });

    test('login unwraps data envelope into payload', () async {
      server.listen((request) async {
        expect(request.uri.path, '/api/exam/login');
        expect(request.method, 'POST');

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

      final client = ExamApiClient(baseUrl: baseUrl);
      final payload = await client.login(
        token: 'token-1',
        deviceFingerprint: 'android:test',
      );

      expect(payload.participantId, 'participant-1');
      expect(payload.session.title, 'Matematika Kelas VIII');
      expect(payload.room?.roomName, 'Lab 1');
    });

    test('getStatus unwraps submission status from data envelope', () async {
      server.listen((request) async {
        expect(request.uri.path, '/api/exam/status');
        expect(request.headers.value('X-Exam-Token'), 'token-1');

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

      final client = ExamApiClient(baseUrl: baseUrl);
      final payload = await client.getStatus('token-1');

      expect(payload.answeredCount, 10);
      expect(payload.totalQuestions, 20);
      expect(payload.timeRemainingSeconds, 600);
      expect(payload.isSubmitted, isTrue);
    });

    test('throws controlled exception when data envelope is missing', () async {
      server.listen((request) async {
        request.response
          ..statusCode = 200
          ..headers.contentType = ContentType.json
          ..write(jsonEncode({'message': 'ok tanpa data'}));
        await request.response.close();
      });

      final client = ExamApiClient(baseUrl: baseUrl);

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

      final client = ExamApiClient(baseUrl: baseUrl);

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

        final client = ExamApiClient(baseUrl: baseUrl);

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
  });
}
