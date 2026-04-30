import 'dart:io';

import 'package:flutter/material.dart';

import '../exam_api.dart';
import '../exam_session_store.dart';
import 'exam_shell_screen.dart';

class ExamLoginScreen extends StatefulWidget {
  const ExamLoginScreen({super.key});

  @override
  State<ExamLoginScreen> createState() => _ExamLoginScreenState();
}

class _ExamLoginScreenState extends State<ExamLoginScreen> {
  final _sessionStore = ExamSessionStore();
  final _tokenController = TextEditingController();
  final _baseUrlController = TextEditingController(
    text: const String.fromEnvironment(
      'API_BASE_URL',
      defaultValue: 'http://10.0.2.2:8080',
    ),
  );

  bool _isSubmitting = false;
  bool _isRestoring = true;
  String? _errorMessage;

  @override
  void initState() {
    super.initState();
    _restoreExamSession();
  }

  @override
  void dispose() {
    _tokenController.dispose();
    _baseUrlController.dispose();
    super.dispose();
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
    });

    final client = ExamApiClient(baseUrl: snapshot.baseUrl);

    try {
      final payload = await client.login(
        token: snapshot.examToken,
        deviceFingerprint: snapshot.deviceFingerprint,
      );
      if (!mounted) {
        return;
      }
      _baseUrlController.text = snapshot.baseUrl;
      _tokenController.text = snapshot.examToken.toUpperCase();
      await Navigator.of(context).push(
        MaterialPageRoute<void>(
          builder: (_) => ExamShellScreen(
            client: client,
            examToken: snapshot.examToken,
            initialPayload: payload,
            deviceFingerprint: snapshot.deviceFingerprint,
            restoredSnapshot: snapshot,
          ),
        ),
      );
    } on ExamApiException catch (_) {
      await _sessionStore.clearSnapshot();
      if (!mounted) {
        return;
      }
      setState(() {
        _errorMessage =
            'Sesi ujian terakhir tidak bisa dipulihkan. Silakan login ulang dengan token aktif.';
      });
    } catch (_) {
      if (!mounted) {
        return;
      }
      setState(() {
        _errorMessage =
            'Tidak dapat memulihkan sesi terakhir karena koneksi ke server gagal.';
      });
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
    final baseUrl = _baseUrlController.text.trim();

    if (token.length != 8) {
      setState(() {
        _errorMessage = 'Token ujian harus terdiri dari 8 karakter.';
      });
      return;
    }

    if (baseUrl.isEmpty) {
      setState(() {
        _errorMessage = 'Alamat server ujian belum diisi.';
      });
      return;
    }

    setState(() {
      _isSubmitting = true;
      _errorMessage = null;
    });

    await _sessionStore.saveBaseUrl(baseUrl);
    final client = ExamApiClient(baseUrl: baseUrl);
    final deviceFingerprint = _deviceFingerprint();

    try {
      final payload = await client.login(
        token: token,
        deviceFingerprint: deviceFingerprint,
      );
      await _sessionStore.saveSnapshot(
        ExamSessionSnapshot(
          baseUrl: baseUrl,
          examToken: token,
          deviceFingerprint: deviceFingerprint,
          currentQuestionIndex: 0,
          answers: const <String, String>{},
          pendingAnswers: const <String, String>{},
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
            initialPayload: payload,
            deviceFingerprint: deviceFingerprint,
          ),
        ),
      );
    } on ExamApiException catch (error) {
      setState(() {
        _errorMessage = error.message;
      });
    } catch (_) {
      setState(() {
        _errorMessage = 'Tidak dapat terhubung ke server ujian.';
      });
    } finally {
      if (mounted) {
        setState(() {
          _isSubmitting = false;
        });
      }
    }
  }

  String _deviceFingerprint() {
    final host = Platform.localHostname;
    return '${Platform.operatingSystem}:$host';
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
              'Masuk ujian dengan token pengawas.',
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
                  title: 'Heartbeat aktif',
                  subtitle: 'Status peserta dipantau berkala',
                ),
                _InfoBadge(
                  icon: Icons.shield_outlined,
                  title: 'Anti-switch dasar',
                  subtitle: 'Keluar aplikasi akan tercatat',
                ),
                _InfoBadge(
                  icon: Icons.save_outlined,
                  title: 'Simpan bertahap',
                  subtitle: 'Jawaban dikirim saat dipilih',
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
                  : 'Masukkan token ujian dari kartu peserta atau pengawas.',
              style: theme.textTheme.bodyMedium,
            ),
            const SizedBox(height: 20),
            TextField(
              controller: _tokenController,
              maxLength: 8,
              textCapitalization: TextCapitalization.characters,
              decoration: const InputDecoration(
                labelText: 'Token ujian',
                hintText: 'Contoh: A1B2C3D4',
                counterText: '',
              ),
            ),
            const SizedBox(height: 14),
            TextField(
              controller: _baseUrlController,
              decoration: const InputDecoration(
                labelText: 'Alamat server API',
                hintText: 'Contoh: http://10.0.2.2:8080',
              ),
            ),
            const SizedBox(height: 16),
            if (_errorMessage != null)
              _MessageBanner(tone: BannerTone.error, message: _errorMessage!),
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
                      ? 'Memeriksa token...'
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
          ],
        ),
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
