import 'package:flutter_test/flutter_test.dart';
import 'package:mobile/src/screens/exam_shell_screen.dart';

void main() {
  test('essay answer character cap stays safely below backend 64 KB JSON budget', () {
    expect(kEssayAnswerMaxChars, lessThanOrEqualTo(16000));
    expect(kEssayAnswerMaxChars, greaterThanOrEqualTo(12000));
  });
}
