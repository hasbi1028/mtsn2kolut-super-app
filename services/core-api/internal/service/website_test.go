package service

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgtype"

	db "mtsn2kolut-super-app/backend/internal/repository/postgres"
)

type fakeWebsiteStore struct {
	listArg  db.ListWebsiteContentsParams
	listRows []db.ListWebsiteContentsRow
	listErr  error

	getID  pgtype.UUID
	getRow db.GetWebsiteContentRow
	getErr error

	createArg db.CreateWebsiteContentParams
	createRow db.WebsiteContent
	createErr error

	updateArg db.UpdateWebsiteContentParams
	updateRow db.WebsiteContent
	updateErr error

	deleteID  pgtype.UUID
	deleteErr error

	publishedArg  db.ListPublishedWebsiteContentsParams
	publishedRows []db.ListPublishedWebsiteContentsRow
	publishedErr  error

	featuredArg  db.ListFeaturedWebsiteContentsParams
	featuredRows []db.ListFeaturedWebsiteContentsRow
	featuredErr  error

	publishedSlugArg db.GetPublishedWebsiteContentBySlugParams
	publishedSlugRow db.GetPublishedWebsiteContentBySlugRow
	publishedSlugErr error
}

func (f *fakeWebsiteStore) ListWebsiteContents(ctx context.Context, arg db.ListWebsiteContentsParams) ([]db.ListWebsiteContentsRow, error) {
	f.listArg = arg
	return f.listRows, f.listErr
}

func (f *fakeWebsiteStore) GetWebsiteContent(ctx context.Context, id pgtype.UUID) (db.GetWebsiteContentRow, error) {
	f.getID = id
	return f.getRow, f.getErr
}

func (f *fakeWebsiteStore) CreateWebsiteContent(ctx context.Context, arg db.CreateWebsiteContentParams) (db.WebsiteContent, error) {
	f.createArg = arg
	return f.createRow, f.createErr
}

func (f *fakeWebsiteStore) UpdateWebsiteContent(ctx context.Context, arg db.UpdateWebsiteContentParams) (db.WebsiteContent, error) {
	f.updateArg = arg
	return f.updateRow, f.updateErr
}

func (f *fakeWebsiteStore) DeleteWebsiteContent(ctx context.Context, id pgtype.UUID) error {
	f.deleteID = id
	return f.deleteErr
}

func (f *fakeWebsiteStore) ListPublishedWebsiteContents(ctx context.Context, arg db.ListPublishedWebsiteContentsParams) ([]db.ListPublishedWebsiteContentsRow, error) {
	f.publishedArg = arg
	return f.publishedRows, f.publishedErr
}

func (f *fakeWebsiteStore) ListFeaturedWebsiteContents(ctx context.Context, arg db.ListFeaturedWebsiteContentsParams) ([]db.ListFeaturedWebsiteContentsRow, error) {
	f.featuredArg = arg
	return f.featuredRows, f.featuredErr
}

func (f *fakeWebsiteStore) GetPublishedWebsiteContentBySlug(ctx context.Context, arg db.GetPublishedWebsiteContentBySlugParams) (db.GetPublishedWebsiteContentBySlugRow, error) {
	f.publishedSlugArg = arg
	return f.publishedSlugRow, f.publishedSlugErr
}

func websiteTestUUID(seed byte) pgtype.UUID {
	return pgtype.UUID{Bytes: [16]byte{seed}, Valid: true}
}

func websiteTestTimestamp(year int, month time.Month, day int, hour int) pgtype.Timestamptz {
	return pgtype.Timestamptz{
		Time:  time.Date(year, month, day, hour, 0, 0, 0, time.UTC),
		Valid: true,
	}
}

func websiteContentFixture(id pgtype.UUID) db.WebsiteContent {
	return db.WebsiteContent{
		ID:              id,
		Kind:            db.WebsiteContentKindPost,
		Title:           "PPDB 2026",
		Slug:            "ppdb-2026",
		Excerpt:         "Informasi PPDB",
		ContentHtml:     "<p>Informasi PPDB</p>",
		CoverImageUrl:   "/uploads/ppdb.jpg",
		Status:          db.WebsiteContentStatusPublished,
		PublishedAt:     websiteTestTimestamp(2026, time.May, 1, 8),
		CreatedBy:       "admin",
		UpdatedBy:       "editor",
		CreatedAt:       websiteTestTimestamp(2026, time.April, 1, 8),
		UpdatedAt:       websiteTestTimestamp(2026, time.May, 1, 9),
		IsFeatured:      true,
		MetaTitle:       "PPDB MTsN 2",
		MetaDescription: "Penerimaan peserta didik baru",
	}
}

func websiteGetRowFromContent(content db.WebsiteContent) db.GetWebsiteContentRow {
	return db.GetWebsiteContentRow{
		ID:              content.ID,
		Kind:            content.Kind,
		Title:           content.Title,
		Slug:            content.Slug,
		Excerpt:         content.Excerpt,
		ContentHtml:     content.ContentHtml,
		CoverImageUrl:   content.CoverImageUrl,
		IsFeatured:      content.IsFeatured,
		MetaTitle:       content.MetaTitle,
		MetaDescription: content.MetaDescription,
		Status:          content.Status,
		PublishedAt:     content.PublishedAt,
		CreatedBy:       content.CreatedBy,
		UpdatedBy:       content.UpdatedBy,
		CreatedAt:       content.CreatedAt,
		UpdatedAt:       content.UpdatedAt,
	}
}

func websiteListRowFromContent(content db.WebsiteContent) db.ListWebsiteContentsRow {
	return db.ListWebsiteContentsRow{
		ID:              content.ID,
		Kind:            content.Kind,
		Title:           content.Title,
		Slug:            content.Slug,
		Excerpt:         content.Excerpt,
		ContentHtml:     content.ContentHtml,
		CoverImageUrl:   content.CoverImageUrl,
		IsFeatured:      content.IsFeatured,
		MetaTitle:       content.MetaTitle,
		MetaDescription: content.MetaDescription,
		Status:          content.Status,
		PublishedAt:     content.PublishedAt,
		CreatedBy:       content.CreatedBy,
		UpdatedBy:       content.UpdatedBy,
		CreatedAt:       content.CreatedAt,
		UpdatedAt:       content.UpdatedAt,
	}
}

func websitePublishedRowFromContent(content db.WebsiteContent) db.ListPublishedWebsiteContentsRow {
	return db.ListPublishedWebsiteContentsRow{
		ID:              content.ID,
		Kind:            content.Kind,
		Title:           content.Title,
		Slug:            content.Slug,
		Excerpt:         content.Excerpt,
		ContentHtml:     content.ContentHtml,
		CoverImageUrl:   content.CoverImageUrl,
		IsFeatured:      content.IsFeatured,
		MetaTitle:       content.MetaTitle,
		MetaDescription: content.MetaDescription,
		Status:          content.Status,
		PublishedAt:     content.PublishedAt,
		CreatedBy:       content.CreatedBy,
		UpdatedBy:       content.UpdatedBy,
		CreatedAt:       content.CreatedAt,
		UpdatedAt:       content.UpdatedAt,
	}
}

func websiteFeaturedRowFromContent(content db.WebsiteContent) db.ListFeaturedWebsiteContentsRow {
	return db.ListFeaturedWebsiteContentsRow{
		ID:              content.ID,
		Kind:            content.Kind,
		Title:           content.Title,
		Slug:            content.Slug,
		Excerpt:         content.Excerpt,
		ContentHtml:     content.ContentHtml,
		CoverImageUrl:   content.CoverImageUrl,
		IsFeatured:      content.IsFeatured,
		MetaTitle:       content.MetaTitle,
		MetaDescription: content.MetaDescription,
		Status:          content.Status,
		PublishedAt:     content.PublishedAt,
		CreatedBy:       content.CreatedBy,
		UpdatedBy:       content.UpdatedBy,
		CreatedAt:       content.CreatedAt,
		UpdatedAt:       content.UpdatedAt,
	}
}

func websiteSlugRowFromContent(content db.WebsiteContent) db.GetPublishedWebsiteContentBySlugRow {
	return db.GetPublishedWebsiteContentBySlugRow{
		ID:              content.ID,
		Kind:            content.Kind,
		Title:           content.Title,
		Slug:            content.Slug,
		Excerpt:         content.Excerpt,
		ContentHtml:     content.ContentHtml,
		CoverImageUrl:   content.CoverImageUrl,
		IsFeatured:      content.IsFeatured,
		MetaTitle:       content.MetaTitle,
		MetaDescription: content.MetaDescription,
		Status:          content.Status,
		PublishedAt:     content.PublishedAt,
		CreatedBy:       content.CreatedBy,
		UpdatedBy:       content.UpdatedBy,
		CreatedAt:       content.CreatedAt,
		UpdatedAt:       content.UpdatedAt,
	}
}

func TestWebsiteNormalizeParseAndTextHelpers(t *testing.T) {
	if got := normalizeWebsiteKind(" POST "); got != "post" {
		t.Fatalf("normalizeWebsiteKind() = %q, want post", got)
	}
	if got := normalizeWebsiteKind("memo"); got != "" {
		t.Fatalf("normalizeWebsiteKind(invalid) = %q, want empty", got)
	}
	if got := normalizeWebsiteStatus(""); got != "draft" {
		t.Fatalf("normalizeWebsiteStatus(empty) = %q, want draft", got)
	}
	if got := normalizeWebsiteStatus(" Published "); got != "published" {
		t.Fatalf("normalizeWebsiteStatus(published) = %q, want published", got)
	}
	if got := normalizeWebsiteStatus("review"); got != "" {
		t.Fatalf("normalizeWebsiteStatus(invalid) = %q, want empty", got)
	}
	if got := normalizeSlug(" MTsN 2: PPDB 2026! "); got != "mtsn-2-ppdb-2026" {
		t.Fatalf("normalizeSlug() = %q, want mtsn-2-ppdb-2026", got)
	}

	parsed, err := parseWebsitePublishedAt("2026-05-01T08:30:00Z")
	if err != nil {
		t.Fatalf("parseWebsitePublishedAt(rfc3339) error = %v", err)
	}
	if !parsed.Valid || parsed.Time.Format(time.RFC3339) != "2026-05-01T08:30:00Z" {
		t.Fatalf("parseWebsitePublishedAt(rfc3339) = %v, want 2026-05-01T08:30:00Z", parsed.Time)
	}
	localParsed, err := parseWebsitePublishedAt("2026-05-01 08:30")
	if err != nil {
		t.Fatalf("parseWebsitePublishedAt(local) error = %v", err)
	}
	if !localParsed.Valid || localParsed.Time.Format("2006-01-02 15:04") != "2026-05-01 08:30" {
		t.Fatalf("parseWebsitePublishedAt(local) = %v, want local datetime", localParsed.Time)
	}
	if _, err := parseWebsitePublishedAt("bad"); err == nil || err.Error() != "jadwal terbit tidak valid" {
		t.Fatalf("parseWebsitePublishedAt(bad) error = %v, want invalid schedule", err)
	}

	cleaned := sanitizeWebsiteHTML(`<p onclick="alert(1)">A<script>x</script><a href="javascript:alert(1)">B</a><img src='javascript:evil()' onerror=boom></p>`)
	for _, blocked := range []string{"script", "onclick", "onerror", "javascript:"} {
		if strings.Contains(strings.ToLower(cleaned), blocked) {
			t.Fatalf("sanitizeWebsiteHTML() = %q, still contains %q", cleaned, blocked)
		}
	}
	if !strings.Contains(cleaned, `href="#"`) || !strings.Contains(cleaned, `src="#"`) {
		t.Fatalf("sanitizeWebsiteHTML() = %q, want dangerous URLs replaced", cleaned)
	}
	if got := plainExcerpt("<p>Satu&nbsp;dua</p><p>tiga</p>"); got != "Satu dua tiga" {
		t.Fatalf("plainExcerpt() = %q, want normalized text", got)
	}
	if got := plainExcerpt(strings.Repeat("a", 181)); !strings.HasSuffix(got, "…") || len([]rune(got)) != 181 {
		t.Fatalf("plainExcerpt(long) length/suffix = %d/%q, want 181 runes with ellipsis", len([]rune(got)), got)
	}

	existing := websiteTestTimestamp(2026, time.May, 1, 8)
	requested := websiteTestTimestamp(2026, time.May, 2, 8)
	if got := publishTime("draft", existing); got.Valid {
		t.Fatalf("publishTime(draft) = %v, want invalid timestamp", got)
	}
	if got := publishTime("published", existing); !got.Valid || !got.Time.Equal(existing.Time) {
		t.Fatalf("publishTime(published existing) = %v, want existing", got)
	}
	if got := publishTime("published", pgtype.Timestamptz{}); !got.Valid {
		t.Fatalf("publishTime(published without existing) = %v, want generated timestamp", got)
	}
	if got := publishTimeFromInput("published", requested, existing); !got.Valid || !got.Time.Equal(requested.Time) {
		t.Fatalf("publishTimeFromInput(requested) = %v, want requested", got)
	}
	if got := publishTimeFromInput("published", pgtype.Timestamptz{}, existing); !got.Valid || !got.Time.Equal(existing.Time) {
		t.Fatalf("publishTimeFromInput(existing) = %v, want existing", got)
	}
}

func TestWebsiteBuildCreateParams(t *testing.T) {
	params, err := buildWebsiteCreateParams(SaveWebsiteContentInput{
		Kind:            " POST ",
		Title:           " PPDB MTsN 2 ",
		ContentHTML:     `<p onclick=x>Isi <b>berita</b></p><script>x</script>`,
		CoverImageURL:   " /uploads/ppdb.jpg ",
		IsFeatured:      true,
		MetaTitle:       " Judul SEO ",
		MetaDescription: " Deskripsi SEO ",
		Status:          "published",
		PublishedAt:     "2026-05-01T09:30:00Z",
		ActorUsername:   " admin ",
	})
	if err != nil {
		t.Fatalf("buildWebsiteCreateParams() error = %v", err)
	}
	if params.Kind != db.WebsiteContentKindPost || params.Title != "PPDB MTsN 2" || params.Slug != "ppdb-mtsn-2" {
		t.Fatalf("buildWebsiteCreateParams() identity = %+v, want normalized kind/title/slug", params)
	}
	if params.Excerpt != "Isi berita" {
		t.Fatalf("buildWebsiteCreateParams() excerpt = %q, want auto excerpt", params.Excerpt)
	}
	if strings.Contains(params.ContentHtml, "script") || strings.Contains(params.ContentHtml, "onclick") {
		t.Fatalf("buildWebsiteCreateParams() content_html = %q, want sanitized HTML", params.ContentHtml)
	}
	if !params.IsFeatured || params.CoverImageUrl != "/uploads/ppdb.jpg" || params.MetaTitle != "Judul SEO" || params.MetaDescription != "Deskripsi SEO" {
		t.Fatalf("buildWebsiteCreateParams() metadata = %+v, want trimmed metadata", params)
	}
	if params.Status != db.WebsiteContentStatusPublished || !params.PublishedAt.Valid || params.PublishedAt.Time.Format(time.RFC3339) != "2026-05-01T09:30:00Z" {
		t.Fatalf("buildWebsiteCreateParams() status/published_at = %s/%v, want published at requested time", params.Status, params.PublishedAt)
	}
	if params.CreatedBy != "admin" || params.UpdatedBy != "admin" {
		t.Fatalf("buildWebsiteCreateParams() actor = %q/%q, want admin", params.CreatedBy, params.UpdatedBy)
	}
}

func TestWebsiteBuildCreateParamsValidationErrors(t *testing.T) {
	tests := []struct {
		name    string
		input   SaveWebsiteContentInput
		wantErr string
	}{
		{
			name:    "invalid kind",
			input:   SaveWebsiteContentInput{Kind: "memo", Title: "Judul", Status: "draft"},
			wantErr: "jenis konten tidak valid",
		},
		{
			name:    "missing title",
			input:   SaveWebsiteContentInput{Kind: "post", Status: "draft"},
			wantErr: "judul wajib diisi",
		},
		{
			name:    "invalid slug",
			input:   SaveWebsiteContentInput{Kind: "post", Title: "!!!", Status: "draft"},
			wantErr: "slug tidak valid",
		},
		{
			name:    "invalid status",
			input:   SaveWebsiteContentInput{Kind: "post", Title: "Judul", Status: "review"},
			wantErr: "status konten tidak valid",
		},
		{
			name:    "invalid published at",
			input:   SaveWebsiteContentInput{Kind: "post", Title: "Judul", Status: "published", PublishedAt: "besok"},
			wantErr: "jadwal terbit tidak valid",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := buildWebsiteCreateParams(tt.input)
			if err == nil || err.Error() != tt.wantErr {
				t.Fatalf("buildWebsiteCreateParams() error = %v, want %q", err, tt.wantErr)
			}
		})
	}
}

func TestWebsiteBuildUpdateParams(t *testing.T) {
	current := websiteContentFixture(websiteTestUUID(7))
	params, err := buildWebsiteUpdateParams(current, SaveWebsiteContentInput{
		Kind:          "announcement",
		Title:         " Pengumuman Rapat ",
		Slug:          " rapat-komite ",
		ContentHTML:   "<p>Rapat komite</p>",
		Excerpt:       " Ringkas ",
		Status:        "published",
		ActorUsername: " operator ",
	})
	if err != nil {
		t.Fatalf("buildWebsiteUpdateParams() error = %v", err)
	}
	if params.ID != current.ID || params.Kind != db.WebsiteContentKindAnnouncement || params.Title != "Pengumuman Rapat" || params.Slug != "rapat-komite" {
		t.Fatalf("buildWebsiteUpdateParams() identity = %+v, want preserved id and normalized fields", params)
	}
	if !params.PublishedAt.Valid || !params.PublishedAt.Time.Equal(current.PublishedAt.Time) {
		t.Fatalf("buildWebsiteUpdateParams() published_at = %v, want existing timestamp", params.PublishedAt)
	}
	if params.UpdatedBy != "operator" || params.Excerpt != "Ringkas" {
		t.Fatalf("buildWebsiteUpdateParams() updated_by/excerpt = %q/%q, want trimmed fields", params.UpdatedBy, params.Excerpt)
	}

	draftParams, err := buildWebsiteUpdateParams(current, SaveWebsiteContentInput{
		Kind:        "post",
		Title:       " Draft Baru ",
		ContentHTML: "<p>Draf</p>",
		Status:      "draft",
	})
	if err != nil {
		t.Fatalf("buildWebsiteUpdateParams(draft) error = %v", err)
	}
	if draftParams.PublishedAt.Valid {
		t.Fatalf("buildWebsiteUpdateParams(draft) published_at = %v, want invalid timestamp", draftParams.PublishedAt)
	}
}

func TestWebsiteBuildUpdateParamsValidationErrors(t *testing.T) {
	current := websiteContentFixture(websiteTestUUID(8))
	tests := []struct {
		name    string
		input   SaveWebsiteContentInput
		wantErr string
	}{
		{
			name:    "invalid kind",
			input:   SaveWebsiteContentInput{Kind: "memo", Title: "Judul", Status: "draft"},
			wantErr: "jenis konten tidak valid",
		},
		{
			name:    "missing title",
			input:   SaveWebsiteContentInput{Kind: "post", Status: "draft"},
			wantErr: "judul wajib diisi",
		},
		{
			name:    "invalid slug",
			input:   SaveWebsiteContentInput{Kind: "post", Title: "!!!", Status: "draft"},
			wantErr: "slug tidak valid",
		},
		{
			name:    "invalid status",
			input:   SaveWebsiteContentInput{Kind: "post", Title: "Judul", Status: "review"},
			wantErr: "status konten tidak valid",
		},
		{
			name:    "invalid published at",
			input:   SaveWebsiteContentInput{Kind: "post", Title: "Judul", Status: "published", PublishedAt: "besok"},
			wantErr: "jadwal terbit tidak valid",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := buildWebsiteUpdateParams(current, tt.input)
			if err == nil || err.Error() != tt.wantErr {
				t.Fatalf("buildWebsiteUpdateParams() error = %v, want %q", err, tt.wantErr)
			}
		})
	}
}

func TestWebsiteServiceListGetAndDelete(t *testing.T) {
	content := websiteContentFixture(websiteTestUUID(1))
	store := &fakeWebsiteStore{
		listRows: []db.ListWebsiteContentsRow{websiteListRowFromContent(content)},
		getRow:   websiteGetRowFromContent(content),
	}
	svc := &Website{q: store}

	items, err := svc.List(context.Background(), " post ", " published ", " ppdb ")
	if err != nil {
		t.Fatalf("List() error = %v", err)
	}
	if store.listArg.KindFilter != "post" || store.listArg.StatusFilter != "published" || store.listArg.SearchQuery != "ppdb" {
		t.Fatalf("List() arg = %+v, want trimmed filters", store.listArg)
	}
	if len(items) != 1 || items[0].ID != content.ID || items[0].Title != content.Title || !items[0].IsFeatured {
		t.Fatalf("List() items = %+v, want mapped content", items)
	}

	got, err := svc.Get(context.Background(), content.ID)
	if err != nil {
		t.Fatalf("Get() error = %v", err)
	}
	if store.getID != content.ID || got.Slug != content.Slug || got.MetaTitle != content.MetaTitle {
		t.Fatalf("Get() = %+v with id %v, want mapped content", got, store.getID)
	}

	if err := svc.Delete(context.Background(), content.ID); err != nil {
		t.Fatalf("Delete() error = %v", err)
	}
	if store.deleteID != content.ID {
		t.Fatalf("Delete() id = %v, want %v", store.deleteID, content.ID)
	}
}

func TestWebsiteServiceReadAndDeletePropagateStoreErrors(t *testing.T) {
	contentID := websiteTestUUID(9)
	expectedErr := errors.New("store failed")

	svc := &Website{q: &fakeWebsiteStore{listErr: expectedErr}}
	if _, err := svc.List(context.Background(), "post", "draft", "ppdb"); !errors.Is(err, expectedErr) {
		t.Fatalf("List() error = %v, want %v", err, expectedErr)
	}

	svc = &Website{q: &fakeWebsiteStore{getErr: expectedErr}}
	if _, err := svc.Get(context.Background(), contentID); !errors.Is(err, expectedErr) {
		t.Fatalf("Get() error = %v, want %v", err, expectedErr)
	}

	svc = &Website{q: &fakeWebsiteStore{deleteErr: expectedErr}}
	if err := svc.Delete(context.Background(), contentID); !errors.Is(err, expectedErr) {
		t.Fatalf("Delete() error = %v, want %v", err, expectedErr)
	}
}

func TestWebsiteServiceCreateAndUpdate(t *testing.T) {
	content := websiteContentFixture(websiteTestUUID(2))
	store := &fakeWebsiteStore{
		createRow: content,
		getRow:    websiteGetRowFromContent(content),
		updateRow: db.WebsiteContent{
			ID:    content.ID,
			Title: "Rapat Komite",
			Slug:  "rapat-komite",
		},
	}
	svc := &Website{q: store}

	created, err := svc.Create(context.Background(), SaveWebsiteContentInput{
		Kind:          "post",
		Title:         " Berita Baru ",
		ContentHTML:   `<p onclick=x>Isi</p>`,
		Status:        "draft",
		ActorUsername: " admin ",
	})
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}
	if created.ID != content.ID {
		t.Fatalf("Create() = %+v, want store row", created)
	}
	if store.createArg.Title != "Berita Baru" || store.createArg.Slug != "berita-baru" || store.createArg.CreatedBy != "admin" {
		t.Fatalf("Create() arg = %+v, want normalized create params", store.createArg)
	}
	if strings.Contains(store.createArg.ContentHtml, "onclick") {
		t.Fatalf("Create() content_html = %q, want sanitized HTML", store.createArg.ContentHtml)
	}

	updated, err := svc.Update(context.Background(), SaveWebsiteContentInput{
		ID:            content.ID,
		Kind:          "announcement",
		Title:         " Rapat Komite ",
		ContentHTML:   "<p>Rapat</p>",
		Status:        "published",
		ActorUsername: " editor ",
	})
	if err != nil {
		t.Fatalf("Update() error = %v", err)
	}
	if updated.Title != "Rapat Komite" {
		t.Fatalf("Update() = %+v, want store update row", updated)
	}
	if store.getID != content.ID || store.updateArg.ID != content.ID {
		t.Fatalf("Update() ids get/update = %v/%v, want %v", store.getID, store.updateArg.ID, content.ID)
	}
	if store.updateArg.Kind != db.WebsiteContentKindAnnouncement || store.updateArg.UpdatedBy != "editor" || !store.updateArg.PublishedAt.Valid {
		t.Fatalf("Update() arg = %+v, want normalized update params", store.updateArg)
	}
}

func TestWebsiteServiceCreateAndUpdatePropagateErrors(t *testing.T) {
	content := websiteContentFixture(websiteTestUUID(10))
	expectedErr := errors.New("store failed")

	svc := &Website{q: &fakeWebsiteStore{createErr: expectedErr}}
	_, err := svc.Create(context.Background(), SaveWebsiteContentInput{
		Kind:        "post",
		Title:       "Berita Baru",
		ContentHTML: "<p>Isi</p>",
		Status:      "draft",
	})
	if !errors.Is(err, expectedErr) {
		t.Fatalf("Create() store error = %v, want %v", err, expectedErr)
	}

	_, err = svc.Create(context.Background(), SaveWebsiteContentInput{Kind: "memo", Title: "Berita Baru", Status: "draft"})
	if err == nil || err.Error() != "jenis konten tidak valid" {
		t.Fatalf("Create() validation error = %v, want invalid kind", err)
	}

	svc = &Website{q: &fakeWebsiteStore{getErr: expectedErr}}
	_, err = svc.Update(context.Background(), SaveWebsiteContentInput{
		ID:          content.ID,
		Kind:        "post",
		Title:       "Berita Baru",
		ContentHTML: "<p>Isi</p>",
		Status:      "draft",
	})
	if !errors.Is(err, expectedErr) {
		t.Fatalf("Update() get error = %v, want %v", err, expectedErr)
	}

	svc = &Website{q: &fakeWebsiteStore{getRow: websiteGetRowFromContent(content)}}
	_, err = svc.Update(context.Background(), SaveWebsiteContentInput{ID: content.ID, Kind: "memo", Title: "Berita Baru", Status: "draft"})
	if err == nil || err.Error() != "jenis konten tidak valid" {
		t.Fatalf("Update() validation error = %v, want invalid kind", err)
	}

	svc = &Website{q: &fakeWebsiteStore{getRow: websiteGetRowFromContent(content), updateErr: expectedErr}}
	_, err = svc.Update(context.Background(), SaveWebsiteContentInput{
		ID:          content.ID,
		Kind:        "post",
		Title:       "Berita Baru",
		ContentHTML: "<p>Isi</p>",
		Status:      "draft",
	})
	if !errors.Is(err, expectedErr) {
		t.Fatalf("Update() store error = %v, want %v", err, expectedErr)
	}
}

func TestWebsiteServicePublishedQueries(t *testing.T) {
	content := websiteContentFixture(websiteTestUUID(3))
	store := &fakeWebsiteStore{
		publishedRows:    []db.ListPublishedWebsiteContentsRow{websitePublishedRowFromContent(content)},
		featuredRows:     []db.ListFeaturedWebsiteContentsRow{websiteFeaturedRowFromContent(content)},
		publishedSlugRow: websiteSlugRowFromContent(content),
	}
	svc := &Website{q: store}

	published, err := svc.ListPublished(context.Background(), " POST ", 0)
	if err != nil {
		t.Fatalf("ListPublished() error = %v", err)
	}
	if store.publishedArg.KindFilter != db.WebsiteContentKindPost || store.publishedArg.LimitCount != 10 {
		t.Fatalf("ListPublished() arg = %+v, want normalized kind and default limit", store.publishedArg)
	}
	if len(published) != 1 || published[0].Title != content.Title {
		t.Fatalf("ListPublished() = %+v, want mapped content", published)
	}
	if _, err := svc.ListPublished(context.Background(), "memo", 10); err == nil || err.Error() != "jenis konten tidak valid" {
		t.Fatalf("ListPublished(invalid) error = %v, want invalid kind", err)
	}

	store.publishedErr = errors.New("published failed")
	if _, err := svc.ListPublished(context.Background(), "post", 2); err == nil || err.Error() != "published failed" {
		t.Fatalf("ListPublished(store error) error = %v, want store error", err)
	}
	store.publishedErr = nil

	featured, err := svc.ListFeatured(context.Background(), "announcement", -1)
	if err != nil {
		t.Fatalf("ListFeatured() error = %v", err)
	}
	if store.featuredArg.KindFilter != db.WebsiteContentKindAnnouncement || store.featuredArg.LimitCount != 6 {
		t.Fatalf("ListFeatured() arg = %+v, want normalized kind and default limit", store.featuredArg)
	}
	if len(featured) != 1 || !featured[0].IsFeatured {
		t.Fatalf("ListFeatured() = %+v, want mapped featured content", featured)
	}
	if _, err := svc.ListFeatured(context.Background(), "memo", 1); err == nil || err.Error() != "jenis konten tidak valid" {
		t.Fatalf("ListFeatured(invalid) error = %v, want invalid kind", err)
	}
	store.featuredErr = errors.New("featured failed")
	if _, err := svc.ListFeatured(context.Background(), "post", 2); err == nil || err.Error() != "featured failed" {
		t.Fatalf("ListFeatured(store error) error = %v, want store error", err)
	}
	store.featuredErr = nil

	got, err := svc.GetPublishedBySlug(context.Background(), " page ", " PPDB 2026!! ")
	if err != nil {
		t.Fatalf("GetPublishedBySlug() error = %v", err)
	}
	if store.publishedSlugArg.KindFilter != db.WebsiteContentKindPage || store.publishedSlugArg.SlugValue != "ppdb-2026" {
		t.Fatalf("GetPublishedBySlug() arg = %+v, want normalized kind and slug", store.publishedSlugArg)
	}
	if got.ID != content.ID || got.MetaDescription != content.MetaDescription {
		t.Fatalf("GetPublishedBySlug() = %+v, want mapped content", got)
	}
	if _, err := svc.GetPublishedBySlug(context.Background(), "memo", "ppdb"); err == nil || err.Error() != "jenis konten tidak valid" {
		t.Fatalf("GetPublishedBySlug(invalid kind) error = %v, want invalid kind", err)
	}
	if _, err := svc.GetPublishedBySlug(context.Background(), "post", "!!!"); err == nil || err.Error() != "slug tidak valid" {
		t.Fatalf("GetPublishedBySlug(invalid slug) error = %v, want invalid slug", err)
	}
	store.publishedSlugErr = errors.New("slug lookup failed")
	if _, err := svc.GetPublishedBySlug(context.Background(), "post", "ppdb-2026"); err == nil || err.Error() != "slug lookup failed" {
		t.Fatalf("GetPublishedBySlug(store error) error = %v, want store error", err)
	}
}
