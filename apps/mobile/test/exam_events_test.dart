import 'package:flutter_test/flutter_test.dart';
import 'package:mobile/src/exam_events.dart';

void main() {
  group('ExamClientEvents', () {
    test('builds BYOD warning taxonomy payloads used by the exam shell', () {
      expect(
        ExamClientEvents.resumeGate(resumeAttemptCount: 2).toJson(),
        <String, Object?>{
          'event_type': 'warning',
          'data': <String, Object?>{
            'reason': 'resume_exam',
            'resume_attempt_count': 2,
          },
        },
      );

      expect(
        ExamClientEvents.submitBlockedPendingSync(
          pendingCount: 3,
          autoSubmit: true,
        ).toJson(),
        <String, Object?>{
          'event_type': 'warning',
          'data': <String, Object?>{
            'reason': 'auto_submit_blocked_pending_sync',
            'pending_count': 3,
          },
        },
      );

      expect(
        ExamClientEvents.staleConnectionEscalated(
          secondsSinceLastContact: 245,
          failureCount: 2,
        ).toJson(),
        <String, Object?>{
          'event_type': 'warning',
          'data': <String, Object?>{
            'reason': 'stale_connection_escalated',
            'seconds_since_last_contact': 245,
            'failure_count': 2,
          },
        },
      );
    });

    test('keeps app switch as a first-class event with fingerprint hint', () {
      expect(
        ExamClientEvents.appSwitch(
          state: 'paused',
          deviceFingerprint: 'android:test',
        ).toJson(),
        <String, Object?>{
          'event_type': 'app_switch',
          'data': <String, Object?>{
            'state': 'paused',
            'device_fingerprint': 'android:test',
          },
        },
      );
    });

    test('strips sensitive fields from nested telemetry payloads', () {
      final event = ExamClientEvents.warning(
        reason: ExamClientEvents.reasonManualSubmit,
        data: <String, Object?>{
          'exam_token': 'secret-token',
          'password': 'secret-password',
          'answer_key': 'A',
          'safe_count': 1,
          'nested': <String, Object?>{
            'refresh_token': 'secret-refresh',
            'status': 'blocked',
          },
          'list': <Object?>[
            <String, Object?>{'token': 'secret-list', 'reason': 'safe'},
          ],
        },
      );

      expect(event.data, <String, Object?>{
        'reason': 'manual_submit',
        'safe_count': 1,
        'nested': <String, Object?>{'status': 'blocked'},
        'list': <Object?>[
          <String, Object?>{'reason': 'safe'},
        ],
      });
    });
  });
}
