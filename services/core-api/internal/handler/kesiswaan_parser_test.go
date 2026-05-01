package handler

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

const kesiswaanParserUUID = "01000000-0000-0000-0000-000000000000"

func TestKesiswaanParseViolationAndAchievementRequests(t *testing.T) {
	rec := httptest.NewRecorder()
	violation, ok := parseViolationRequest(rec, kesiswaanViolationRequest{
		StudentID:    kesiswaanParserUUID,
		CategoryID:   kesiswaanParserUUID,
		IncidentDate: "2026-05-01",
		Points:       15,
		Description:  "Terlambat",
		ActionTaken:  "Pembinaan",
		Status:       "open",
	})
	if !ok || !violation.StudentID.Valid || !violation.CategoryID.Valid || violation.Points != 15 || violation.Status != "open" {
		t.Fatalf("parseViolationRequest() = %+v/%v, want mapped violation", violation, ok)
	}
	if !violation.IncidentDate.Valid || violation.IncidentDate.Time.Format("2006-01-02") != "2026-05-01" {
		t.Fatalf("parseViolationRequest() incident_date = %v, want 2026-05-01", violation.IncidentDate)
	}
	rec = httptest.NewRecorder()
	if _, ok := parseViolationRequest(rec, kesiswaanViolationRequest{StudentID: "bad"}); ok || rec.Code != http.StatusBadRequest {
		t.Fatalf("parseViolationRequest(bad uuid) ok/status = %v/%d, want false/400", ok, rec.Code)
	}
	rec = httptest.NewRecorder()
	if _, ok := parseViolationRequest(rec, kesiswaanViolationRequest{StudentID: kesiswaanParserUUID, CategoryID: kesiswaanParserUUID, IncidentDate: "bad"}); ok || rec.Code != http.StatusBadRequest {
		t.Fatalf("parseViolationRequest(bad date) ok/status = %v/%d, want false/400", ok, rec.Code)
	}

	rec = httptest.NewRecorder()
	achievement, ok := parseAchievementRequest(rec, kesiswaanAchievementRequest{
		StudentID:       kesiswaanParserUUID,
		AchievementDate: "2026-05-02",
		Title:           "Juara Olimpiade",
		Level:           "kabupaten",
		Category:        "akademik",
		Organizer:       "Kemenag",
		Description:     "Juara 1",
		DocumentUrl:     "/bukti",
	})
	if !ok || !achievement.StudentID.Valid || achievement.Title != "Juara Olimpiade" || achievement.Level != "kabupaten" {
		t.Fatalf("parseAchievementRequest() = %+v/%v, want mapped achievement", achievement, ok)
	}
	rec = httptest.NewRecorder()
	if _, ok := parseAchievementRequest(rec, kesiswaanAchievementRequest{StudentID: kesiswaanParserUUID, AchievementDate: "bad"}); ok || rec.Code != http.StatusBadRequest {
		t.Fatalf("parseAchievementRequest(bad date) ok/status = %v/%d, want false/400", ok, rec.Code)
	}
}

func TestKesiswaanParseExtracurricularCounselingAndTransferRequests(t *testing.T) {
	rec := httptest.NewRecorder()
	member, ok := parseExtracurricularMemberRequest(rec, kesiswaanExtracurricularMemberRequest{
		ExtracurricularID: kesiswaanParserUUID,
		StudentID:         kesiswaanParserUUID,
		JoinedAt:          "2026-05-03",
		Role:              "member",
		Status:            "active",
		Notes:             "aktif latihan",
	})
	if !ok || !member.ExtracurricularID.Valid || !member.StudentID.Valid || member.Role != "member" || member.Status != "active" {
		t.Fatalf("parseExtracurricularMemberRequest() = %+v/%v, want mapped member", member, ok)
	}
	if !member.JoinedAt.Valid || member.JoinedAt.Time.Format("2006-01-02") != "2026-05-03" {
		t.Fatalf("parseExtracurricularMemberRequest() joined_at = %v, want 2026-05-03", member.JoinedAt)
	}
	rec = httptest.NewRecorder()
	if _, ok := parseExtracurricularMemberRequest(rec, kesiswaanExtracurricularMemberRequest{ExtracurricularID: "bad"}); ok || rec.Code != http.StatusBadRequest {
		t.Fatalf("parseExtracurricularMemberRequest(bad uuid) ok/status = %v/%d, want false/400", ok, rec.Code)
	}

	rec = httptest.NewRecorder()
	counseling, ok := parseCounselingRequest(rec, kesiswaanCounselingRequest{
		StudentID:      kesiswaanParserUUID,
		SessionDate:    "2026-05-04",
		Topic:          "Belajar",
		Summary:        "Perlu pendampingan",
		FollowUp:       "Jadwal ulang",
		Status:         "open",
		IsConfidential: true,
	})
	if !ok || !counseling.StudentID.Valid || counseling.Topic != "Belajar" || !counseling.IsConfidential {
		t.Fatalf("parseCounselingRequest() = %+v/%v, want mapped confidential counseling", counseling, ok)
	}
	rec = httptest.NewRecorder()
	if _, ok := parseCounselingRequest(rec, kesiswaanCounselingRequest{StudentID: kesiswaanParserUUID, SessionDate: "bad"}); ok || rec.Code != http.StatusBadRequest {
		t.Fatalf("parseCounselingRequest(bad date) ok/status = %v/%d, want false/400", ok, rec.Code)
	}

	rec = httptest.NewRecorder()
	transfer, ok := parseStudentTransferRequest(rec, kesiswaanTransferRequest{
		StudentID:         kesiswaanParserUUID,
		TransferDate:      "2026-05-05",
		TransferType:      "out",
		PreviousSchool:    "MTsN 2",
		DestinationSchool: "MTsN lain",
		Reason:            "Pindah domisili",
		DocumentRef:       "SURAT-1",
		Notes:             "lengkap",
	})
	if !ok || !transfer.StudentID.Valid || transfer.TransferType != "out" || transfer.DocumentRef != "SURAT-1" {
		t.Fatalf("parseStudentTransferRequest() = %+v/%v, want mapped transfer", transfer, ok)
	}
	rec = httptest.NewRecorder()
	if _, ok := parseStudentTransferRequest(rec, kesiswaanTransferRequest{StudentID: "bad"}); ok || rec.Code != http.StatusBadRequest {
		t.Fatalf("parseStudentTransferRequest(bad uuid) ok/status = %v/%d, want false/400", ok, rec.Code)
	}
}
