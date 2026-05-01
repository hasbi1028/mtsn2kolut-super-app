import 'dart:async';

import 'package:flutter/material.dart';

import '../exam_api.dart';
import '../exam_error_messages.dart';
import '../exam_session_store.dart';
import '../models.dart';
import 'exam_completed_screen.dart';
import 'exam_shell_connection.dart';
import 'exam_status_guide_screen.dart';
import 'exam_shell_widgets.dart';
import '../widgets/audio_prompt_card.dart';
import '../widgets/rich_exam_text.dart';

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
    _currentQuestionIndex = widget.restoredSnapshot?.currentQuestionIndex ?? 0;
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
      if (question.isEssay) {
        _essayControllers[i].text = _answers[question.id] ?? '';
      }
    }
    if (widget.autoStartRuntime) {
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
        widget.client.sendEvent(
          token: widget.examToken,
          eventType: 'app_switch',
          data: <String, Object?>{
            'state': state.name,
            'device_fingerprint': widget.deviceFingerprint,
          },
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

  Future<void> _syncStatus() async {
    setState(() {
      _isSyncingStatus = true;
    });
    try {
      final status = await widget.client.getStatus(widget.examToken);
      if (!mounted) {
        return;
      }
      setState(() {
        _answeredCount = status.answeredCount;
        _timeRemainingSeconds = status.timeRemainingSeconds;
        _isSubmitted = status.isSubmitted;
      });
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
        _errorMessage = 'Status server belum bisa diperbarui.';
      });
      unawaited(_handleConnectionAttentionSignals());
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
        .sendEvent(
          token: widget.examToken,
          eventType: 'warning',
          data: <String, Object?>{
            'reason': 'resume_exam',
            'resume_attempt_count': _resumeAttemptCount + 1,
          },
        )
        .catchError((_) {});
    _resumeAttemptCount += 1;
    if (_resumeAttemptCount > 1) {
      await widget.client
          .sendEvent(
            token: widget.examToken,
            eventType: 'warning',
            data: <String, Object?>{
              'reason': 'repeat_resume_attempt',
              'resume_attempt_count': _resumeAttemptCount,
            },
          )
          .catchError((_) {});
    }
    await _syncStatus();
    if (!mounted) {
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

  Future<void> _flushPendingAnswers() async {
    if (_pendingAnswers.isEmpty || _isSubmitted) {
      return;
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
      return;
    }

    setState(() {
      _statusMessage = _pendingAnswers.isEmpty
          ? 'Semua jawaban lokal berhasil disinkronkan ke server.'
          : '$syncedCount jawaban lokal berhasil disinkronkan. Sisanya akan dicoba lagi.';
    });
  }

  Future<void> _selectOption(ExamQuestion question, String answer) async {
    if (_isSubmitted) {
      return;
    }
    final wasAnswered = _answers.containsKey(question.id);
    setState(() {
      _answers[question.id] = answer;
      if (!wasAnswered) {
        _answeredCount += 1;
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
      setState(() {
        _consecutiveSyncFailures += 1;
        _lastSyncFailureAt = DateTime.now();
        _pendingAnswers[question.id] = answer;
        _errorMessage = null;
        _statusMessage = answerFailureMessage(error);
        _serverNotice = answerFailureNotice(error);
      });
      await widget.client
          .sendEvent(
            token: widget.examToken,
            eventType: 'warning',
            data: <String, Object?>{
              'reason': 'answer_saved_local_only',
              'question_id': question.id,
            },
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

  Future<void> _saveEssayAnswer() async {
    final question = widget.initialPayload.questions[_currentQuestionIndex];
    final answer = _essayControllers[_currentQuestionIndex].text.trim();

    if (answer.isEmpty) {
      setState(() {
        _errorMessage = 'Isi jawaban uraian terlebih dahulu.';
      });
      return;
    }

    await _selectOption(question, answer);
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
    if (_isSubmitting || _isSubmitted) {
      return;
    }

    if (_pendingAnswers.isNotEmpty) {
      await _flushPendingAnswers();
    }

    if (!autoSubmit && _pendingAnswers.isNotEmpty) {
      setState(() {
        _errorMessage =
            'Masih ada jawaban yang belum tersinkron ke server. Tunggu koneksi stabil lalu coba kirim lagi.';
      });
      await widget.client
          .sendEvent(
            token: widget.examToken,
            eventType: 'warning',
            data: <String, Object?>{
              'reason': 'submit_blocked_pending_sync',
              'pending_count': _pendingAnswers.length,
            },
          )
          .catchError((_) {});
      return;
    }

    if (!autoSubmit && _isDegradedMode) {
      setState(() {
        _errorMessage =
            'Mode koneksi menurun sedang aktif. Perbarui status dan tunggu sinkron pulih sebelum mengirim ujian.';
      });
      await widget.client
          .sendEvent(
            token: widget.examToken,
            eventType: 'warning',
            data: <String, Object?>{
              'reason': 'submit_blocked_degraded_mode',
              'failure_count': _consecutiveSyncFailures,
            },
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
      setState(() {
        _isSubmitted = true;
        _statusMessage = 'Ujian berhasil dikirim.';
        _serverNotice = null;
      });
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
                widget.initialPayload.session.scheduledStart
                    ?.toIso8601String() ??
                '',
            scheduledEndIso:
                widget.initialPayload.session.scheduledEnd?.toIso8601String() ??
                '',
            durationMinutes: widget.initialPayload.session.durationMinutes,
            answeredCount: _answeredCount,
            totalQuestions: widget.initialPayload.totalQuestions,
            wasAutoSubmitted: autoSubmit,
          ),
        ),
      );
      if (!autoSubmit) {
        await widget.client.sendEvent(
          token: widget.examToken,
          eventType: 'warning',
          data: const <String, Object?>{'reason': 'manual_submit'},
        );
      }
    } on ExamApiException catch (error) {
      if (!mounted) {
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
        .sendEvent(
          token: widget.examToken,
          eventType: 'warning',
          data: <String, Object?>{
            'reason': 'degraded_mode_entered',
            'failure_count': _consecutiveSyncFailures,
          },
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
        .sendEvent(
          token: widget.examToken,
          eventType: 'warning',
          data: <String, Object?>{
            'reason': 'stale_connection_attention',
            'seconds_since_last_contact': DateTime.now()
                .difference(_lastServerContactAt!)
                .inSeconds,
            'failure_count': _consecutiveSyncFailures,
          },
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
        .sendEvent(
          token: widget.examToken,
          eventType: 'warning',
          data: <String, Object?>{
            'reason': 'stale_connection_escalated',
            'seconds_since_last_contact': DateTime.now()
                .difference(_lastServerContactAt!)
                .inSeconds,
            'failure_count': _consecutiveSyncFailures,
          },
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

  @override
  Widget build(BuildContext context) {
    final payload = widget.initialPayload;
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
        await widget.client.sendEvent(
          token: widget.examToken,
          eventType: 'warning',
          data: const <String, Object?>{'reason': 'back_button_attempt'},
        );
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
              QuestionMediaCard(url: question.stimulusMediaUrl),
              const SizedBox(height: 16),
            ],
            if (question.stimulusAudioUrl.trim().isNotEmpty) ...[
              AudioPromptCard(
                url: question.stimulusAudioUrl,
                label: 'Audio stimulus',
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
              QuestionMediaCard(url: question.stemMediaUrl),
            ],
            if (question.stemAudioUrl.trim().isNotEmpty) ...[
              const SizedBox(height: 16),
              AudioPromptCard(
                url: question.stemAudioUrl,
                label: 'Audio soal',
                hasBeenPlayed: audioPlayed,
                onPlayed: () => _markQuestionAudioPlayed(question.id),
              ),
            ],
            const SizedBox(height: 22),
            Expanded(
              child: question.isEssay
                  ? _buildEssayQuestion(theme, question)
                  : _buildMultipleChoiceQuestion(theme, question),
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

  Widget _buildMultipleChoiceQuestion(ThemeData theme, ExamQuestion question) {
    return ListView.separated(
      itemCount: question.options.length,
      separatorBuilder: (_, _) => const SizedBox(height: 12),
      itemBuilder: (context, index) {
        final option = question.options[index];
        final selectedAnswer = _answers[question.id];
        final selected = selectedAnswer == option.label;

        return InkWell(
          onTap: () => _selectOption(question, option.label),
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
              ],
            ),
          ),
        );
      },
    );
  }

  Widget _buildEssayQuestion(ThemeData theme, ExamQuestion question) {
    final controller = _essayControllers[_currentQuestionIndex];
    final savedValue = _answers[question.id];
    if (savedValue != null && controller.text != savedValue) {
      controller.text = savedValue;
      controller.selection = TextSelection.collapsed(
        offset: controller.text.length,
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
            decoration: const InputDecoration(
              alignLabelWithHint: true,
              labelText: 'Jawaban uraian',
              hintText: 'Tulis jawaban Anda di sini...',
            ),
          ),
        ),
        const SizedBox(height: 14),
        FilledButton.icon(
          onPressed: _isSavingAnswer || _isSubmitted ? null : _saveEssayAnswer,
          icon: const Icon(Icons.save_outlined),
          label: const Text('Simpan Jawaban'),
        ),
      ],
    );
  }
}
