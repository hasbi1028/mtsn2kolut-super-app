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
import 'package:mobile/src/screens/exam_completed_screen.dart';
import 'package:mobile/src/screens/exam_status_guide_screen.dart';

void main() {
  testWidgets('login screen renders exam shell entry', (tester) async {
    await tester.pumpWidget(const MtsnMobileApp());

    expect(find.text('Masuk Ujian'), findsWidgets);
    expect(find.text('Token ujian'), findsOneWidget);
    expect(find.text('Pengaturan Operator'), findsOneWidget);
    expect(find.text('Alamat server API'), findsNothing);
  });

  test('essay answer cap stays below backend serialized 64 KiB budget', () {
    final maxVisibleAnswer = List.filled(kEssayAnswerMaxChars, 'a').join();
    expect(
      ExamApiClient.isAnswerBodyWithinLimit(
        questionId: 'question-short-1',
        answer: maxVisibleAnswer,
      ),
      isTrue,
    );

    // Character-count limits alone are not enough: escaped/control-heavy input
    // can exceed the exact serialized JSON body budget while still being below
    // the visible character cap. The API helper must catch that before submit.
    final escapedHeavyAnswer = List.filled(
      kEssayAnswerMaxChars - 1,
      '\u0000',
    ).join();
    expect(escapedHeavyAnswer.length, lessThanOrEqualTo(kEssayAnswerMaxChars));
    expect(
      ExamApiClient.isAnswerBodyWithinLimit(
        questionId: 'question-short-1',
        answer: escapedHeavyAnswer,
      ),
      isFalse,
    );
  });

  testWidgets('login token field accepts 32-character backend hex token', (
    tester,
  ) async {
    const token = '0123456789abcdef0123456789abcdef';

    await tester.pumpWidget(
      const _TestApp(child: ExamLoginScreen(autoRestore: false)),
    );
    await tester.pump();

    await tester.enterText(
      find.widgetWithText(TextField, 'Token ujian'),
      token,
    );
    await tester.pump();

    final editable = tester.widget<EditableText>(
      find.byType(EditableText).first,
    );
    expect(editable.controller.text, token);
    expect(
      find.text('Contoh: 32 karakter heksadesimal dari kartu ujian'),
      findsOneWidget,
    );
  });

  testWidgets('login screen reveals operator server field on demand', (
    tester,
  ) async {
    tester.view.physicalSize = const Size(1440, 2200);
    tester.view.devicePixelRatio = 1.0;
    addTearDown(tester.view.resetPhysicalSize);
    addTearDown(tester.view.resetDevicePixelRatio);

    await tester.pumpWidget(
      const _TestApp(child: ExamLoginScreen(autoRestore: false)),
    );

    await tester.pump();
    expect(find.text('Alamat server API'), findsNothing);

    await tester.ensureVisible(find.text('Tampilkan'));
    await tester.tap(find.text('Tampilkan'));
    await tester.pumpAndSettle();

    expect(find.text('Alamat server API'), findsOneWidget);
    expect(
      find.text('HTTP hanya untuk uji lokal atau jaringan privat'),
      findsOneWidget,
    );
    expect(
      find.textContaining('Jangan pakai HTTP untuk ujian produksi'),
      findsOneWidget,
    );
  });

  testWidgets('login screen shows HTTPS guidance only inside operator panel', (
    tester,
  ) async {
    tester.view.physicalSize = const Size(1440, 2200);
    tester.view.devicePixelRatio = 1.0;
    addTearDown(tester.view.resetPhysicalSize);
    addTearDown(tester.view.resetDevicePixelRatio);

    await tester.pumpWidget(
      const _TestApp(child: ExamLoginScreen(autoRestore: false)),
    );

    await tester.pump();
    expect(find.text('HTTPS siap untuk server produksi'), findsNothing);

    await tester.ensureVisible(find.text('Tampilkan'));
    await tester.tap(find.text('Tampilkan'));
    await tester.pumpAndSettle();
    await tester.enterText(
      find.widgetWithText(TextField, 'Alamat server API'),
      'https://cbt.mtsn2kolut.sch.id',
    );
    await tester.pump();

    expect(find.text('HTTPS siap untuk server produksi'), findsOneWidget);
    expect(find.textContaining('sesuai untuk sesi produksi'), findsOneWidget);
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

  testWidgets('login screen renders transport guidance notice', (tester) async {
    await tester.pumpWidget(
      const _TestApp(
        child: ExamLoginScreen(
          autoRestore: false,
          initialErrorMessage:
              'Perangkat belum bisa terhubung ke server ujian. Periksa alamat server dan koneksi yang sedang dipakai.',
          initialErrorNotice: ExamGuidanceNotice(
            title: 'Server ujian belum terjangkau',
            message:
                'Peserta tidak perlu terus menekan login. Periksa koneksi perangkat atau alamat server, lalu coba lagi setelah pengawas memastikan jaringan siap.',
            tone: ExamGuidanceTone.warning,
          ),
        ),
      ),
    );

    await tester.pump();

    expect(find.text('Server ujian belum terjangkau'), findsOneWidget);
    expect(
      find.text(
        'Peserta tidak perlu terus menekan login. Periksa koneksi perangkat atau alamat server, lalu coba lagi setelah pengawas memastikan jaringan siap.',
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

  testWidgets('login screen renders stable restore health label', (
    tester,
  ) async {
    await tester.pumpWidget(
      _TestApp(
        child: ExamLoginScreen(
          autoRestore: false,
          previewSnapshot: _sampleSnapshot(
            lastServerContactIso: '2026-05-01T08:44:00+08:00',
          ),
        ),
      ),
    );

    await tester.pump();

    expect(find.text('Terakhir stabil'), findsOneWidget);
    expect(find.textContaining('Kontak server 08:44'), findsOneWidget);
    expect(find.textContaining('Gangguan '), findsNothing);
  });

  testWidgets('login screen renders disturbed restore health label', (
    tester,
  ) async {
    await tester.pumpWidget(
      _TestApp(
        child: ExamLoginScreen(
          autoRestore: false,
          previewSnapshot: _sampleSnapshot(
            lastSyncFailureIso: '2026-05-01T08:46:00+08:00',
          ),
        ),
      ),
    );

    await tester.pump();

    expect(find.text('Pernah terganggu'), findsOneWidget);
    expect(find.textContaining('Kontak server '), findsNothing);
    expect(find.textContaining('Gangguan 08:46'), findsOneWidget);
  });

  testWidgets('login screen renders empty restore health label', (
    tester,
  ) async {
    await tester.pumpWidget(
      _TestApp(
        child: ExamLoginScreen(
          autoRestore: false,
          previewSnapshot: _sampleSnapshot(),
        ),
      ),
    );

    await tester.pump();

    expect(find.text('Belum ada riwayat koneksi'), findsOneWidget);
    expect(find.textContaining('Kontak server '), findsNothing);
    expect(find.textContaining('Gangguan '), findsNothing);
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

  testWidgets('restore failed screen renders transport guidance notice', (
    tester,
  ) async {
    tester.view.physicalSize = const Size(1280, 2200);
    tester.view.devicePixelRatio = 1.0;
    addTearDown(tester.view.resetPhysicalSize);
    addTearDown(tester.view.resetDevicePixelRatio);

    await tester.pumpWidget(
      _TestApp(
        child: ExamRestoreFailedScreen(
          snapshot: _sampleSnapshot(),
          message:
              'Sesi lama belum bisa dipulihkan karena perangkat belum terhubung ke server ujian. Coba lagi setelah koneksi membaik atau hubungi pengawas.',
          notice: const ExamGuidanceNotice(
            title: 'Restore tertunda karena koneksi',
            message:
                'Pengawas perlu memastikan perangkat sudah kembali terhubung ke server sebelum peserta mencoba memulihkan sesi lama lagi.',
            tone: ExamGuidanceTone.warning,
          ),
        ),
      ),
    );

    await tester.pump();

    expect(find.text('Restore tertunda karena koneksi'), findsOneWidget);
    expect(
      find.text(
        'Pengawas perlu memastikan perangkat sudah kembali terhubung ke server sebelum peserta mencoba memulihkan sesi lama lagi.',
      ),
      findsOneWidget,
    );
  });

  testWidgets('restore failed screen renders local answer counts', (
    tester,
  ) async {
    tester.view.physicalSize = const Size(1280, 2200);
    tester.view.devicePixelRatio = 1.0;
    addTearDown(tester.view.resetPhysicalSize);
    addTearDown(tester.view.resetDevicePixelRatio);

    await tester.pumpWidget(
      _TestApp(
        child: ExamRestoreFailedScreen(
          snapshot: _sampleSnapshot(
            answers: const <String, String>{'q1': 'A', 'q2': 'B'},
            pendingAnswers: const <String, String>{'q2': 'B'},
          ),
          message: 'Sesi lama belum bisa dipulihkan.',
        ),
      ),
    );

    await tester.pump();

    expect(find.text('2 jawaban lokal'), findsOneWidget);
    expect(find.text('1 belum tersinkron'), findsOneWidget);
  });

  testWidgets('restore failed screen renders stable restore health label', (
    tester,
  ) async {
    tester.view.physicalSize = const Size(1280, 2200);
    tester.view.devicePixelRatio = 1.0;
    addTearDown(tester.view.resetPhysicalSize);
    addTearDown(tester.view.resetDevicePixelRatio);

    await tester.pumpWidget(
      _TestApp(
        child: ExamRestoreFailedScreen(
          snapshot: _sampleSnapshot(
            lastServerContactIso: '2026-05-01T08:44:00+08:00',
          ),
          message: 'Sesi lama belum bisa dipulihkan.',
        ),
      ),
    );

    await tester.pump();

    expect(find.text('Terakhir stabil'), findsOneWidget);
    expect(find.textContaining('Kontak server 08:44'), findsOneWidget);
    expect(find.textContaining('Gangguan '), findsNothing);
  });

  testWidgets('restore failed screen renders disturbed restore health label', (
    tester,
  ) async {
    tester.view.physicalSize = const Size(1280, 2200);
    tester.view.devicePixelRatio = 1.0;
    addTearDown(tester.view.resetPhysicalSize);
    addTearDown(tester.view.resetDevicePixelRatio);

    await tester.pumpWidget(
      _TestApp(
        child: ExamRestoreFailedScreen(
          snapshot: _sampleSnapshot(
            lastSyncFailureIso: '2026-05-01T08:46:00+08:00',
          ),
          message: 'Sesi lama belum bisa dipulihkan.',
        ),
      ),
    );

    await tester.pump();

    expect(find.text('Pernah terganggu'), findsOneWidget);
    expect(find.textContaining('Kontak server '), findsNothing);
    expect(find.textContaining('Gangguan 08:46'), findsOneWidget);
  });

  testWidgets('restore failed screen renders empty restore health label', (
    tester,
  ) async {
    tester.view.physicalSize = const Size(1280, 2200);
    tester.view.devicePixelRatio = 1.0;
    addTearDown(tester.view.resetPhysicalSize);
    addTearDown(tester.view.resetDevicePixelRatio);

    await tester.pumpWidget(
      _TestApp(
        child: ExamRestoreFailedScreen(
          snapshot: _sampleSnapshot(),
          message: 'Sesi lama belum bisa dipulihkan.',
        ),
      ),
    );

    await tester.pump();

    expect(find.text('Belum ada riwayat koneksi'), findsOneWidget);
    expect(find.textContaining('Kontak server '), findsNothing);
    expect(find.textContaining('Gangguan '), findsNothing);
  });

  testWidgets('completed screen renders manual submit summary', (tester) async {
    tester.view.physicalSize = const Size(1440, 2600);
    tester.view.devicePixelRatio = 1.0;
    addTearDown(tester.view.resetPhysicalSize);
    addTearDown(tester.view.resetDevicePixelRatio);

    await tester.pumpWidget(
      const _TestApp(
        child: ExamCompletedScreen(
          studentName: 'Siti Aminah',
          studentNis: '24001',
          sessionTitle: 'Matematika Kelas VIII',
          roomName: 'Lab 1',
          scheduledStartIso: '2026-05-01T08:00:00+08:00',
          scheduledEndIso: '2026-05-01T09:30:00+08:00',
          durationMinutes: 90,
          answeredCount: 18,
          totalQuestions: 20,
          wasAutoSubmitted: false,
        ),
      ),
    );

    await tester.pump();

    expect(find.text('Ujian berhasil dikirim.'), findsOneWidget);
    expect(
      find.text(
        'Jawaban Anda sudah diterima server. Silakan menunggu arahan pengawas.',
      ),
      findsOneWidget,
    );
    expect(find.text('18 / 20 soal'), findsOneWidget);
    expect(find.text('Matematika Kelas VIII'), findsOneWidget);
  });

  testWidgets('completed screen renders auto submit summary', (tester) async {
    tester.view.physicalSize = const Size(1440, 2600);
    tester.view.devicePixelRatio = 1.0;
    addTearDown(tester.view.resetPhysicalSize);
    addTearDown(tester.view.resetDevicePixelRatio);

    await tester.pumpWidget(
      const _TestApp(
        child: ExamCompletedScreen(
          studentName: 'Siti Aminah',
          studentNis: '24001',
          sessionTitle: 'Bahasa Indonesia Kelas VIII',
          roomName: 'Lab 2',
          scheduledStartIso: '2026-05-01T10:00:00+08:00',
          scheduledEndIso: '2026-05-01T11:30:00+08:00',
          durationMinutes: 90,
          answeredCount: 20,
          totalQuestions: 20,
          wasAutoSubmitted: true,
        ),
      ),
    );

    await tester.pump();

    expect(find.text('Ujian ditutup otomatis.'), findsOneWidget);
    expect(
      find.text(
        'Waktu ujian telah habis dan jawaban Anda sudah dikirim ke server.',
      ),
      findsOneWidget,
    );
    expect(find.text('20 / 20 soal'), findsOneWidget);
    expect(find.text('Bahasa Indonesia Kelas VIII'), findsOneWidget);
  });

  testWidgets('status guide screen renders all byod status cards', (
    tester,
  ) async {
    tester.view.physicalSize = const Size(1440, 2600);
    tester.view.devicePixelRatio = 1.0;
    addTearDown(tester.view.resetPhysicalSize);
    addTearDown(tester.view.resetDevicePixelRatio);

    await tester.pumpWidget(const _TestApp(child: ExamStatusGuideScreen()));

    await tester.pump();

    expect(find.text('Panduan Status Ujian'), findsOneWidget);
    expect(find.text('Tersambung'), findsOneWidget);
    expect(find.text('Lokal'), findsOneWidget);
    expect(find.text('Gangguan'), findsOneWidget);
    await tester.scrollUntilVisible(find.text('Menurun'), 200);
    expect(find.text('Menurun'), findsOneWidget);
    expect(find.text('Catatan untuk pengawas'), findsOneWidget);
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

  testWidgets('exam shell renders transport guidance notice', (tester) async {
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
            title: 'Jawaban tersimpan lokal',
            message:
                'Perangkat belum bisa menjangkau server, tetapi jawaban peserta masih aman di perangkat ini. Pengawas perlu membantu memulihkan koneksi sebelum sinkron ulang.',
            tone: ExamGuidanceTone.warning,
          ),
          initialPayload: _sampleLoginPayload(),
        ),
      ),
    );

    await tester.pump();

    expect(find.text('Jawaban tersimpan lokal'), findsOneWidget);
    expect(
      find.text(
        'Perangkat belum bisa menjangkau server, tetapi jawaban peserta masih aman di perangkat ini. Pengawas perlu membantu memulihkan koneksi sebelum sinkron ulang.',
      ),
      findsOneWidget,
    );
  });

  testWidgets('exam shell still blocks back when telemetry fails', (
    tester,
  ) async {
    final client = _RecordingExamApiClient(
      baseUrl: 'http://127.0.0.1:65535',
      throwOnEvent: true,
    );
    await tester.pumpWidget(
      _TestApp(
        child: ExamShellScreen(
          client: client,
          examToken: 'abc12345',
          deviceFingerprint: 'android:test',
          autoStartRuntime: false,
          initialPayload: _sampleLoginPayload(),
        ),
      ),
    );

    await tester.pump();
    await tester.binding.handlePopRoute();
    await tester.pumpAndSettle();

    expect(client.eventCount, 1);
    expect(
      find.text('Tombol kembali dinonaktifkan selama ujian berlangsung.'),
      findsOneWidget,
    );
    expect(find.text('Soal 1'), findsOneWidget);
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

  testWidgets('exam shell shows sync chip tersambung for healthy state', (
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
          initialPayload: _sampleLoginPayload(),
        ),
      ),
    );

    await tester.pump();

    expect(find.text('Tersambung'), findsWidgets);
  });

  testWidgets('exam shell shows sync chip lokal when pending answers exist', (
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
            pendingAnswers: const <String, String>{'question-1': 'B'},
          ),
          initialPayload: _sampleLoginPayload(),
        ),
      ),
    );

    await tester.pump();

    expect(find.text('Lokal'), findsWidgets);
  });

  testWidgets('exam shell shows sync chip waspada for stale contact', (
    tester,
  ) async {
    tester.view.physicalSize = const Size(1440, 2200);
    tester.view.devicePixelRatio = 1.0;
    addTearDown(tester.view.resetPhysicalSize);
    addTearDown(tester.view.resetDevicePixelRatio);

    final staleContact = DateTime.now()
        .subtract(const Duration(minutes: 3))
        .toIso8601String();

    await tester.pumpWidget(
      _TestApp(
        child: ExamShellScreen(
          client: ExamApiClient(baseUrl: 'http://127.0.0.1:65535'),
          examToken: 'abc12345',
          deviceFingerprint: 'android:test',
          autoStartRuntime: false,
          restoredSnapshot: _sampleSnapshot(lastServerContactIso: staleContact),
          initialPayload: _sampleLoginPayload(),
        ),
      ),
    );

    await tester.pump();

    expect(find.text('Waspada'), findsWidgets);
  });

  testWidgets('exam shell shows sync chip menurun for degraded mode', (
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
            lastServerContactIso: now.toIso8601String(),
            lastSyncFailureIso: now.toIso8601String(),
            consecutiveSyncFailures: 3,
          ),
          initialPayload: _sampleLoginPayload(),
        ),
      ),
    );

    await tester.pump();

    expect(find.text('Menurun'), findsWidgets);
  });

  testWidgets('exam shell shows sync chip sinkron while syncing status', (
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
          initialIsSyncingStatus: true,
          initialPayload: _sampleLoginPayload(),
        ),
      ),
    );

    await tester.pump();

    expect(find.text('Sinkron'), findsOneWidget);
  });

  testWidgets('exam shell shows sync chip cek ulang during resume gate', (
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
          initialPayload: _sampleLoginPayload(),
        ),
      ),
    );

    await tester.pump();

    expect(find.text('Cek Ulang'), findsOneWidget);
  });

  testWidgets('exam shell shows sync chip gangguan for generic error state', (
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
          initialErrorMessage: 'Status server belum bisa diperbarui.',
          initialPayload: _sampleLoginPayload(),
        ),
      ),
    );

    await tester.pump();

    expect(find.text('Gangguan'), findsWidgets);
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

  testWidgets('exam shell resolves relative media and audio URLs', (
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
          initialPayload: _sampleRelativeMediaLoginPayload(),
        ),
      ),
    );

    await tester.pump();

    final image = tester.widget<Image>(find.byType(Image).first);
    final provider = image.image as NetworkImage;
    expect(provider.url, 'http://127.0.0.1:65535/uploads/questions/q1.png');
    expect(
      find.text('http://127.0.0.1:65535/uploads/audio/q1.mp3'),
      findsOneWidget,
    );
  });

  testWidgets('exam shell renders rich stimulus and stem content', (
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
          initialPayload: _sampleRichContentLoginPayload(),
        ),
      ),
    );

    await tester.pump();

    expect(find.text('Bacalah teks berikut dengan saksama.'), findsOneWidget);
    expect(find.textContaining('Kalimat pertama'), findsOneWidget);
    expect(find.textContaining('Kalimat kedua'), findsOneWidget);
    expect(
      find.textContaining('Apa gagasan utama paragraf di atas?'),
      findsOneWidget,
    );
  });

  testWidgets('exam shell renders multiple answer as checkbox options', (
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
          initialPayload: _sampleMultipleAnswerPayload(),
        ),
      ),
    );

    await tester.pump();

    expect(find.byType(Checkbox), findsNWidgets(4));
    expect(find.text('Pilihan A'), findsOneWidget);
  });

  testWidgets('exam shell renders true/false fallback choices', (tester) async {
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
          initialPayload: _sampleTrueFalsePayload(),
        ),
      ),
    );

    await tester.pump();

    expect(find.text('Benar'), findsOneWidget);
    expect(find.text('Salah'), findsOneWidget);
    expect(find.text('true'), findsOneWidget);
    expect(find.text('false'), findsOneWidget);
    expect(find.byType(Checkbox), findsNothing);
  });

  testWidgets('exam shell saves true/false fallback answer', (tester) async {
    tester.view.physicalSize = const Size(1440, 2200);
    tester.view.devicePixelRatio = 1.0;
    addTearDown(tester.view.resetPhysicalSize);
    addTearDown(tester.view.resetDevicePixelRatio);
    final client = _RecordingExamApiClient(baseUrl: 'http://127.0.0.1:65535');

    await tester.pumpWidget(
      _TestApp(
        child: ExamShellScreen(
          client: client,
          examToken: 'abc12345',
          deviceFingerprint: 'android:test',
          autoStartRuntime: false,
          initialPayload: _sampleTrueFalsePayload(),
        ),
      ),
    );
    await tester.pump();

    expect(find.text('0 / 1'), findsOneWidget);

    await tester.tap(find.text('Benar'));
    await tester.pump();
    await tester.pump(const Duration(milliseconds: 100));

    expect(client.saveAnswerCount, 1);
    expect(client.lastSavedQuestionId, 'question-true-false-1');
    expect(client.lastSavedAnswer, 'true');
    expect(find.text('1 / 1'), findsOneWidget);
  });

  testWidgets('exam shell restores true/false saved answer', (tester) async {
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
            answers: const <String, String>{'question-true-false-1': 'false'},
          ),
          initialPayload: _sampleTrueFalsePayload(),
        ),
      ),
    );

    await tester.pump();

    expect(find.text('Benar'), findsOneWidget);
    expect(find.text('Salah'), findsOneWidget);
    expect(find.text('1 / 1'), findsOneWidget);
  });

  testWidgets('exam shell keeps backend true/false options compatible', (
    tester,
  ) async {
    tester.view.physicalSize = const Size(1440, 2200);
    tester.view.devicePixelRatio = 1.0;
    addTearDown(tester.view.resetPhysicalSize);
    addTearDown(tester.view.resetDevicePixelRatio);
    final client = _RecordingExamApiClient(baseUrl: 'http://127.0.0.1:65535');

    await tester.pumpWidget(
      _TestApp(
        child: ExamShellScreen(
          client: client,
          examToken: 'abc12345',
          deviceFingerprint: 'android:test',
          autoStartRuntime: false,
          initialPayload: _sampleTrueFalseWithBackendOptionsPayload(),
        ),
      ),
    );
    await tester.pump();

    expect(find.text('true'), findsNothing);
    expect(find.text('false'), findsNothing);
    expect(find.text('Benar dari naskah'), findsOneWidget);
    expect(find.text('Salah dari naskah'), findsOneWidget);

    await tester.tap(find.text('Salah dari naskah'));
    await tester.pump();
    await tester.pump(const Duration(milliseconds: 100));

    expect(client.saveAnswerCount, 1);
    expect(client.lastSavedQuestionId, 'question-true-false-2');
    expect(client.lastSavedAnswer, 'B');
  });

  testWidgets('exam shell renders short answer as compact text input', (
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
          initialPayload: _sampleShortAnswerPayload(),
        ),
      ),
    );

    await tester.pump();

    expect(find.text('Jawaban singkat'), findsOneWidget);
    expect(find.text('Jawaban uraian'), findsNothing);
  });

  testWidgets('exam shell autosaves text answers locally before server save', (
    tester,
  ) async {
    tester.view.physicalSize = const Size(1440, 2200);
    tester.view.devicePixelRatio = 1.0;
    addTearDown(tester.view.resetPhysicalSize);
    addTearDown(tester.view.resetDevicePixelRatio);
    final store = _MemoryExamSessionStore();
    final client = _RecordingExamApiClient(baseUrl: 'http://127.0.0.1:65535');

    await tester.pumpWidget(
      _TestApp(
        child: ExamShellScreen(
          client: client,
          examToken: 'abc12345',
          deviceFingerprint: 'android:test',
          autoStartRuntime: false,
          sessionStore: store,
          initialPayload: _sampleShortAnswerPayload(),
        ),
      ),
    );
    await tester.pump();

    await tester.enterText(
      find.widgetWithText(TextField, 'Jawaban singkat'),
      'Fotosintesis',
    );
    await tester.pump(const Duration(milliseconds: 450));

    final snapshot = await store.loadSnapshot();
    expect(snapshot?.answers, containsPair('question-short-1', 'Fotosintesis'));
    expect(snapshot?.pendingAnswers, isEmpty);
    expect(client.saveAnswerCount, 0);
    expect(find.text('1 / 1'), findsOneWidget);
  });

  testWidgets(
    'exam shell does not queue answer after already-submitted conflict',
    (tester) async {
      tester.view.physicalSize = const Size(1440, 2200);
      tester.view.devicePixelRatio = 1.0;
      addTearDown(tester.view.resetPhysicalSize);
      addTearDown(tester.view.resetDevicePixelRatio);
      final store = _MemoryExamSessionStore();
      final client = _RecordingExamApiClient(
        baseUrl: 'http://127.0.0.1:65535',
        saveAnswerError: const ExamApiException(
          'already submitted',
          statusCode: 409,
        ),
        statusPayload: const ExamStatusPayload(
          answeredCount: 1,
          totalQuestions: 1,
          timeRemainingSeconds: 0,
          isSubmitted: true,
        ),
      );

      await tester.pumpWidget(
        _TestApp(
          child: ExamShellScreen(
            client: client,
            examToken: 'abc12345',
            deviceFingerprint: 'android:test',
            autoStartRuntime: false,
            sessionStore: store,
            initialPayload: _sampleLoginPayload(),
          ),
        ),
      );
      await tester.pump();

      await tester.tap(find.text('4'));
      await tester.pumpAndSettle();

      expect(client.saveAnswerCount, 1);
      expect(find.text('Ujian berhasil dikirim.'), findsOneWidget);
      expect(await store.loadSnapshot(), isNull);
    },
  );

  testWidgets('exam shell preserves local answer on device-mismatch conflict', (
    tester,
  ) async {
    tester.view.physicalSize = const Size(1440, 2200);
    tester.view.devicePixelRatio = 1.0;
    addTearDown(tester.view.resetPhysicalSize);
    addTearDown(tester.view.resetDevicePixelRatio);
    final store = _MemoryExamSessionStore();
    final client = _RecordingExamApiClient(
      baseUrl: 'http://127.0.0.1:65535',
      saveAnswerError: const ExamApiException(
        'token already bound to another device',
        statusCode: 409,
      ),
    );

    await tester.pumpWidget(
      _TestApp(
        child: ExamShellScreen(
          client: client,
          examToken: 'abc12345',
          deviceFingerprint: 'android:test',
          autoStartRuntime: false,
          sessionStore: store,
          initialPayload: _sampleLoginPayload(),
        ),
      ),
    );
    await tester.pump();

    await tester.tap(find.text('4'));
    await tester.pumpAndSettle();

    final snapshot = await store.loadSnapshot();
    expect(client.saveAnswerCount, 1);
    expect(find.text('Ujian berhasil dikirim.'), findsNothing);
    expect(snapshot?.answers, containsPair('question-1', 'B'));
    expect(snapshot?.pendingAnswers, containsPair('question-1', 'B'));
  });

  testWidgets(
    'exam shell treats submitted status as terminal and clears snapshot',
    (tester) async {
      tester.view.physicalSize = const Size(1440, 2200);
      tester.view.devicePixelRatio = 1.0;
      addTearDown(tester.view.resetPhysicalSize);
      addTearDown(tester.view.resetDevicePixelRatio);
      final store = _MemoryExamSessionStore(
        initialSnapshot: _sampleSnapshot(
          answers: const <String, String>{'question-1': 'B'},
        ),
      );
      final client = _RecordingExamApiClient(
        baseUrl: 'http://127.0.0.1:65535',
        statusPayload: const ExamStatusPayload(
          answeredCount: 1,
          totalQuestions: 1,
          timeRemainingSeconds: 0,
          isSubmitted: true,
        ),
      );

      await tester.pumpWidget(
        _TestApp(
          child: ExamShellScreen(
            client: client,
            examToken: 'abc12345',
            deviceFingerprint: 'android:test',
            sessionStore: store,
            restoredSnapshot: _sampleSnapshot(
              answers: const <String, String>{'question-1': 'B'},
            ),
            initialPayload: _sampleLoginPayload(),
          ),
        ),
      );
      await tester.pumpAndSettle();

      expect(find.text('Ujian berhasil dikirim.'), findsOneWidget);
      expect(await store.loadSnapshot(), isNull);
    },
  );

  testWidgets('exam shell treats submit 409 as terminal and clears snapshot', (
    tester,
  ) async {
    tester.view.physicalSize = const Size(1440, 2200);
    tester.view.devicePixelRatio = 1.0;
    addTearDown(tester.view.resetPhysicalSize);
    addTearDown(tester.view.resetDevicePixelRatio);
    final store = _MemoryExamSessionStore(
      initialSnapshot: _sampleSnapshot(
        answers: const <String, String>{'question-1': 'B'},
      ),
    );
    final client = _RecordingExamApiClient(
      baseUrl: 'http://127.0.0.1:65535',
      submitError: const ExamApiException('already submitted', statusCode: 409),
    );

    await tester.pumpWidget(
      _TestApp(
        child: ExamShellScreen(
          client: client,
          examToken: 'abc12345',
          deviceFingerprint: 'android:test',
          autoStartRuntime: false,
          sessionStore: store,
          restoredSnapshot: _sampleSnapshot(
            answers: const <String, String>{'question-1': 'B'},
          ),
          initialPayload: _sampleLoginPayload(),
        ),
      ),
    );
    await tester.pump();

    await tester.tap(find.text('Kirim Ujian'));
    await tester.pumpAndSettle();
    await tester.tap(find.text('Kirim Ujian').last);
    await tester.pumpAndSettle();

    expect(find.text('Ujian berhasil dikirim.'), findsOneWidget);
    expect(await store.loadSnapshot(), isNull);
  });

  testWidgets('exam shell recomputes progress after pending answer flush', (
    tester,
  ) async {
    tester.view.physicalSize = const Size(1440, 2200);
    tester.view.devicePixelRatio = 1.0;
    addTearDown(tester.view.resetPhysicalSize);
    addTearDown(tester.view.resetDevicePixelRatio);
    final store = _MemoryExamSessionStore(
      initialSnapshot: _sampleSnapshot(
        answers: const <String, String>{'question-1': 'B'},
        pendingAnswers: const <String, String>{'question-1': 'B'},
      ),
    );
    final client = _RecordingExamApiClient(
      baseUrl: 'http://127.0.0.1:65535',
      statusPayload: const ExamStatusPayload(
        answeredCount: 0,
        totalQuestions: 1,
        timeRemainingSeconds: 1800,
        isSubmitted: false,
      ),
    );

    await tester.pumpWidget(
      _TestApp(
        child: ExamShellScreen(
          client: client,
          examToken: 'abc12345',
          deviceFingerprint: 'android:test',
          sessionStore: store,
          restoredSnapshot: _sampleSnapshot(
            answers: const <String, String>{'question-1': 'B'},
            pendingAnswers: const <String, String>{'question-1': 'B'},
          ),
          initialPayload: _sampleLoginPayload(),
        ),
      ),
    );
    await tester.pumpAndSettle();

    expect(client.saveAnswerCount, 1);
    expect(find.text('1 / 1'), findsOneWidget);
    expect((await store.loadSnapshot())?.pendingAnswers, isEmpty);
  });

  testWidgets('exam shell renders matching question with pair selectors', (
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
          initialPayload: _sampleMatchingPayload(),
        ),
      ),
    );

    await tester.pump();

    expect(find.text('Pilihan pasangan'), findsOneWidget);
    expect(find.byType(DropdownButtonFormField<String>), findsNWidgets(2));
    expect(find.text('Fotosintesis'), findsOneWidget);
    expect(find.text('Proses membuat makanan'), findsOneWidget);
    expect(find.text('Distraktor kanan'), findsOneWidget);
  });

  testWidgets('exam shell renders ordering question with reorder controls', (
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
          initialPayload: _sampleOrderingPayload(),
        ),
      ),
    );

    await tester.pump();

    expect(find.text('Urutan jawaban'), findsOneWidget);
    expect(find.text('Langkah pertama'), findsOneWidget);
    expect(find.text('Langkah kedua'), findsOneWidget);
    expect(find.text('Langkah ketiga'), findsOneWidget);
    expect(find.byTooltip('Turunkan A'), findsOneWidget);
    expect(find.byTooltip('Naikkan B'), findsOneWidget);
  });

  testWidgets('exam shell sends ordering answer as comma-separated labels', (
    tester,
  ) async {
    tester.view.physicalSize = const Size(1440, 2200);
    tester.view.devicePixelRatio = 1.0;
    addTearDown(tester.view.resetPhysicalSize);
    addTearDown(tester.view.resetDevicePixelRatio);
    final client = _RecordingExamApiClient(baseUrl: 'http://127.0.0.1:65535');

    await tester.pumpWidget(
      _TestApp(
        child: ExamShellScreen(
          client: client,
          examToken: 'abc12345',
          deviceFingerprint: 'android:test',
          autoStartRuntime: false,
          initialPayload: _sampleOrderingPayload(),
        ),
      ),
    );
    await tester.pump();

    expect(find.text('0 / 1'), findsOneWidget);

    await tester.tap(find.byTooltip('Turunkan A'));
    await tester.pump();
    await tester.pump(const Duration(milliseconds: 100));

    expect(client.saveAnswerCount, 1);
    expect(client.lastSavedQuestionId, 'question-ordering-1');
    expect(client.lastSavedAnswer, 'B,A,C');
    expect(find.text('1 / 1'), findsOneWidget);
  });

  testWidgets('exam shell saves current ordering answer without moving option', (
    tester,
  ) async {
    tester.view.physicalSize = const Size(1440, 2200);
    tester.view.devicePixelRatio = 1.0;
    addTearDown(tester.view.resetPhysicalSize);
    addTearDown(tester.view.resetDevicePixelRatio);
    final client = _RecordingExamApiClient(baseUrl: 'http://127.0.0.1:65535');

    await tester.pumpWidget(
      _TestApp(
        child: ExamShellScreen(
          client: client,
          examToken: 'abc12345',
          deviceFingerprint: 'android:test',
          autoStartRuntime: false,
          initialPayload: _sampleOrderingPayload(),
        ),
      ),
    );
    await tester.pump();

    expect(find.text('0 / 1'), findsOneWidget);

    await tester.tap(find.text('Simpan urutan saat ini'));
    await tester.pump();
    await tester.pump(const Duration(milliseconds: 100));

    expect(client.saveAnswerCount, 1);
    expect(client.lastSavedQuestionId, 'question-ordering-1');
    expect(client.lastSavedAnswer, 'A,B,C');
    expect(find.text('1 / 1'), findsOneWidget);
  });

  testWidgets('exam shell restores ordering answer in saved order', (
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
            answers: const <String, String>{'question-ordering-1': 'C,A,B'},
          ),
          initialPayload: _sampleOrderingPayload(),
        ),
      ),
    );

    await tester.pump();

    final thirdTop = tester.getTopLeft(find.text('Langkah ketiga')).dy;
    final firstTop = tester.getTopLeft(find.text('Langkah pertama')).dy;
    final secondTop = tester.getTopLeft(find.text('Langkah kedua')).dy;
    expect(thirdTop, lessThan(firstTop));
    expect(firstTop, lessThan(secondTop));
    expect(find.text('1 / 1'), findsOneWidget);
  });

  testWidgets('exam shell keeps partial ordering answer incomplete', (
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
            answers: const <String, String>{'question-ordering-1': 'B,A'},
          ),
          initialPayload: _sampleOrderingPayload(),
        ),
      ),
    );

    await tester.pump();

    expect(find.text('0 / 1'), findsOneWidget);
  });

  testWidgets('exam shell keeps duplicate ordering answer incomplete', (
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
            answers: const <String, String>{'question-ordering-1': 'A,A,B'},
          ),
          initialPayload: _sampleOrderingPayload(),
        ),
      ),
    );

    await tester.pump();

    expect(find.text('0 / 1'), findsOneWidget);
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

  testWidgets('exam shell keeps resume gate when status refresh fails', (
    tester,
  ) async {
    tester.view.physicalSize = const Size(1440, 2200);
    tester.view.devicePixelRatio = 1.0;
    addTearDown(tester.view.resetPhysicalSize);
    addTearDown(tester.view.resetDevicePixelRatio);

    final client = _RecordingExamApiClient(
      baseUrl: 'http://127.0.0.1:65535',
      statusError: const ExamApiException('transport'),
    );

    await tester.pumpWidget(
      _TestApp(
        child: ExamShellScreen(
          client: client,
          examToken: 'abc12345',
          deviceFingerprint: 'android:test',
          autoStartRuntime: false,
          initialResumeCheckRequired: true,
          initialPayload: _sampleLoginPayload(),
        ),
      ),
    );

    await tester.pump();
    await tester.tap(find.text('Lanjutkan dengan pengecekan'));
    await tester.pumpAndSettle();

    expect(client.statusCount, 1);
    expect(find.text('Mode ujian diamankan'), findsOneWidget);
    expect(find.text('Lanjutkan dengan pengecekan'), findsOneWidget);
    expect(
      find.textContaining('Status ujian belum berhasil dicek ulang'),
      findsOneWidget,
    );
  });

  testWidgets('exam shell treats pending answer flush 409 as terminal', (
    tester,
  ) async {
    tester.view.physicalSize = const Size(1440, 2200);
    tester.view.devicePixelRatio = 1.0;
    addTearDown(tester.view.resetPhysicalSize);
    addTearDown(tester.view.resetDevicePixelRatio);
    final store = _MemoryExamSessionStore(
      initialSnapshot: _sampleSnapshot(
        answers: const <String, String>{'question-1': 'B'},
        pendingAnswers: const <String, String>{'question-1': 'B'},
      ),
    );
    final client = _RecordingExamApiClient(
      baseUrl: 'http://127.0.0.1:65535',
      saveAnswerError: const ExamApiException(
        'already submitted',
        statusCode: 409,
      ),
      statusPayload: const ExamStatusPayload(
        answeredCount: 1,
        totalQuestions: 1,
        timeRemainingSeconds: 0,
        isSubmitted: true,
      ),
    );

    await tester.pumpWidget(
      _TestApp(
        child: ExamShellScreen(
          client: client,
          examToken: 'abc12345',
          deviceFingerprint: 'android:test',
          autoStartRuntime: false,
          sessionStore: store,
          restoredSnapshot: _sampleSnapshot(
            answers: const <String, String>{'question-1': 'B'},
            pendingAnswers: const <String, String>{'question-1': 'B'},
          ),
          initialPayload: _sampleLoginPayload(),
        ),
      ),
    );

    await tester.pump();
    await tester.tap(find.text('Kirim Ujian'));
    await tester.pumpAndSettle();

    expect(client.saveAnswerCount, 1);
    expect(client.statusCount, 1);
    expect(find.text('Ujian berhasil dikirim.'), findsOneWidget);
    expect(await store.loadSnapshot(), isNull);
  });

  testWidgets('exam shell preserves pending flush on device mismatch 409', (
    tester,
  ) async {
    tester.view.physicalSize = const Size(1440, 2200);
    tester.view.devicePixelRatio = 1.0;
    addTearDown(tester.view.resetPhysicalSize);
    addTearDown(tester.view.resetDevicePixelRatio);
    final store = _MemoryExamSessionStore(
      initialSnapshot: _sampleSnapshot(
        answers: const <String, String>{'question-1': 'B'},
        pendingAnswers: const <String, String>{'question-1': 'B'},
      ),
    );
    final client = _RecordingExamApiClient(
      baseUrl: 'http://127.0.0.1:65535',
      saveAnswerError: const ExamApiException(
        'token already bound to another device',
        statusCode: 409,
      ),
    );

    await tester.pumpWidget(
      _TestApp(
        child: ExamShellScreen(
          client: client,
          examToken: 'abc12345',
          deviceFingerprint: 'android:test',
          autoStartRuntime: false,
          sessionStore: store,
          restoredSnapshot: _sampleSnapshot(
            answers: const <String, String>{'question-1': 'B'},
            pendingAnswers: const <String, String>{'question-1': 'B'},
          ),
          initialPayload: _sampleLoginPayload(),
        ),
      ),
    );

    await tester.pump();
    await tester.tap(find.text('Kirim Ujian'));
    await tester.pumpAndSettle();

    final snapshot = await store.loadSnapshot();
    expect(client.saveAnswerCount, 1);
    expect(find.text('Ujian berhasil dikirim.'), findsNothing);
    expect(snapshot?.pendingAnswers, containsPair('question-1', 'B'));
    expect(
      find.textContaining('Masih ada jawaban yang belum tersinkron'),
      findsWidgets,
    );
  });

  testWidgets('exam shell blocks empty question payload safely', (
    tester,
  ) async {
    final client = _RecordingExamApiClient(baseUrl: 'http://127.0.0.1:65535');

    await tester.pumpWidget(
      _TestApp(
        child: ExamShellScreen(
          client: client,
          examToken: 'abc12345',
          deviceFingerprint: 'android:test',
          initialPayload: _emptyQuestionPayload(),
        ),
      ),
    );

    await tester.pump();
    await tester.pump(const Duration(seconds: 2));

    expect(find.text('Paket soal belum tersedia'), findsOneWidget);
    expect(
      find.textContaining('Server mengirim sesi ujian tanpa daftar soal'),
      findsOneWidget,
    );
    expect(find.text('Perbarui Status'), findsOneWidget);
    expect(client.heartbeatCount, 0);
    expect(client.statusCount, 0);
    expect(client.submitCount, 0);
  });

  testWidgets('exam shell clamps restored question index', (tester) async {
    await tester.pumpWidget(
      _TestApp(
        child: ExamShellScreen(
          client: ExamApiClient(baseUrl: 'http://127.0.0.1:65535'),
          examToken: 'abc12345',
          deviceFingerprint: 'android:test',
          autoStartRuntime: false,
          initialPayload: _sampleLoginPayload(),
          restoredSnapshot: _sampleSnapshot(currentQuestionIndex: 99),
        ),
      ),
    );

    await tester.pump();

    expect(find.text('Soal 1'), findsOneWidget);
    expect(find.text('2 + 2 = ...'), findsOneWidget);
  });
}

class _RecordingExamApiClient extends ExamApiClient {
  _RecordingExamApiClient({
    required super.baseUrl,
    this.throwOnEvent = false,
    this.statusPayload = const ExamStatusPayload(
      answeredCount: 0,
      totalQuestions: 0,
      timeRemainingSeconds: 0,
      isSubmitted: false,
    ),
    this.statusError,
    this.submitError,
    this.saveAnswerError,
  });

  int heartbeatCount = 0;
  int statusCount = 0;
  int submitCount = 0;
  int eventCount = 0;
  int saveAnswerCount = 0;
  String? lastSavedQuestionId;
  String? lastSavedAnswer;
  final bool throwOnEvent;
  final ExamStatusPayload statusPayload;
  final ExamApiException? statusError;
  final ExamApiException? submitError;
  final ExamApiException? saveAnswerError;

  @override
  Future<void> sendHeartbeat(String token) async {
    heartbeatCount += 1;
  }

  @override
  Future<ExamStatusPayload> getStatus(String token) async {
    statusCount += 1;
    final error = statusError;
    if (error != null) {
      throw error;
    }
    return statusPayload;
  }

  @override
  Future<void> submit(String token) async {
    submitCount += 1;
    final error = submitError;
    if (error != null) {
      throw error;
    }
  }

  @override
  Future<void> saveAnswer({
    required String token,
    required String questionId,
    required String answer,
  }) async {
    saveAnswerCount += 1;
    lastSavedQuestionId = questionId;
    lastSavedAnswer = answer;
    final error = saveAnswerError;
    if (error != null) {
      throw error;
    }
  }

  @override
  Future<void> sendEvent({
    required String token,
    required String eventType,
    Map<String, Object?> data = const <String, Object?>{},
  }) async {
    eventCount += 1;
    if (throwOnEvent) {
      throw const ExamApiException('transport');
    }
  }
}

class _MemoryExamSessionStore extends ExamSessionStore {
  _MemoryExamSessionStore({ExamSessionSnapshot? initialSnapshot})
    : _snapshot = initialSnapshot;

  ExamSessionSnapshot? _snapshot;
  String? _baseUrl;

  @override
  Future<void> saveBaseUrl(String baseUrl) async {
    _baseUrl = baseUrl;
  }

  @override
  Future<String?> loadBaseUrl() async {
    return _baseUrl;
  }

  @override
  Future<void> saveSnapshot(ExamSessionSnapshot snapshot) async {
    _snapshot = snapshot;
    _baseUrl = snapshot.baseUrl;
  }

  @override
  Future<ExamSessionSnapshot?> loadSnapshot() async {
    return _snapshot;
  }

  @override
  Future<void> clearSnapshot() async {
    _snapshot = null;
  }
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

ExamLoginPayload _emptyQuestionPayload() {
  return ExamLoginPayload(
    participantId: 'participant-empty-1',
    student: const ExamStudent(nis: '24001', nama: 'Siti Aminah'),
    session: ExamSession(
      id: 'session-empty-1',
      title: 'Matematika Kelas VIII',
      scheduledStart: DateTime.parse('2026-05-01T08:00:00+08:00'),
      scheduledEnd: DateTime.parse('2026-05-01T09:30:00+08:00'),
      durationMinutes: 90,
    ),
    room: const ExamRoom(roomName: 'Lab 1'),
    questions: const [],
    answeredCount: 0,
    totalQuestions: 0,
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

ExamLoginPayload _sampleRelativeMediaLoginPayload() {
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
        stemMediaUrl: '/uploads/questions/q1.png',
        stimulusMediaUrl: '',
        stemAudioUrl: 'uploads/audio/q1.mp3',
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

ExamLoginPayload _sampleRichContentLoginPayload() {
  return ExamLoginPayload(
    participantId: 'participant-rich-1',
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
        id: 'question-rich-1',
        questionText: '',
        stemHtml: '<p>Apa gagasan utama paragraf di atas?</p>',
        stimulusHtml:
            '<p>Bacalah teks berikut dengan saksama.</p><ul><li>Kalimat pertama</li><li>Kalimat kedua</li></ul>',
        stemMediaUrl: '',
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

ExamLoginPayload _sampleMultipleAnswerPayload() {
  return ExamLoginPayload(
    participantId: 'participant-multiple-1',
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
        id: 'question-multiple-1',
        questionType: 'multiple_answer',
        questionText: 'Pilih semua pernyataan yang benar.',
        stemHtml: '',
        stimulusHtml: '',
        stemMediaUrl: '',
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

ExamLoginPayload _sampleTrueFalsePayload() {
  return ExamLoginPayload(
    participantId: 'participant-true-false-1',
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
        id: 'question-true-false-1',
        questionType: 'true_false',
        questionText: 'Fotosintesis menghasilkan oksigen.',
        stemHtml: '',
        stimulusHtml: '',
        stemMediaUrl: '',
        stimulusMediaUrl: '',
        stemAudioUrl: '',
        stimulusAudioUrl: '',
        options: [],
      ),
    ],
    answeredCount: 0,
    totalQuestions: 1,
    timeRemainingSeconds: 1800,
  );
}

ExamLoginPayload _sampleTrueFalseWithBackendOptionsPayload() {
  return ExamLoginPayload(
    participantId: 'participant-true-false-2',
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
        id: 'question-true-false-2',
        questionType: 'true_false',
        questionText: 'Air mendidih pada suhu ruang.',
        stemHtml: '',
        stimulusHtml: '',
        stemMediaUrl: '',
        stimulusMediaUrl: '',
        stemAudioUrl: '',
        stimulusAudioUrl: '',
        options: [
          ExamOption(label: 'A', text: 'Benar dari naskah'),
          ExamOption(label: 'B', text: 'Salah dari naskah'),
        ],
      ),
    ],
    answeredCount: 0,
    totalQuestions: 1,
    timeRemainingSeconds: 1800,
  );
}

ExamLoginPayload _sampleShortAnswerPayload() {
  return ExamLoginPayload(
    participantId: 'participant-short-1',
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
        id: 'question-short-1',
        questionType: 'short_answer',
        questionText: 'Proses tumbuhan membuat makanan disebut...',
        stemHtml: '',
        stimulusHtml: '',
        stemMediaUrl: '',
        stimulusMediaUrl: '',
        stemAudioUrl: '',
        stimulusAudioUrl: '',
        options: [],
      ),
    ],
    answeredCount: 0,
    totalQuestions: 1,
    timeRemainingSeconds: 1800,
  );
}

ExamLoginPayload _sampleMatchingPayload() {
  return ExamLoginPayload(
    participantId: 'participant-matching-1',
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
        id: 'question-matching-1',
        questionType: 'matching',
        questionText: 'Jodohkan istilah berikut dengan pengertiannya.',
        stemHtml: '',
        stimulusHtml: '',
        stemMediaUrl: '',
        stimulusMediaUrl: '',
        stemAudioUrl: '',
        stimulusAudioUrl: '',
        options: [
          ExamOption(
            label: 'A',
            text: 'Fotosintesis',
            matchLabel: '1',
            matchText: 'Proses membuat makanan',
          ),
          ExamOption(
            label: 'B',
            text: 'Evaporasi',
            matchLabel: '2',
            matchText: 'Penguapan',
          ),
          ExamOption(
            label: '',
            text: '',
            matchLabel: '3',
            matchText: 'Distraktor kanan',
            isDistractor: true,
          ),
        ],
      ),
    ],
    answeredCount: 0,
    totalQuestions: 1,
    timeRemainingSeconds: 1800,
  );
}

ExamLoginPayload _sampleOrderingPayload() {
  return ExamLoginPayload(
    participantId: 'participant-ordering-1',
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
        id: 'question-ordering-1',
        questionType: 'ordering',
        questionText: 'Urutkan langkah kerja ilmiah berikut.',
        stemHtml: '',
        stimulusHtml: '',
        stemMediaUrl: '',
        stimulusMediaUrl: '',
        stemAudioUrl: '',
        stimulusAudioUrl: '',
        options: [
          ExamOption(label: 'A', text: 'Langkah pertama'),
          ExamOption(label: 'B', text: 'Langkah kedua'),
          ExamOption(label: 'C', text: 'Langkah ketiga'),
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
  Map<String, String> answers = const <String, String>{},
  Map<String, String> pendingAnswers = const <String, String>{},
  List<String> playedAudioQuestionIds = const <String>[],
  int currentQuestionIndex = 0,
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
    currentQuestionIndex: currentQuestionIndex,
    answers: answers,
    pendingAnswers: pendingAnswers,
    playedAudioQuestionIds: playedAudioQuestionIds,
    lastServerContactIso: lastServerContactIso,
    lastSyncFailureIso: lastSyncFailureIso,
    consecutiveSyncFailures: consecutiveSyncFailures,
  );
}
