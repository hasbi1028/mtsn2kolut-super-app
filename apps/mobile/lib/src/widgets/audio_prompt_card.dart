import 'package:audioplayers/audioplayers.dart';
import 'package:flutter/material.dart';

class AudioPromptCard extends StatefulWidget {
  const AudioPromptCard({
    super.key,
    required this.url,
    required this.label,
    this.hasBeenPlayed = false,
    this.onPlayed,
  });

  final String url;
  final String label;
  final bool hasBeenPlayed;
  final VoidCallback? onPlayed;

  @override
  State<AudioPromptCard> createState() => _AudioPromptCardState();
}

class _AudioPromptCardState extends State<AudioPromptCard> {
  late final AudioPlayer _player;
  bool _hasPlayedOnce = false;
  bool _isPlaying = false;
  bool _isLoading = false;
  String? _errorMessage;

  @override
  void initState() {
    super.initState();
    _player = AudioPlayer();
    _player.onPlayerComplete.listen((_) {
      if (!mounted) {
        return;
      }
      setState(() {
        _isPlaying = false;
        _isLoading = false;
      });
    });
  }

  @override
  void dispose() {
    _player.dispose();
    super.dispose();
  }

  Future<void> _togglePlayback() async {
    if (_isPlaying) {
      await _player.pause();
      if (!mounted) {
        return;
      }
      setState(() {
        _isPlaying = false;
      });
      return;
    }

    setState(() {
      _isLoading = true;
      _errorMessage = null;
    });

    try {
      await _player.play(UrlSource(widget.url));
      if (!mounted) {
        return;
      }
      setState(() {
        _isPlaying = true;
        _hasPlayedOnce = true;
      });
      widget.onPlayed?.call();
    } catch (_) {
      if (!mounted) {
        return;
      }
      setState(() {
        _errorMessage = 'Audio tidak dapat diputar di perangkat ini.';
      });
    } finally {
      if (mounted) {
        setState(() {
          _isLoading = false;
        });
      }
    }
  }

  @override
  Widget build(BuildContext context) {
    final theme = Theme.of(context);

    return Container(
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
            widget.label,
            style: theme.textTheme.labelLarge?.copyWith(
              color: theme.colorScheme.primary,
              fontWeight: FontWeight.w800,
            ),
          ),
          const SizedBox(height: 6),
          Wrap(
            spacing: 8,
            runSpacing: 8,
            children: [
              _AudioStatusBadge(
                icon: Icons.graphic_eq,
                label: _effectivePlayedState
                    ? 'Audio sudah diputar'
                    : 'Audio belum diputar',
                highlighted: _effectivePlayedState,
              ),
            ],
          ),
          const SizedBox(height: 10),
          Row(
            children: [
              FilledButton.icon(
                onPressed: _isLoading ? null : _togglePlayback,
                icon: _isLoading
                    ? const SizedBox(
                        width: 16,
                        height: 16,
                        child: CircularProgressIndicator(strokeWidth: 2),
                      )
                    : Icon(_isPlaying ? Icons.pause : Icons.play_arrow),
                label: Text(_isPlaying ? 'Jeda Audio' : 'Putar Audio'),
              ),
              const SizedBox(width: 12),
              Expanded(
                child: Text(
                  widget.url,
                  style: theme.textTheme.bodySmall,
                  maxLines: 2,
                  overflow: TextOverflow.ellipsis,
                ),
              ),
            ],
          ),
          if (_errorMessage != null) ...[
            const SizedBox(height: 10),
            Text(
              _errorMessage!,
              style: theme.textTheme.bodySmall?.copyWith(
                color: Colors.red.shade700,
                fontWeight: FontWeight.w700,
              ),
            ),
          ],
        ],
      ),
    );
  }

  bool get _effectivePlayedState => widget.hasBeenPlayed || _hasPlayedOnce;
}

class _AudioStatusBadge extends StatelessWidget {
  const _AudioStatusBadge({
    required this.icon,
    required this.label,
    required this.highlighted,
  });

  final IconData icon;
  final String label;
  final bool highlighted;

  @override
  Widget build(BuildContext context) {
    final theme = Theme.of(context);
    final background = highlighted
        ? theme.colorScheme.primary.withValues(alpha: 0.12)
        : const Color(0xFFEAEFE3);
    final foreground = highlighted
        ? theme.colorScheme.primary
        : theme.colorScheme.onSurfaceVariant;

    return Container(
      padding: const EdgeInsets.symmetric(horizontal: 10, vertical: 7),
      decoration: BoxDecoration(
        color: background,
        borderRadius: BorderRadius.circular(999),
      ),
      child: Row(
        mainAxisSize: MainAxisSize.min,
        children: [
          Icon(icon, size: 16, color: foreground),
          const SizedBox(width: 6),
          Text(
            label,
            style: theme.textTheme.labelSmall?.copyWith(
              color: foreground,
              fontWeight: FontWeight.w700,
            ),
          ),
        ],
      ),
    );
  }
}
