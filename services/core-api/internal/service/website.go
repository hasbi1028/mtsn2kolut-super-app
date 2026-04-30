package service

import (
	"context"
	"fmt"
	"html"
	"regexp"
	"strings"
	"time"

	"github.com/jackc/pgx/v5/pgtype"
	db "mtsn2kolut-super-app/backend/internal/repository/postgres"
)

var websiteStripHTMLTags = regexp.MustCompile(`(?s)<[^>]*>`)
var websiteStripDangerousBlocks = []*regexp.Regexp{
	regexp.MustCompile(`(?is)<script[^>]*>.*?</script>`),
	regexp.MustCompile(`(?is)<style[^>]*>.*?</style>`),
	regexp.MustCompile(`(?is)<iframe[^>]*>.*?</iframe>`),
	regexp.MustCompile(`(?is)<object[^>]*>.*?</object>`),
	regexp.MustCompile(`(?is)<embed[^>]*>.*?</embed>`),
}
var websiteStripEventHandlers = regexp.MustCompile(`(?i)\s+on[a-z]+\s*=\s*(".*?"|'.*?'|[^\s>]+)`)
var websiteStripDangerousURLs = regexp.MustCompile(`(?i)\s(href|src)\s*=\s*(['"])\s*javascript:[^'"]*['"]`)
var websiteSlugNoise = regexp.MustCompile(`[^a-z0-9]+`)

type websiteStore interface {
	ListWebsiteContents(ctx context.Context, arg db.ListWebsiteContentsParams) ([]db.ListWebsiteContentsRow, error)
	GetWebsiteContent(ctx context.Context, id pgtype.UUID) (db.GetWebsiteContentRow, error)
	CreateWebsiteContent(ctx context.Context, arg db.CreateWebsiteContentParams) (db.WebsiteContent, error)
	UpdateWebsiteContent(ctx context.Context, arg db.UpdateWebsiteContentParams) (db.WebsiteContent, error)
	DeleteWebsiteContent(ctx context.Context, id pgtype.UUID) error
	ListPublishedWebsiteContents(ctx context.Context, arg db.ListPublishedWebsiteContentsParams) ([]db.ListPublishedWebsiteContentsRow, error)
	ListFeaturedWebsiteContents(ctx context.Context, arg db.ListFeaturedWebsiteContentsParams) ([]db.ListFeaturedWebsiteContentsRow, error)
	GetPublishedWebsiteContentBySlug(ctx context.Context, arg db.GetPublishedWebsiteContentBySlugParams) (db.GetPublishedWebsiteContentBySlugRow, error)
}

type Website struct{ q websiteStore }

func NewWebsite(q *db.Queries) *Website { return &Website{q: q} }

type SaveWebsiteContentInput struct {
	ID              pgtype.UUID
	Kind            string
	Title           string
	Slug            string
	Excerpt         string
	ContentHTML     string
	CoverImageURL   string
	IsFeatured      bool
	MetaTitle       string
	MetaDescription string
	Status          string
	ActorUsername   string
}

func (s *Website) List(ctx context.Context, kind, status, search string) ([]db.WebsiteContent, error) {
	rows, err := s.q.ListWebsiteContents(ctx, db.ListWebsiteContentsParams{
		KindFilter:   strings.TrimSpace(kind),
		StatusFilter: strings.TrimSpace(status),
		SearchQuery:  strings.TrimSpace(search),
	})
	if err != nil {
		return nil, err
	}
	items := make([]db.WebsiteContent, 0, len(rows))
	for _, row := range rows {
		items = append(items, mapWebsiteContentFromList(row))
	}
	return items, nil
}

func (s *Website) Get(ctx context.Context, id pgtype.UUID) (db.WebsiteContent, error) {
	row, err := s.q.GetWebsiteContent(ctx, id)
	if err != nil {
		return db.WebsiteContent{}, err
	}
	return mapWebsiteContentFromGet(row), nil
}

func (s *Website) Create(ctx context.Context, in SaveWebsiteContentInput) (db.WebsiteContent, error) {
	params, err := buildWebsiteCreateParams(in)
	if err != nil {
		return db.WebsiteContent{}, err
	}
	return s.q.CreateWebsiteContent(ctx, params)
}

func (s *Website) Update(ctx context.Context, in SaveWebsiteContentInput) (db.WebsiteContent, error) {
	current, err := s.q.GetWebsiteContent(ctx, in.ID)
	if err != nil {
		return db.WebsiteContent{}, err
	}
	params, err := buildWebsiteUpdateParams(mapWebsiteContentFromGet(current), in)
	if err != nil {
		return db.WebsiteContent{}, err
	}
	return s.q.UpdateWebsiteContent(ctx, params)
}

func (s *Website) Delete(ctx context.Context, id pgtype.UUID) error {
	return s.q.DeleteWebsiteContent(ctx, id)
}

func (s *Website) ListPublished(ctx context.Context, kind string, limit int32) ([]db.WebsiteContent, error) {
	if limit <= 0 {
		limit = 10
	}
	kind = normalizeWebsiteKind(kind)
	if kind == "" {
		return nil, fmt.Errorf("jenis konten tidak valid")
	}
	rows, err := s.q.ListPublishedWebsiteContents(ctx, db.ListPublishedWebsiteContentsParams{
		KindFilter: db.WebsiteContentKind(kind),
		LimitCount: limit,
	})
	if err != nil {
		return nil, err
	}
	items := make([]db.WebsiteContent, 0, len(rows))
	for _, row := range rows {
		items = append(items, mapWebsiteContentFromPublished(row))
	}
	return items, nil
}

func (s *Website) ListFeatured(ctx context.Context, kind string, limit int32) ([]db.WebsiteContent, error) {
	if limit <= 0 {
		limit = 6
	}
	kind = normalizeWebsiteKind(kind)
	if kind == "" {
		return nil, fmt.Errorf("jenis konten tidak valid")
	}
	rows, err := s.q.ListFeaturedWebsiteContents(ctx, db.ListFeaturedWebsiteContentsParams{
		KindFilter: db.WebsiteContentKind(kind),
		LimitCount: limit,
	})
	if err != nil {
		return nil, err
	}
	items := make([]db.WebsiteContent, 0, len(rows))
	for _, row := range rows {
		items = append(items, mapWebsiteContentFromFeatured(row))
	}
	return items, nil
}

func (s *Website) GetPublishedBySlug(ctx context.Context, kind, slug string) (db.WebsiteContent, error) {
	kind = normalizeWebsiteKind(kind)
	if kind == "" {
		return db.WebsiteContent{}, fmt.Errorf("jenis konten tidak valid")
	}
	slug = normalizeSlug(slug)
	if slug == "" {
		return db.WebsiteContent{}, fmt.Errorf("slug tidak valid")
	}
	row, err := s.q.GetPublishedWebsiteContentBySlug(ctx, db.GetPublishedWebsiteContentBySlugParams{
		KindFilter: db.WebsiteContentKind(kind),
		SlugValue:  slug,
	})
	if err != nil {
		return db.WebsiteContent{}, err
	}
	return mapWebsiteContentFromPublishedBySlug(row), nil
}

func mapWebsiteContentFromGet(row db.GetWebsiteContentRow) db.WebsiteContent {
	return db.WebsiteContent{
		ID:              row.ID,
		Kind:            row.Kind,
		Title:           row.Title,
		Slug:            row.Slug,
		Excerpt:         row.Excerpt,
		ContentHtml:     row.ContentHtml,
		CoverImageUrl:   row.CoverImageUrl,
		IsFeatured:      row.IsFeatured,
		MetaTitle:       row.MetaTitle,
		MetaDescription: row.MetaDescription,
		Status:          row.Status,
		PublishedAt:     row.PublishedAt,
		CreatedBy:       row.CreatedBy,
		UpdatedBy:       row.UpdatedBy,
		CreatedAt:       row.CreatedAt,
		UpdatedAt:       row.UpdatedAt,
	}
}

func mapWebsiteContentFromList(row db.ListWebsiteContentsRow) db.WebsiteContent {
	return db.WebsiteContent{
		ID:              row.ID,
		Kind:            row.Kind,
		Title:           row.Title,
		Slug:            row.Slug,
		Excerpt:         row.Excerpt,
		ContentHtml:     row.ContentHtml,
		CoverImageUrl:   row.CoverImageUrl,
		IsFeatured:      row.IsFeatured,
		MetaTitle:       row.MetaTitle,
		MetaDescription: row.MetaDescription,
		Status:          row.Status,
		PublishedAt:     row.PublishedAt,
		CreatedBy:       row.CreatedBy,
		UpdatedBy:       row.UpdatedBy,
		CreatedAt:       row.CreatedAt,
		UpdatedAt:       row.UpdatedAt,
	}
}

func mapWebsiteContentFromPublished(row db.ListPublishedWebsiteContentsRow) db.WebsiteContent {
	return db.WebsiteContent{
		ID:              row.ID,
		Kind:            row.Kind,
		Title:           row.Title,
		Slug:            row.Slug,
		Excerpt:         row.Excerpt,
		ContentHtml:     row.ContentHtml,
		CoverImageUrl:   row.CoverImageUrl,
		IsFeatured:      row.IsFeatured,
		MetaTitle:       row.MetaTitle,
		MetaDescription: row.MetaDescription,
		Status:          row.Status,
		PublishedAt:     row.PublishedAt,
		CreatedBy:       row.CreatedBy,
		UpdatedBy:       row.UpdatedBy,
		CreatedAt:       row.CreatedAt,
		UpdatedAt:       row.UpdatedAt,
	}
}

func mapWebsiteContentFromFeatured(row db.ListFeaturedWebsiteContentsRow) db.WebsiteContent {
	return db.WebsiteContent{
		ID:              row.ID,
		Kind:            row.Kind,
		Title:           row.Title,
		Slug:            row.Slug,
		Excerpt:         row.Excerpt,
		ContentHtml:     row.ContentHtml,
		CoverImageUrl:   row.CoverImageUrl,
		IsFeatured:      row.IsFeatured,
		MetaTitle:       row.MetaTitle,
		MetaDescription: row.MetaDescription,
		Status:          row.Status,
		PublishedAt:     row.PublishedAt,
		CreatedBy:       row.CreatedBy,
		UpdatedBy:       row.UpdatedBy,
		CreatedAt:       row.CreatedAt,
		UpdatedAt:       row.UpdatedAt,
	}
}

func mapWebsiteContentFromPublishedBySlug(row db.GetPublishedWebsiteContentBySlugRow) db.WebsiteContent {
	return db.WebsiteContent{
		ID:              row.ID,
		Kind:            row.Kind,
		Title:           row.Title,
		Slug:            row.Slug,
		Excerpt:         row.Excerpt,
		ContentHtml:     row.ContentHtml,
		CoverImageUrl:   row.CoverImageUrl,
		IsFeatured:      row.IsFeatured,
		MetaTitle:       row.MetaTitle,
		MetaDescription: row.MetaDescription,
		Status:          row.Status,
		PublishedAt:     row.PublishedAt,
		CreatedBy:       row.CreatedBy,
		UpdatedBy:       row.UpdatedBy,
		CreatedAt:       row.CreatedAt,
		UpdatedAt:       row.UpdatedAt,
	}
}

func normalizeWebsiteKind(value string) string {
	switch strings.TrimSpace(strings.ToLower(value)) {
	case "page", "post", "announcement":
		return strings.TrimSpace(strings.ToLower(value))
	default:
		return ""
	}
}

func normalizeWebsiteStatus(value string) string {
	switch strings.TrimSpace(strings.ToLower(value)) {
	case "", "draft":
		return "draft"
	case "published":
		return "published"
	default:
		return ""
	}
}

func sanitizeWebsiteHTML(raw string) string {
	cleaned := strings.TrimSpace(raw)
	for _, pattern := range websiteStripDangerousBlocks {
		cleaned = pattern.ReplaceAllString(cleaned, "")
	}
	cleaned = websiteStripEventHandlers.ReplaceAllString(cleaned, "")
	cleaned = websiteStripDangerousURLs.ReplaceAllString(cleaned, ` $1="#"`)
	return strings.TrimSpace(cleaned)
}

func plainExcerpt(raw string) string {
	text := html.UnescapeString(websiteStripHTMLTags.ReplaceAllString(raw, " "))
	text = strings.Join(strings.Fields(text), " ")
	if len(text) > 180 {
		return strings.TrimSpace(text[:180]) + "…"
	}
	return text
}

func normalizeSlug(value string) string {
	value = strings.TrimSpace(strings.ToLower(value))
	value = websiteSlugNoise.ReplaceAllString(value, "-")
	value = strings.Trim(value, "-")
	return value
}

func publishTime(status string, existing pgtype.Timestamptz) pgtype.Timestamptz {
	if status != "published" {
		return pgtype.Timestamptz{}
	}
	if existing.Valid {
		return existing
	}
	var ts pgtype.Timestamptz
	_ = ts.Scan(time.Now())
	return ts
}

func buildWebsiteCreateParams(in SaveWebsiteContentInput) (db.CreateWebsiteContentParams, error) {
	kind := normalizeWebsiteKind(in.Kind)
	if kind == "" {
		return db.CreateWebsiteContentParams{}, fmt.Errorf("jenis konten tidak valid")
	}
	title := strings.TrimSpace(in.Title)
	if title == "" {
		return db.CreateWebsiteContentParams{}, fmt.Errorf("judul wajib diisi")
	}
	slug := normalizeSlug(in.Slug)
	if slug == "" {
		slug = normalizeSlug(title)
	}
	if slug == "" {
		return db.CreateWebsiteContentParams{}, fmt.Errorf("slug tidak valid")
	}
	status := normalizeWebsiteStatus(in.Status)
	if status == "" {
		return db.CreateWebsiteContentParams{}, fmt.Errorf("status konten tidak valid")
	}
	contentHTML := sanitizeWebsiteHTML(in.ContentHTML)
	excerpt := strings.TrimSpace(in.Excerpt)
	if excerpt == "" {
		excerpt = plainExcerpt(contentHTML)
	}
	return db.CreateWebsiteContentParams{
		Kind:            db.WebsiteContentKind(kind),
		Title:           title,
		Slug:            slug,
		Excerpt:         excerpt,
		ContentHtml:     contentHTML,
		CoverImageUrl:   strings.TrimSpace(in.CoverImageURL),
		IsFeatured:      in.IsFeatured,
		MetaTitle:       strings.TrimSpace(in.MetaTitle),
		MetaDescription: strings.TrimSpace(in.MetaDescription),
		Status:          db.WebsiteContentStatus(status),
		PublishedAt:     publishTime(status, pgtype.Timestamptz{}),
		CreatedBy:       strings.TrimSpace(in.ActorUsername),
		UpdatedBy:       strings.TrimSpace(in.ActorUsername),
	}, nil
}

func buildWebsiteUpdateParams(current db.WebsiteContent, in SaveWebsiteContentInput) (db.UpdateWebsiteContentParams, error) {
	kind := normalizeWebsiteKind(in.Kind)
	if kind == "" {
		return db.UpdateWebsiteContentParams{}, fmt.Errorf("jenis konten tidak valid")
	}
	title := strings.TrimSpace(in.Title)
	if title == "" {
		return db.UpdateWebsiteContentParams{}, fmt.Errorf("judul wajib diisi")
	}
	slug := normalizeSlug(in.Slug)
	if slug == "" {
		slug = normalizeSlug(title)
	}
	if slug == "" {
		return db.UpdateWebsiteContentParams{}, fmt.Errorf("slug tidak valid")
	}
	status := normalizeWebsiteStatus(in.Status)
	if status == "" {
		return db.UpdateWebsiteContentParams{}, fmt.Errorf("status konten tidak valid")
	}
	contentHTML := sanitizeWebsiteHTML(in.ContentHTML)
	excerpt := strings.TrimSpace(in.Excerpt)
	if excerpt == "" {
		excerpt = plainExcerpt(contentHTML)
	}
	return db.UpdateWebsiteContentParams{
		ID:              current.ID,
		Kind:            db.WebsiteContentKind(kind),
		Title:           title,
		Slug:            slug,
		Excerpt:         excerpt,
		ContentHtml:     contentHTML,
		CoverImageUrl:   strings.TrimSpace(in.CoverImageURL),
		IsFeatured:      in.IsFeatured,
		MetaTitle:       strings.TrimSpace(in.MetaTitle),
		MetaDescription: strings.TrimSpace(in.MetaDescription),
		Status:          db.WebsiteContentStatus(status),
		PublishedAt:     publishTime(status, current.PublishedAt),
		UpdatedBy:       strings.TrimSpace(in.ActorUsername),
	}, nil
}
