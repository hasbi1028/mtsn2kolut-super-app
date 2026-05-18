import 'package:flutter/widgets.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:mobile/src/anti_cheat_guard.dart';
import 'package:mobile/src/exam_events.dart';

void main() {
  group('AntiCheatWindowState', () {
    test('parses native Android window state map', () {
      final state = AntiCheatWindowState.fromMap(const <Object?, Object?>{
        'isMultiWindow': true,
        'isPictureInPicture': false,
        'hasWindowFocus': false,
        'secureFlagEnabled': true,
      });

      expect(state.isMultiWindow, isTrue);
      expect(state.isPictureInPicture, isFalse);
      expect(state.hasWindowFocus, isFalse);
      expect(state.secureFlagEnabled, isTrue);
    });

    test('parses native Windows window state map', () {
      final state = AntiCheatWindowState.fromMap(const <Object?, Object?>{
        'platform': 'windows',
        'isFullscreen': false,
        'isMinimized': true,
        'hasWindowFocus': false,
        'secureFlagEnabled': true,
      });

      expect(state.platform, 'windows');
      expect(state.isFullscreen, isFalse);
      expect(state.isMinimized, isTrue);
      expect(state.hasWindowFocus, isFalse);
      expect(state.secureFlagEnabled, isTrue);
    });
  });

  group('AntiCheatSnapshot', () {
    test('blocks split screen and emits sanitized warning event for the first strike', () {
      const snapshot = AntiCheatSnapshot(
        windowState: AntiCheatWindowState(
          isMultiWindow: true,
          secureFlagEnabled: true,
        ),
        violationCount: 1,
      );

      expect(snapshot.shouldBlockInteraction, isTrue);
      expect(snapshot.primaryReason, 'split_screen_detected');
      final event = snapshot.toWarningEvent();
      expect(event.eventType, ExamClientEvents.typeAntiCheatViolation);
      expect(event.data['reason'], 'split_screen_detected');
      expect(event.data['violation_count'], 1);
      expect(event.data['severity'], 'warning');
      expect(event.data.containsKey('device_fingerprint'), isFalse);
      expect(event.data.containsKey('token'), isFalse);
    });

    test('escalates the second anti-cheat strike to high severity', () {
      const snapshot = AntiCheatSnapshot(
        windowState: AntiCheatWindowState(isMultiWindow: true),
        violationCount: 2,
      );

      final event = snapshot.toWarningEvent();
      expect(event.data['severity'], 'high');
    });

    test('locks locally after configured max violations', () {
      const snapshot = AntiCheatSnapshot(
        windowState: AntiCheatWindowState(isPictureInPicture: true),
        violationCount: 3,
        locked: true,
      );

      expect(snapshot.shouldBlockInteraction, isTrue);
      expect(snapshot.primaryReason, 'anti_cheat_local_lock');
      final event = snapshot.toWarningEvent();
      expect(event.data['severity'], 'critical');
    });

    test('treats background lifecycle as blocking', () {
      const snapshot = AntiCheatSnapshot(
        lifecycleState: AppLifecycleState.paused,
      );

      expect(snapshot.shouldBlockInteraction, isTrue);
      expect(snapshot.primaryReason, 'app_backgrounded');
    });

    test('treats Windows non-fullscreen as blocking', () {
      const snapshot = AntiCheatSnapshot(
        windowState: AntiCheatWindowState(
          platform: 'windows',
          isFullscreen: false,
        ),
      );

      expect(snapshot.shouldBlockInteraction, isTrue);
      expect(snapshot.primaryReason, 'windows_not_fullscreen');
    });

    test('treats Windows minimized state as backgrounded', () {
      const snapshot = AntiCheatSnapshot(
        windowState: AntiCheatWindowState(
          platform: 'windows',
          isFullscreen: true,
          isMinimized: true,
        ),
      );

      expect(snapshot.shouldBlockInteraction, isTrue);
      expect(snapshot.primaryReason, 'app_minimized');
    });
  });
}
