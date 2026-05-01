package handler

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/golang-jwt/jwt/v5"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"

	db "mtsn2kolut-super-app/backend/internal/repository/postgres"
	"mtsn2kolut-super-app/backend/internal/service"
)

type fakeWebsiteService struct {
	*service.Website

	listKind          string
	listStatus        string
	listSearch        string
	listRows          []db.WebsiteContent
	listErr           error
	createInput       service.SaveWebsiteContentInput
	createRow         db.WebsiteContent
	createErr         error
	updateInput       service.SaveWebsiteContentInput
	updateRow         db.WebsiteContent
	updateErr         error
	deleteID          pgtype.UUID
	deleteErr         error
	featuredKind      string
	featuredLimit     int32
	featuredRows      []db.WebsiteContent
	featuredErr       error
	publishedKind     string
	publishedLimit    int32
	publishedRows     []db.WebsiteContent
	publishedErr      error
	publishedSlugKind string
	publishedSlug     string
	publishedRow      db.WebsiteContent
	publishedSlugErr  error
}

func (f *fakeWebsiteService) List(_ context.Context, kind, status, search string) ([]db.WebsiteContent, error) {
	f.listKind = kind
	f.listStatus = status
	f.listSearch = search
	return f.listRows, f.listErr
}

func (f *fakeWebsiteService) Create(_ context.Context, in service.SaveWebsiteContentInput) (db.WebsiteContent, error) {
	f.createInput = in
	return f.createRow, f.createErr
}

func (f *fakeWebsiteService) Update(_ context.Context, in service.SaveWebsiteContentInput) (db.WebsiteContent, error) {
	f.updateInput = in
	return f.updateRow, f.updateErr
}

func (f *fakeWebsiteService) Delete(_ context.Context, id pgtype.UUID) error {
	f.deleteID = id
	return f.deleteErr
}

func (f *fakeWebsiteService) ListFeatured(_ context.Context, kind string, limit int32) ([]db.WebsiteContent, error) {
	f.featuredKind = kind
	f.featuredLimit = limit
	return f.featuredRows, f.featuredErr
}

func (f *fakeWebsiteService) ListPublished(_ context.Context, kind string, limit int32) ([]db.WebsiteContent, error) {
	f.publishedKind = kind
	f.publishedLimit = limit
	return f.publishedRows, f.publishedErr
}

func (f *fakeWebsiteService) GetPublishedBySlug(_ context.Context, kind, slug string) (db.WebsiteContent, error) {
	f.publishedSlugKind = kind
	f.publishedSlug = slug
	return f.publishedRow, f.publishedSlugErr
}

func websiteContent(id pgtype.UUID, kind, title string) db.WebsiteContent {
	return db.WebsiteContent{ID: id, Kind: db.WebsiteContentKind(kind), Title: title, Slug: "konten", Status: db.WebsiteContentStatusPublished}
}

func websiteAdminRequest(method, target, body string) *http.Request {
	req := httptest.NewRequest(method, target, strings.NewReader(body))
	return withClaims(req, jwt.MapClaims{"roles": []any{"admin"}, "usr": "editor"})
}

func TestWebsiteSuccessHandlersForwardPayloads(t *testing.T) {
	id := handlerTestUUID(146)
	fake := &fakeWebsiteService{
		Website:       &service.Website{},
		listRows:      []db.WebsiteContent{websiteContent(id, "post", "Berita")},
		createRow:     websiteContent(id, "post", "Berita Baru"),
		updateRow:     websiteContent(id, "post", "Berita Revisi"),
		featuredRows:  []db.WebsiteContent{websiteContent(id, "post", "Unggulan")},
		publishedRows: []db.WebsiteContent{websiteContent(id, "post", "Terbit")},
		publishedRow:  websiteContent(id, "post", "Detail"),
	}
	h := &Website{svc: fake}

	rec := httptest.NewRecorder()
	h.List(rec, websiteAdminRequest(http.MethodGet, "/api/website/content?kind=post&status=published&search=ujian", ""))
	if rec.Code != http.StatusOK {
		t.Fatalf("List status = %d, want 200; body=%s", rec.Code, rec.Body.String())
	}
	if fake.listKind != "post" || fake.listStatus != "published" || fake.listSearch != "ujian" {
		t.Fatalf("List filters = (%q, %q, %q), want query filters", fake.listKind, fake.listStatus, fake.listSearch)
	}

	body := `{"kind":"post","title":"Berita Baru","slug":"berita-baru","excerpt":"Ringkas","content_html":"<p>Isi</p>","cover_image_url":"/a.jpg","is_featured":true,"meta_title":"Meta","meta_description":"Desc","status":"published","published_at":"2026-05-01T08:00:00Z"}`
	rec = httptest.NewRecorder()
	h.Create(rec, websiteAdminRequest(http.MethodPost, "/api/website/content", body))
	if rec.Code != http.StatusCreated {
		t.Fatalf("Create status = %d, want 201; body=%s", rec.Code, rec.Body.String())
	}
	if fake.createInput.Title != "Berita Baru" || fake.createInput.ActorUsername != "editor" || !fake.createInput.IsFeatured || fake.createInput.PublishedAt == "" {
		t.Fatalf("Create input = %+v, want decoded payload and actor", fake.createInput)
	}

	rec = httptest.NewRecorder()
	h.Update(rec, withRouteParam(websiteAdminRequest(http.MethodPatch, "/api/website/content/"+id.String(), body), "id", id.String()))
	if rec.Code != http.StatusOK {
		t.Fatalf("Update status = %d, want 200; body=%s", rec.Code, rec.Body.String())
	}
	if fake.updateInput.ID != id || fake.updateInput.ActorUsername != "editor" || fake.updateInput.Title != "Berita Baru" {
		t.Fatalf("Update input = %+v, want route id and actor", fake.updateInput)
	}

	rec = httptest.NewRecorder()
	h.Delete(rec, withRouteParam(websiteAdminRequest(http.MethodDelete, "/api/website/content/"+id.String(), ""), "id", id.String()))
	if rec.Code != http.StatusNoContent || fake.deleteID != id {
		t.Fatalf("Delete status/id = %d/%v, want 204/%v", rec.Code, fake.deleteID, id)
	}

	rec = httptest.NewRecorder()
	h.ListFeaturedPosts(rec, httptest.NewRequest(http.MethodGet, "/berita/featured?limit=8", nil))
	if rec.Code != http.StatusOK || fake.featuredKind != "post" || fake.featuredLimit != 8 {
		t.Fatalf("ListFeaturedPosts status/kind/limit = %d/%q/%d", rec.Code, fake.featuredKind, fake.featuredLimit)
	}

	rec = httptest.NewRecorder()
	h.ListPublishedPosts(rec, httptest.NewRequest(http.MethodGet, "/berita?limit=25", nil))
	if rec.Code != http.StatusOK || fake.publishedKind != "post" || fake.publishedLimit != 25 {
		t.Fatalf("ListPublishedPosts status/kind/limit = %d/%q/%d", rec.Code, fake.publishedKind, fake.publishedLimit)
	}

	rec = httptest.NewRecorder()
	h.GetPublishedPost(rec, withRouteParam(httptest.NewRequest(http.MethodGet, "/berita/konten", nil), "slug", "konten"))
	if rec.Code != http.StatusOK || fake.publishedSlugKind != "post" || fake.publishedSlug != "konten" {
		t.Fatalf("GetPublishedPost status/kind/slug = %d/%q/%q", rec.Code, fake.publishedSlugKind, fake.publishedSlug)
	}

	rec = httptest.NewRecorder()
	h.ListPublishedAnnouncements(rec, httptest.NewRequest(http.MethodGet, "/pengumuman?limit=9", nil))
	if rec.Code != http.StatusOK || fake.publishedKind != "announcement" || fake.publishedLimit != 9 {
		t.Fatalf("ListPublishedAnnouncements status/kind/limit = %d/%q/%d", rec.Code, fake.publishedKind, fake.publishedLimit)
	}

	rec = httptest.NewRecorder()
	h.GetPublishedAnnouncement(rec, withRouteParam(httptest.NewRequest(http.MethodGet, "/pengumuman/konten", nil), "slug", "konten"))
	if rec.Code != http.StatusOK || fake.publishedSlugKind != "announcement" {
		t.Fatalf("GetPublishedAnnouncement status/kind = %d/%q", rec.Code, fake.publishedSlugKind)
	}

	rec = httptest.NewRecorder()
	h.GetPublishedPage(rec, withRouteParam(httptest.NewRequest(http.MethodGet, "/profil", nil), "slug", "profil"))
	if rec.Code != http.StatusOK || fake.publishedSlugKind != "page" || fake.publishedSlug != "profil" {
		t.Fatalf("GetPublishedPage status/kind/slug = %d/%q/%q", rec.Code, fake.publishedSlugKind, fake.publishedSlug)
	}
}

func TestWebsiteHandlersMapServiceErrors(t *testing.T) {
	id := handlerTestUUID(147)
	errDB := errors.New("db down")
	tests := []struct {
		name       string
		handler    func(*Website, http.ResponseWriter, *http.Request)
		svc        *fakeWebsiteService
		req        *http.Request
		wantStatus int
	}{
		{name: "list", handler: (*Website).List, svc: &fakeWebsiteService{Website: &service.Website{}, listErr: errDB}, req: websiteAdminRequest(http.MethodGet, "/api/website/content", ""), wantStatus: http.StatusInternalServerError},
		{name: "create validation", handler: (*Website).Create, svc: &fakeWebsiteService{Website: &service.Website{}, createErr: errors.New("judul wajib diisi")}, req: websiteAdminRequest(http.MethodPost, "/api/website/content", `{}`), wantStatus: http.StatusBadRequest},
		{name: "update not found", handler: (*Website).Update, svc: &fakeWebsiteService{Website: &service.Website{}, updateErr: pgx.ErrNoRows}, req: withRouteParam(websiteAdminRequest(http.MethodPatch, "/api/website/content/"+id.String(), `{}`), "id", id.String()), wantStatus: http.StatusNotFound},
		{name: "update validation", handler: (*Website).Update, svc: &fakeWebsiteService{Website: &service.Website{}, updateErr: errors.New("status konten tidak valid")}, req: withRouteParam(websiteAdminRequest(http.MethodPatch, "/api/website/content/"+id.String(), `{}`), "id", id.String()), wantStatus: http.StatusBadRequest},
		{name: "delete", handler: (*Website).Delete, svc: &fakeWebsiteService{Website: &service.Website{}, deleteErr: errDB}, req: withRouteParam(websiteAdminRequest(http.MethodDelete, "/api/website/content/"+id.String(), ""), "id", id.String()), wantStatus: http.StatusInternalServerError},
		{name: "featured", handler: (*Website).ListFeaturedPosts, svc: &fakeWebsiteService{Website: &service.Website{}, featuredErr: errDB}, req: httptest.NewRequest(http.MethodGet, "/berita/featured", nil), wantStatus: http.StatusInternalServerError},
		{name: "published posts", handler: (*Website).ListPublishedPosts, svc: &fakeWebsiteService{Website: &service.Website{}, publishedErr: errDB}, req: httptest.NewRequest(http.MethodGet, "/berita", nil), wantStatus: http.StatusInternalServerError},
		{name: "published post not found", handler: (*Website).GetPublishedPost, svc: &fakeWebsiteService{Website: &service.Website{}, publishedSlugErr: pgx.ErrNoRows}, req: withRouteParam(httptest.NewRequest(http.MethodGet, "/berita/missing", nil), "slug", "missing"), wantStatus: http.StatusNotFound},
		{name: "published announcement", handler: (*Website).ListPublishedAnnouncements, svc: &fakeWebsiteService{Website: &service.Website{}, publishedErr: errDB}, req: httptest.NewRequest(http.MethodGet, "/pengumuman", nil), wantStatus: http.StatusInternalServerError},
		{name: "published announcement not found", handler: (*Website).GetPublishedAnnouncement, svc: &fakeWebsiteService{Website: &service.Website{}, publishedSlugErr: pgx.ErrNoRows}, req: withRouteParam(httptest.NewRequest(http.MethodGet, "/pengumuman/missing", nil), "slug", "missing"), wantStatus: http.StatusNotFound},
		{name: "published page not found", handler: (*Website).GetPublishedPage, svc: &fakeWebsiteService{Website: &service.Website{}, publishedSlugErr: pgx.ErrNoRows}, req: withRouteParam(httptest.NewRequest(http.MethodGet, "/profil/missing", nil), "slug", "missing"), wantStatus: http.StatusNotFound},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rec := httptest.NewRecorder()
			tt.handler(&Website{svc: tt.svc}, rec, tt.req)
			if rec.Code != tt.wantStatus {
				t.Fatalf("status = %d, want %d; body=%s", rec.Code, tt.wantStatus, rec.Body.String())
			}
		})
	}
}
