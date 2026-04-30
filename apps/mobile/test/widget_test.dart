import 'package:flutter_test/flutter_test.dart';

import 'package:mobile/src/app.dart';

void main() {
  testWidgets('login screen renders exam shell entry', (tester) async {
    await tester.pumpWidget(const MtsnMobileApp());

    expect(find.text('Masuk Ujian'), findsWidgets);
    expect(find.text('Token ujian'), findsOneWidget);
    expect(find.text('Alamat server API'), findsOneWidget);
  });
}
