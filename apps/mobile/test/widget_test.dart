import 'package:flutter/material.dart';
import 'package:flutter_test/flutter_test.dart';

import 'package:mobile/src/app.dart';
import 'package:mobile/src/exam_error_messages.dart';
import 'package:mobile/src/exam_session_store.dart';
import 'package:mobile/src/screens/exam_login_screen.dart';
import 'package:mobile/src/screens/exam_restore_failed_screen.dart';

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

    await tester.pumpAndSettle();

    expect(find.text('Masuk Ujian'), findsWidgets);
    expect(find.text('Token sudah terikat ke perangkat lain'), findsOneWidget);
    expect(
      find.text(
        'Jangan terus mencoba login dari perangkat ini. Gunakan perangkat yang sama seperti sebelumnya atau minta pengawas memverifikasi token.',
      ),
      findsOneWidget,
    );
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

    await tester.pumpAndSettle();

    expect(find.text('Sesi lama tidak bisa dipulihkan'), findsOneWidget);
    expect(find.text('Sesi lama aktif di perangkat lain'), findsOneWidget);
    expect(
      find.text(
        'Peserta tidak perlu terus mencoba restore di perangkat ini. Pengawas sebaiknya mengarahkan peserta kembali ke perangkat awal atau memeriksa status token.',
      ),
      findsOneWidget,
    );
    expect(find.text('Matematika Kelas VIII'), findsOneWidget);
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
