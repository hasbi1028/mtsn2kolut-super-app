import 'package:flutter/material.dart';

class RichExamText extends StatelessWidget {
  const RichExamText({super.key, required this.content, this.style});

  final String content;
  final TextStyle? style;

  @override
  Widget build(BuildContext context) {
    final cleaned = normalizeExamText(content);
    if (cleaned.isEmpty) {
      return const SizedBox.shrink();
    }

    final paragraphs = cleaned
        .split('\n\n')
        .map((part) => part.trim())
        .where((part) => part.isNotEmpty)
        .toList();

    return Column(
      crossAxisAlignment: CrossAxisAlignment.start,
      children: [
        for (var i = 0; i < paragraphs.length; i++) ...[
          Text(paragraphs[i], style: style),
          if (i != paragraphs.length - 1) const SizedBox(height: 10),
        ],
      ],
    );
  }
}

String normalizeExamText(String raw) {
  var text = raw.trim();
  if (text.isEmpty) {
    return '';
  }

  text = text
      .replaceAll(RegExp(r'<br\s*/?>', caseSensitive: false), '\n')
      .replaceAll(RegExp(r'</p>', caseSensitive: false), '\n\n')
      .replaceAll(RegExp(r'</div>', caseSensitive: false), '\n')
      .replaceAll(RegExp(r'</li>', caseSensitive: false), '\n')
      .replaceAll(RegExp(r'<li[^>]*>', caseSensitive: false), '• ')
      .replaceAll(RegExp(r'<[^>]+>'), '');

  const entities = <String, String>{
    '&nbsp;': ' ',
    '&amp;': '&',
    '&lt;': '<',
    '&gt;': '>',
    '&quot;': '"',
    '&#39;': "'",
  };

  entities.forEach((key, value) {
    text = text.replaceAll(key, value);
  });

  return text
      .replaceAll(RegExp(r'\r\n?'), '\n')
      .replaceAll(RegExp(r'\n{3,}'), '\n\n')
      .replaceAll(RegExp(r'[ \t]{2,}'), ' ')
      .trim();
}
