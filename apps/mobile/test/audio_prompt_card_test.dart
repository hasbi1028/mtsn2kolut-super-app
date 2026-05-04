import 'dart:convert';
import 'dart:io';

import 'package:flutter/material.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:mobile/src/widgets/audio_prompt_card.dart';

void main() {
  group('AudioPromptCard', () {
    late HttpServer server;
    late String baseUrl;

    setUp(() async {
      server = await HttpServer.bind(InternetAddress.loopbackIPv4, 0);
      baseUrl = 'http://${server.address.address}:${server.port}';
    });

    tearDown(() async {
      await server.close(force: true);
    });

    testWidgets('shows an error for over-limit audio responses', (
      tester,
    ) async {
      server.listen((request) async {
        final bytes = utf8.encode('too-large!!!');
        request.response
          ..statusCode = 200
          ..headers.contentType = ContentType('audio', 'mpeg')
          ..contentLength = bytes.length
          ..add(bytes);
        await request.response.close();
      });

      await _pumpAudioCard(tester, url: '$baseUrl/audio.mp3', maxAudioBytes: 4);
      await tester.tap(find.text('Putar Audio'));
      await tester.pumpAndSettle();

      expect(
        find.text('Audio tidak dapat diputar di perangkat ini.'),
        findsOneWidget,
      );
    });

    testWidgets('shows an error for wrong MIME audio responses', (
      tester,
    ) async {
      server.listen((request) async {
        request.response
          ..statusCode = 200
          ..headers.contentType = ContentType.text
          ..write('not audio');
        await request.response.close();
      });

      await _pumpAudioCard(tester, url: '$baseUrl/not-audio.txt');
      await tester.tap(find.text('Putar Audio'));
      await tester.pumpAndSettle();

      expect(
        find.text('Audio tidak dapat diputar di perangkat ini.'),
        findsOneWidget,
      );
    });

    testWidgets('shows an error for slow audio responses', (tester) async {
      server.listen((request) async {
        request.response
          ..statusCode = 200
          ..headers.contentType = ContentType('audio', 'mpeg');
        await Future<void>.delayed(const Duration(milliseconds: 80));
        request.response.write(utf8.encode('slow-audio'));
        await request.response.close();
      });

      await _pumpAudioCard(
        tester,
        url: '$baseUrl/slow-audio.mp3',
        audioTimeout: const Duration(milliseconds: 20),
      );
      await tester.tap(find.text('Putar Audio'));
      await tester.pumpAndSettle();

      expect(
        find.text('Audio tidak dapat diputar di perangkat ini.'),
        findsOneWidget,
      );
    });
  });
}

Future<void> _pumpAudioCard(
  WidgetTester tester, {
  required String url,
  int maxAudioBytes = 8 * 1024 * 1024,
  Duration audioTimeout = const Duration(seconds: 20),
}) async {
  await tester.pumpWidget(
    MaterialApp(
      home: Scaffold(
        body: AudioPromptCard(
          url: url,
          label: 'Audio soal',
          maxAudioBytes: maxAudioBytes,
          audioTimeout: audioTimeout,
        ),
      ),
    ),
  );
}
