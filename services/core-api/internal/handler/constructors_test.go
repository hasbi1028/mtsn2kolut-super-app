package handler

import "testing"

func TestHandlerConstructorsReturnHandlers(t *testing.T) {
	if NewExam(nil) == nil {
		t.Fatal("NewExam() = nil")
	}
	if NewGrade(nil) == nil {
		t.Fatal("NewGrade() = nil")
	}
	if NewHealth(nil, nil, nil) == nil {
		t.Fatal("NewHealth() = nil")
	}
	if NewLetter(nil) == nil {
		t.Fatal("NewLetter() = nil")
	}
	if NewLibrary(nil) == nil {
		t.Fatal("NewLibrary() = nil")
	}
	if NewNotification(nil) == nil {
		t.Fatal("NewNotification() = nil")
	}
}
