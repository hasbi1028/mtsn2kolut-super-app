import 'package:flutter_test/flutter_test.dart';
import 'package:mobile/src/widgets/rich_exam_text.dart';

void main() {
  group('normalizeExamText', () {
    test('returns empty string for blank content', () {
      expect(normalizeExamText('   '), '');
    });

    test('normalizes paragraphs and line breaks', () {
      final result = normalizeExamText(
        '<p>Paragraf satu</p><p>Paragraf dua<br>baris berikut</p>',
      );

      expect(result, 'Paragraf satu\n\nParagraf dua\nbaris berikut');
    });

    test('normalizes list items into bullets', () {
      final result = normalizeExamText(
        '<ul><li>Pilihan A</li><li>Pilihan B</li></ul>',
      );

      expect(result, '• Pilihan A\n• Pilihan B');
    });

    test('decodes common html entities', () {
      final result = normalizeExamText(
        '<div>&lt;teks&gt; &amp; &quot;contoh&quot; &#39;uji&#39;</div>',
      );

      expect(result, '<teks> & "contoh" \'uji\'');
    });

    test('collapses excessive spacing and newlines', () {
      final result = normalizeExamText(
        '<div>Kalimat   pertama</div>\n\n\n<div>Kalimat kedua</div>',
      );

      expect(result, 'Kalimat pertama\n\nKalimat kedua');
    });
  });
}
