import 'package:flutter_test/flutter_test.dart';
import 'package:mobile/src/exam_api.dart';
import 'package:mobile/src/exam_error_messages.dart';

void main() {
  group('exam error messages', () {
    test('login failure maps common status codes', () {
      expect(
        loginFailureMessage(
          const ExamApiException('transport', statusCode: null),
        ),
        'Perangkat belum bisa terhubung ke server ujian. Periksa alamat server dan koneksi yang sedang dipakai.',
      );
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

      final transportNotice = loginFailureNotice(
        const ExamApiException('transport', statusCode: null),
      );
      expect(transportNotice?.title, 'Server ujian belum terjangkau');
      expect(transportNotice?.tone, ExamGuidanceTone.warning);

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
          const ExamApiException('transport', statusCode: null),
        ),
        'Sesi lama belum bisa dipulihkan karena perangkat belum terhubung ke server ujian. Coba lagi setelah koneksi membaik atau hubungi pengawas.',
      );
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

      final transportNotice = restoreFailureNotice(
        const ExamApiException('transport', statusCode: null),
      );
      expect(transportNotice?.title, 'Restore tertunda karena koneksi');
      expect(transportNotice?.tone, ExamGuidanceTone.warning);

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

      expect(
        shouldClearSnapshotAfterRestoreFailure(
          const ExamApiException('transport', statusCode: null),
        ),
        isFalse,
      );
      expect(
        shouldClearSnapshotAfterRestoreFailure(
          const ExamApiException('backend', statusCode: 409),
        ),
        isFalse,
      );
      expect(
        shouldClearSnapshotAfterRestoreFailure(
          const ExamApiException('backend', statusCode: 404),
        ),
        isTrue,
      );
    });

    test('status failure maps transport and participant-context states', () {
      expect(
        statusFailureMessage(
          const ExamApiException('transport', statusCode: null),
        ),
        'Status server belum bisa diperbarui karena perangkat belum terhubung. Tetap di layar ujian dan minta pengawas memeriksa koneksi.',
      );
      expect(
        statusFailureMessage(const ExamApiException('backend', statusCode: 401)),
        'Konteks sesi perangkat tidak sah. Minta pengawas memeriksa token dan perangkat sebelum melanjutkan.',
      );
      expect(
        statusFailureMessage(const ExamApiException('backend', statusCode: 403)),
        'Sesi ujian tidak lagi aktif menurut server. Tunggu arahan pengawas sebelum melanjutkan.',
      );
      expect(
        statusFailureMessage(const ExamApiException('backend', statusCode: 409)),
        'Token sesi ini terdeteksi aktif di perangkat lain. Jangan lanjutkan dari perangkat ini sebelum pengawas memverifikasi.',
      );

      final transportNotice = statusFailureNotice(
        const ExamApiException('transport', statusCode: null),
      );
      expect(transportNotice?.title, 'Status belum tersinkron');
      expect(transportNotice?.tone, ExamGuidanceTone.warning);

      final unauthorizedNotice = statusFailureNotice(
        const ExamApiException('backend', statusCode: 401),
      );
      expect(unauthorizedNotice?.title, 'Konteks peserta tidak sah');
      expect(unauthorizedNotice?.tone, ExamGuidanceTone.danger);

      final mismatchNotice = statusFailureNotice(
        const ExamApiException('backend', statusCode: 409),
      );
      expect(mismatchNotice?.title, 'Perangkat berbeda terdeteksi');
      expect(mismatchNotice?.tone, ExamGuidanceTone.danger);
    });

    test('answer failure maps common status codes', () {
      expect(
        answerFailureMessage(
          const ExamApiException('transport', statusCode: null),
        ),
        'Perangkat sedang kehilangan koneksi ke server ujian. Jawaban tetap disimpan di perangkat dan akan dicoba sinkron ulang.',
      );
      expect(
        answerFailureMessage(
          const ExamApiException('backend', statusCode: 401),
        ),
        'Sesi perangkat belum sah untuk mengirim jawaban. Jawaban tetap disimpan lokal sambil menunggu pemeriksaan pengawas.',
      );
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

      final transportNotice = answerFailureNotice(
        const ExamApiException('transport', statusCode: null),
      );
      expect(transportNotice?.title, 'Jawaban tersimpan lokal');
      expect(transportNotice?.tone, ExamGuidanceTone.warning);

      final unauthorizedNotice = answerFailureNotice(
        const ExamApiException('backend', statusCode: 401),
      );
      expect(unauthorizedNotice?.title, 'Sesi perangkat perlu diverifikasi');
      expect(unauthorizedNotice?.tone, ExamGuidanceTone.danger);

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
          const ExamApiException('transport', statusCode: null),
          autoSubmit: false,
        ),
        'Perangkat belum bisa terhubung ke server ujian. Jangan tinggalkan layar ini sebelum pengawas memastikan koneksi kembali.',
      );
      expect(
        submitFailureMessage(
          const ExamApiException('transport', statusCode: null),
          autoSubmit: true,
        ),
        'Submit otomatis belum bisa dikirim karena perangkat kehilangan koneksi ke server ujian. Segera minta pengawas memeriksa jaringan.',
      );
      expect(
        submitFailureMessage(
          const ExamApiException('backend', statusCode: 401),
          autoSubmit: false,
        ),
        'Sesi perangkat belum sah untuk submit. Tetap di layar ini dan minta pengawas memeriksa token atau reset akses.',
      );
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

      final transportNotice = submitFailureNotice(
        const ExamApiException('transport', statusCode: null),
        autoSubmit: true,
      );
      expect(transportNotice?.title, 'Submit otomatis tertunda karena koneksi');
      expect(transportNotice?.tone, ExamGuidanceTone.warning);

      final unauthorizedNotice = submitFailureNotice(
        const ExamApiException('backend', statusCode: 401),
        autoSubmit: false,
      );
      expect(unauthorizedNotice?.title, 'Submit ditahan karena konteks peserta');
      expect(unauthorizedNotice?.tone, ExamGuidanceTone.danger);

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
