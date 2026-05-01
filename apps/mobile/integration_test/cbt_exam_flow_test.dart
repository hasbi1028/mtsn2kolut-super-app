import 'dart:async';
import 'dart:convert';
import 'dart:io';

import 'package:flutter/material.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:integration_test/integration_test.dart';
import 'package:mobile/src/exam_session_store.dart';
import 'package:mobile/src/screens/exam_login_screen.dart';

void main() {
  IntegrationTestWidgetsFlutterBinding.ensureInitialized();

  testWidgets('student can login, answer, and submit an exam', (tester) async {
    final server = await _MockExamServer.start();
    addTearDown(server.close);

    final metadataStore = _MemorySnapshotValueStore();
    final secureStore = _MemorySnapshotValueStore();
    final sessionStore = ExamSessionStore(
      metadataStore: metadataStore,
      secureStore: secureStore,
    );

    await tester.pumpWidget(
      _IntegrationTestApp(
        child: ExamLoginScreen(autoRestore: false, sessionStore: sessionStore),
      ),
    );
    await tester.pump();

    await tester.enterText(find.byType(TextField).first, 'a1b2c3d4');
    await _tapVisible(tester, find.text('Tampilkan'));
    await tester.enterText(find.byType(TextField).last, server.baseUrl);
    await _tapVisible(tester, find.widgetWithText(FilledButton, 'Masuk Ujian'));

    await _pumpUntilFound(tester, find.text('Siti Rahma Integrasi'));
    expect(server.loginRequests, hasLength(1));
    expect(server.loginRequests.single['token'], 'a1b2c3d4');
    expect(
      server.loginRequests.single['device_fingerprint'],
      isA<String>().having((value) => value, 'value', isNotEmpty),
    );

    await _pumpUntilFound(tester, find.text('2 + 2 = ...'));
    await _tapVisible(tester, find.text('4'));

    final multipleChoiceAnswer = await server.nextAnswer();
    expect(multipleChoiceAnswer['question_id'], 'question-pg-1');
    expect(multipleChoiceAnswer['answer'], 'B');

    await _tapVisible(tester, find.widgetWithText(FilledButton, 'Berikutnya'));
    await _pumpUntilFound(tester, find.text('Soal 2'));
    await tester.enterText(
      find.byType(TextField).last,
      'Tetap di aplikasi dan lapor ke pengawas.',
    );
    await _tapVisible(
      tester,
      find.widgetWithText(FilledButton, 'Simpan Jawaban'),
    );

    final essayAnswer = await server.nextAnswer();
    expect(essayAnswer['question_id'], 'question-essay-1');
    expect(essayAnswer['answer'], 'Tetap di aplikasi dan lapor ke pengawas.');

    await _tapVisible(tester, find.widgetWithText(FilledButton, 'Kirim Ujian'));
    await _pumpUntilFound(tester, find.text('Kirim jawaban akhir?'));
    await tester.tap(find.widgetWithText(FilledButton, 'Kirim Ujian').last);
    await tester.pump();

    await server.nextSubmit();
    await _pumpUntilFound(tester, find.text('Ujian berhasil dikirim.'));
    expect(find.text('2 / 2 soal'), findsOneWidget);
    expect(server.statusRequestCount, greaterThan(0));
    expect(server.examTokenHeaders, everyElement('a1b2c3d4'));
  });
}

class _EventQueue<T> {
  final List<T> _items = <T>[];
  final List<Completer<T>> _waiters = <Completer<T>>[];
  bool _closed = false;

  void add(T item) {
    if (_closed) {
      return;
    }
    if (_waiters.isNotEmpty) {
      _waiters.removeAt(0).complete(item);
      return;
    }
    _items.add(item);
  }

  Future<T> next() {
    if (_items.isNotEmpty) {
      return Future<T>.value(_items.removeAt(0));
    }
    if (_closed) {
      return Future<T>.error(StateError('queue closed'));
    }
    final completer = Completer<T>();
    _waiters.add(completer);
    return completer.future;
  }

  void close() {
    _closed = true;
    for (final waiter in _waiters) {
      if (!waiter.isCompleted) {
        waiter.completeError(StateError('queue closed'));
      }
    }
    _waiters.clear();
    _items.clear();
  }
}

class _IntegrationTestApp extends StatelessWidget {
  const _IntegrationTestApp({required this.child});

  final Widget child;

  @override
  Widget build(BuildContext context) {
    return MaterialApp(debugShowCheckedModeBanner: false, home: child);
  }
}

class _MemorySnapshotValueStore implements SnapshotValueStore {
  final Map<String, String> _values = <String, String>{};

  @override
  Future<void> delete(String key) async {
    _values.remove(key);
  }

  @override
  Future<String?> read(String key) async => _values[key];

  @override
  Future<void> write(String key, String value) async {
    _values[key] = value;
  }
}

class _MockExamServer {
  _MockExamServer._(this._server);

  final HttpServer _server;
  final List<Map<String, dynamic>> loginRequests = <Map<String, dynamic>>[];
  final List<String?> examTokenHeaders = <String?>[];
  final _answerQueue = _EventQueue<Map<String, dynamic>>();
  final _submitQueue = _EventQueue<void>();
  int statusRequestCount = 0;

  String get baseUrl => 'http://${_server.address.address}:${_server.port}';

  static Future<_MockExamServer> start() async {
    final server = await HttpServer.bind(InternetAddress.loopbackIPv4, 0);
    final mock = _MockExamServer._(server);
    unawaited(mock._serve());
    return mock;
  }

  Future<Map<String, dynamic>> nextAnswer() {
    return _answerQueue.next().timeout(const Duration(seconds: 5));
  }

  Future<void> nextSubmit() {
    return _submitQueue.next().timeout(const Duration(seconds: 5));
  }

  Future<void> close() async {
    _answerQueue.close();
    _submitQueue.close();
    await _server.close(force: true);
  }

  Future<void> _serve() async {
    await for (final request in _server) {
      final body = await utf8.decoder.bind(request).join();
      final jsonBody = body.trim().isEmpty
          ? <String, dynamic>{}
          : jsonDecode(body) as Map<String, dynamic>;

      switch ((request.method, request.uri.path)) {
        case ('POST', '/api/exam/login'):
          loginRequests.add(jsonBody);
          await _writeJson(request, _loginResponse());
        case ('GET', '/api/exam/status'):
          statusRequestCount += 1;
          examTokenHeaders.add(request.headers.value('X-Exam-Token'));
          await _writeJson(request, _statusResponse());
        case ('POST', '/api/exam/heartbeat'):
          examTokenHeaders.add(request.headers.value('X-Exam-Token'));
          await _writeJson(request, const {
            'data': {'status': 'ok'},
          });
        case ('POST', '/api/exam/event'):
          examTokenHeaders.add(request.headers.value('X-Exam-Token'));
          await _writeJson(request, const {
            'data': {'status': 'recorded'},
          });
        case ('POST', '/api/exam/answer'):
          examTokenHeaders.add(request.headers.value('X-Exam-Token'));
          _answerQueue.add(jsonBody);
          await _writeJson(request, const {
            'data': {'status': 'recorded'},
          });
        case ('POST', '/api/exam/submit'):
          examTokenHeaders.add(request.headers.value('X-Exam-Token'));
          _submitQueue.add(null);
          await _writeJson(request, const {
            'data': {'status': 'submitted'},
          });
        default:
          await _writeJson(request, const {
            'message': 'not found',
          }, statusCode: HttpStatus.notFound);
      }
    }
  }

  Future<void> _writeJson(
    HttpRequest request,
    Object payload, {
    int statusCode = HttpStatus.ok,
  }) async {
    request.response.statusCode = statusCode;
    request.response.headers.contentType = ContentType.json;
    request.response.write(jsonEncode(payload));
    await request.response.close();
  }

  Map<String, Object?> _loginResponse() {
    final now = DateTime.now();
    return <String, Object?>{
      'data': <String, Object?>{
        'participant_id': 'participant-integration-1',
        'student': <String, Object?>{
          'nis': 'CBTINT001',
          'nama': 'Siti Rahma Integrasi',
        },
        'session': <String, Object?>{
          'id': 'session-integration-1',
          'title': 'Simulasi Integrasi CBT',
          'scheduled_start': now
              .subtract(const Duration(minutes: 5))
              .toIso8601String(),
          'scheduled_end': now.add(const Duration(hours: 1)).toIso8601String(),
          'duration_minutes': 60,
        },
        'room': <String, Object?>{'room_name': 'Lab Integrasi'},
        'questions': <Object?>[
          <String, Object?>{
            'id': 'question-pg-1',
            'question_text': '2 + 2 = ...',
            'question_type': 'multiple_choice',
            'stem_html': '',
            'stimulus_html': '',
            'stem_media_url': '',
            'stimulus_media_url': '',
            'stem_audio_url': '',
            'stimulus_audio_url': '',
            'options': <Object?>[
              <String, Object?>{'label': 'A', 'text': '3'},
              <String, Object?>{'label': 'B', 'text': '4'},
              <String, Object?>{'label': 'C', 'text': '5'},
              <String, Object?>{'label': 'D', 'text': '6'},
            ],
          },
          <String, Object?>{
            'id': 'question-essay-1',
            'question_text':
                'Apa yang harus dilakukan jika koneksi ujian menurun?',
            'question_type': 'essay',
            'stem_html': '',
            'stimulus_html': '',
            'stem_media_url': '',
            'stimulus_media_url': '',
            'stem_audio_url': '',
            'stimulus_audio_url': '',
            'options': <Object?>[],
          },
        ],
        'answered_count': 0,
        'total_questions': 2,
        'time_remaining_seconds': 3600,
      },
    };
  }

  Map<String, Object?> _statusResponse() {
    return <String, Object?>{
      'data': <String, Object?>{
        'answered_count': 0,
        'total_questions': 2,
        'time_remaining_seconds': 3600,
        'is_submitted': false,
      },
    };
  }
}

Future<void> _tapVisible(WidgetTester tester, Finder finder) async {
  await tester.ensureVisible(finder);
  await tester.pump();
  await tester.tap(finder);
  await tester.pump();
}

Future<void> _pumpUntilFound(
  WidgetTester tester,
  Finder finder, {
  Duration timeout = const Duration(seconds: 8),
}) async {
  final end = DateTime.now().add(timeout);
  while (DateTime.now().isBefore(end)) {
    await tester.pump(const Duration(milliseconds: 100));
    if (finder.evaluate().isNotEmpty) {
      return;
    }
  }
  fail('Timed out waiting for $finder');
}
