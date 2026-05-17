import 'package:flutter/material.dart';

import '../device_fingerprint.dart';
import '../exam_api.dart';
import '../exam_error_messages.dart';
import '../exam_format.dart';
import '../exam_session_store.dart';
import 'exam_shell_screen.dart';
import 'exam_restore_failed_screen.dart';

class ExamLoginScreen extends StatefulWidget {
  const ExamLoginScreen({
    super.key,
    this.autoRestore = true,
    this.initialErrorMessage,
    this.initialErrorNotice,
    this.previewSnapshot,
    this.sessionStore,
    this.deviceFingerprintStore,
  });

  final bool autoRestore;
  final String? initialErrorMessage;
  final ExamGuidanceNotice? initialErrorNotice;
  final ExamSessionSnapshot? previewSnapshot;
  final ExamSessionStore? sessionStore;
  final DeviceFingerprintStore? deviceFingerprintStore;

  @override
  State<ExamLoginScreen> createState() => _ExamLoginScreenState();
}

class _ExamLoginScreenState extends State<ExamLoginScreen> {
  late final ExamSessionStore _sessionStore;
  late final DeviceFingerprintStore _deviceFingerprintStore;
  final _tokenController = TextEditingController();
  final _roomTokenController = TextEditingController();
  final _baseUrlController = TextEditingController(
    text: const String.fromEnvironment(
      'API_BASE_URL',
      defaultValue: 'http://10.0.2.2:8080',
    ),
  );

  bool _isSubmitting = false;
  bool _isRestoring = true;
  bool _showOperatorSettings = false;
  String? _errorMessage;
  ExamGuidanceNotice? _errorNotice;

  @override
  void initState() {
    super.initState();
    _sessionStore = widget.sessionStore ?? ExamSessionStore();
    _deviceFingerprintStore =
        widget.deviceFingerprintStore ?? DeviceFingerprintStore();
    _errorMessage = widget.initialErrorMessage;
    _errorNotice = widget.initialErrorNotice;
    _baseUrlController.addListener(_refreshOperatorBaseUrlGuidance);
    if (widget.autoRestore) {
      _restoreExamSession();
    } else {
      _isRestoring = false;
    }
  }

  @override
  void dispose() {
    _baseUrlController.removeListener(_refreshOperatorBaseUrlGuidance);
    _tokenController.dispose();
    _roomTokenController.dispose();
    _baseUrlController.dispose();
    super.dispose();
  }

  void _refreshOperatorBaseUrlGuidance() {
    if (!_showOperatorSettings || !mounted) {
      return;
    }
    setState(() {});
  }

  Future<void> _restoreExamSession() async {
    final storedBaseUrl = await _sessionStore.loadBaseUrl();
    if (storedBaseUrl != null && storedBaseUrl.isNotEmpty) {
      _baseUrlController.text = storedBaseUrl;
    }

    final snapshot = await _sessionStore.loadSnapshot();
    if (!mounted) {
      return;
    }

    if (snapshot == null) {
      setState(() {
        _isRestoring = false;
      });
      return;
    }

    setState(() {
      _isSubmitting = true;
      _isRestoring = true;
      _errorMessage = null;
      _errorNotice = null;
    });

    try {
      final client = ExamApiClient(
        baseUrl: snapshot.baseUrl,
        deviceFingerprint: snapshot.deviceFingerprint,
      );
      final payload = await client.login(
        token: snapshot.examToken,
        roomToken: snapshot.roomToken,
        deviceFingerprint: snapshot.deviceFingerprint,
      );
      if (!mounted) {
        return;
      }
      _baseUrlController.text = snapshot.baseUrl;
      _tokenController.text = snapshot.examToken.toUpperCase();
      _roomTokenController.text = snapshot.roomToken.toUpperCase();
      await Navigator.of(context).push(
        MaterialPageRoute<void>(
          builder: (_) => ExamShellScreen(
            client: client,
            examToken: snapshot.examToken,
            roomToken: snapshot.roomToken,
            initialPayload: payload,
            deviceFingerprint: snapshot.deviceFingerprint,
            restoredSnapshot: snapshot,
            initialResumeCheckRequired: true,
            sessionStore: _sessionStore,
          ),
        ),
      );
    } on ExamApiException catch (error) {
      if (shouldClearSnapshotAfterRestoreFailure(error)) {
        await _sessionStore.clearSnapshot();
      }
      if (!mounted) {
        return;
      }
      await Navigator.of(context).push(
        MaterialPageRoute<void>(
          builder: (_) => ExamRestoreFailedScreen(
            snapshot: snapshot,
            message: restoreFailureMessage(error),
            notice: restoreFailureNotice(error),
          ),
        ),
      );
    } catch (_) {
      if (!mounted) {
        return;
      }
      await Navigator.of(context).push(
        MaterialPageRoute<void>(
          builder: (_) => ExamRestoreFailedScreen(
            snapshot: snapshot,
            message:
                'Koneksi ke server gagal saat mencoba memulihkan sesi. Pastikan jaringan stabil lalu login ulang dengan token aktif jika diperlukan.',
          ),
        ),
      );
    } finally {
      if (mounted) {
        setState(() {
          _isSubmitting = false;
          _isRestoring = false;
        });
      }
    }
  }

  Future<void> _submit() async {
    final token = _tokenController.text.trim().toLowerCase();
    final roomToken = _roomTokenController.text.trim();
    final baseUrl = _baseUrlController.text.trim();

    if (token.length < 8 || token.length > 64) {
      setState(() {
        _errorMessage = 'Token ujian tidak valid. Periksa kembali kartu ujian.';
        _errorNotice = null;
      });
      return;
    }

    if (roomToken.length < 4 || roomToken.length > 64) {
      setState(() {
        _errorMessage = 'Token ruang tidak valid. Minta token ruang kepada pengawas.';
        _errorNotice = null;
      });
      return;
    }

    if (baseUrl.isEmpty) {
      setState(() {
        _errorMessage = 'Alamat server ujian belum diisi.';
        _errorNotice = null;
      });
      return;
    }

    setState(() {
      _isSubmitting = true;
      _errorMessage = null;
      _errorNotice = null;
    });

    try {
      final normalizedBaseUrl = ExamApiClient.normalizeBaseUrl(baseUrl);
      await _sessionStore.saveBaseUrl(normalizedBaseUrl);
      final deviceFingerprint = await _deviceFingerprint();
      final client = ExamApiClient(
        baseUrl: normalizedBaseUrl,
        deviceFingerprint: deviceFingerprint,
      );
      final payload = await client.login(
        token: token,
        roomToken: roomToken,
        deviceFingerprint: deviceFingerprint,
      );
      await _sessionStore.saveSnapshot(
        ExamSessionSnapshot(
          baseUrl: normalizedBaseUrl,
          examToken: token,
          roomToken: roomToken,
          deviceFingerprint: deviceFingerprint,
          studentName: payload.student.nama,
          studentNis: payload.student.nis,
          sessionTitle: payload.session.title,
          roomName: payload.room?.roomName ?? '-',
          scheduledStartIso:
              payload.session.scheduledStart?.toIso8601String() ?? '',
          scheduledEndIso:
              payload.session.scheduledEnd?.toIso8601String() ?? '',
          durationMinutes: payload.session.durationMinutes,
          currentQuestionIndex: 0,
          answers: const <String, String>{},
          pendingAnswers: const <String, String>{},
          playedAudioQuestionIds: const <String>[],
          lastServerContactIso: '',
          lastSyncFailureIso: '',
          consecutiveSyncFailures: 0,
        ),
      );
      if (!mounted) {
        return;
      }
      await Navigator.of(context).push(
        MaterialPageRoute<void>(
          builder: (_) => ExamShellScreen(
            client: client,
            examToken: token,
            roomToken: roomToken,
            initialPayload: payload,
            deviceFingerprint: deviceFingerprint,
            sessionStore: _sessionStore,
          ),
        ),
      );
    } on ExamApiException catch (error) {
      setState(() {
        _errorMessage = loginFailureMessage(error);
        _errorNotice = loginFailureNotice(error);
      });
    } catch (_) {
      setState(() {
        _errorMessage = 'Tidak dapat terhubung ke server ujian.';
        _errorNotice = null;
      });
    } finally {
      if (mounted) {
        setState(() {
          _isSubmitting = false;
        });
      }
    }
  }

  Future<String> _deviceFingerprint() {
    return _deviceFingerprintStore.loadFingerprint();
  }

  @override
  Widget build(BuildContext context) {
    final theme = Theme.of(context);

    return Scaffold(
      body: SafeArea(
        child: LayoutBuilder(
          builder: (context, constraints) {
            final wide = constraints.maxWidth >= 960;

            return Center(
              child: ConstrainedBox(
                constraints: const BoxConstraints(maxWidth: 1120),
                child: Padding(
                  padding: const EdgeInsets.all(24),
                  child: wide
                      ? Row(
                          children: [
                            Expanded(child: _buildIntro(theme)),
                            const SizedBox(width: 24),
                            SizedBox(width: 420, child: _buildFormCard(theme)),
                          ],
                        )
                      : SingleChildScrollView(
                          child: Column(
                            children: [
                              _buildIntro(theme),
                              const SizedBox(height: 20),
                              _buildFormCard(theme),
                            ],
                          ),
                        ),
                ),
              ),
            );
          },
        ),
      ),
    );
  }

  Widget _buildIntro(ThemeData theme) {
    return Card(
      child: Padding(
        padding: const EdgeInsets.all(28),
        child: Column(
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            Container(
              padding: const EdgeInsets.symmetric(horizontal: 14, vertical: 8),
              decoration: BoxDecoration(
                color: theme.colorScheme.primary.withValues(alpha: 0.12),
                borderRadius: BorderRadius.circular(999),
              ),
              child: Text(
                'CBT MTsN 2 Kolaka Utara',
                style: theme.textTheme.labelLarge?.copyWith(
                  color: theme.colorScheme.primary,
                  fontWeight: FontWeight.w700,
                ),
              ),
            ),
            const SizedBox(height: 20),
            Text(
              'Masuk ujian dengan Token Ujian dan Token Ruang.',
              style: theme.textTheme.headlineMedium?.copyWith(
                fontWeight: FontWeight.w800,
                color: const Color(0xFF14361D),
              ),
            ),
            const SizedBox(height: 12),
            Text(
              'Aplikasi ini dipakai siswa untuk mengerjakan ujian CBT secara fokus, '
              'dengan penyimpanan jawaban bertahap dan pemantauan status perangkat.',
              style: theme.textTheme.bodyLarge?.copyWith(height: 1.55),
            ),
            const SizedBox(height: 28),
            Wrap(
              spacing: 12,
              runSpacing: 12,
              children: const [
                _InfoBadge(
                  icon: Icons.lock_clock_outlined,
                  title: 'Status terpantau',
                  subtitle: 'Koneksi ujian dicek berkala',
                ),
                _InfoBadge(
                  icon: Icons.shield_outlined,
                  title: 'Tetap di layar ujian',
                  subtitle: 'Perpindahan aplikasi akan dilaporkan',
                ),
                _InfoBadge(
                  icon: Icons.save_outlined,
                  title: 'Jawaban aman',
                  subtitle: 'Jawaban disimpan saat dipilih',
                ),
              ],
            ),
          ],
        ),
      ),
    );
  }

  Widget _buildFormCard(ThemeData theme) {
    return Card(
      child: Padding(
        padding: const EdgeInsets.all(24),
        child: Column(
          crossAxisAlignment: CrossAxisAlignment.start,
          mainAxisSize: MainAxisSize.min,
          children: [
            Text(
              'Masuk Ujian',
              style: theme.textTheme.headlineSmall?.copyWith(
                fontWeight: FontWeight.w800,
              ),
            ),
            const SizedBox(height: 8),
            Text(
              _isRestoring
                  ? 'Memeriksa apakah ada sesi ujian yang masih bisa dipulihkan.'
                  : 'Masukkan Token Ujian dari kartu peserta dan Token Ruang dari pengawas.',
              style: theme.textTheme.bodyMedium,
            ),
            const SizedBox(height: 20),
            if (_isRestoring)
              _buildRestoreHintCard(theme)
            else if (widget.previewSnapshot != null)
              Padding(
                padding: const EdgeInsets.only(bottom: 20),
                child: _buildRestoreHintCard(
                  theme,
                  cached: widget.previewSnapshot,
                ),
              )
            else
              FutureBuilder<ExamSessionSnapshot?>(
                future: _sessionStore.loadSnapshot(),
                builder: (context, snapshot) {
                  final cached = snapshot.data;
                  if (cached == null) {
                    return const SizedBox.shrink();
                  }
                  return Padding(
                    padding: const EdgeInsets.only(bottom: 20),
                    child: _buildRestoreHintCard(theme, cached: cached),
                  );
                },
              ),
            TextField(
              controller: _tokenController,
              maxLength: 64,
              textCapitalization: TextCapitalization.characters,
              decoration: const InputDecoration(
                labelText: 'Token Ujian',
                hintText: 'Masukkan Token Ujian dari kartu peserta',
                counterText: '',
              ),
            ),
            const SizedBox(height: 14),
            TextField(
              controller: _roomTokenController,
              maxLength: 64,
              textCapitalization: TextCapitalization.characters,
              decoration: const InputDecoration(
                labelText: 'Token Ruang',
                hintText: 'Minta Token Ruang kepada pengawas',
                counterText: '',
              ),
            ),
            const SizedBox(height: 14),
            _buildOperatorSettings(theme),
            const SizedBox(height: 16),
            if (_errorMessage != null)
              _MessageBanner(tone: BannerTone.error, message: _errorMessage!),
            if (_errorNotice != null) ...[
              const SizedBox(height: 12),
              _ExamGuidancePanel(notice: _errorNotice!),
            ],
            const SizedBox(height: 12),
            SizedBox(
              width: double.infinity,
              child: FilledButton.icon(
                onPressed: _isSubmitting ? null : _submit,
                icon: _isSubmitting
                    ? const SizedBox(
                        width: 18,
                        height: 18,
                        child: CircularProgressIndicator(strokeWidth: 2),
                      )
                    : const Icon(Icons.login),
                label: Text(
                  _isRestoring
                      ? 'Memulihkan sesi...'
                      : _isSubmitting
                      ? 'Memeriksa Token Ujian...'
                      : 'Masuk Ujian',
                ),
              ),
            ),
            const SizedBox(height: 12),
            Text(
              'Gunakan perangkat yang akan dipakai sampai ujian selesai. '
              'Perpindahan aplikasi akan tercatat ke server.',
              style: theme.textTheme.bodySmall?.copyWith(height: 1.5),
            ),
            const SizedBox(height: 8),
            Text(
              'Jika ujian terputus, jangan panik. Tetap gunakan perangkat ini dan minta pengawas membantu memulihkan sesi.',
              style: theme.textTheme.bodySmall?.copyWith(
                height: 1.45,
                color: theme.colorScheme.onSurfaceVariant,
              ),
            ),
          ],
        ),
      ),
    );
  }

  Widget _buildOperatorSettings(ThemeData theme) {
    return Container(
      width: double.infinity,
      decoration: BoxDecoration(
        color: const Color(0xFFF8FAF7),
        borderRadius: BorderRadius.circular(18),
        border: Border.all(color: theme.colorScheme.outlineVariant),
      ),
      child: Padding(
        padding: const EdgeInsets.all(16),
        child: Column(
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            Row(
              children: [
                Expanded(
                  child: Column(
                    crossAxisAlignment: CrossAxisAlignment.start,
                    children: [
                      Text(
                        'Pengaturan Operator',
                        style: theme.textTheme.titleSmall?.copyWith(
                          fontWeight: FontWeight.w800,
                        ),
                      ),
                      const SizedBox(height: 4),
                      Text(
                        _showOperatorSettings
                            ? 'Hanya ubah alamat server atas arahan operator atau pengawas ruang.'
                            : 'Disembunyikan saat mode siswa biasa agar peserta tidak mudah salah mengubah alamat server.',
                        style: theme.textTheme.bodySmall?.copyWith(
                          height: 1.45,
                        ),
                      ),
                    ],
                  ),
                ),
                TextButton.icon(
                  onPressed: () {
                    setState(() {
                      _showOperatorSettings = !_showOperatorSettings;
                    });
                  },
                  icon: Icon(
                    _showOperatorSettings
                        ? Icons.expand_less
                        : Icons.tune_outlined,
                  ),
                  label: Text(
                    _showOperatorSettings ? 'Sembunyikan' : 'Tampilkan',
                  ),
                ),
              ],
            ),
            if (_showOperatorSettings) ...[
              const SizedBox(height: 14),
              TextField(
                controller: _baseUrlController,
                decoration: const InputDecoration(
                  labelText: 'Alamat server ujian',
                  hintText: 'Masukkan alamat dari operator',
                  helperText: 'Diisi hanya jika pengawas/operator meminta.',
                ),
              ),
              const SizedBox(height: 10),
              _BaseUrlGuidance(baseUrl: _baseUrlController.text),
            ],
          ],
        ),
      ),
    );
  }

  Widget _buildRestoreHintCard(ThemeData theme, {ExamSessionSnapshot? cached}) {
    return Container(
      width: double.infinity,
      padding: const EdgeInsets.all(16),
      decoration: BoxDecoration(
        color: const Color(0xFFF6F8F3),
        borderRadius: BorderRadius.circular(18),
        border: Border.all(color: theme.colorScheme.outlineVariant),
      ),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          Text(
            _isRestoring
                ? 'Memeriksa sesi terakhir...'
                : 'Sesi terakhir terdeteksi',
            style: theme.textTheme.titleMedium?.copyWith(
              fontWeight: FontWeight.w800,
            ),
          ),
          const SizedBox(height: 8),
          Text(
            cached == null
                ? 'Jika sebelumnya ujian terputus, aplikasi akan mencoba memulihkannya otomatis.'
                : '${cached.sessionTitle}\n${cached.studentName} • ${cached.studentNis} • Ruang ${cached.roomName}\n${formatExamSchedule(scheduledStartIso: cached.scheduledStartIso, scheduledEndIso: cached.scheduledEndIso, durationMinutes: cached.durationMinutes)}',
            style: theme.textTheme.bodyMedium?.copyWith(height: 1.5),
          ),
          if (cached != null) ...[
            const SizedBox(height: 12),
            Wrap(
              spacing: 8,
              runSpacing: 8,
              children: [
                _RestoreMetaChip(
                  label: formatRestoreHealthLabel(
                    lastServerContactIso: cached.lastServerContactIso,
                    lastSyncFailureIso: cached.lastSyncFailureIso,
                    consecutiveSyncFailures: cached.consecutiveSyncFailures,
                  ),
                  warning: cached.consecutiveSyncFailures >= 3,
                ),
                if (cached.lastServerContactIso.trim().isNotEmpty)
                  _RestoreMetaChip(
                    label:
                        'Kontak server ${formatRestoreClock(cached.lastServerContactIso)}',
                  ),
                if (cached.lastSyncFailureIso.trim().isNotEmpty)
                  _RestoreMetaChip(
                    label:
                        'Gangguan ${formatRestoreClock(cached.lastSyncFailureIso)}',
                    warning: true,
                  ),
              ],
            ),
          ],
        ],
      ),
    );
  }
}

class _ExamGuidancePanel extends StatelessWidget {
  const _ExamGuidancePanel({required this.notice});

  final ExamGuidanceNotice notice;

  @override
  Widget build(BuildContext context) {
    final theme = Theme.of(context);
    final (background, border, foreground, icon) = switch (notice.tone) {
      ExamGuidanceTone.info => (
        const Color(0xFFEAF4EB),
        const Color(0xFF7FB08A),
        const Color(0xFF1E5B2F),
        Icons.info_outline,
      ),
      ExamGuidanceTone.warning => (
        const Color(0xFFFFF3D8),
        const Color(0xFFF2C46D),
        const Color(0xFF9A6700),
        Icons.warning_amber_rounded,
      ),
      ExamGuidanceTone.danger => (
        const Color(0xFFFDE8E8),
        const Color(0xFFE8A4A4),
        const Color(0xFF9F2F2F),
        Icons.gpp_bad_outlined,
      ),
    };

    return Container(
      width: double.infinity,
      padding: const EdgeInsets.all(16),
      decoration: BoxDecoration(
        color: background,
        borderRadius: BorderRadius.circular(18),
        border: Border.all(color: border),
      ),
      child: Row(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          Icon(icon, color: foreground),
          const SizedBox(width: 12),
          Expanded(
            child: Column(
              crossAxisAlignment: CrossAxisAlignment.start,
              children: [
                Text(
                  notice.title,
                  style: theme.textTheme.titleSmall?.copyWith(
                    color: foreground,
                    fontWeight: FontWeight.w800,
                  ),
                ),
                const SizedBox(height: 6),
                Text(
                  notice.message,
                  style: theme.textTheme.bodyMedium?.copyWith(
                    color: foreground,
                    height: 1.45,
                  ),
                ),
              ],
            ),
          ),
        ],
      ),
    );
  }
}

class _RestoreMetaChip extends StatelessWidget {
  const _RestoreMetaChip({required this.label, this.warning = false});

  final String label;
  final bool warning;

  @override
  Widget build(BuildContext context) {
    final theme = Theme.of(context);
    final background = warning
        ? const Color(0xFFFFF3D8)
        : theme.colorScheme.primary.withValues(alpha: 0.12);
    final foreground = warning
        ? const Color(0xFF9A6700)
        : theme.colorScheme.primary;

    return Container(
      padding: const EdgeInsets.symmetric(horizontal: 10, vertical: 7),
      decoration: BoxDecoration(
        color: background,
        borderRadius: BorderRadius.circular(999),
      ),
      child: Text(
        label,
        style: theme.textTheme.labelSmall?.copyWith(
          color: foreground,
          fontWeight: FontWeight.w700,
        ),
      ),
    );
  }
}

class _BaseUrlGuidance extends StatelessWidget {
  const _BaseUrlGuidance({required this.baseUrl});

  final String baseUrl;

  @override
  Widget build(BuildContext context) {
    final theme = Theme.of(context);
    final normalized = baseUrl.trim().toLowerCase();
    final isHttps = normalized.startsWith('https://');
    final isHttp = normalized.startsWith('http://');
    final title = isHttps
        ? 'Alamat server siap digunakan'
        : isHttp
        ? 'Alamat perlu dicek pengawas/operator'
        : 'Alamat server belum lengkap';
    final message = isHttps
        ? 'Lanjutkan hanya jika alamat ini sesuai dengan arahan pengawas atau operator.'
        : isHttp
        ? 'Pengawas/operator wajib memastikan alamat ini benar sebelum peserta login.'
        : 'Minta pengawas/operator memeriksa alamat server sebelum peserta login.';
    final color = isHttps
        ? const Color(0xFF1E6B36)
        : isHttp
        ? const Color(0xFF8A5A00)
        : theme.colorScheme.error;

    return Container(
      width: double.infinity,
      padding: const EdgeInsets.all(12),
      decoration: BoxDecoration(
        color: color.withValues(alpha: 0.08),
        borderRadius: BorderRadius.circular(14),
        border: Border.all(color: color.withValues(alpha: 0.28)),
      ),
      child: Row(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          Icon(
            isHttps ? Icons.https_outlined : Icons.info_outline,
            color: color,
          ),
          const SizedBox(width: 10),
          Expanded(
            child: Column(
              crossAxisAlignment: CrossAxisAlignment.start,
              children: [
                Text(
                  title,
                  style: theme.textTheme.labelLarge?.copyWith(
                    color: color,
                    fontWeight: FontWeight.w800,
                  ),
                ),
                const SizedBox(height: 4),
                Text(
                  message,
                  style: theme.textTheme.bodySmall?.copyWith(height: 1.4),
                ),
              ],
            ),
          ),
        ],
      ),
    );
  }
}

class _InfoBadge extends StatelessWidget {
  const _InfoBadge({
    required this.icon,
    required this.title,
    required this.subtitle,
  });

  final IconData icon;
  final String title;
  final String subtitle;

  @override
  Widget build(BuildContext context) {
    final theme = Theme.of(context);
    return Container(
      width: 220,
      padding: const EdgeInsets.all(16),
      decoration: BoxDecoration(
        color: Colors.white,
        borderRadius: BorderRadius.circular(20),
        border: Border.all(color: theme.colorScheme.outlineVariant),
      ),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          Icon(icon, color: theme.colorScheme.primary),
          const SizedBox(height: 12),
          Text(
            title,
            style: theme.textTheme.titleMedium?.copyWith(
              fontWeight: FontWeight.w700,
            ),
          ),
          const SizedBox(height: 6),
          Text(subtitle, style: theme.textTheme.bodySmall),
        ],
      ),
    );
  }
}

enum BannerTone { error, success, info }

class _MessageBanner extends StatelessWidget {
  const _MessageBanner({required this.tone, required this.message});

  final BannerTone tone;
  final String message;

  @override
  Widget build(BuildContext context) {
    final theme = Theme.of(context);
    final color = switch (tone) {
      BannerTone.error => Colors.red.shade700,
      BannerTone.success => const Color(0xFF0B7A3B),
      BannerTone.info => theme.colorScheme.primary,
    };

    return Container(
      width: double.infinity,
      padding: const EdgeInsets.all(14),
      decoration: BoxDecoration(
        color: color.withValues(alpha: 0.09),
        borderRadius: BorderRadius.circular(16),
        border: Border.all(color: color.withValues(alpha: 0.25)),
      ),
      child: Text(
        message,
        style: theme.textTheme.bodyMedium?.copyWith(
          color: color,
          fontWeight: FontWeight.w600,
        ),
      ),
    );
  }
}
