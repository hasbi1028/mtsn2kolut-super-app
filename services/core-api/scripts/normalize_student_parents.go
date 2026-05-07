//go:build ignore

package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"regexp"
	"strings"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
)

type studentParentSource struct {
	ID          pgtype.UUID
	Nama        string
	ParentName  string
	ParentPhone string
	Address     string
}

type parsedParent struct {
	Relationship string
	Name         string
}

var (
	spacePattern    = regexp.MustCompile(`\s+`)
	separatorRegexp = regexp.MustCompile(`[;\n|]+`)
)

func main() {
	apply := flag.Bool("apply", false, "write parent and parent_students rows")
	dryRun := flag.Bool("dry-run", true, "preview changes without writing")
	flag.Parse()

	if *apply {
		*dryRun = false
	}

	databaseURL := strings.TrimSpace(os.Getenv("DATABASE_URL"))
	if databaseURL == "" {
		log.Fatal("DATABASE_URL is required")
	}

	ctx := context.Background()
	pool, err := pgxpool.New(ctx, databaseURL)
	if err != nil {
		log.Fatal("connect database")
	}
	defer pool.Close()

	students, err := loadStudents(ctx, pool)
	if err != nil {
		log.Fatal("load students")
	}

	var parsedCount, createdParents, reusedParents, linkedRelationships int
	for _, student := range students {
		parents := parseParents(student.ParentName)
		if len(parents) == 0 {
			continue
		}
		parsedCount += len(parents)
		if *dryRun {
			fmt.Printf("DRY-RUN siswa=%s parsed_parents=%d\n", student.Nama, len(parents))
			continue
		}
		err := applyStudentParents(ctx, pool, student, parents, &createdParents, &reusedParents, &linkedRelationships)
		if err != nil {
			log.Fatalf("apply parents for one student failed: %v", err)
		}
	}

	mode := "dry-run"
	if !*dryRun {
		mode = "apply"
	}
	fmt.Printf("mode=%s students_scanned=%d parsed_relationships=%d parents_created=%d parents_reused=%d links_upserted=%d\n",
		mode, len(students), parsedCount, createdParents, reusedParents, linkedRelationships)
}

func loadStudents(ctx context.Context, pool *pgxpool.Pool) ([]studentParentSource, error) {
	rows, err := pool.Query(ctx, `
		SELECT id, nama, parent_name, parent_phone, alamat
		FROM students
		WHERE BTRIM(parent_name) <> ''
		ORDER BY nama ASC
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var students []studentParentSource
	for rows.Next() {
		var item studentParentSource
		if err := rows.Scan(&item.ID, &item.Nama, &item.ParentName, &item.ParentPhone, &item.Address); err != nil {
			return nil, err
		}
		students = append(students, item)
	}
	return students, rows.Err()
}

func applyStudentParents(ctx context.Context, pool *pgxpool.Pool, student studentParentSource, parents []parsedParent, createdParents, reusedParents, linkedRelationships *int) error {
	tx, err := pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer func() {
		_ = tx.Rollback(ctx)
	}()

	for _, parent := range parents {
		parentID, reused, err := findOrCreateParent(ctx, tx, parent.Name, student.ParentPhone, student.Address)
		if err != nil {
			return err
		}
		if reused {
			(*reusedParents)++
		} else {
			(*createdParents)++
		}

		primary, err := shouldMarkPrimary(ctx, tx, student.ID)
		if err != nil {
			return err
		}
		if err := upsertRelationship(ctx, tx, parentID, student.ID, parent.Relationship, primary); err != nil {
			return err
		}
		(*linkedRelationships)++
	}

	return tx.Commit(ctx)
}

func findOrCreateParent(ctx context.Context, tx pgx.Tx, name, phone, address string) (pgtype.UUID, bool, error) {
	var id pgtype.UUID
	err := tx.QueryRow(ctx, `
		SELECT id
		FROM parents
		WHERE LOWER(REGEXP_REPLACE(BTRIM(nama), '\s+', ' ', 'g')) = LOWER(REGEXP_REPLACE(BTRIM($1), '\s+', ' ', 'g'))
		  AND (
		    BTRIM(COALESCE($2, '')) = ''
		    OR phone = ''
		    OR REGEXP_REPLACE(phone, '\D', '', 'g') = REGEXP_REPLACE($2, '\D', '', 'g')
		  )
		  AND (
		    BTRIM(COALESCE($3, '')) = ''
		    OR address = ''
		    OR LOWER(REGEXP_REPLACE(BTRIM(address), '\s+', ' ', 'g')) = LOWER(REGEXP_REPLACE(BTRIM($3), '\s+', ' ', 'g'))
		  )
		ORDER BY
		  CASE WHEN REGEXP_REPLACE(phone, '\D', '', 'g') <> '' AND REGEXP_REPLACE(phone, '\D', '', 'g') = REGEXP_REPLACE($2, '\D', '', 'g') THEN 0 ELSE 1 END,
		  CASE WHEN address <> '' AND LOWER(REGEXP_REPLACE(BTRIM(address), '\s+', ' ', 'g')) = LOWER(REGEXP_REPLACE(BTRIM($3), '\s+', ' ', 'g')) THEN 0 ELSE 1 END,
		  created_at ASC
		LIMIT 1
	`, name, phone, address).Scan(&id)
	if err == nil {
		return id, true, nil
	}
	if err != pgx.ErrNoRows {
		return pgtype.UUID{}, false, err
	}
	err = tx.QueryRow(ctx, `
		INSERT INTO parents (nama, phone, address)
		VALUES ($1, $2, $3)
		RETURNING id
	`, name, phone, address).Scan(&id)
	return id, false, err
}

func shouldMarkPrimary(ctx context.Context, tx pgx.Tx, studentID pgtype.UUID) (bool, error) {
	var exists bool
	err := tx.QueryRow(ctx, `
		SELECT EXISTS (
			SELECT 1
			FROM parent_students
			WHERE student_id = $1 AND is_primary_contact = TRUE
		)
	`, studentID).Scan(&exists)
	return !exists, err
}

func upsertRelationship(ctx context.Context, tx pgx.Tx, parentID, studentID pgtype.UUID, relationship string, primary bool) error {
	_, err := tx.Exec(ctx, `
		INSERT INTO parent_students (parent_id, student_id, relationship, is_primary_contact, notes)
		VALUES ($1, $2, $3, $4, '')
		ON CONFLICT (parent_id, student_id) DO UPDATE
		SET relationship = EXCLUDED.relationship,
		    is_primary_contact = CASE
		        WHEN parent_students.is_primary_contact THEN TRUE
		        ELSE EXCLUDED.is_primary_contact
		    END,
		    updated_at = NOW()
	`, parentID, studentID, relationship, primary)
	return err
}

func parseParents(raw string) []parsedParent {
	cleaned := strings.TrimSpace(raw)
	if cleaned == "" {
		return nil
	}

	parts := separatorRegexp.Split(cleaned, -1)
	parents := make([]parsedParent, 0, len(parts))
	for _, part := range parts {
		parent := parseParentPart(part)
		if parent.Name == "" {
			continue
		}
		parents = append(parents, parent)
	}
	if len(parents) == 0 {
		parent := parsedParent{Relationship: "wali", Name: normalizeName(cleaned)}
		if parent.Name != "" {
			parents = append(parents, parent)
		}
	}
	return dedupeParsedParents(parents)
}

func parseParentPart(part string) parsedParent {
	value := strings.TrimSpace(part)
	if value == "" {
		return parsedParent{}
	}
	label, name, ok := strings.Cut(value, ":")
	if !ok {
		label, name, ok = strings.Cut(value, "-")
	}
	if !ok {
		return parsedParent{Relationship: "wali", Name: normalizeName(value)}
	}
	return parsedParent{
		Relationship: normalizeRelationship(label),
		Name:         normalizeName(name),
	}
}

func normalizeRelationship(label string) string {
	value := strings.ToLower(strings.TrimSpace(label))
	switch {
	case strings.Contains(value, "ayah") || strings.Contains(value, "bapak") || strings.Contains(value, "abi"):
		return "ayah"
	case strings.Contains(value, "ibu") || strings.Contains(value, "umi") || strings.Contains(value, "ummi"):
		return "ibu"
	case strings.Contains(value, "wali"):
		return "wali"
	default:
		return "lainnya"
	}
}

func normalizeName(value string) string {
	name := spacePattern.ReplaceAllString(strings.TrimSpace(value), " ")
	lower := strings.ToLower(name)
	if name == "" || name == "-" || lower == "tidak ada" || lower == "alm" || lower == "almarhum" || lower == "almarhumah" {
		return ""
	}
	return name
}

func dedupeParsedParents(parents []parsedParent) []parsedParent {
	seen := map[string]bool{}
	out := make([]parsedParent, 0, len(parents))
	for _, parent := range parents {
		key := parent.Relationship + ":" + strings.ToLower(parent.Name)
		if seen[key] {
			continue
		}
		seen[key] = true
		out = append(out, parent)
	}
	return out
}
