package db

import (
	"context"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
)

func TestIntegrationStudentParentAndRombelQueries(t *testing.T) {
	t.Parallel()

	tdb := setupIntegrationTestDB(t)
	ctx := context.Background()
	q := tdb.Q

	suffix := integrationSuffix()
	year, class, inactiveClass := seedIntegrationRombel(t, ctx, q, suffix)

	activeStudent, err := q.CreateStudent(ctx, CreateStudentParams{
		Nis:         "IT-NIS-" + suffix + "-1",
		Nisn:        "IT-NISN-" + suffix + "-1",
		Nama:        "Integration Student A " + suffix,
		Gender:      GenderEnumL,
		ParentName:  "Legacy Parent " + suffix,
		ParentPhone: "080000" + suffix,
		ClassID:     class.ID,
		IsActive:    true,
		Status:      StudentStatusEnumActive,
	})
	if err != nil {
		t.Fatalf("create active student: %v", err)
	}
	inactiveStudent, err := q.CreateStudent(ctx, CreateStudentParams{
		Nis:         "IT-NIS-" + suffix + "-2",
		Nisn:        "IT-NISN-" + suffix + "-2",
		Nama:        "Integration Student Z " + suffix,
		Gender:      GenderEnumP,
		ParentName:  "Legacy Parent " + suffix,
		ParentPhone: "080001" + suffix,
		ClassID:     class.ID,
		IsActive:    false,
		Status:      StudentStatusEnumAlumni,
	})
	if err != nil {
		t.Fatalf("create inactive student: %v", err)
	}

	parent, err := q.CreateParent(ctx, CreateParentParams{
		Nama:    "Integration Parent " + suffix,
		Phone:   "+62 811 " + suffix,
		Address: "Integration Address " + suffix,
	})
	if err != nil {
		t.Fatalf("create parent: %v", err)
	}
	otherParent, err := q.CreateParent(ctx, CreateParentParams{
		Nama:    "Integration Other Parent " + suffix,
		Phone:   "+62 812 " + suffix,
		Address: "Integration Other Address " + suffix,
	})
	if err != nil {
		t.Fatalf("create other parent: %v", err)
	}
	if _, err := q.UpsertParentStudentRelationship(ctx, UpsertParentStudentRelationshipParams{
		ParentID:         parent.ID,
		StudentID:        activeStudent.ID,
		Relationship:     "ayah",
		IsPrimaryContact: true,
		Notes:            "primary contact " + suffix,
	}); err != nil {
		t.Fatalf("upsert parent-student relationship: %v", err)
	}
	if err := q.LinkParentStudent(ctx, LinkParentStudentParams{ParentID: otherParent.ID, StudentID: activeStudent.ID}); err != nil {
		t.Fatalf("link other parent student: %v", err)
	}

	t.Run("school class and rombel listing queries", func(t *testing.T) {
		classes, err := q.ListSchoolClasses(ctx)
		if err != nil {
			t.Fatalf("list school classes: %v", err)
		}
		listed := findSchoolClassForIntegration(classes, class.ID)
		if listed == nil {
			t.Fatalf("created class %v not returned by ListSchoolClasses", class.ID)
		}
		if listed.Code != class.Code || listed.Name != class.Name || listed.Level != class.Level || listed.AcademicYearID != year.ID || listed.AcademicYearName != year.Name {
			t.Fatalf("listed class = %#v, want code %q name %q level %q year %v/%q", listed, class.Code, class.Name, class.Level, year.ID, year.Name)
		}

		matrixClasses, err := q.ListSubjectAssignmentMatrixClasses(ctx, year.ID)
		if err != nil {
			t.Fatalf("list subject assignment matrix classes: %v", err)
		}
		if got := findMatrixClassForIntegration(matrixClasses, class.ID); got == nil || got.Code != class.Code || got.Name != class.Name || got.Level != class.Level {
			t.Fatalf("active class not returned with expected fields by ListSubjectAssignmentMatrixClasses: got %#v", got)
		}
		if got := findMatrixClassForIntegration(matrixClasses, inactiveClass.ID); got != nil {
			t.Fatalf("inactive class %v unexpectedly returned by ListSubjectAssignmentMatrixClasses", inactiveClass.ID)
		}

		rolloverStudents, err := q.ListYearRolloverStudents(ctx, year.ID)
		if err != nil {
			t.Fatalf("list year rollover students: %v", err)
		}
		rollover := findYearRolloverStudentForIntegration(rolloverStudents, activeStudent.ID)
		if rollover == nil {
			t.Fatalf("active student %v not returned by ListYearRolloverStudents", activeStudent.ID)
		}
		if rollover.ClassID != class.ID || rollover.ClassCode != class.Code || rollover.ClassName != class.Name || rollover.ClassLevel != class.Level {
			t.Fatalf("rollover student class fields = %#v, want class %#v", rollover, class)
		}
		if got := findYearRolloverStudentForIntegration(rolloverStudents, inactiveStudent.ID); got != nil {
			t.Fatalf("inactive/non-active-status student %v unexpectedly returned by ListYearRolloverStudents", inactiveStudent.ID)
		}
	})

	t.Run("student listing and detail queries", func(t *testing.T) {
		got, err := q.GetStudentByID(ctx, activeStudent.ID)
		if err != nil {
			t.Fatalf("get student by id: %v", err)
		}
		assertStudentProjection(t, got.ID, got.Nis, got.Nisn, got.Nama, got.Gender, got.ClassID, got.ClassName.String, got.ClassCode.String, got.LinkedParentNames, got.LinkedParentCount, activeStudent.ID, activeStudent.Nis, activeStudent.Nisn, activeStudent.Nama, activeStudent.Gender, class.ID, class.Name, class.Code, 2, []string{parent.Nama, otherParent.Nama})

		students, err := q.ListStudents(ctx)
		if err != nil {
			t.Fatalf("list students: %v", err)
		}
		listed := findStudentForIntegration(students, activeStudent.ID)
		if listed == nil {
			t.Fatalf("active student %v not returned by ListStudents", activeStudent.ID)
		}
		assertStudentProjection(t, listed.ID, listed.Nis, listed.Nisn, listed.Nama, listed.Gender, listed.ClassID, listed.ClassName.String, listed.ClassCode.String, listed.LinkedParentNames, listed.LinkedParentCount, activeStudent.ID, activeStudent.Nis, activeStudent.Nisn, activeStudent.Nama, activeStudent.Gender, class.ID, class.Name, class.Code, 2, []string{parent.Nama, otherParent.Nama})

		activeByClass, err := q.ListActiveStudentsByClassID(ctx, class.ID)
		if err != nil {
			t.Fatalf("list active students by class id: %v", err)
		}
		if got := findActiveStudentByClassForIntegration(activeByClass, activeStudent.ID); got == nil || got.Nis != activeStudent.Nis || got.Nisn != activeStudent.Nisn || got.Nama != activeStudent.Nama || got.Gender != activeStudent.Gender {
			t.Fatalf("active student by class = %#v, want %v/%q", got, activeStudent.ID, activeStudent.Nama)
		}
		if got := findActiveStudentByClassForIntegration(activeByClass, inactiveStudent.ID); got != nil {
			t.Fatalf("inactive student %v unexpectedly returned by ListActiveStudentsByClassID", inactiveStudent.ID)
		}

		portalPreview, err := q.ListStudentPortalPreviewStudents(ctx)
		if err != nil {
			t.Fatalf("list student portal preview students: %v", err)
		}
		preview := findStudentPortalPreviewForIntegration(portalPreview, activeStudent.ID)
		if preview == nil || preview.ClassID != class.ID || preview.ClassName.String != class.Name || preview.ClassCode.String != class.Code || preview.Status != StudentStatusEnumActive || !preview.IsActive {
			t.Fatalf("student portal preview = %#v, want active student class fields", preview)
		}
		if got := findStudentPortalPreviewForIntegration(portalPreview, inactiveStudent.ID); got != nil {
			t.Fatalf("inactive student %v unexpectedly returned by ListStudentPortalPreviewStudents", inactiveStudent.ID)
		}

		imports, err := q.ListAcademicImportStudents(ctx)
		if err != nil {
			t.Fatalf("list academic import students: %v", err)
		}
		if got := findAcademicImportStudentForIntegration(imports, activeStudent.ID); got == nil || got.ClassID != class.ID || !got.IsActive || got.Status != StudentStatusEnumActive {
			t.Fatalf("academic import active student = %#v", got)
		}
		if got := findAcademicImportStudentForIntegration(imports, inactiveStudent.ID); got == nil || got.ClassID != class.ID || got.IsActive || got.Status != StudentStatusEnumAlumni {
			t.Fatalf("academic import inactive student = %#v", got)
		}
	})

	t.Run("parent and parent portal queries", func(t *testing.T) {
		gotParent, err := q.GetParent(ctx, parent.ID)
		if err != nil {
			t.Fatalf("get parent: %v", err)
		}
		if gotParent.ID != parent.ID || gotParent.Nama != parent.Nama || gotParent.Phone != parent.Phone || gotParent.Address != parent.Address {
			t.Fatalf("get parent = %#v, want %#v", gotParent, parent)
		}

		parents, err := q.ListParents(ctx)
		if err != nil {
			t.Fatalf("list parents: %v", err)
		}
		if got := findParentForIntegration(parents, parent.ID); got == nil || got.Nama != parent.Nama || got.Phone != parent.Phone || got.Address != parent.Address {
			t.Fatalf("listed parent = %#v, want %#v", got, parent)
		}

		foundNormalized, err := q.FindParentForNormalization(ctx, FindParentForNormalizationParams{
			Nama:    "  integration   parent " + suffix + "  ",
			Phone:   "+62 811 " + suffix,
			Address: "integration address " + suffix,
		})
		if err != nil {
			t.Fatalf("find parent for normalization: %v", err)
		}
		if foundNormalized.ID != parent.ID {
			t.Fatalf("normalized parent id = %v, want %v", foundNormalized.ID, parent.ID)
		}

		studentParents, err := q.ListStudentParents(ctx, activeStudent.ID)
		if err != nil {
			t.Fatalf("list student parents: %v", err)
		}
		studentParent := findStudentParentForIntegration(studentParents, parent.ID)
		if studentParent == nil || studentParent.Nama != parent.Nama || fmt.Sprint(studentParent.Relationship) != "ayah" || !studentParent.IsPrimaryContact || studentParent.Notes != "primary contact "+suffix {
			t.Fatalf("student parent = %#v", studentParent)
		}

		parentChildren, err := q.ListParentChildren(ctx, parent.ID)
		if err != nil {
			t.Fatalf("list parent children: %v", err)
		}
		child := findParentChildForIntegration(parentChildren, activeStudent.ID)
		if child == nil || child.Nis != activeStudent.Nis || child.Nama != activeStudent.Nama || child.ClassID != class.ID || child.ClassName.String != class.Name || fmt.Sprint(child.Relationship) != "ayah" || !child.IsPrimaryContact || child.Notes != "primary contact "+suffix {
			t.Fatalf("parent child = %#v", child)
		}

		previewParents, err := q.ListParentPortalPreviewParents(ctx)
		if err != nil {
			t.Fatalf("list parent portal preview parents: %v", err)
		}
		preview := findParentPortalPreviewForIntegration(previewParents, parent.ID)
		if preview == nil || preview.Nama != parent.Nama || preview.Phone != parent.Phone || preview.LinkedStudentCount != 1 {
			t.Fatalf("parent portal preview = %#v", preview)
		}

		access, err := q.GetParentPortalChildAccess(ctx, GetParentPortalChildAccessParams{ParentID: parent.ID, StudentID: activeStudent.ID})
		if err != nil {
			t.Fatalf("get parent portal child access: %v", err)
		}
		if access != activeStudent.ID {
			t.Fatalf("parent portal child access = %v, want %v", access, activeStudent.ID)
		}
		if _, err := q.GetParentPortalChildAccess(ctx, GetParentPortalChildAccessParams{ParentID: parent.ID, StudentID: inactiveStudent.ID}); err != pgx.ErrNoRows {
			t.Fatalf("unlinked parent portal child access err = %v, want pgx.ErrNoRows", err)
		}

		profile, err := q.GetParentPortalChildProfile(ctx, GetParentPortalChildProfileParams{ParentID: parent.ID, StudentID: activeStudent.ID})
		if err != nil {
			t.Fatalf("get parent portal child profile: %v", err)
		}
		assertStudentProjection(t, profile.ID, profile.Nis, profile.Nisn, profile.Nama, profile.Gender, profile.ClassID, profile.ClassName.String, profile.ClassCode.String, profile.LinkedParentNames, profile.LinkedParentCount, activeStudent.ID, activeStudent.Nis, activeStudent.Nisn, activeStudent.Nama, activeStudent.Gender, class.ID, class.Name, class.Code, 2, []string{parent.Nama, otherParent.Nama})
		if _, err := q.GetParentPortalChildProfile(ctx, GetParentPortalChildProfileParams{ParentID: parent.ID, StudentID: inactiveStudent.ID}); err != pgx.ErrNoRows {
			t.Fatalf("unlinked parent portal child profile err = %v, want pgx.ErrNoRows", err)
		}
	})
}

func seedIntegrationRombel(t *testing.T, ctx context.Context, q *Queries, suffix string) (AcademicYear, SchoolClass, SchoolClass) {
	t.Helper()

	year, err := q.CreateAcademicYear(ctx, CreateAcademicYearParams{
		Name:      "Integration Student Year " + suffix,
		StartDate: pgDate(2026, time.July, 1),
		EndDate:   pgDate(2027, time.June, 30),
		IsActive:  false,
	})
	if err != nil {
		t.Fatalf("create academic year: %v", err)
	}
	class, err := q.CreateSchoolClass(ctx, CreateSchoolClassParams{
		AcademicYearID: year.ID,
		Code:           "IT-STU-" + suffix,
		Name:           "Integration Student Class " + suffix,
		Level:          "8",
		IsActive:       true,
	})
	if err != nil {
		t.Fatalf("create active school class: %v", err)
	}
	inactiveClass, err := q.CreateSchoolClass(ctx, CreateSchoolClassParams{
		AcademicYearID: year.ID,
		Code:           "IT-STU-OFF-" + suffix,
		Name:           "Integration Inactive Student Class " + suffix,
		Level:          "9",
		IsActive:       false,
	})
	if err != nil {
		t.Fatalf("create inactive school class: %v", err)
	}
	return year, class, inactiveClass
}

func assertStudentProjection(t *testing.T, id pgtype.UUID, nis, nisn, nama string, gender GenderEnum, classID pgtype.UUID, className, classCode string, linkedParentNames []byte, linkedParentCount int64, wantID pgtype.UUID, wantNIS, wantNISN, wantNama string, wantGender GenderEnum, wantClassID pgtype.UUID, wantClassName, wantClassCode string, wantLinkedParentCount int64, wantLinkedParentNames []string) {
	t.Helper()

	if id != wantID || nis != wantNIS || nisn != wantNISN || nama != wantNama || gender != wantGender || classID != wantClassID || className != wantClassName || classCode != wantClassCode || linkedParentCount != wantLinkedParentCount {
		t.Fatalf("student projection = id %v nis %q nisn %q nama %q gender %q class %v/%q/%q parents %d/%q, want id %v nis %q nisn %q nama %q gender %q class %v/%q/%q parents %d/%v", id, nis, nisn, nama, gender, classID, className, classCode, linkedParentCount, string(linkedParentNames), wantID, wantNIS, wantNISN, wantNama, wantGender, wantClassID, wantClassName, wantClassCode, wantLinkedParentCount, wantLinkedParentNames)
	}
	for _, wantName := range wantLinkedParentNames {
		if !strings.Contains(string(linkedParentNames), wantName) {
			t.Fatalf("linked parent names %q does not contain %q", string(linkedParentNames), wantName)
		}
	}
}

func findSchoolClassForIntegration(classes []ListSchoolClassesRow, id pgtype.UUID) *ListSchoolClassesRow {
	for i := range classes {
		if classes[i].ID == id {
			return &classes[i]
		}
	}
	return nil
}

func findMatrixClassForIntegration(classes []ListSubjectAssignmentMatrixClassesRow, id pgtype.UUID) *ListSubjectAssignmentMatrixClassesRow {
	for i := range classes {
		if classes[i].ID == id {
			return &classes[i]
		}
	}
	return nil
}

func findYearRolloverStudentForIntegration(students []ListYearRolloverStudentsRow, id pgtype.UUID) *ListYearRolloverStudentsRow {
	for i := range students {
		if students[i].ID == id {
			return &students[i]
		}
	}
	return nil
}

func findStudentForIntegration(students []ListStudentsRow, id pgtype.UUID) *ListStudentsRow {
	for i := range students {
		if students[i].ID == id {
			return &students[i]
		}
	}
	return nil
}

func findActiveStudentByClassForIntegration(students []ListActiveStudentsByClassIDRow, id pgtype.UUID) *ListActiveStudentsByClassIDRow {
	for i := range students {
		if students[i].ID == id {
			return &students[i]
		}
	}
	return nil
}

func findStudentPortalPreviewForIntegration(students []ListStudentPortalPreviewStudentsRow, id pgtype.UUID) *ListStudentPortalPreviewStudentsRow {
	for i := range students {
		if students[i].ID == id {
			return &students[i]
		}
	}
	return nil
}

func findAcademicImportStudentForIntegration(students []ListAcademicImportStudentsRow, id pgtype.UUID) *ListAcademicImportStudentsRow {
	for i := range students {
		if students[i].ID == id {
			return &students[i]
		}
	}
	return nil
}

func findParentForIntegration(parents []Parent, id pgtype.UUID) *Parent {
	for i := range parents {
		if parents[i].ID == id {
			return &parents[i]
		}
	}
	return nil
}

func findStudentParentForIntegration(parents []ListStudentParentsRow, id pgtype.UUID) *ListStudentParentsRow {
	for i := range parents {
		if parents[i].ID == id {
			return &parents[i]
		}
	}
	return nil
}

func findParentChildForIntegration(children []ListParentChildrenRow, id pgtype.UUID) *ListParentChildrenRow {
	for i := range children {
		if children[i].ID == id {
			return &children[i]
		}
	}
	return nil
}

func findParentPortalPreviewForIntegration(parents []ListParentPortalPreviewParentsRow, id pgtype.UUID) *ListParentPortalPreviewParentsRow {
	for i := range parents {
		if parents[i].ID == id {
			return &parents[i]
		}
	}
	return nil
}
