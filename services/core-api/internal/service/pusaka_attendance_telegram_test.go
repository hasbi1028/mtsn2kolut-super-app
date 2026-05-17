package service

import (
	"bytes"
	"image/color"
	"image/png"
	"strings"
	"testing"
	"time"
)

func TestPusakaAttendanceTelegramPresentationHelpers(t *testing.T) {
	date := time.Date(2026, 5, 17, 0, 0, 0, 0, time.UTC)
	if got := datePg(date); !got.Valid || got.Time.Location() != time.UTC || got.Time.Hour() != 0 || got.Time.Day() != 17 {
		t.Fatalf("datePg() = %+v", got)
	}
	for input, want := range map[string]string{
		" 07:15:30 WITA ":         "07:15:30",
		"masuk 06:57:00.123 WITA": "06:57:00",
		"08:05":                   "08:05",
		"":                        "-",
	} {
		if got := cleanTime(input); got != want {
			t.Fatalf("cleanTime(%q) = %q, want %q", input, got, want)
		}
	}
	for input, want := range map[string]string{
		"123":         "***",
		" 987654321 ": "*****4321",
		"":            "",
	} {
		if got := maskChatID(input); got != want {
			t.Fatalf("maskChatID(%q) = %q, want %q", input, got, want)
		}
	}
	if got := truncate("Siti Aminah", 20); got != "Siti Aminah" {
		t.Fatalf("truncate short = %q", got)
	}
	if got := truncate("Nama Pegawai Sangat Panjang", 12); got != "Nama Pegawa…" {
		t.Fatalf("truncate long = %q", got)
	}
}

func TestPusakaAttendanceCaptionAndPNGRendering(t *testing.T) {
	report := attendanceReport{
		Date:           time.Date(2026, 5, 17, 0, 0, 0, 0, time.UTC),
		GeneratedAt:    time.Date(2026, 5, 17, 10, 11, 12, 0, time.FixedZone("WITA", 8*3600)),
		TotalEmployees: 2,
		CheckedIn:      1,
		NotCheckedIn:   1,
		CheckedOut:     1,
		Rows: []attendanceReportRow{
			{No: 1, EmployeeName: "Andi", CheckIn: "07:01", CheckOut: "15:30", Status: "lengkap"},
			{No: 2, EmployeeName: "Budi", CheckIn: "-", CheckOut: "-", Status: "belum"},
		},
	}
	caption := buildAttendanceCaption(report)
	for _, want := range []string{"Daftar Hadir Ringkas", "Tanggal: 17 May 2026", "Total Pegawai: 2", "Sudah Masuk: 1", "Belum Masuk: 1", "Sudah Pulang: 1"} {
		if !strings.Contains(caption, want) {
			t.Fatalf("caption %q missing %q", caption, want)
		}
	}

	pngBytes, err := renderAttendancePNG(report)
	if err != nil {
		t.Fatalf("renderAttendancePNG() error = %v", err)
	}
	if !bytes.HasPrefix(pngBytes, []byte("\x89PNG\r\n\x1a\n")) {
		t.Fatalf("renderAttendancePNG() did not return PNG bytes")
	}
	img, err := png.Decode(bytes.NewReader(pngBytes))
	if err != nil {
		t.Fatalf("decode PNG: %v", err)
	}
	if got, want := img.Bounds().Dx(), 1200; got != want {
		t.Fatalf("PNG width = %d, want %d", got, want)
	}
	if got, want := img.Bounds().Dy(), 220+len(report.Rows)*30+50; got != want {
		t.Fatalf("PNG height = %d, want %d", got, want)
	}
}

func TestPusakaAttendanceStatusColors(t *testing.T) {
	cases := []struct {
		status string
		want   color.RGBA
	}{
		{"lengkap", color.RGBA{22, 101, 52, 255}},
		{"MASUK", color.RGBA{180, 83, 9, 255}},
		{"belum", color.RGBA{100, 116, 139, 255}},
	}
	for _, tc := range cases {
		got, ok := statusColor(tc.status).(color.RGBA)
		if !ok || got != tc.want {
			t.Fatalf("statusColor(%q) = %#v, want %#v", tc.status, statusColor(tc.status), tc.want)
		}
	}
}
