import 'dart:async';

import 'package:flutter/material.dart';

import '../exam_api.dart';
import '../exam_error_messages.dart';
import '../exam_events.dart';
import '../exam_session_store.dart';
import '../models.dart';
import 'exam_completed_screen.dart';
import 'exam_shell_connection.dart';
import 'exam_status_guide_screen.dart';
import 'exam_shell_widgets.dart';
import '../widgets/audio_prompt_card.dart';
import '../widgets/rich_exam_text.dart';

// Caps to keep answers within the backend MaxBytesReader budget (64 KB) and
// give students a clear UX before the server rejects an oversized payload.
// Essay answers use a conservative character cap because UTF-8 and JSON escaping
// can make serialized payloads larger than the visible character count.
const int kShortAnswerMaxChars = 256;
const int kEssayAnswerMaxChars = 15000;

class ExamShellScreen extends StatefulWidget {
  const ExamShellScreen({
    super.key,
    required this.client,
    required this.examToken,
    required this.initialPayload,
    required this.deviceFingerprint,
    this.restoredSnapshot,
    this.autoStartRuntime = true,
    this.initialServerNotice,
    this.initialResumeCheckRequired = false,
    this.initialIsResumingExam = false,
    this.initialIsSyncingStatus = false,
    this.initialIsSavingAnswer = false,
    this.initialErrorMessage,
    this.sessionStore,
  });

  final ExamApiClient client;
  final String examToken;
  final ExamLoginPayload initialPayload;
  final String deviceFingerprint;
  final ExamSessionSnapshot? restoredSnapshot;
  final bool autoStartRuntime;
  final ExamGuidanceNotice? initialServerNotice;
  final bool initialResumeCheckRequired;
  final bool initialIsResumingExam;
  final bool initialIsSyncingStatus;
  final bool initialIsSavingAnswer;
  final String? initialErrorMessage;
  final ExamSessionStore? sessionStore;

  @override
  State<ExamShellScreen> createState() => _ExamShellScreenState();
}

class _ExamShellScreenState extends State<ExamShellScreen>
    with WidgetsBindingObserver {
  late final ExamSessionStore _sessionStore;
  late final Map<String, String> _answers;
  late final Map<String, String> _pendingAnswers;
  late final List<TextEditingController> _essayControllers;
  late final Set<String> _playedAudioQuestionIds;

  int _currentQuestionIndex = 0;
  int _answeredCount = 0;
  int _timeRemainingSeconds = 0;
  bool _isSavingAnswer = false;
  bool _isSubmitting = false;
  bool _isSyncingStatus = false;
  bool _isSubmitted = false;
  bool _isSubmitPendingIntervention = false;
  bool _resumeCheckRequired = false;
  bool _isResumingExam = false;
  bool _hasReportedDegradedMode = false;
  bool _hasReportedStaleAttention = false;
  bool _hasReportedEscalatedStaleAttention = false;
  int _resumeAttemptCount = 0;
  int _consecutiveSyncFailures = 0;
  DateTime? _lastServerContactAt;
  DateTime? _lastSyncFailureAt;
  String? _statusMessage;
  String? _errorMessage;
  ExamGuidanceNotice? _serverNotice;
  Timer? _countdownTimer;
  Timer? _heartbeatTimer;
  Timer? _textAutosaveTimer;

  ExamShellConnectionViewModel get _connectionState =>
      ExamShellConnectionViewModel(
        isSubmitted: _isSubmitted,
        pendingAnswerCount: _pendingAnswers.length,
        consecutiveSyncFailures: _consecutiveSyncFailures,
        hasError: _errorMessage != null,
        lastServerContactAt: _lastServerContactAt,
        lastSyncFailureAt: _lastSyncFailureAt,
        isSavingAnswer: _isSavingAnswer,
        isSyncingStatus: _isSyncingStatus,
        isResumingExam: _isResumingExam,
        resumeCheckRequired: _resumeCheckRequired,
      );

  @override
  void initState() {
    super.initState();
    _sessionStore = widget.sessionStore ?? ExamSessionStore();
    WidgetsBinding.instance.addObserver(this);
    _answers = Map<String, String>.from(widget.restoredSnapshot?.answers ?? {});
    _pendingAnswers = Map<String, String>.from(
      widget.restoredSnapshot?.pendingAnswers ?? {},
    );
    _playedAudioQuestionIds = Set<String>.from(
      widget.restoredSnapshot?.playedAudioQuestionIds ?? const <String>[],
    );
    _essayControllers = widget.initialPayload.questions
        .map((_) => TextEditingController())
        .toList();
    _answeredCount = widget.initialPayload.answeredCount;
    _timeRemainingSeconds = widget.initialPayload.timeRemainingSeconds;
    _currentQuestionIndex = _clampQuestionIndex(
      widget.restoredSnapshot?.currentQuestionIndex ?? 0,
    );
    _lastServerContactAt = DateTime.tryParse(
      widget.restoredSnapshot?.lastServerContactIso ?? '',
    );
    _lastSyncFailureAt = DateTime.tryParse(
      widget.restoredSnapshot?.lastSyncFailureIso ?? '',
    );
    _consecutiveSyncFailures =
        widget.restoredSnapshot?.consecutiveSyncFailures ?? 0;
    _serverNotice = widget.initialServerNotice;
    _resumeCheckRequired = widget.initialResumeCheckRequired;
    _isResumingExam = widget.initialIsResumingExam;
    _isSyncingStatus = widget.initialIsSyncingStatus;
    _isSavingAnswer = widget.initialIsSavingAnswer;
    _errorMessage = widget.initialErrorMessage;
    for (var i = 0; i < widget.initialPayload.questions.length; i++) {
      final question = widget.initialPayload.questions[i];
      if (question.isTextAnswer) {
        _essayControllers[i].text = _answers[question.id] ?? '';
      }
    }
    if (_answers.isNotEmpty) {
      final localAnsweredCount = _calculateAnsweredCount();
      if (localAnsweredCount > _answeredCount) {
        _answeredCount = localAnsweredCount;
      }
    }
    if (widget.autoStartRuntime && widget.initialPayload.questions.isNotEmpty) {
      _startCountdown();
      _startHeartbeat();
      _syncStatus();
    }
    unawaited(_persistSnapshot());
  }

  @override
  void dispose() {
    WidgetsBinding.instance.removeObserver(this);
    _countdownTimer?.cancel();
    _heartbeatTimer?.cancel();
    _textAutosaveTimer?.cancel();
    for (final controller in _essayControllers) {
      controller.dispose();
    }
    super.dispose();
  }

  @override
  void didChangeAppLifecycleState(AppLifecycleState state) {
    if (state == AppLifecycleState.inactive ||
        state == AppLifecycleState.paused ||
        state == AppLifecycleState.detached) {
      if (mounted) {
        setState(() {
          _resumeCheckRequired = true;
          _statusMessage =
              'Aplikasi meninggalkan mode ujian. Status akan dicek ulang saat kembali.';
        });
      }
      unawaited(
        widget.client.sendExamEvent(
          token: widget.examToken,
          event: ExamClientEvents.appSwitch(
            state: state.name,
            deviceFingerprint: widget.deviceFingerprint,
          ),
        ),
      );
      unawaited(_persistSnapshot());
    } else if (state == AppLifecycleState.resumed && !_isSubmitted) {
      unawaited(_handleResumeCheck());
    }
  }

  void _startCountdown() {
    _countdownTimer?.cancel();
    _countdownTimer = Timer.periodic(const Duration(seconds: 1), (timer) {
      if (!mounted || _isSubmitted) {
        timer.cancel();
        return;
      }
      if (_timeRemainingSeconds <= 0) {
        timer.cancel();
        _submit(autoSubmit: true);
        return;
      }
      setState(() {
        _timeRemainingSeconds -= 1;
      });
      unawaited(_handleConnectionAttentionSignals());
    });
  }

  void _startHeartbeat() {
    _heartbeatTimer?.cancel();
    _heartbeatTimer = Timer.periodic(const Duration(seconds: 45), (_) async {
      if (_isSubmitted) {
        return;
      }
      try {
        await widget.client.sendHeartbeat(widget.examToken);
        _markServerContact();
        if (_pendingAnswers.isNotEmpty) {
          await _flushPendingAnswers();
        }
      } catch (_) {
        if (!mounted) {
          return;
        }
        setState(() {
          _consecutiveSyncFailures += 1;
          _lastSyncFailureAt = DateTime.now();
          _errorMessage = 'Koneksi ke server ujian sempat terputus.';
        });
        unawaited(_handleConnectionAttentionSignals());
      }
    });
  }

  Future<bool> _syncStatus() async {
    setState(() {
      _isSyncingStatus = true;
    });
    try {
      final status = await widget.client.getStatus(widget.examToken);
      if (!mounted) {
        return false;
      }
      setState(() {
        _answeredCount = status.answeredCount;
        _timeRemainingSeconds = status.timeRemainingSeconds;
        _isSubmitted = status.isSubmitted;
      });
      _markServerContact();
      if (status.isSubmitted) {
        await _finishExam(wasAutoSubmitted: false);
        return true;
      }
      if (_pendingAnswers.isNotEmpty) {
        await _flushPendingAnswers();
      }
      return true;
    } on ExamApiException catch (error) {
      if (!mounted) {
        return false;
      }
      setState(() {
        _consecutiveSyncFailures += 1;
        _lastSyncFailureAt = DateTime.now();
        _errorMessage = statusFailureMessage(error);
        _serverNotice = statusFailureNotice(error);
      });
      unawaited(_handleConnectionAttentionSignals());
      return false;
    } catch (_) {
      if (!mounted) {
        return false;
      }
      setState(() {
        _consecutiveSyncFailures += 1;
        _lastSyncFailureAt = DateTime.now();
        _errorMessage = 'Status server belum bisa diperbarui.';
      });
      unawaited(_handleConnectionAttentionSignals());
      return false;
    } finally {
      if (mounted) {
        setState(() {
          _isSyncingStatus = false;
        });
      }
    }
  }

  Future<void> _handleResumeCheck() async {
    if (_isResumingExam || _isSubmitted) {
      return;
    }
    if (mounted) {
      setState(() {
        _isResumingExam = true;
        _statusMessage =
            'Memeriksa ulang status ujian setelah aplikasi dibuka kembali...';
        _errorMessage = null;
      });
    }
    await widget.client
        .sendExamEvent(
          token: widget.examToken,
          event: ExamClientEvents.resumeGate(
            resumeAttemptCount: _resumeAttemptCount + 1,
          ),
        )
        .catchError((_) {});
    _resumeAttemptCount += 1;
    if (_resumeAttemptCount > 1) {
      await widget.client
          .sendExamEvent(
            token: widget.examToken,
            event: ExamClientEvents.repeatResumeAttempt(
              resumeAttemptCount: _resumeAttemptCount,
            ),
          )
          .catchError((_) {});
    }
    final statusRefreshed = await _syncStatus();
    if (!mounted) {
      return;
    }
    if (!statusRefreshed) {
      setState(() {
        _isResumingExam = false;
        _resumeCheckRequired = true;
        _statusMessage =
            'Status ujian belum berhasil dicek ulang. Tetap di mode aman dan minta pengawas membantu koneksi.';
      });
      return;
    }
    setState(() {
      _isResumingExam = false;
      _resumeCheckRequired = false;
      _statusMessage = _pendingAnswers.isEmpty
          ? 'Status ujian sudah diperbarui. Anda dapat melanjutkan.'
          : 'Status ujian diperbarui, tetapi masih ada jawaban lokal yang menunggu sinkron.';
    });
  }

  int _clampQuestionIndex(int index) {
    final questionCount = widget.initialPayload.questions.length;
    if (questionCount == 0) {
      return 0;
    }
    if (index < 0) {
      return 0;
    }
    if (index >= questionCount) {
      return questionCount - 1;
    }
    return index;
  }

  Future<bool> _flushPendingAnswers() async {
    if (_pendingAnswers.isEmpty || _isSubmitted) {
      return _pendingAnswers.isEmpty;
    }

    final entries = Map<String, String>.from(_pendingAnswers).entries.toList();
    var syncedCount = 0;

    for (final entry in entries) {
      try {
        await widget.client.saveAnswer(
          token: widget.examToken,
          questionId: entry.key,
          answer: entry.value,
        );
        _pendingAnswers.remove(entry.key);
        syncedCount += 1;
        _markServerContact();
      } on ExamApiException catch (error) {
        if (error.isAlreadySubmittedConflict) {
          final statusRefreshed = await _syncStatus();
          if (_isSubmitted) {
            _pendingAnswers.remove(entry.key);
            return true;
          }
          return statusRefreshed && _pendingAnswers.isEmpty;
        }
        if (mounted) {
          setState(() {
            _consecutiveSyncFailures += 1;
            _lastSyncFailureAt = DateTime.now();
          });
        }
        await _handleConnectionAttentionSignals();
        break;
      } catch (_) {
        if (mounted) {
          setState(() {
            _consecutiveSyncFailures += 1;
            _lastSyncFailureAt = DateTime.now();
          });
        }
        await _handleConnectionAttentionSignals();
        break;
      }
    }

    await _persistSnapshot();

    if (!mounted || syncedCount == 0) {
      return _pendingAnswers.isEmpty;
    }

    setState(() {
      _answeredCount = _calculateAnsweredCount();
      _statusMessage = _pendingAnswers.isEmpty
          ? 'Semua jawaban lokal berhasil disinkronkan ke server.'
          : '$syncedCount jawaban lokal berhasil disinkronkan. Sisanya akan dicoba lagi.';
      if (_pendingAnswers.isEmpty) {
        _isSubmitPendingIntervention = false;
      }
    });
    return _pendingAnswers.isEmpty;
  }

  int _calculateAnsweredCount() {
    return widget.initialPayload.questions
        .where(
          (question) =>
              _isAnswerComplete(question, _answers[question.id] ?? ''),
        )
        .length;
  }

  Future<void> _selectOption(ExamQuestion question, String answer) async {
    if (_isSubmitted) {
      return;
    }
    final wasAnswered = _isAnswerComplete(
      question,
      _answers[question.id] ?? '',
    );
    final isAnswered = _isAnswerComplete(question, answer);
    setState(() {
      _answers[question.id] = answer;
      if (!wasAnswered && isAnswered) {
        _answeredCount += 1;
      } else if (wasAnswered && !isAnswered && _answeredCount > 0) {
        _answeredCount -= 1;
      }
      _isSavingAnswer = true;
      _statusMessage = 'Jawaban sedang disimpan...';
      _errorMessage = null;
    });

    try {
      await widget.client.saveAnswer(
        token: widget.examToken,
        questionId: question.id,
        answer: answer,
      );
      if (!mounted) {
        return;
      }
      setState(() {
        _pendingAnswers.remove(question.id);
        _statusMessage = 'Jawaban tersimpan ke server.';
        _serverNotice = null;
      });
      _markServerContact();
      await _persistSnapshot();
    } on ExamApiException catch (error) {
      if (!mounted) {
        return;
      }
      if (error.isAlreadySubmittedConflict) {
        setState(() {
          _pendingAnswers[question.id] = answer;
          _errorMessage = null;
          _statusMessage = answerFailureMessage(error);
          _serverNotice = answerFailureNotice(error);
        });
        await _syncStatus();
        if (_isSubmitted) {
          _pendingAnswers.remove(question.id);
        }
        if (!_isSubmitted) {
          await _persistSnapshot();
        }
        return;
      }
      setState(() {
        _consecutiveSyncFailures += 1;
        _lastSyncFailureAt = DateTime.now();
        _pendingAnswers[question.id] = answer;
        _errorMessage = null;
        _statusMessage = answerFailureMessage(error);
        _serverNotice = answerFailureNotice(error);
      });
      await widget.client
          .sendExamEvent(
            token: widget.examToken,
            event: ExamClientEvents.answerSavedLocalOnly(
              questionId: question.id,
            ),
          )
          .catchError((_) {});
      await _handleConnectionAttentionSignals();
      await _persistSnapshot();
    } finally {
      if (mounted) {
        setState(() {
          _isSavingAnswer = false;
        });
      }
    }
  }

  Future<void> _saveTextAnswer() async {
    final question = widget.initialPayload.questions[_currentQuestionIndex];
    final answer = _essayControllers[_currentQuestionIndex].text.trim();

    if (answer.isEmpty) {
      setState(() {
        _errorMessage = question.isShortAnswer
            ? 'Isi jawaban singkat terlebih dahulu.'
            : 'Isi jawaban uraian terlebih dahulu.';
      });
      return;
    }

    if (!ExamApiClient.isAnswerBodyWithinLimit(
      questionId: question.id,
      answer: answer,
    )) {
      setState(() {
        _errorMessage =
            'Jawaban terlalu panjang untuk dikirim ke server. Ringkas jawaban sebelum menyimpan.';
        _serverNotice = const ExamGuidanceNotice(
          title: 'Jawaban melebihi batas kirim',
          message:
              'Aplikasi menahan jawaban ini agar tidak menjadi pending permanen. Ringkas isi jawaban, lalu simpan ulang sebelum mengirim ujian.',
          tone: ExamGuidanceTone.warning,
        );
      });
      return;
    }

    await _selectOption(question, answer);
  }

  void _autosaveTextAnswer(ExamQuestion question, String rawAnswer) {
    if (_isSubmitted) {
      return;
    }
    final answer = rawAnswer.trim();
    setState(() {
      if (answer.isEmpty) {
        _answers.remove(question.id);
      } else {
        _answers[question.id] = answer;
      }
      _answeredCount = _calculateAnsweredCount();
      _statusMessage = 'Draft jawaban tersimpan lokal.';
      _errorMessage = null;
    });
    _textAutosaveTimer?.cancel();
    _textAutosaveTimer = Timer(const Duration(milliseconds: 350), () {
      unawaited(_persistSnapshot());
    });
  }

  Future<void> _persistSnapshot() async {
    await _sessionStore.saveSnapshot(
      ExamSessionSnapshot(
        baseUrl: widget.client.baseUrl,
        examToken: widget.examToken,
        deviceFingerprint: widget.deviceFingerprint,
        studentName: widget.initialPayload.student.nama,
        studentNis: widget.initialPayload.student.nis,
        sessionTitle: widget.initialPayload.session.title,
        roomName: widget.initialPayload.room?.roomName ?? '-',
        scheduledStartIso:
            widget.initialPayload.session.scheduledStart?.toIso8601String() ??
            '',
        scheduledEndIso:
            widget.initialPayload.session.scheduledEnd?.toIso8601String() ?? '',
        durationMinutes: widget.initialPayload.session.durationMinutes,
        currentQuestionIndex: _currentQuestionIndex,
        answers: Map<String, String>.from(_answers),
        pendingAnswers: Map<String, String>.from(_pendingAnswers),
        playedAudioQuestionIds: _playedAudioQuestionIds.toList()..sort(),
        lastServerContactIso: _lastServerContactAt?.toIso8601String() ?? '',
        lastSyncFailureIso: _lastSyncFailureAt?.toIso8601String() ?? '',
        consecutiveSyncFailures: _consecutiveSyncFailures,
      ),
    );
  }

  Future<void> _submit({bool autoSubmit = false}) async {
    if (widget.initialPayload.questions.isEmpty) {
      return;
    }

    if (_isSubmitting || _isSubmitted) {
      return;
    }

    if (_pendingAnswers.isNotEmpty) {
      final flushed = await _flushPendingAnswers();
      if (!flushed || _pendingAnswers.isNotEmpty) {
        if (!mounted) {
          return;
        }
        setState(() {
          _isSubmitPendingIntervention = true;
          _errorMessage = autoSubmit
              ? 'Waktu habis, tetapi masih ada jawaban lokal yang belum diterima server. Tetap di layar ini, minta pengawas memeriksa koneksi, lalu tekan perbarui status atau coba kirim ulang setelah sinkron pulih.'
              : 'Masih ada jawaban yang belum tersinkron ke server. Tunggu koneksi stabil lalu coba kirim lagi.';
          _serverNotice = ExamGuidanceNotice(
            title: autoSubmit
                ? 'Submit otomatis ditahan'
                : 'Submit ditahan sementara',
            message:
                'Snapshot jawaban lokal tetap disimpan di perangkat ini. Jangan menutup aplikasi sampai pengawas memastikan sinkronisasi pulih atau memberikan instruksi lanjutan.',
            tone: ExamGuidanceTone.danger,
          );
        });
        await _persistSnapshot();
        await widget.client
            .sendExamEvent(
              token: widget.examToken,
              event: ExamClientEvents.submitBlockedPendingSync(
                pendingCount: _pendingAnswers.length,
                autoSubmit: autoSubmit,
              ),
            )
            .catchError((_) {});
        return;
      }
    }

    if (!autoSubmit && _isDegradedMode) {
      setState(() {
        _errorMessage =
            'Mode koneksi menurun sedang aktif. Perbarui status dan tunggu sinkron pulih sebelum mengirim ujian.';
      });
      await widget.client
          .sendExamEvent(
            token: widget.examToken,
            event: ExamClientEvents.submitBlockedDegradedMode(
              failureCount: _consecutiveSyncFailures,
            ),
          )
          .catchError((_) {});
      return;
    }

    if (!autoSubmit) {
      if (!mounted) {
        return;
      }
      final confirmed = await showDialog<bool>(
        context: context,
        builder: (context) {
          return AlertDialog(
            title: const Text('Kirim jawaban akhir?'),
            content: Text(
              'Jawaban yang sudah dikirim tidak bisa diubah lagi. '
              'Anda sudah menjawab $_answeredCount dari ${widget.initialPayload.totalQuestions} soal.',
            ),
            actions: [
              TextButton(
                onPressed: () => Navigator.of(context).pop(false),
                child: const Text('Batal'),
              ),
              FilledButton(
                onPressed: () => Navigator.of(context).pop(true),
                child: const Text('Kirim Ujian'),
              ),
            ],
          );
        },
      );

      if (confirmed != true) {
        return;
      }
    }

    setState(() {
      _isSubmitting = true;
      _isSubmitPendingIntervention = false;
      _statusMessage = autoSubmit
          ? 'Waktu habis. Jawaban sedang dikirim otomatis...'
          : 'Jawaban akhir sedang dikirim...';
      _errorMessage = null;
    });

    try {
      await widget.client.submit(widget.examToken);
      if (!mounted) {
        return;
      }
      await _finishExam(wasAutoSubmitted: autoSubmit);
      if (!autoSubmit) {
        await widget.client.sendExamEvent(
          token: widget.examToken,
          event: ExamClientEvents.manualSubmit(),
        );
      }
    } on ExamApiException catch (error) {
      if (!mounted) {
        return;
      }
      if (error.isAlreadySubmittedConflict) {
        await _finishExam(wasAutoSubmitted: autoSubmit);
        return;
      }
      setState(() {
        _consecutiveSyncFailures += 1;
        _lastSyncFailureAt = DateTime.now();
        _errorMessage = submitFailureMessage(error, autoSubmit: autoSubmit);
        _serverNotice = submitFailureNotice(error, autoSubmit: autoSubmit);
      });
      await _handleConnectionAttentionSignals();
    } finally {
      if (mounted) {
        setState(() {
          _isSubmitting = false;
        });
      }
    }
  }

  void _markServerContact() {
    if (!mounted) {
      return;
    }
    setState(() {
      _lastServerContactAt = DateTime.now();
      _lastSyncFailureAt = null;
      _consecutiveSyncFailures = 0;
      _serverNotice = null;
    });
    _hasReportedDegradedMode = false;
    _hasReportedStaleAttention = false;
    _hasReportedEscalatedStaleAttention = false;
  }

  bool get _isDegradedMode => _connectionState.isDegradedMode;

  Future<void> _handleConnectionAttentionSignals() async {
    await _handlePotentialDegradedMode();
    await _handlePotentialStaleAttention();
    await _handlePotentialEscalatedStaleAttention();
  }

  Future<void> _handlePotentialDegradedMode() async {
    if (!_isDegradedMode || _hasReportedDegradedMode) {
      return;
    }
    _hasReportedDegradedMode = true;
    await widget.client
        .sendExamEvent(
          token: widget.examToken,
          event: ExamClientEvents.degradedModeEntered(
            failureCount: _consecutiveSyncFailures,
          ),
        )
        .catchError((_) {});
  }

  Future<void> _handlePotentialStaleAttention() async {
    if (!_connectionState.needsSupervisorAttention ||
        _hasReportedStaleAttention) {
      return;
    }
    _hasReportedStaleAttention = true;
    await widget.client
        .sendExamEvent(
          token: widget.examToken,
          event: ExamClientEvents.staleConnectionAttention(
            secondsSinceLastContact: DateTime.now()
                .difference(_lastServerContactAt!)
                .inSeconds,
            failureCount: _consecutiveSyncFailures,
          ),
        )
        .catchError((_) {});
  }

  Future<void> _handlePotentialEscalatedStaleAttention() async {
    if (!_connectionState.needsEscalatedSupervisorAttention ||
        _hasReportedEscalatedStaleAttention) {
      return;
    }
    _hasReportedEscalatedStaleAttention = true;
    await widget.client
        .sendExamEvent(
          token: widget.examToken,
          event: ExamClientEvents.staleConnectionEscalated(
            secondsSinceLastContact: DateTime.now()
                .difference(_lastServerContactAt!)
                .inSeconds,
            failureCount: _consecutiveSyncFailures,
          ),
        )
        .catchError((_) {});
  }

  void _markQuestionAudioPlayed(String questionId) {
    if (_playedAudioQuestionIds.contains(questionId)) {
      return;
    }
    setState(() {
      _playedAudioQuestionIds.add(questionId);
      _statusMessage = 'Audio soal telah diputar di perangkat ini.';
    });
    unawaited(_persistSnapshot());
  }

  bool _questionHasAudio(ExamQuestion question) {
    return question.stimulusAudioUrl.trim().isNotEmpty ||
        question.stemAudioUrl.trim().isNotEmpty;
  }

  bool _questionAudioPlayed(ExamQuestion question) {
    return _playedAudioQuestionIds.contains(question.id);
  }

  Map<String, String> _examAssetHeadersForUrl(String url) {
    return widget.client.examAssetHeadersForUrl(widget.examToken, url);
  }

  String _resolveExamAssetUrl(String url) {
    return widget.client.resolveAssetUrl(url);
  }

  Future<void> _finishExam({required bool wasAutoSubmitted}) async {
    if (!mounted) {
      return;
    }
    setState(() {
      _isSubmitted = true;
      _isSubmitPendingIntervention = false;
      _statusMessage =
          'Ujian sudah selesai dan akan diverifikasi pada layar akhir.';
      _serverNotice = null;
    });
    _countdownTimer?.cancel();
    _heartbeatTimer?.cancel();
    _textAutosaveTimer?.cancel();
    await _sessionStore.clearSnapshot();
    if (!mounted) {
      return;
    }
    Navigator.of(context).pushReplacement(
      MaterialPageRoute<void>(
        builder: (_) => ExamCompletedScreen(
          studentName: widget.initialPayload.student.nama,
          studentNis: widget.initialPayload.student.nis,
          sessionTitle: widget.initialPayload.session.title,
          roomName: widget.initialPayload.room?.roomName ?? '-',
          scheduledStartIso:
              widget.initialPayload.session.scheduledStart?.toIso8601String() ??
              '',
          scheduledEndIso:
              widget.initialPayload.session.scheduledEnd?.toIso8601String() ??
              '',
          durationMinutes: widget.initialPayload.session.durationMinutes,
          answeredCount: _answeredCount,
          totalQuestions: widget.initialPayload.totalQuestions,
          wasAutoSubmitted: wasAutoSubmitted,
        ),
      ),
    );
  }

  @override
  Widget build(BuildContext context) {
    final payload = widget.initialPayload;
    if (payload.questions.isEmpty) {
      return _buildEmptyExamPayloadScaffold(context, payload);
    }
    final currentQuestion = payload.questions[_currentQuestionIndex];
    final theme = Theme.of(context);
    final connection = _connectionState;

    return PopScope(
      canPop: false,
      onPopInvokedWithResult: (didPop, _) async {
        if (didPop || _isSubmitted) {
          return;
        }
        final messenger = ScaffoldMessenger.of(context);
        try {
          await widget.client.sendExamEvent(
            token: widget.examToken,
            event: ExamClientEvents.backButtonAttempt(),
          );
        } catch (_) {
          // Back blocking is a local safety control; telemetry must not break it.
        }
        if (!mounted) {
          return;
        }
        messenger.showSnackBar(
          const SnackBar(
            content: Text(
              'Tombol kembali dinonaktifkan selama ujian berlangsung.',
            ),
          ),
        );
      },
      child: Scaffold(
        appBar: AppBar(
          title: Text(payload.session.title),
          actions: [
            Padding(
              padding: const EdgeInsets.only(right: 8),
              child: Center(
                child: SyncStatusChip(
                  label: connection.syncStatusLabel,
                  tone: connection.syncStatusTone,
                ),
              ),
            ),
            IconButton(
              onPressed: () {
                Navigator.of(context).push(
                  MaterialPageRoute<void>(
                    builder: (_) => const ExamStatusGuideScreen(),
                  ),
                );
              },
              icon: const Icon(Icons.info_outline),
              tooltip: 'Panduan status',
            ),
            IconButton(
              onPressed: _isSyncingStatus ? null : _syncStatus,
              icon: _isSyncingStatus
                  ? const SizedBox(
                      width: 18,
                      height: 18,
                      child: CircularProgressIndicator(strokeWidth: 2),
                    )
                  : const Icon(Icons.sync),
              tooltip: 'Perbarui status',
            ),
          ],
        ),
        body: SafeArea(
          child: Stack(
            children: [
              LayoutBuilder(
                builder: (context, constraints) {
                  final wide = constraints.maxWidth >= 1080;
                  final sidePanel = _buildSidePanel(theme, payload, wide: wide);
                  final content = _buildQuestionArea(theme, currentQuestion);

                  return Padding(
                    padding: const EdgeInsets.all(16),
                    child: wide
                        ? Row(
                            crossAxisAlignment: CrossAxisAlignment.start,
                            children: [
                              SizedBox(width: 290, child: sidePanel),
                              const SizedBox(width: 16),
                              Expanded(child: content),
                            ],
                          )
                        : SingleChildScrollView(
                            child: Column(
                              children: [
                                sidePanel,
                                const SizedBox(height: 16),
                                SizedBox(
                                  height: constraints.maxHeight * 0.82,
                                  child: content,
                                ),
                              ],
                            ),
                          ),
                  );
                },
              ),
              if (_resumeCheckRequired || _isResumingExam)
                Positioned.fill(
                  child: ColoredBox(
                    color: Colors.black.withValues(alpha: 0.45),
                    child: Center(
                      child: ConstrainedBox(
                        constraints: const BoxConstraints(maxWidth: 420),
                        child: Card(
                          child: Padding(
                            padding: const EdgeInsets.all(24),
                            child: Column(
                              mainAxisSize: MainAxisSize.min,
                              crossAxisAlignment: CrossAxisAlignment.start,
                              children: [
                                Text(
                                  'Mode ujian diamankan',
                                  style: theme.textTheme.headlineSmall
                                      ?.copyWith(fontWeight: FontWeight.w800),
                                ),
                                const SizedBox(height: 10),
                                Text(
                                  _isResumingExam
                                      ? 'Sistem sedang memeriksa ulang status peserta dan mencoba menyinkronkan jawaban lokal.'
                                      : 'Aplikasi mendeteksi perpindahan dari mode ujian. Lanjutkan hanya jika pengawas mengizinkan.',
                                  style: theme.textTheme.bodyMedium?.copyWith(
                                    height: 1.5,
                                  ),
                                ),
                                if (_pendingAnswers.isNotEmpty) ...[
                                  const SizedBox(height: 12),
                                  Text(
                                    '${_pendingAnswers.length} jawaban lokal menunggu sinkron.',
                                    style: theme.textTheme.bodySmall?.copyWith(
                                      color: theme.colorScheme.primary,
                                      fontWeight: FontWeight.w700,
                                    ),
                                  ),
                                ],
                                const SizedBox(height: 18),
                                SizedBox(
                                  width: double.infinity,
                                  child: FilledButton.icon(
                                    onPressed: _isResumingExam
                                        ? null
                                        : _handleResumeCheck,
                                    icon: _isResumingExam
                                        ? const SizedBox(
                                            width: 18,
                                            height: 18,
                                            child: CircularProgressIndicator(
                                              strokeWidth: 2,
                                            ),
                                          )
                                        : const Icon(
                                            Icons.verified_user_outlined,
                                          ),
                                    label: Text(
                                      _isResumingExam
                                          ? 'Memeriksa status...'
                                          : 'Lanjutkan dengan pengecekan',
                                    ),
                                  ),
                                ),
                              ],
                            ),
                          ),
                        ),
                      ),
                    ),
                  ),
                ),
            ],
          ),
        ),
      ),
    );
  }

  Widget _buildEmptyExamPayloadScaffold(
    BuildContext context,
    ExamLoginPayload payload,
  ) {
    final theme = Theme.of(context);
    return PopScope(
      canPop: false,
      child: Scaffold(
        appBar: AppBar(title: Text(payload.session.title)),
        body: SafeArea(
          child: Center(
            child: ConstrainedBox(
              constraints: const BoxConstraints(maxWidth: 520),
              child: Card(
                child: Padding(
                  padding: const EdgeInsets.all(24),
                  child: Column(
                    mainAxisSize: MainAxisSize.min,
                    crossAxisAlignment: CrossAxisAlignment.start,
                    children: [
                      Text(
                        'Paket soal belum tersedia',
                        style: theme.textTheme.headlineSmall?.copyWith(
                          fontWeight: FontWeight.w800,
                        ),
                      ),
                      const SizedBox(height: 12),
                      Text(
                        'Server mengirim sesi ujian tanpa daftar soal. Jangan melakukan submit dari perangkat ini. Minta pengawas memeriksa paket ujian dan coba perbarui status setelah diperbaiki.',
                        style: theme.textTheme.bodyMedium?.copyWith(
                          height: 1.5,
                        ),
                      ),
                      const SizedBox(height: 18),
                      SizedBox(
                        width: double.infinity,
                        child: FilledButton.icon(
                          onPressed: _isSyncingStatus ? null : _syncStatus,
                          icon: _isSyncingStatus
                              ? const SizedBox(
                                  width: 18,
                                  height: 18,
                                  child: CircularProgressIndicator(
                                    strokeWidth: 2,
                                  ),
                                )
                              : const Icon(Icons.sync),
                          label: Text(
                            _isSyncingStatus
                                ? 'Memperbarui status...'
                                : 'Perbarui Status',
                          ),
                        ),
                      ),
                    ],
                  ),
                ),
              ),
            ),
          ),
        ),
      ),
    );
  }

  Widget _buildSidePanel(
    ThemeData theme,
    ExamLoginPayload payload, {
    required bool wide,
  }) {
    final connection = _connectionState;
    final questionGrid = GridView.builder(
      itemCount: payload.questions.length,
      gridDelegate: const SliverGridDelegateWithFixedCrossAxisCount(
        crossAxisCount: 4,
        crossAxisSpacing: 10,
        mainAxisSpacing: 10,
        childAspectRatio: 1,
      ),
      itemBuilder: (context, index) {
        final question = payload.questions[index];
        final selected = index == _currentQuestionIndex;
        final answered = _answers.containsKey(question.id);
        final hasAudio = _questionHasAudio(question);
        final audioPlayed = _questionAudioPlayed(question);
        final background = selected
            ? theme.colorScheme.primary
            : answered
            ? theme.colorScheme.primary.withValues(alpha: 0.12)
            : const Color(0xFFF1F4ED);
        final foreground = selected
            ? theme.colorScheme.onPrimary
            : answered
            ? theme.colorScheme.primary
            : theme.colorScheme.onSurface;

        return InkWell(
          onTap: () {
            setState(() {
              _currentQuestionIndex = index;
            });
            unawaited(_persistSnapshot());
          },
          borderRadius: BorderRadius.circular(16),
          child: Ink(
            decoration: BoxDecoration(
              color: background,
              borderRadius: BorderRadius.circular(16),
            ),
            child: Center(
              child: Column(
                mainAxisAlignment: MainAxisAlignment.center,
                children: [
                  Text(
                    '${index + 1}',
                    style: theme.textTheme.titleMedium?.copyWith(
                      color: foreground,
                      fontWeight: FontWeight.w800,
                    ),
                  ),
                  if (hasAudio) ...[
                    const SizedBox(height: 4),
                    Icon(
                      audioPlayed ? Icons.headset_mic : Icons.headset_off,
                      size: 16,
                      color: foreground,
                    ),
                  ],
                ],
              ),
            ),
          ),
        );
      },
    );

    return Card(
      child: SingleChildScrollView(
        child: Padding(
          padding: const EdgeInsets.all(20),
          child: Column(
            crossAxisAlignment: CrossAxisAlignment.start,
            children: [
              Text(
                payload.student.nama,
                style: theme.textTheme.titleLarge?.copyWith(
                  fontWeight: FontWeight.w800,
                ),
              ),
              const SizedBox(height: 4),
              Text('NIS ${payload.student.nis}'),
              const SizedBox(height: 4),
              Text('Ruang ${payload.room?.roomName ?? '-'}'),
              const SizedBox(height: 20),
              StatTile(
                label: 'Sisa waktu',
                value: ExamShellConnectionViewModel.formatDuration(
                  _timeRemainingSeconds,
                ),
                accent: _timeRemainingSeconds <= 300,
              ),
              const SizedBox(height: 12),
              StatTile(
                label: 'Progres',
                value: '$_answeredCount / ${payload.totalQuestions}',
              ),
              const SizedBox(height: 12),
              StatTile(
                label: 'Kontak server terakhir',
                value: connection.formatClock(_lastServerContactAt),
              ),
              const SizedBox(height: 12),
              ConnectionHealthCard(
                label: connection.connectionHealthTitle,
                description: connection.connectionHealthDescription,
                tone: connection.connectionHealthTone,
              ),
              if (_lastSyncFailureAt != null) ...[
                const SizedBox(height: 12),
                StatTile(
                  label: 'Gangguan terakhir',
                  value: connection.formatClock(_lastSyncFailureAt),
                  accent: true,
                ),
              ],
              const SizedBox(height: 20),
              if (_consecutiveSyncFailures >= 2) ...[
                ConnectionWarningCard(
                  failureCount: _consecutiveSyncFailures,
                  lastFailureAt: connection.formatClock(_lastSyncFailureAt),
                  onRetry: _isSyncingStatus ? null : _syncStatus,
                ),
                const SizedBox(height: 14),
              ],
              if (connection.needsSupervisorAttention) ...[
                SupervisorAttentionCard(
                  lastContactAt: connection.formatClock(_lastServerContactAt),
                  staleDuration: connection.staleAttentionDurationLabel,
                  escalationThreshold: connection.staleEscalationThresholdLabel,
                  escalated: connection.needsEscalatedSupervisorAttention,
                  onRetry: _isSyncingStatus ? null : _syncStatus,
                ),
                const SizedBox(height: 14),
              ],
              if (_isDegradedMode) ...[
                DegradedModeCard(
                  pendingCount: _pendingAnswers.length,
                  failureCount: _consecutiveSyncFailures,
                  onRetry: _isSyncingStatus ? null : _syncStatus,
                ),
                const SizedBox(height: 14),
              ],
              if (_pendingAnswers.isNotEmpty) ...[
                StatTile(
                  label: 'Jawaban lokal',
                  value: '${_pendingAnswers.length} menunggu sinkron',
                  accent: true,
                ),
                const SizedBox(height: 12),
              ],
              if (_serverNotice != null) ...[
                ExamGuidanceCard(notice: _serverNotice!),
                const SizedBox(height: 12),
              ],
              if (_isSubmitPendingIntervention) ...[
                InlineMessage(
                  tone: BannerTone.error,
                  message:
                      'Submit final sedang ditahan sampai semua jawaban lokal tersinkron ke server.',
                ),
                const SizedBox(height: 12),
              ],
              if (_statusMessage != null)
                InlineMessage(
                  tone: BannerTone.success,
                  message: _statusMessage!,
                ),
              if (_errorMessage != null) ...[
                if (_statusMessage != null) const SizedBox(height: 10),
                InlineMessage(tone: BannerTone.error, message: _errorMessage!),
              ],
              const SizedBox(height: 20),
              Text(
                'Navigasi soal',
                style: theme.textTheme.titleMedium?.copyWith(
                  fontWeight: FontWeight.w700,
                ),
              ),
              const SizedBox(height: 12),
              SizedBox(height: wide ? 156 : 88, child: questionGrid),
              const SizedBox(height: 16),
              SizedBox(
                width: double.infinity,
                child: FilledButton.icon(
                  onPressed: _isSubmitting || _isSubmitted || _isDegradedMode
                      ? null
                      : _submit,
                  icon: _isSubmitting
                      ? const SizedBox(
                          width: 18,
                          height: 18,
                          child: CircularProgressIndicator(strokeWidth: 2),
                        )
                      : Icon(
                          _isDegradedMode ? Icons.sync_problem : Icons.task_alt,
                        ),
                  label: Text(
                    _isSubmitted
                        ? 'Ujian Terkirim'
                        : _isSubmitPendingIntervention
                        ? 'Coba kirim ulang setelah sinkron'
                        : _isDegradedMode
                        ? 'Kirim ditahan saat koneksi menurun'
                        : 'Kirim Ujian',
                  ),
                ),
              ),
            ],
          ),
        ),
      ),
    );
  }

  Widget _buildQuestionArea(ThemeData theme, ExamQuestion question) {
    final hasAudio = _questionHasAudio(question);
    final audioPlayed = _questionAudioPlayed(question);
    return Card(
      child: Padding(
        padding: const EdgeInsets.all(22),
        child: Column(
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            if (_errorMessage != null) ...[
              InlineMessage(tone: BannerTone.error, message: _errorMessage!),
              const SizedBox(height: 14),
            ] else if (_statusMessage != null) ...[
              InlineMessage(tone: BannerTone.info, message: _statusMessage!),
              const SizedBox(height: 14),
            ],
            Text(
              'Soal ${_currentQuestionIndex + 1}',
              style: theme.textTheme.labelLarge?.copyWith(
                color: theme.colorScheme.primary,
                fontWeight: FontWeight.w800,
              ),
            ),
            if (hasAudio) ...[
              const SizedBox(height: 10),
              QuestionAudioStatusChip(hasPlayed: audioPlayed),
            ],
            const SizedBox(height: 10),
            if (question.stimulusHtml.trim().isNotEmpty) ...[
              Container(
                width: double.infinity,
                padding: const EdgeInsets.all(16),
                decoration: BoxDecoration(
                  color: const Color(0xFFF6F8F3),
                  borderRadius: BorderRadius.circular(18),
                  border: Border.all(color: theme.colorScheme.outlineVariant),
                ),
                child: RichExamText(
                  content: question.stimulusHtml,
                  style: theme.textTheme.bodyLarge?.copyWith(height: 1.55),
                ),
              ),
              const SizedBox(height: 16),
            ],
            if (question.stimulusMediaUrl.trim().isNotEmpty) ...[
              QuestionMediaCard(
                url: _resolveExamAssetUrl(question.stimulusMediaUrl),
                headers: _examAssetHeadersForUrl(question.stimulusMediaUrl),
              ),
              const SizedBox(height: 16),
            ],
            if (question.stimulusAudioUrl.trim().isNotEmpty) ...[
              AudioPromptCard(
                url: _resolveExamAssetUrl(question.stimulusAudioUrl),
                label: 'Audio stimulus',
                headers: _examAssetHeadersForUrl(question.stimulusAudioUrl),
                hasBeenPlayed: audioPlayed,
                onPlayed: () => _markQuestionAudioPlayed(question.id),
              ),
              const SizedBox(height: 16),
            ],
            RichExamText(
              content: question.stemHtml.trim().isNotEmpty
                  ? question.stemHtml
                  : question.questionText.isEmpty
                  ? 'Soal belum memiliki teks.'
                  : question.questionText,
              style: theme.textTheme.titleLarge?.copyWith(
                height: 1.45,
                fontWeight: FontWeight.w700,
              ),
            ),
            if (question.stemMediaUrl.trim().isNotEmpty) ...[
              const SizedBox(height: 16),
              QuestionMediaCard(
                url: _resolveExamAssetUrl(question.stemMediaUrl),
                headers: _examAssetHeadersForUrl(question.stemMediaUrl),
              ),
            ],
            if (question.stemAudioUrl.trim().isNotEmpty) ...[
              const SizedBox(height: 16),
              AudioPromptCard(
                url: _resolveExamAssetUrl(question.stemAudioUrl),
                label: 'Audio soal',
                headers: _examAssetHeadersForUrl(question.stemAudioUrl),
                hasBeenPlayed: audioPlayed,
                onPlayed: () => _markQuestionAudioPlayed(question.id),
              ),
            ],
            const SizedBox(height: 22),
            Expanded(
              child: question.isTextAnswer
                  ? _buildTextAnswerQuestion(theme, question)
                  : question.isOrdering
                  ? _buildOrderingQuestion(theme, question)
                  : question.isMatching
                  ? _buildMatchingQuestion(theme, question)
                  : _buildObjectiveQuestion(theme, question),
            ),
            const SizedBox(height: 16),
            Row(
              children: [
                OutlinedButton.icon(
                  onPressed: _currentQuestionIndex == 0
                      ? null
                      : () {
                          setState(() {
                            _currentQuestionIndex -= 1;
                          });
                          unawaited(_persistSnapshot());
                        },
                  icon: const Icon(Icons.chevron_left),
                  label: const Text('Sebelumnya'),
                ),
                const SizedBox(width: 12),
                FilledButton.icon(
                  onPressed:
                      _currentQuestionIndex ==
                          widget.initialPayload.questions.length - 1
                      ? null
                      : () {
                          setState(() {
                            _currentQuestionIndex += 1;
                          });
                          unawaited(_persistSnapshot());
                        },
                  icon: const Icon(Icons.chevron_right),
                  label: const Text('Berikutnya'),
                ),
                const Spacer(),
                if (_isSavingAnswer)
                  Row(
                    children: [
                      const SizedBox(
                        width: 16,
                        height: 16,
                        child: CircularProgressIndicator(strokeWidth: 2),
                      ),
                      const SizedBox(width: 8),
                      Text('Menyimpan...', style: theme.textTheme.bodySmall),
                    ],
                  ),
              ],
            ),
          ],
        ),
      ),
    );
  }

  List<String> _selectedOptionLabels(ExamQuestion question) {
    final raw = _answers[question.id] ?? '';
    return raw
        .split(',')
        .map((label) => label.trim())
        .where((label) => label.isNotEmpty)
        .toSet()
        .toList()
      ..sort();
  }

  Map<String, String> _selectedMatchingPairs(ExamQuestion question) {
    return _selectedMatchingPairsFromRaw(_answers[question.id] ?? '');
  }

  String _encodeMatchingPairs(Map<String, String> pairs) {
    final entries = pairs.entries.toList()
      ..sort((a, b) => a.key.compareTo(b.key));
    return entries.map((entry) => '${entry.key}=${entry.value}').join(';');
  }

  bool _isAnswerComplete(ExamQuestion question, String answer) {
    if (answer.trim().isEmpty) {
      return false;
    }
    if (question.isOrdering) {
      return _isOrderingAnswerComplete(question, answer);
    }
    if (!question.isMatching) {
      return true;
    }
    final selected = _selectedMatchingPairsFromRaw(answer);
    final requiredLabels = question.options
        .where((option) => !option.isDistractor)
        .map((option) => option.label)
        .toSet();
    return requiredLabels.isNotEmpty &&
        requiredLabels.every(
          (label) => selected[label]?.trim().isNotEmpty ?? false,
        );
  }

  Map<String, String> _selectedMatchingPairsFromRaw(String raw) {
    final pairs = <String, String>{};
    for (final part in raw.split(';')) {
      final pieces = part.split('=');
      if (pieces.length != 2) {
        continue;
      }
      final left = pieces[0].trim();
      final right = pieces[1].trim();
      if (left.isEmpty || right.isEmpty) {
        continue;
      }
      pairs[left] = right;
    }
    return pairs;
  }

  List<String> _selectedOrderingLabelsFromRaw(String raw) {
    return raw
        .split(',')
        .map((label) => label.trim())
        .where((label) => label.isNotEmpty)
        .toList();
  }

  List<String> _orderingOptionLabels(ExamQuestion question) {
    return question.options
        .map((option) => option.label.trim())
        .where((label) => label.isNotEmpty)
        .toList();
  }

  bool _isOrderingAnswerComplete(ExamQuestion question, String answer) {
    final requiredLabels = _orderingOptionLabels(question);
    final requiredSet = requiredLabels.toSet();
    if (requiredLabels.isEmpty || requiredSet.length != requiredLabels.length) {
      return false;
    }

    final selectedLabels = _selectedOrderingLabelsFromRaw(answer);
    final selectedSet = selectedLabels.toSet();
    return selectedLabels.length == requiredLabels.length &&
        selectedSet.length == selectedLabels.length &&
        requiredSet.every(selectedSet.contains);
  }

  List<ExamOption> _orderedOptions(ExamQuestion question) {
    final optionsByLabel = <String, ExamOption>{};
    for (final option in question.options) {
      final label = option.label.trim();
      if (label.isNotEmpty && !optionsByLabel.containsKey(label)) {
        optionsByLabel[label] = option;
      }
    }

    final orderedOptions = <ExamOption>[];
    final usedLabels = <String>{};
    for (final label in _selectedOrderingLabelsFromRaw(
      _answers[question.id] ?? '',
    )) {
      final option = optionsByLabel[label];
      if (option != null && usedLabels.add(label)) {
        orderedOptions.add(option);
      }
    }

    for (final option in question.options) {
      final label = option.label.trim();
      if (label.isNotEmpty && usedLabels.add(label)) {
        orderedOptions.add(option);
      }
    }
    return orderedOptions;
  }

  String _encodeOrderingOptions(List<ExamOption> options) {
    return options
        .map((option) => option.label.trim())
        .where((label) => label.isNotEmpty)
        .join(',');
  }

  Future<void> _moveOrderingOption(
    ExamQuestion question,
    int fromIndex,
    int delta,
  ) async {
    final orderedOptions = _orderedOptions(question);
    final targetIndex = fromIndex + delta;
    if (fromIndex < 0 ||
        fromIndex >= orderedOptions.length ||
        targetIndex < 0 ||
        targetIndex >= orderedOptions.length) {
      return;
    }

    final moved = orderedOptions.removeAt(fromIndex);
    orderedOptions.insert(targetIndex, moved);
    await _selectOption(question, _encodeOrderingOptions(orderedOptions));
  }

  Future<void> _selectMatchingPair(
    ExamQuestion question,
    String leftLabel,
    String rightLabel,
  ) async {
    final selected = _selectedMatchingPairs(question);
    selected[leftLabel] = rightLabel;
    await _selectOption(question, _encodeMatchingPairs(selected));
  }

  Future<void> _toggleObjectiveOption(
    ExamQuestion question,
    String label,
  ) async {
    if (!question.isMultipleAnswer) {
      await _selectOption(question, label);
      return;
    }
    final selected = _selectedOptionLabels(question);
    if (selected.contains(label)) {
      if (selected.length == 1) {
        setState(() {
          _statusMessage = 'Pilih minimal satu opsi jawaban.';
        });
        return;
      }
      selected.remove(label);
    } else {
      selected.add(label);
    }
    selected.sort();
    await _selectOption(question, selected.join(','));
  }

  Widget _buildObjectiveQuestion(ThemeData theme, ExamQuestion question) {
    return ListView.separated(
      itemCount: question.options.length,
      separatorBuilder: (_, _) => const SizedBox(height: 12),
      itemBuilder: (context, index) {
        final option = question.options[index];
        final selected = question.isMultipleAnswer
            ? _selectedOptionLabels(question).contains(option.label)
            : _answers[question.id] == option.label;

        return InkWell(
          onTap: () => _toggleObjectiveOption(question, option.label),
          borderRadius: BorderRadius.circular(18),
          child: Ink(
            padding: const EdgeInsets.all(16),
            decoration: BoxDecoration(
              color: selected
                  ? theme.colorScheme.primary.withValues(alpha: 0.12)
                  : const Color(0xFFF6F8F3),
              borderRadius: BorderRadius.circular(18),
              border: Border.all(
                color: selected
                    ? theme.colorScheme.primary
                    : theme.colorScheme.outlineVariant,
              ),
            ),
            child: Row(
              crossAxisAlignment: CrossAxisAlignment.start,
              children: [
                CircleAvatar(
                  radius: 18,
                  backgroundColor: selected
                      ? theme.colorScheme.primary
                      : Colors.white,
                  foregroundColor: selected
                      ? theme.colorScheme.onPrimary
                      : theme.colorScheme.primary,
                  child: Text(option.label),
                ),
                const SizedBox(width: 14),
                Expanded(
                  child: RichExamText(
                    content: option.text,
                    style: theme.textTheme.bodyLarge?.copyWith(height: 1.5),
                  ),
                ),
                if (question.isMultipleAnswer)
                  Checkbox(
                    value: selected,
                    onChanged: (_) =>
                        _toggleObjectiveOption(question, option.label),
                  ),
              ],
            ),
          ),
        );
      },
    );
  }

  Widget _buildMatchingQuestion(ThemeData theme, ExamQuestion question) {
    final selected = _selectedMatchingPairs(question);
    final rightOptions = question.options
        .where((option) => option.matchLabel.trim().isNotEmpty)
        .toList();
    final leftOptions = question.options
        .where((option) => !option.isDistractor)
        .toList();

    return ListView(
      children: [
        Container(
          width: double.infinity,
          padding: const EdgeInsets.all(14),
          decoration: BoxDecoration(
            color: const Color(0xFFF6F8F3),
            borderRadius: BorderRadius.circular(18),
            border: Border.all(color: theme.colorScheme.outlineVariant),
          ),
          child: Column(
            crossAxisAlignment: CrossAxisAlignment.start,
            children: [
              Text(
                'Pilihan pasangan',
                style: theme.textTheme.labelLarge?.copyWith(
                  color: theme.colorScheme.primary,
                  fontWeight: FontWeight.w800,
                ),
              ),
              const SizedBox(height: 10),
              ...rightOptions.map(
                (option) => Padding(
                  padding: const EdgeInsets.only(bottom: 8),
                  child: Row(
                    crossAxisAlignment: CrossAxisAlignment.start,
                    children: [
                      SizedBox(
                        width: 28,
                        child: Text(
                          '${option.matchLabel}.',
                          style: theme.textTheme.bodyMedium?.copyWith(
                            fontWeight: FontWeight.w800,
                          ),
                        ),
                      ),
                      Expanded(
                        child: Text(
                          option.matchText,
                          style: theme.textTheme.bodyMedium?.copyWith(
                            height: 1.4,
                          ),
                        ),
                      ),
                    ],
                  ),
                ),
              ),
            ],
          ),
        ),
        const SizedBox(height: 16),
        ...leftOptions.map(
          (option) => Padding(
            padding: const EdgeInsets.only(bottom: 14),
            child: Container(
              padding: const EdgeInsets.all(14),
              decoration: BoxDecoration(
                color: Colors.white,
                borderRadius: BorderRadius.circular(18),
                border: Border.all(color: theme.colorScheme.outlineVariant),
              ),
              child: Column(
                crossAxisAlignment: CrossAxisAlignment.start,
                children: [
                  Row(
                    crossAxisAlignment: CrossAxisAlignment.start,
                    children: [
                      CircleAvatar(
                        radius: 17,
                        backgroundColor: theme.colorScheme.primary,
                        foregroundColor: theme.colorScheme.onPrimary,
                        child: Text(option.label),
                      ),
                      const SizedBox(width: 12),
                      Expanded(
                        child: RichExamText(
                          content: option.text,
                          style: theme.textTheme.bodyLarge?.copyWith(
                            height: 1.45,
                            fontWeight: FontWeight.w600,
                          ),
                        ),
                      ),
                    ],
                  ),
                  const SizedBox(height: 12),
                  Builder(
                    builder: (context) {
                      final currentValue =
                          rightOptions.any(
                            (right) =>
                                right.matchLabel == selected[option.label],
                          )
                          ? selected[option.label]
                          : null;
                      return DropdownButtonFormField<String>(
                        initialValue: currentValue,
                        decoration: const InputDecoration(
                          labelText: 'Pilih pasangan kanan',
                          border: OutlineInputBorder(),
                        ),
                        items: rightOptions
                            .map(
                              (right) => DropdownMenuItem<String>(
                                value: right.matchLabel,
                                child: Text(
                                  '${right.matchLabel}. ${right.matchText}',
                                  overflow: TextOverflow.ellipsis,
                                ),
                              ),
                            )
                            .toList(),
                        onChanged: _isSubmitted
                            ? null
                            : (value) {
                                if (value == null) {
                                  return;
                                }
                                _selectMatchingPair(
                                  question,
                                  option.label,
                                  value,
                                );
                              },
                      );
                    },
                  ),
                ],
              ),
            ),
          ),
        ),
      ],
    );
  }

  Widget _buildOrderingQuestion(ThemeData theme, ExamQuestion question) {
    final orderedOptions = _orderedOptions(question);

    return ListView(
      children: [
        Container(
          width: double.infinity,
          padding: const EdgeInsets.all(14),
          decoration: BoxDecoration(
            color: const Color(0xFFF6F8F3),
            borderRadius: BorderRadius.circular(18),
            border: Border.all(color: theme.colorScheme.outlineVariant),
          ),
          child: Column(
            crossAxisAlignment: CrossAxisAlignment.start,
            children: [
              Text(
                'Urutan jawaban',
                style: theme.textTheme.labelLarge?.copyWith(
                  color: theme.colorScheme.primary,
                  fontWeight: FontWeight.w800,
                ),
              ),
              const SizedBox(height: 6),
              Text(
                'Susun pilihan dari atas ke bawah sesuai urutan jawaban.',
                style: theme.textTheme.bodyMedium?.copyWith(height: 1.4),
              ),
            ],
          ),
        ),
        const SizedBox(height: 12),
        SizedBox(
          width: double.infinity,
          child: FilledButton.icon(
            onPressed: _isSavingAnswer || _isSubmitted
                ? null
                : () => _selectOption(
                    question,
                    _encodeOrderingOptions(_orderedOptions(question)),
                  ),
            icon: const Icon(Icons.save_outlined),
            label: const Text('Simpan urutan saat ini'),
          ),
        ),
        const SizedBox(height: 16),
        ...List.generate(orderedOptions.length, (index) {
          final option = orderedOptions[index];
          final label = option.label.trim();
          return Padding(
            padding: const EdgeInsets.only(bottom: 12),
            child: Container(
              padding: const EdgeInsets.all(14),
              decoration: BoxDecoration(
                color: Colors.white,
                borderRadius: BorderRadius.circular(18),
                border: Border.all(color: theme.colorScheme.outlineVariant),
              ),
              child: Row(
                crossAxisAlignment: CrossAxisAlignment.start,
                children: [
                  CircleAvatar(
                    radius: 18,
                    backgroundColor: theme.colorScheme.primary,
                    foregroundColor: theme.colorScheme.onPrimary,
                    child: Text('${index + 1}'),
                  ),
                  const SizedBox(width: 12),
                  CircleAvatar(
                    radius: 18,
                    backgroundColor: const Color(0xFFF6F8F3),
                    foregroundColor: theme.colorScheme.primary,
                    child: Text(label),
                  ),
                  const SizedBox(width: 14),
                  Expanded(
                    child: RichExamText(
                      content: option.text,
                      style: theme.textTheme.bodyLarge?.copyWith(height: 1.45),
                    ),
                  ),
                  const SizedBox(width: 10),
                  Column(
                    mainAxisSize: MainAxisSize.min,
                    children: [
                      IconButton(
                        tooltip: 'Naikkan $label',
                        onPressed: _isSubmitted || index == 0
                            ? null
                            : () => _moveOrderingOption(question, index, -1),
                        icon: const Icon(Icons.keyboard_arrow_up),
                      ),
                      IconButton(
                        tooltip: 'Turunkan $label',
                        onPressed:
                            _isSubmitted || index == orderedOptions.length - 1
                            ? null
                            : () => _moveOrderingOption(question, index, 1),
                        icon: const Icon(Icons.keyboard_arrow_down),
                      ),
                    ],
                  ),
                ],
              ),
            ),
          );
        }),
      ],
    );
  }

  Widget _buildTextAnswerQuestion(ThemeData theme, ExamQuestion question) {
    final controller = _essayControllers[_currentQuestionIndex];
    final savedValue = _answers[question.id];
    if (savedValue != null && controller.text != savedValue) {
      controller.text = savedValue;
      controller.selection = TextSelection.collapsed(
        offset: controller.text.length,
      );
    }

    if (question.isShortAnswer) {
      return ListView(
        children: [
          TextField(
            controller: controller,
            textInputAction: TextInputAction.done,
            maxLength: kShortAnswerMaxChars,
            decoration: const InputDecoration(
              labelText: 'Jawaban singkat',
              hintText: 'Tulis jawaban singkat Anda...',
            ),
            onSubmitted: (_) => _saveTextAnswer(),
            onChanged: (value) => _autosaveTextAnswer(question, value),
          ),
          const SizedBox(height: 14),
          FilledButton.icon(
            onPressed: _isSavingAnswer || _isSubmitted ? null : _saveTextAnswer,
            icon: const Icon(Icons.save_outlined),
            label: const Text('Simpan Jawaban'),
          ),
        ],
      );
    }

    return Column(
      crossAxisAlignment: CrossAxisAlignment.start,
      children: [
        Expanded(
          child: TextField(
            controller: controller,
            maxLines: null,
            expands: true,
            maxLength: kEssayAnswerMaxChars,
            decoration: const InputDecoration(
              alignLabelWithHint: true,
              labelText: 'Jawaban uraian',
              hintText: 'Tulis jawaban Anda di sini...',
            ),
            onChanged: (value) => _autosaveTextAnswer(question, value),
          ),
        ),
        const SizedBox(height: 14),
        FilledButton.icon(
          onPressed: _isSavingAnswer || _isSubmitted ? null : _saveTextAnswer,
          icon: const Icon(Icons.save_outlined),
          label: const Text('Simpan Jawaban'),
        ),
      ],
    );
  }
}
