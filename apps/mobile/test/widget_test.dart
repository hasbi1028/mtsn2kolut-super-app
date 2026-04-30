import 'package:flutter/material.dart';
import 'package:flutter_test/flutter_test.dart';

import 'package:mobile/src/app.dart';
import 'package:mobile/src/exam_error_messages.dart';
import 'package:mobile/src/exam_api.dart';
import 'package:mobile/src/models.dart';
import 'package:mobile/src/exam_session_store.dart';
import 'package:mobile/src/screens/exam_login_screen.dart';
import 'package:mobile/src/screens/exam_restore_failed_screen.dart';
import 'package:mobile/src/screens/exam_shell_screen.dart';

void main() {
  testWidgets('login screen renders exam shell entry', (tester) async {
    await tester.pumpWidget(const MtsnMobileApp());

    expect(find.text('Masuk Ujian'), findsWidgets);
    expect(find.text('Token ujian'), findsOneWidget);
    expect(find.text('Alamat server API'), findsOneWidget);
  });

  testWidgets('login screen renders persistent guidance notice', (
    tester,
  ) async {
    await tester.pumpWidget(
      const _TestApp(
        child: ExamLoginScreen(
          autoRestore: false,
          initialErrorMessage:
              'Token ini sudah terhubung dengan perangkat lain. Gunakan perangkat yang sama atau minta bantuan pengawas.',
          initialErrorNotice: ExamGuidanceNotice(
            title: 'Token sudah terikat ke perangkat lain',
            message:
                'Jangan terus mencoba login dari perangkat ini. Gunakan perangkat yang sama seperti sebelumnya atau minta pengawas memverifikasi token.',
            tone: ExamGuidanceTone.danger,
          ),
        ),
      ),
    );

    await tester.pump();

    expect(find.text('Masuk Ujian'), findsWidgets);
    expect(find.text('Token sudah terikat ke perangkat lain'), findsOneWidget);
    expect(
      find.text(
        'Jangan terus mencoba login dari perangkat ini. Gunakan perangkat yang sama seperti sebelumnya atau minta pengawas memverifikasi token.',
      ),
      findsOneWidget,
    );
  });

  testWidgets('login screen renders cached restore snapshot card', (
    tester,
  ) async {
    await tester.pumpWidget(
      _TestApp(
        child: ExamLoginScreen(
          autoRestore: false,
          previewSnapshot: _sampleSnapshot(
            lastServerContactIso: '2026-05-01T08:44:00+08:00',
            lastSyncFailureIso: '2026-05-01T08:46:00+08:00',
            consecutiveSyncFailures: 3,
          ),
        ),
      ),
    );

    await tester.pump();

    expect(find.text('Sesi terakhir terdeteksi'), findsOneWidget);
    expect(find.textContaining('Matematika Kelas VIII'), findsOneWidget);
    expect(find.text('Perlu perhatian koneksi'), findsOneWidget);
    expect(find.textContaining('Kontak server 08:44'), findsOneWidget);
    expect(find.textContaining('Gangguan 08:46'), findsOneWidget);
  });

  testWidgets('restore failed screen renders persistent guidance notice', (
    tester,
  ) async {
    tester.view.physicalSize = const Size(1280, 2200);
    tester.view.devicePixelRatio = 1.0;
    addTearDown(tester.view.resetPhysicalSize);
    addTearDown(tester.view.resetDevicePixelRatio);

    await tester.pumpWidget(
      _TestApp(
        child: ExamRestoreFailedScreen(
          snapshot: ExamSessionSnapshot(
            baseUrl: 'http://10.0.2.2:8080',
            examToken: 'abc12345',
            deviceFingerprint: 'android:test',
            studentName: 'Siti Aminah',
            studentNis: '24001',
            sessionTitle: 'Matematika Kelas VIII',
            roomName: 'Lab 1',
            scheduledStartIso: '2026-05-01T08:00:00+08:00',
            scheduledEndIso: '2026-05-01T09:30:00+08:00',
            durationMinutes: 90,
            currentQuestionIndex: 4,
            answers: const <String, String>{},
            pendingAnswers: const <String, String>{},
            playedAudioQuestionIds: const <String>[],
            lastServerContactIso: '2026-05-01T08:44:00+08:00',
            lastSyncFailureIso: '2026-05-01T08:46:00+08:00',
            consecutiveSyncFailures: 3,
          ),
          message:
              'Sesi lama terikat ke perangkat lain. Gunakan perangkat yang sama seperti sebelumnya atau minta bantuan pengawas.',
          notice: ExamGuidanceNotice(
            title: 'Sesi lama aktif di perangkat lain',
            message:
                'Peserta tidak perlu terus mencoba restore di perangkat ini. Pengawas sebaiknya mengarahkan peserta kembali ke perangkat awal atau memeriksa status token.',
            tone: ExamGuidanceTone.danger,
          ),
        ),
      ),
    );

    await tester.pump();

    expect(find.text('Sesi lama tidak bisa dipulihkan'), findsOneWidget);
    expect(find.text('Sesi lama aktif di perangkat lain'), findsOneWidget);
    expect(
      find.text(
        'Peserta tidak perlu terus mencoba restore di perangkat ini. Pengawas sebaiknya mengarahkan peserta kembali ke perangkat awal atau memeriksa status token.',
      ),
      findsOneWidget,
    );
    expect(find.text('Matematika Kelas VIII'), findsOneWidget);
    expect(find.text('Perlu perhatian koneksi'), findsOneWidget);
    expect(find.textContaining('Kontak server 08:44'), findsOneWidget);
    expect(find.textContaining('Gangguan 08:46'), findsOneWidget);
  });

  testWidgets('exam shell renders warning guidance notice', (tester) async {
    tester.view.physicalSize = const Size(1440, 2200);
    tester.view.devicePixelRatio = 1.0;
    addTearDown(tester.view.resetPhysicalSize);
    addTearDown(tester.view.resetDevicePixelRatio);

    await tester.pumpWidget(
      _TestApp(
        child: ExamShellScreen(
          client: ExamApiClient(baseUrl: 'http://127.0.0.1:65535'),
          examToken: 'abc12345',
          deviceFingerprint: 'android:test',
          autoStartRuntime: false,
          initialServerNotice: const ExamGuidanceNotice(
            title: 'Waktu ujian sudah berakhir',
            message:
                'Jawaban lokal masih aman di perangkat ini, tetapi pengawas perlu memastikan apakah sesi masih bisa dipulihkan atau harus diakhiri.',
            tone: ExamGuidanceTone.warning,
          ),
          initialPayload: _sampleLoginPayload(),
        ),
      ),
    );

    await tester.pump();

    expect(find.text('Waktu ujian sudah berakhir'), findsOneWidget);
    expect(
      find.text(
        'Jawaban lokal masih aman di perangkat ini, tetapi pengawas perlu memastikan apakah sesi masih bisa dipulihkan atau harus diakhiri.',
      ),
      findsOneWidget,
    );
  });

  testWidgets('exam shell renders danger guidance notice', (tester) async {
    tester.view.physicalSize = const Size(1440, 2200);
    tester.view.devicePixelRatio = 1.0;
    addTearDown(tester.view.resetPhysicalSize);
    addTearDown(tester.view.resetDevicePixelRatio);

    await tester.pumpWidget(
      _TestApp(
        child: ExamShellScreen(
          client: ExamApiClient(baseUrl: 'http://127.0.0.1:65535'),
          examToken: 'abc12345',
          deviceFingerprint: 'android:test',
          autoStartRuntime: false,
          initialServerNotice: const ExamGuidanceNotice(
            title: 'Ujian sudah selesai di server',
            message:
                'Perangkat ini tidak dapat mengirim jawaban baru lagi. Pengawas sebaiknya mengecek apakah submit sebelumnya sudah final.',
            tone: ExamGuidanceTone.danger,
          ),
          initialPayload: _sampleLoginPayload(),
        ),
      ),
    );

    await tester.pump();

    expect(find.text('Ujian sudah selesai di server'), findsOneWidget);
    expect(
      find.text(
        'Perangkat ini tidak dapat mengirim jawaban baru lagi. Pengawas sebaiknya mengecek apakah submit sebelumnya sudah final.',
      ),
      findsOneWidget,
    );
  });

  testWidgets('exam shell shows audio not-played state', (tester) async {
    tester.view.physicalSize = const Size(1440, 2200);
    tester.view.devicePixelRatio = 1.0;
    addTearDown(tester.view.resetPhysicalSize);
    addTearDown(tester.view.resetDevicePixelRatio);

    await tester.pumpWidget(
      _TestApp(
        child: ExamShellScreen(
          client: ExamApiClient(baseUrl: 'http://127.0.0.1:65535'),
          examToken: 'abc12345',
          deviceFingerprint: 'android:test',
          autoStartRuntime: false,
          initialPayload: _sampleAudioLoginPayload(),
        ),
      ),
    );

    await tester.pump();

    expect(find.text('Audio soal belum diputar'), findsOneWidget);
  });

  testWidgets('exam shell shows audio played state from restored snapshot', (
    tester,
  ) async {
    tester.view.physicalSize = const Size(1440, 2200);
    tester.view.devicePixelRatio = 1.0;
    addTearDown(tester.view.resetPhysicalSize);
    addTearDown(tester.view.resetDevicePixelRatio);

    await tester.pumpWidget(
      _TestApp(
        child: ExamShellScreen(
          client: ExamApiClient(baseUrl: 'http://127.0.0.1:65535'),
          examToken: 'abc12345',
          deviceFingerprint: 'android:test',
          autoStartRuntime: false,
          restoredSnapshot: _sampleSnapshot(
            playedAudioQuestionIds: const <String>['question-audio-1'],
          ),
          initialPayload: _sampleAudioLoginPayload(),
        ),
      ),
    );

    await tester.pump();

    expect(find.text('Audio soal sudah diputar'), findsOneWidget);
  });

  testWidgets('exam shell renders media card when question has media url', (
    tester,
  ) async {
    tester.view.physicalSize = const Size(1440, 2200);
    tester.view.devicePixelRatio = 1.0;
    addTearDown(tester.view.resetPhysicalSize);
    addTearDown(tester.view.resetDevicePixelRatio);

    await tester.pumpWidget(
      _TestApp(
        child: ExamShellScreen(
          client: ExamApiClient(baseUrl: 'http://127.0.0.1:65535'),
          examToken: 'abc12345',
          deviceFingerprint: 'android:test',
          autoStartRuntime: false,
          initialPayload: _sampleMediaLoginPayload(),
        ),
      ),
    );

    await tester.pump();

    expect(find.text('Media soal'), findsOneWidget);
  });

  testWidgets('exam shell renders stale supervisor attention panel', (
    tester,
  ) async {
    tester.view.physicalSize = const Size(1440, 2200);
    tester.view.devicePixelRatio = 1.0;
    addTearDown(tester.view.resetPhysicalSize);
    addTearDown(tester.view.resetDevicePixelRatio);

    final now = DateTime.now();
    final staleContact = now.subtract(const Duration(minutes: 3));

    await tester.pumpWidget(
      _TestApp(
        child: ExamShellScreen(
          client: ExamApiClient(baseUrl: 'http://127.0.0.1:65535'),
          examToken: 'abc12345',
          deviceFingerprint: 'android:test',
          autoStartRuntime: false,
          restoredSnapshot: _sampleSnapshot(
            lastServerContactIso: staleContact.toIso8601String(),
            consecutiveSyncFailures: 0,
          ),
          initialPayload: _sampleLoginPayload(),
        ),
      ),
    );

    await tester.pump();

    expect(find.text('Perlu intervensi pengawas'), findsOneWidget);
    expect(
      find.textContaining('Status koneksi berada di level waspada'),
      findsOneWidget,
    );
    expect(
      find.textContaining('Jika kondisi ini bertahan sampai 4 menit'),
      findsOneWidget,
    );
  });

  testWidgets('exam shell renders degraded mode panel', (tester) async {
    tester.view.physicalSize = const Size(1440, 2200);
    tester.view.devicePixelRatio = 1.0;
    addTearDown(tester.view.resetPhysicalSize);
    addTearDown(tester.view.resetDevicePixelRatio);

    final now = DateTime.now();

    await tester.pumpWidget(
      _TestApp(
        child: ExamShellScreen(
          client: ExamApiClient(baseUrl: 'http://127.0.0.1:65535'),
          examToken: 'abc12345',
          deviceFingerprint: 'android:test',
          autoStartRuntime: false,
          restoredSnapshot: _sampleSnapshot(
            lastServerContactIso: now
                .subtract(const Duration(minutes: 1))
                .toIso8601String(),
            lastSyncFailureIso: now.toIso8601String(),
            consecutiveSyncFailures: 3,
            pendingAnswers: const <String, String>{'question-1': 'B'},
          ),
          initialPayload: _sampleLoginPayload(),
        ),
      ),
    );

    await tester.pump();

    expect(find.text('Mode koneksi menurun aktif'), findsOneWidget);
    expect(
      find.textContaining('Sinkron gagal 3 kali berturut-turut'),
      findsOneWidget,
    );
    expect(find.text('Pulihkan Sinkron'), findsOneWidget);
  });

  testWidgets('exam shell renders repeated connection warning panel', (
    tester,
  ) async {
    tester.view.physicalSize = const Size(1440, 2200);
    tester.view.devicePixelRatio = 1.0;
    addTearDown(tester.view.resetPhysicalSize);
    addTearDown(tester.view.resetDevicePixelRatio);

    final now = DateTime.now();

    await tester.pumpWidget(
      _TestApp(
        child: ExamShellScreen(
          client: ExamApiClient(baseUrl: 'http://127.0.0.1:65535'),
          examToken: 'abc12345',
          deviceFingerprint: 'android:test',
          autoStartRuntime: false,
          restoredSnapshot: _sampleSnapshot(
            lastServerContactIso: now
                .subtract(const Duration(seconds: 30))
                .toIso8601String(),
            lastSyncFailureIso: now.toIso8601String(),
            consecutiveSyncFailures: 2,
          ),
          initialPayload: _sampleLoginPayload(),
        ),
      ),
    );

    await tester.pumpAndSettle();

    expect(find.text('Koneksi perlu diperhatikan'), findsOneWidget);
    expect(
      find.textContaining(
        'Perangkat mengalami 2 gangguan sinkron berturut-turut',
      ),
      findsOneWidget,
    );
    expect(find.text('Coba Sinkron Ulang'), findsOneWidget);
  });

  testWidgets('exam shell renders secured mode overlay before resume check', (
    tester,
  ) async {
    tester.view.physicalSize = const Size(1440, 2200);
    tester.view.devicePixelRatio = 1.0;
    addTearDown(tester.view.resetPhysicalSize);
    addTearDown(tester.view.resetDevicePixelRatio);

    await tester.pumpWidget(
      _TestApp(
        child: ExamShellScreen(
          client: ExamApiClient(baseUrl: 'http://127.0.0.1:65535'),
          examToken: 'abc12345',
          deviceFingerprint: 'android:test',
          autoStartRuntime: false,
          initialResumeCheckRequired: true,
          restoredSnapshot: _sampleSnapshot(
            pendingAnswers: const <String, String>{'question-1': 'B'},
          ),
          initialPayload: _sampleLoginPayload(),
        ),
      ),
    );

    await tester.pump();

    expect(find.text('Mode ujian diamankan'), findsOneWidget);
    expect(
      find.text(
        'Aplikasi mendeteksi perpindahan dari mode ujian. Lanjutkan hanya jika pengawas mengizinkan.',
      ),
      findsOneWidget,
    );
    expect(find.text('1 jawaban lokal menunggu sinkron.'), findsOneWidget);
    expect(find.text('Lanjutkan dengan pengecekan'), findsOneWidget);
  });

  testWidgets('exam shell renders secured mode overlay while resuming', (
    tester,
  ) async {
    tester.view.physicalSize = const Size(1440, 2200);
    tester.view.devicePixelRatio = 1.0;
    addTearDown(tester.view.resetPhysicalSize);
    addTearDown(tester.view.resetDevicePixelRatio);

    await tester.pumpWidget(
      _TestApp(
        child: ExamShellScreen(
          client: ExamApiClient(baseUrl: 'http://127.0.0.1:65535'),
          examToken: 'abc12345',
          deviceFingerprint: 'android:test',
          autoStartRuntime: false,
          initialResumeCheckRequired: true,
          initialIsResumingExam: true,
          restoredSnapshot: _sampleSnapshot(
            pendingAnswers: const <String, String>{'question-1': 'B'},
          ),
          initialPayload: _sampleLoginPayload(),
        ),
      ),
    );

    await tester.pump();

    expect(find.text('Mode ujian diamankan'), findsOneWidget);
    expect(
      find.text(
        'Sistem sedang memeriksa ulang status peserta dan mencoba menyinkronkan jawaban lokal.',
      ),
      findsOneWidget,
    );
    expect(find.text('1 jawaban lokal menunggu sinkron.'), findsOneWidget);
    expect(find.text('Memeriksa status...'), findsOneWidget);
  });
}

class _TestApp extends StatelessWidget {
  const _TestApp({required this.child});

  final Widget child;

  @override
  Widget build(BuildContext context) {
    return MaterialApp(home: child);
  }
}

ExamLoginPayload _sampleLoginPayload() {
  return ExamLoginPayload(
    participantId: 'participant-1',
    student: const ExamStudent(nis: '24001', nama: 'Siti Aminah'),
    session: ExamSession(
      id: 'session-1',
      title: 'Matematika Kelas VIII',
      scheduledStart: DateTime.parse('2026-05-01T08:00:00+08:00'),
      scheduledEnd: DateTime.parse('2026-05-01T09:30:00+08:00'),
      durationMinutes: 90,
    ),
    room: const ExamRoom(roomName: 'Lab 1'),
    questions: const [
      ExamQuestion(
        id: 'question-1',
        questionText: '2 + 2 = ...',
        stemHtml: '',
        stimulusHtml: '',
        stemMediaUrl: '',
        stimulusMediaUrl: '',
        stemAudioUrl: '',
        stimulusAudioUrl: '',
        options: [
          ExamOption(label: 'A', text: '3'),
          ExamOption(label: 'B', text: '4'),
          ExamOption(label: 'C', text: '5'),
          ExamOption(label: 'D', text: '6'),
        ],
      ),
    ],
    answeredCount: 0,
    totalQuestions: 1,
    timeRemainingSeconds: 1800,
  );
}

ExamLoginPayload _sampleAudioLoginPayload() {
  return ExamLoginPayload(
    participantId: 'participant-audio-1',
    student: const ExamStudent(nis: '24001', nama: 'Siti Aminah'),
    session: ExamSession(
      id: 'session-1',
      title: 'Bahasa Indonesia Kelas VIII',
      scheduledStart: DateTime.parse('2026-05-01T08:00:00+08:00'),
      scheduledEnd: DateTime.parse('2026-05-01T09:30:00+08:00'),
      durationMinutes: 90,
    ),
    room: const ExamRoom(roomName: 'Lab 1'),
    questions: const [
      ExamQuestion(
        id: 'question-audio-1',
        questionText: 'Dengarkan audio berikut lalu pilih jawaban yang benar.',
        stemHtml: '',
        stimulusHtml: '',
        stemMediaUrl: '',
        stimulusMediaUrl: '',
        stemAudioUrl: 'https://cdn.example.com/audio/question-1.mp3',
        stimulusAudioUrl: '',
        options: [
          ExamOption(label: 'A', text: 'Pilihan A'),
          ExamOption(label: 'B', text: 'Pilihan B'),
          ExamOption(label: 'C', text: 'Pilihan C'),
          ExamOption(label: 'D', text: 'Pilihan D'),
        ],
      ),
    ],
    answeredCount: 0,
    totalQuestions: 1,
    timeRemainingSeconds: 1800,
  );
}

ExamLoginPayload _sampleMediaLoginPayload() {
  return ExamLoginPayload(
    participantId: 'participant-media-1',
    student: const ExamStudent(nis: '24001', nama: 'Siti Aminah'),
    session: ExamSession(
      id: 'session-1',
      title: 'IPA Kelas VIII',
      scheduledStart: DateTime.parse('2026-05-01T08:00:00+08:00'),
      scheduledEnd: DateTime.parse('2026-05-01T09:30:00+08:00'),
      durationMinutes: 90,
    ),
    room: const ExamRoom(roomName: 'Lab 1'),
    questions: const [
      ExamQuestion(
        id: 'question-media-1',
        questionText: 'Perhatikan gambar berikut.',
        stemHtml: '',
        stimulusHtml: '',
        stemMediaUrl: 'https://cdn.example.com/images/question-1.png',
        stimulusMediaUrl: '',
        stemAudioUrl: '',
        stimulusAudioUrl: '',
        options: [
          ExamOption(label: 'A', text: 'Pilihan A'),
          ExamOption(label: 'B', text: 'Pilihan B'),
          ExamOption(label: 'C', text: 'Pilihan C'),
          ExamOption(label: 'D', text: 'Pilihan D'),
        ],
      ),
    ],
    answeredCount: 0,
    totalQuestions: 1,
    timeRemainingSeconds: 1800,
  );
}

ExamSessionSnapshot _sampleSnapshot({
  String lastServerContactIso = '',
  String lastSyncFailureIso = '',
  int consecutiveSyncFailures = 0,
  Map<String, String> pendingAnswers = const <String, String>{},
  List<String> playedAudioQuestionIds = const <String>[],
}) {
  return ExamSessionSnapshot(
    baseUrl: 'http://10.0.2.2:8080',
    examToken: 'abc12345',
    deviceFingerprint: 'android:test',
    studentName: 'Siti Aminah',
    studentNis: '24001',
    sessionTitle: 'Matematika Kelas VIII',
    roomName: 'Lab 1',
    scheduledStartIso: '2026-05-01T08:00:00+08:00',
    scheduledEndIso: '2026-05-01T09:30:00+08:00',
    durationMinutes: 90,
    currentQuestionIndex: 0,
    answers: const <String, String>{},
    pendingAnswers: pendingAnswers,
    playedAudioQuestionIds: playedAudioQuestionIds,
    lastServerContactIso: lastServerContactIso,
    lastSyncFailureIso: lastSyncFailureIso,
    consecutiveSyncFailures: consecutiveSyncFailures,
  );
}
