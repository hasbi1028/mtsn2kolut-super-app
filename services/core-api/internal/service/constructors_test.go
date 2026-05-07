package service

import "testing"

func TestServiceConstructorsReturnUsableInstances(t *testing.T) {
	if svc := NewArchive(nil, ""); svc == nil || svc.storageDir != "data/archives" {
		t.Fatalf("NewArchive(empty) = %+v, want default storage dir", svc)
	}
	if svc := NewArchive(nil, "tmp/archives"); svc == nil || svc.storageDir != "tmp/archives" {
		t.Fatalf("NewArchive(custom) = %+v, want custom storage dir", svc)
	}
	if svc := NewAudit(nil); svc == nil {
		t.Fatal("NewAudit(nil) = nil, want service")
	}
	if svc := NewAuth(nil, "secret", "admin-password"); svc == nil || string(svc.jwtSecret) != "secret" || svc.adminPassword != "admin-password" {
		t.Fatalf("NewAuth() = %+v, want configured auth service", svc)
	}
	if svc := NewClassJournal(nil); svc == nil {
		t.Fatal("NewClassJournal(nil) = nil, want service")
	}
	if svc := NewClassJournalWithPool(nil); svc == nil || svc.q == nil {
		t.Fatal("NewClassJournalWithPool(nil) = nil, want service with query store")
	}
	if svc := NewDocumentCycle(nil); svc == nil {
		t.Fatal("NewDocumentCycle(nil) = nil, want service")
	}
	if svc := NewInventory(nil); svc == nil {
		t.Fatal("NewInventory(nil) = nil, want service")
	}
	if svc := NewPusakaJob(nil); svc == nil {
		t.Fatal("NewPusakaJob(nil) = nil, want service")
	}
	if svc := NewPusakaJobWithPool(nil); svc == nil || svc.q == nil {
		t.Fatal("NewPusakaJobWithPool(nil) = nil, want service with query store")
	}
	if svc := NewKesiswaan(nil, ""); svc == nil || svc.photoDir != "data/student-photos" {
		t.Fatalf("NewKesiswaan(empty) = %+v, want default photo dir", svc)
	}
	if svc := NewKesiswaan(nil, "tmp/photos"); svc == nil || svc.photoDir != "tmp/photos" {
		t.Fatalf("NewKesiswaan(custom) = %+v, want custom photo dir", svc)
	}
	if svc := NewKesiswaanWithPool(nil, ""); svc == nil || svc.photoDir != "data/student-photos" || svc.q == nil {
		t.Fatalf("NewKesiswaanWithPool(empty) = %+v, want default dir and query store", svc)
	}
	if svc := NewLetter(nil); svc == nil {
		t.Fatal("NewLetter(nil) = nil, want service")
	}
	if svc := NewLibrary(nil); svc == nil {
		t.Fatal("NewLibrary(nil) = nil, want service")
	}
	if svc := NewNotification(nil); svc == nil {
		t.Fatal("NewNotification(nil) = nil, want service")
	}
	if svc := NewParent(nil); svc == nil {
		t.Fatal("NewParent(nil) = nil, want service")
	}
	if svc := NewPortal(nil); svc == nil {
		t.Fatal("NewPortal(nil) = nil, want service")
	}
	if svc := NewPusakaSchedule(nil); svc == nil {
		t.Fatal("NewPusakaSchedule(nil) = nil, want service")
	}
	if svc := NewSetting(nil); svc == nil {
		t.Fatal("NewSetting(nil) = nil, want service")
	}
	if svc := NewStudent(nil); svc == nil {
		t.Fatal("NewStudent(nil) = nil, want service")
	}
	if svc := NewStudentCertificate(nil); svc == nil || svc.q == nil {
		t.Fatalf("NewStudentCertificate(nil) = %+v, want service with query store", svc)
	}
	if svc := NewWebsite(nil); svc == nil {
		t.Fatal("NewWebsite(nil) = nil, want service")
	}
}
