class ExamClientEvent {
  const ExamClientEvent({required this.eventType, required this.data});

  final String eventType;
  final Map<String, Object?> data;

  Map<String, Object?> toJson() {
    return <String, Object?>{'event_type': eventType, 'data': data};
  }
}

class ExamClientEvents {
  const ExamClientEvents._();

  static const String typeAppSwitch = 'app_switch';
  static const String typeWarning = 'warning';
  static const String typeScreenshotAttempt = 'screenshot_attempt';
  static const String typeAntiCheatViolation = 'anti_cheat_violation';

  static const String reasonResumeExam = 'resume_exam';
  static const String reasonRepeatResumeAttempt = 'repeat_resume_attempt';
  static const String reasonAnswerSavedLocalOnly = 'answer_saved_local_only';
  static const String reasonSubmitBlockedPendingSync =
      'submit_blocked_pending_sync';
  static const String reasonAutoSubmitBlockedPendingSync =
      'auto_submit_blocked_pending_sync';
  static const String reasonSubmitBlockedDegradedMode =
      'submit_blocked_degraded_mode';
  static const String reasonDegradedModeEntered = 'degraded_mode_entered';
  static const String reasonStaleConnectionAttention =
      'stale_connection_attention';
  static const String reasonStaleConnectionEscalated =
      'stale_connection_escalated';
  static const String reasonBackButtonAttempt = 'back_button_attempt';
  static const String reasonManualSubmit = 'manual_submit';

  static ExamClientEvent appSwitch({
    required String state,
    required String deviceFingerprint,
  }) {
    return ExamClientEvent(
      eventType: typeAppSwitch,
      data: _sanitizeData(<String, Object?>{
        'state': state,
        // Device fingerprint is used only for runtime binding checks. Do not
        // persist it into proctor/evidence telemetry payloads.
        'device_fingerprint': deviceFingerprint,
      }),
    );
  }

  static ExamClientEvent warning({
    required String reason,
    Map<String, Object?> data = const <String, Object?>{},
  }) {
    return ExamClientEvent(
      eventType: typeWarning,
      data: _sanitizeData(<String, Object?>{'reason': reason, ...data}),
    );
  }

  static ExamClientEvent antiCheatViolation({
    required String reason,
    required int violationCount,
    required int maxViolationsBeforeLock,
    Map<String, Object?> data = const <String, Object?>{},
  }) {
    return ExamClientEvent(
      eventType: typeAntiCheatViolation,
      data: _sanitizeData(<String, Object?>{
        'reason': reason,
        'violation_count': violationCount,
        'max_violations_before_lock': maxViolationsBeforeLock,
        'severity': _antiCheatSeverity(
          violationCount: violationCount,
          maxViolationsBeforeLock: maxViolationsBeforeLock,
        ),
        ...data,
      }),
    );
  }

  static String _antiCheatSeverity({
    required int violationCount,
    required int maxViolationsBeforeLock,
  }) {
    if (violationCount >= maxViolationsBeforeLock) return 'critical';
    if (violationCount >= 2) return 'high';
    return 'warning';
  }

  static ExamClientEvent resumeGate({required int resumeAttemptCount}) {
    return warning(
      reason: reasonResumeExam,
      data: <String, Object?>{'resume_attempt_count': resumeAttemptCount},
    );
  }

  static ExamClientEvent repeatResumeAttempt({
    required int resumeAttemptCount,
  }) {
    return warning(
      reason: reasonRepeatResumeAttempt,
      data: <String, Object?>{'resume_attempt_count': resumeAttemptCount},
    );
  }

  static ExamClientEvent answerSavedLocalOnly({required String questionId}) {
    return warning(
      reason: reasonAnswerSavedLocalOnly,
      data: <String, Object?>{'question_id': questionId},
    );
  }

  static ExamClientEvent submitBlockedPendingSync({
    required int pendingCount,
    required bool autoSubmit,
  }) {
    return warning(
      reason: autoSubmit
          ? reasonAutoSubmitBlockedPendingSync
          : reasonSubmitBlockedPendingSync,
      data: <String, Object?>{'pending_count': pendingCount},
    );
  }

  static ExamClientEvent submitBlockedDegradedMode({
    required int failureCount,
  }) {
    return warning(
      reason: reasonSubmitBlockedDegradedMode,
      data: <String, Object?>{'failure_count': failureCount},
    );
  }

  static ExamClientEvent degradedModeEntered({required int failureCount}) {
    return warning(
      reason: reasonDegradedModeEntered,
      data: <String, Object?>{'failure_count': failureCount},
    );
  }

  static ExamClientEvent staleConnectionAttention({
    required int secondsSinceLastContact,
    required int failureCount,
  }) {
    return warning(
      reason: reasonStaleConnectionAttention,
      data: <String, Object?>{
        'seconds_since_last_contact': secondsSinceLastContact,
        'failure_count': failureCount,
      },
    );
  }

  static ExamClientEvent staleConnectionEscalated({
    required int secondsSinceLastContact,
    required int failureCount,
  }) {
    return warning(
      reason: reasonStaleConnectionEscalated,
      data: <String, Object?>{
        'seconds_since_last_contact': secondsSinceLastContact,
        'failure_count': failureCount,
      },
    );
  }

  static ExamClientEvent backButtonAttempt() {
    return warning(reason: reasonBackButtonAttempt);
  }

  static ExamClientEvent manualSubmit() {
    return warning(reason: reasonManualSubmit);
  }

  static Map<String, Object?> _sanitizeData(Map<String, Object?> data) {
    final sanitized = <String, Object?>{};
    for (final entry in data.entries) {
      final key = entry.key.trim();
      if (key.isEmpty || _isSensitiveKey(key)) {
        continue;
      }
      sanitized[key] = _sanitizeValue(entry.value);
    }
    return sanitized;
  }

  static Object? _sanitizeValue(Object? value) {
    if (value is Map) {
      final nested = <String, Object?>{};
      for (final entry in value.entries) {
        final key = entry.key.toString().trim();
        if (key.isEmpty || _isSensitiveKey(key)) {
          continue;
        }
        nested[key] = _sanitizeValue(entry.value);
      }
      return nested;
    }
    if (value is Iterable) {
      return value.map(_sanitizeValue).toList(growable: false);
    }
    return value;
  }

  static bool _isSensitiveKey(String key) {
    final normalized = key.toLowerCase().replaceAll(RegExp(r'[^a-z0-9]'), '');
    return normalized.contains('token') ||
        normalized.contains('password') ||
        normalized.contains('answerkey') ||
        normalized.contains('devicefingerprint');
  }
}
