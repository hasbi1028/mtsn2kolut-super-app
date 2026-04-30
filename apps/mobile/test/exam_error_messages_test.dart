import 'package:flutter_test/flutter_test.dart';
import 'package:mobile/src/exam_api.dart';
import 'package:mobile/src/exam_error_messages.dart';

void main() {
  group('exam error messages', () {
    test('login failure maps common status codes', () {
      expect(
        loginFailureMessage(const ExamApiException('backend', statusCode: 404)),
        'Token ujian tidak ditemukan. Periksa kembali token dari pengawas atau kartu ujian.',
      );
      expect(
        loginFailureMessage(const ExamApiException('backend', statusCode: 403)),
        'Sesi ujian belum aktif atau sudah berakhir. Hubungi pengawas untuk memastikan jadwal sesi.',
      );
      expect(
        loginFailureMessage(const ExamApiException('backend', statusCode: 409)),
        'Token ini sudah terhubung dengan perangkat lain. Gunakan perangkat yang sama atau minta bantuan pengawas.',
      );
      expect(
        loginFailureMessage(
          const ExamApiException('Pesan asli backend', statusCode: 500),
        ),
        'Pesan asli backend',
      );

      final warningNotice = loginFailureNotice(
        const ExamApiException('backend', statusCode: 403),
      );
      expect(warningNotice?.title, 'Sesi belum bisa dimasuki');
      expect(warningNotice?.tone, ExamGuidanceTone.warning);

      final dangerNotice = loginFailureNotice(
        const ExamApiException('backend', statusCode: 409),
      );
      expect(dangerNotice?.title, 'Token sudah terikat ke perangkat lain');
      expect(dangerNotice?.tone, ExamGuidanceTone.danger);
    });

    test('restore failure maps common status codes', () {
      expect(
        restoreFailureMessage(
          const ExamApiException('backend', statusCode: 404),
        ),
        'Token sesi lama sudah tidak ditemukan lagi di server. Login ulang dengan token aktif dari pengawas jika sesi masih berlangsung.',
      );
      expect(
        restoreFailureMessage(
          const ExamApiException('backend', statusCode: 403),
        ),
        'Sesi lama tidak bisa dipulihkan karena ujian belum aktif lagi atau sudah ditutup. Periksa status sesi dengan pengawas.',
      );
      expect(
        restoreFailureMessage(
          const ExamApiException('backend', statusCode: 409),
        ),
        'Sesi lama terikat ke perangkat lain. Gunakan perangkat yang sama seperti sebelumnya atau minta bantuan pengawas.',
      );
      expect(
        restoreFailureMessage(
          const ExamApiException('Pesan asli backend', statusCode: 500),
        ),
        'Sesi lama tidak bisa dipulihkan. Pesan asli backend',
      );

      final warningNotice = restoreFailureNotice(
        const ExamApiException('backend', statusCode: 403),
      );
      expect(warningNotice?.title, 'Sesi lama belum bisa dipulihkan');
      expect(warningNotice?.tone, ExamGuidanceTone.warning);

      final dangerNotice = restoreFailureNotice(
        const ExamApiException('backend', statusCode: 409),
      );
      expect(dangerNotice?.title, 'Sesi lama aktif di perangkat lain');
      expect(dangerNotice?.tone, ExamGuidanceTone.danger);
    });

    test('answer failure maps common status codes', () {
      expect(
        answerFailureMessage(
          const ExamApiException('backend', statusCode: 403),
        ),
        'Waktu ujian sudah berakhir. Jawaban tetap disimpan di perangkat ini, tetapi pengawas perlu memastikan apakah sesi masih bisa dipulihkan.',
      );
      expect(
        answerFailureMessage(
          const ExamApiException('backend', statusCode: 409),
        ),
        'Ujian ini sudah dinyatakan selesai di server. Jawaban baru tidak bisa dikirim lagi.',
      );
      expect(
        answerFailureMessage(
          const ExamApiException('Pesan asli backend', statusCode: 500),
        ),
        'Pesan asli backend Jawaban tetap disimpan di perangkat dan akan dicoba sinkron ulang.',
      );

      final warningNotice = answerFailureNotice(
        const ExamApiException('backend', statusCode: 403),
      );
      expect(warningNotice?.title, 'Waktu ujian sudah berakhir');
      expect(warningNotice?.tone, ExamGuidanceTone.warning);

      final dangerNotice = answerFailureNotice(
        const ExamApiException('backend', statusCode: 409),
      );
      expect(dangerNotice?.title, 'Ujian sudah selesai di server');
      expect(dangerNotice?.tone, ExamGuidanceTone.danger);
    });

    test('submit failure maps common status codes', () {
      expect(
        submitFailureMessage(
          const ExamApiException('backend', statusCode: 403),
          autoSubmit: false,
        ),
        'Waktu ujian sudah berakhir menurut server. Hubungi pengawas untuk memastikan status kirim ujian.',
      );
      expect(
        submitFailureMessage(
          const ExamApiException('backend', statusCode: 403),
          autoSubmit: true,
        ),
        'Waktu ujian sudah habis, tetapi server belum menerima submit otomatis. Segera minta pengawas memeriksa koneksi dan status sesi.',
      );
      expect(
        submitFailureMessage(
          const ExamApiException('backend', statusCode: 409),
          autoSubmit: false,
        ),
        'Ujian ini sudah tercatat selesai di server. Tidak perlu menekan kirim lagi.',
      );
      expect(
        submitFailureMessage(
          const ExamApiException('Pesan asli backend', statusCode: 500),
          autoSubmit: false,
        ),
        'Pesan asli backend',
      );

      final warningNotice = submitFailureNotice(
        const ExamApiException('backend', statusCode: 403),
        autoSubmit: true,
      );
      expect(warningNotice?.tone, ExamGuidanceTone.warning);

      final infoNotice = submitFailureNotice(
        const ExamApiException('backend', statusCode: 409),
        autoSubmit: false,
      );
      expect(infoNotice?.title, 'Submit sudah tercatat');
      expect(infoNotice?.tone, ExamGuidanceTone.info);
    });
  });
}
