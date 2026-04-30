package handler

import (
	"net/http/httptest"
	"testing"

	"mtsn2kolut-super-app/backend/internal/service"
)

func TestAbsolutizeExamAssetURLHonorsForwardedHeaders(t *testing.T) {
	req := httptest.NewRequest("POST", "http://internal/api/exam/login", nil)
	req.Host = "internal:8080"
	req.Header.Set("X-Forwarded-Proto", "https")
	req.Header.Set("X-Forwarded-Host", "cbt.mtsn2kolut.sch.id")

	got := absolutizeExamAssetURL(req, "/api/cbt/assets/asset-1/file")
	want := "https://cbt.mtsn2kolut.sch.id/api/cbt/assets/asset-1/file"
	if got != want {
		t.Fatalf("absolutizeExamAssetURL() = %q, want %q", got, want)
	}
}

func TestAbsolutizeExamAssetURLFallsBackToRequestHost(t *testing.T) {
	req := httptest.NewRequest("POST", "http://localhost:8080/api/exam/login", nil)
	req.Host = "localhost:8080"

	got := absolutizeExamAssetURL(req, "api/cbt/assets/asset-2/file")
	want := "http://localhost:8080/api/cbt/assets/asset-2/file"
	if got != want {
		t.Fatalf("absolutizeExamAssetURL() = %q, want %q", got, want)
	}
}

func TestAbsolutizeExamAssetURLLeavesAbsoluteURLUntouched(t *testing.T) {
	req := httptest.NewRequest("POST", "http://localhost:8080/api/exam/login", nil)

	got := absolutizeExamAssetURL(req, "https://cdn.example.com/file.png")
	want := "https://cdn.example.com/file.png"
	if got != want {
		t.Fatalf("absolutizeExamAssetURL() = %q, want %q", got, want)
	}
}

func TestAbsolutizeExamLoginResultRewritesAllMediaFields(t *testing.T) {
	req := httptest.NewRequest("POST", "http://internal/api/exam/login", nil)
	req.Header.Set("X-Forwarded-Proto", "https")
	req.Header.Set("X-Forwarded-Host", "mobile-api.example.sch.id")

	result := service.LoginResult{
		Questions: []service.ExamQuestion{
			{
				StemMediaURL:     "/api/cbt/assets/image-1/file",
				StimulusMediaURL: "/api/cbt/assets/image-2/file",
				StemAudioURL:     "/api/cbt/assets/audio-1/file",
				StimulusAudioURL: "/api/cbt/assets/audio-2/file",
			},
		},
	}

	absolutizeExamLoginResult(req, &result)

	question := result.Questions[0]
	if question.StemMediaURL != "https://mobile-api.example.sch.id/api/cbt/assets/image-1/file" {
		t.Fatalf("StemMediaURL = %q", question.StemMediaURL)
	}
	if question.StimulusMediaURL != "https://mobile-api.example.sch.id/api/cbt/assets/image-2/file" {
		t.Fatalf("StimulusMediaURL = %q", question.StimulusMediaURL)
	}
	if question.StemAudioURL != "https://mobile-api.example.sch.id/api/cbt/assets/audio-1/file" {
		t.Fatalf("StemAudioURL = %q", question.StemAudioURL)
	}
	if question.StimulusAudioURL != "https://mobile-api.example.sch.id/api/cbt/assets/audio-2/file" {
		t.Fatalf("StimulusAudioURL = %q", question.StimulusAudioURL)
	}
}
