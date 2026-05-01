package service

import (
	"testing"
	"time"
)

func TestBuildWebsiteCreateParamsKeepsFuturePublishSchedule(t *testing.T) {
	params, err := buildWebsiteCreateParams(SaveWebsiteContentInput{
		Kind:        "post",
		Title:       "Agenda Ujian",
		ContentHTML: "<p>Isi agenda ujian</p>",
		Status:      "published",
		PublishedAt: "2030-05-01T09:30",
	})
	if err != nil {
		t.Fatalf("buildWebsiteCreateParams() error = %v", err)
	}
	if !params.PublishedAt.Valid {
		t.Fatal("published_at should be valid for scheduled publish")
	}
	scheduled := params.PublishedAt.Time.In(time.Local)
	if scheduled.Year() != 2030 || scheduled.Month() != time.May || scheduled.Day() != 1 || scheduled.Hour() != 9 || scheduled.Minute() != 30 {
		t.Fatalf("published_at = %v, want 2030-05-01 09:30 local", scheduled)
	}
}

func TestBuildWebsiteCreateParamsRejectsInvalidSchedule(t *testing.T) {
	_, err := buildWebsiteCreateParams(SaveWebsiteContentInput{
		Kind:        "post",
		Title:       "Agenda Ujian",
		ContentHTML: "<p>Isi agenda ujian</p>",
		Status:      "published",
		PublishedAt: "jadwal-rusak",
	})
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestPublishTimeFromInputClearsDraftSchedule(t *testing.T) {
	requested, err := parseWebsitePublishedAt("2030-05-01T09:30")
	if err != nil {
		t.Fatalf("parseWebsitePublishedAt() error = %v", err)
	}
	publishedAt := publishTimeFromInput("draft", requested, requested)
	if publishedAt.Valid {
		t.Fatal("draft content should not keep a publish schedule")
	}
}
