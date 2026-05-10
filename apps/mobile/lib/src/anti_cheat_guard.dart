import 'dart:async';

import 'package:flutter/services.dart';
import 'package:flutter/widgets.dart';

import 'exam_events.dart';

const String kAntiCheatMethodChannel = 'id.sch.mtsn2kolutara.mobile/anti_cheat';
const String kAntiCheatEventChannel =
    'id.sch.mtsn2kolutara.mobile/anti_cheat_events';

@immutable
class AntiCheatWindowState {
  const AntiCheatWindowState({
    this.isMultiWindow = false,
    this.isPictureInPicture = false,
    this.hasWindowFocus = true,
    this.secureFlagEnabled = false,
  });

  factory AntiCheatWindowState.fromMap(Map<Object?, Object?> map) {
    bool readBool(String key, bool fallback) {
      final value = map[key];
      return value is bool ? value : fallback;
    }

    return AntiCheatWindowState(
      isMultiWindow: readBool('isMultiWindow', false),
      isPictureInPicture: readBool('isPictureInPicture', false),
      hasWindowFocus: readBool('hasWindowFocus', true),
      secureFlagEnabled: readBool('secureFlagEnabled', false),
    );
  }

  final bool isMultiWindow;
  final bool isPictureInPicture;
  final bool hasWindowFocus;
  final bool secureFlagEnabled;

  Map<String, Object?> toJson() => <String, Object?>{
    'is_multi_window': isMultiWindow,
    'is_picture_in_picture': isPictureInPicture,
    'has_window_focus': hasWindowFocus,
    'secure_flag_enabled': secureFlagEnabled,
  };
}

@immutable
class AntiCheatSnapshot {
  const AntiCheatSnapshot({
    this.windowState = const AntiCheatWindowState(),
    this.lifecycleState = AppLifecycleState.resumed,
    this.violationCount = 0,
    this.locked = false,
    this.maxViolationsBeforeLock = 3,
  });

  final AntiCheatWindowState windowState;
  final AppLifecycleState lifecycleState;
  final int violationCount;
  final bool locked;
  final int maxViolationsBeforeLock;

  bool get isBackgrounded => lifecycleState != AppLifecycleState.resumed;
  bool get isSplitScreen => windowState.isMultiWindow;
  bool get isPictureInPicture => windowState.isPictureInPicture;
  bool get hasFocusLost => !windowState.hasWindowFocus;

  bool get shouldBlockInteraction =>
      locked ||
      isSplitScreen ||
      isPictureInPicture ||
      hasFocusLost ||
      isBackgrounded;

  String get primaryReason {
    if (locked) return 'anti_cheat_local_lock';
    if (isSplitScreen) return 'split_screen_detected';
    if (isPictureInPicture) return 'picture_in_picture_detected';
    if (isBackgrounded) return 'app_backgrounded';
    if (hasFocusLost) return 'window_focus_lost';
    return 'secure';
  }

  String get title {
    if (locked) return 'Ujian dikunci sementara';
    if (isSplitScreen) return 'Split screen tidak diizinkan';
    if (isPictureInPicture) return 'Picture-in-picture tidak diizinkan';
    if (isBackgrounded) return 'Aplikasi ujian harus tetap aktif';
    if (hasFocusLost) return 'Fokus aplikasi ujian hilang';
    return 'Mode ujian aman';
  }

  String get message {
    if (locked) {
      return 'Sistem mencatat beberapa pelanggaran mode ujian. Minta pengawas memeriksa perangkat dan melakukan reset akses bila siswa diizinkan lanjut.';
    }
    if (isSplitScreen) {
      return 'Tutup mode layar terbagi/multi-window, lalu kembalikan aplikasi CBT ke layar penuh sebelum melanjutkan.';
    }
    if (isPictureInPicture) {
      return 'Tutup mode picture-in-picture dan gunakan aplikasi CBT dalam layar penuh.';
    }
    if (isBackgrounded) {
      return 'Jangan membuka aplikasi lain selama ujian. Kembali ke aplikasi CBT dan tunggu pengecekan status.';
    }
    if (hasFocusLost) {
      return 'Pastikan tidak ada jendela melayang, pop-up, atau overlay lain di atas aplikasi CBT.';
    }
    return 'Perangkat berada pada mode ujian yang diizinkan.';
  }

  AntiCheatSnapshot copyWith({
    AntiCheatWindowState? windowState,
    AppLifecycleState? lifecycleState,
    int? violationCount,
    bool? locked,
    int? maxViolationsBeforeLock,
  }) {
    return AntiCheatSnapshot(
      windowState: windowState ?? this.windowState,
      lifecycleState: lifecycleState ?? this.lifecycleState,
      violationCount: violationCount ?? this.violationCount,
      locked: locked ?? this.locked,
      maxViolationsBeforeLock:
          maxViolationsBeforeLock ?? this.maxViolationsBeforeLock,
    );
  }

  ExamClientEvent toWarningEvent() {
    return ExamClientEvents.antiCheatViolation(
      reason: primaryReason,
      violationCount: violationCount,
      maxViolationsBeforeLock: maxViolationsBeforeLock,
      data: <String, Object?>{
        'lifecycle_state': lifecycleState.name,
        ...windowState.toJson(),
      },
    );
  }
}

class AntiCheatGuard {
  AntiCheatGuard({
    MethodChannel methodChannel = const MethodChannel(kAntiCheatMethodChannel),
    EventChannel eventChannel = const EventChannel(kAntiCheatEventChannel),
  }) : _methodChannel = methodChannel,
       _eventChannel = eventChannel;

  final MethodChannel _methodChannel;
  final EventChannel _eventChannel;

  Stream<AntiCheatWindowState> get windowStateStream {
    return _eventChannel.receiveBroadcastStream().map((event) {
      if (event is Map) {
        return AntiCheatWindowState.fromMap(event);
      }
      return const AntiCheatWindowState();
    });
  }

  Future<void> enableSecureFlag() async {
    try {
      await _methodChannel.invokeMethod<void>('enableSecureFlag');
    } on PlatformException {
      // FLAG_SECURE is best-effort. Native Android sets it during Activity start.
    } on MissingPluginException {
      // Non-Android tests/platforms do not expose this channel.
    }
  }

  Future<AntiCheatWindowState> getWindowState() async {
    try {
      final result = await _methodChannel.invokeMapMethod<Object?, Object?>(
        'getWindowState',
      );
      return AntiCheatWindowState.fromMap(result ?? const <Object?, Object?>{});
    } on PlatformException {
      return const AntiCheatWindowState();
    } on MissingPluginException {
      return const AntiCheatWindowState();
    }
  }
}
