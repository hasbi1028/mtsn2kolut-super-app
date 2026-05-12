package handler

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"mtsn2kolut-super-app/backend/internal/api"
	db "mtsn2kolut-super-app/backend/internal/repository/postgres"
	"mtsn2kolut-super-app/backend/internal/service"
)

type Website struct{ svc websiteService }

type websiteService interface {
	List(ctx context.Context, kind, status, search string) ([]db.WebsiteContent, error)
	Create(ctx context.Context, in service.SaveWebsiteContentInput) (db.WebsiteContent, error)
	Update(ctx context.Context, in service.SaveWebsiteContentInput) (db.WebsiteContent, error)
	Delete(ctx context.Context, id pgtype.UUID) error
	ListFeatured(ctx context.Context, kind string, limit int32) ([]db.WebsiteContent, error)
	ListPublished(ctx context.Context, kind string, limit int32) ([]db.WebsiteContent, error)
	GetPublishedBySlug(ctx context.Context, kind, slug string) (db.WebsiteContent, error)
}

func NewWebsite(svc *service.Website) *Website { return &Website{svc: svc} }

func websiteActorUsername(r *http.Request) string {
	if claims, ok := api.ClaimsFromContext(r.Context()); ok {
		if usr, ok := claims["usr"].(string); ok && usr != "" {
			return usr
		}
		if sub, ok := claims["sub"].(string); ok {
			return sub
		}
	}
	return ""
}

func (h *Website) List(w http.ResponseWriter, r *http.Request) {
	if !adminAccessAllowed(r) {
		api.Forbidden(w)
		return
	}
	rows, err := h.svc.List(
		r.Context(),
		r.URL.Query().Get("kind"),
		r.URL.Query().Get("status"),
		r.URL.Query().Get("search"),
	)
	if err != nil {
		api.Internal(w, err)
		return
	}
	api.OK(w, rows)
}

func (h *Website) Create(w http.ResponseWriter, r *http.Request) {
	if !adminAccessAllowed(r) {
		api.Forbidden(w)
		return
	}
	r.Body = http.MaxBytesReader(w, r.Body, 256<<10)
	var body struct {
		Kind            string `json:"kind"`
		Title           string `json:"title"`
		Slug            string `json:"slug"`
		Excerpt         string `json:"excerpt"`
		ContentHTML     string `json:"content_html"`
		CoverImageURL   string `json:"cover_image_url"`
		IsFeatured      bool   `json:"is_featured"`
		MetaTitle       string `json:"meta_title"`
		MetaDescription string `json:"meta_description"`
		Status          string `json:"status"`
		PublishedAt     string `json:"published_at"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		api.BadRequest(w, "Data yang dikirim tidak valid")
		return
	}
	row, err := h.svc.Create(r.Context(), service.SaveWebsiteContentInput{
		Kind:            body.Kind,
		Title:           body.Title,
		Slug:            body.Slug,
		Excerpt:         body.Excerpt,
		ContentHTML:     body.ContentHTML,
		CoverImageURL:   body.CoverImageURL,
		IsFeatured:      body.IsFeatured,
		MetaTitle:       body.MetaTitle,
		MetaDescription: body.MetaDescription,
		Status:          body.Status,
		PublishedAt:     body.PublishedAt,
		ActorUsername:   websiteActorUsername(r),
	})
	if err != nil {
		writeClientError(w, err, "Konten website tidak valid")
		return
	}
	api.Created(w, row)
}

func (h *Website) Update(w http.ResponseWriter, r *http.Request) {
	if !adminAccessAllowed(r) {
		api.Forbidden(w)
		return
	}
	id, err := parseUUID(chi.URLParam(r, "id"))
	if err != nil {
		api.BadRequest(w, "ID data tidak valid")
		return
	}
	r.Body = http.MaxBytesReader(w, r.Body, 256<<10)
	var body struct {
		Kind            string `json:"kind"`
		Title           string `json:"title"`
		Slug            string `json:"slug"`
		Excerpt         string `json:"excerpt"`
		ContentHTML     string `json:"content_html"`
		CoverImageURL   string `json:"cover_image_url"`
		IsFeatured      bool   `json:"is_featured"`
		MetaTitle       string `json:"meta_title"`
		MetaDescription string `json:"meta_description"`
		Status          string `json:"status"`
		PublishedAt     string `json:"published_at"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		api.BadRequest(w, "Data yang dikirim tidak valid")
		return
	}
	row, err := h.svc.Update(r.Context(), service.SaveWebsiteContentInput{
		ID:              id,
		Kind:            body.Kind,
		Title:           body.Title,
		Slug:            body.Slug,
		Excerpt:         body.Excerpt,
		ContentHTML:     body.ContentHTML,
		CoverImageURL:   body.CoverImageURL,
		IsFeatured:      body.IsFeatured,
		MetaTitle:       body.MetaTitle,
		MetaDescription: body.MetaDescription,
		Status:          body.Status,
		PublishedAt:     body.PublishedAt,
		ActorUsername:   websiteActorUsername(r),
	})
	if errors.Is(err, pgx.ErrNoRows) {
		api.NotFound(w)
		return
	}
	if err != nil {
		writeClientError(w, err, "Perubahan konten website tidak valid")
		return
	}
	api.OK(w, row)
}

func (h *Website) Delete(w http.ResponseWriter, r *http.Request) {
	if !adminAccessAllowed(r) {
		api.Forbidden(w)
		return
	}
	id, err := parseUUID(chi.URLParam(r, "id"))
	if err != nil {
		api.BadRequest(w, "ID data tidak valid")
		return
	}
	if err := h.svc.Delete(r.Context(), id); err != nil {
		api.Internal(w, err)
		return
	}
	api.NoContent(w)
}

func (h *Website) ListFeaturedPosts(w http.ResponseWriter, r *http.Request) {
	limit := int32(6)
	if raw := r.URL.Query().Get("limit"); raw != "" {
		if parsed, err := strconv.Atoi(raw); err == nil && parsed > 0 && parsed <= 20 {
			limit = int32(parsed)
		}
	}
	rows, err := h.svc.ListFeatured(r.Context(), "post", limit)
	if err != nil {
		api.Internal(w, err)
		return
	}
	api.OK(w, rows)
}

func (h *Website) ListPublishedPosts(w http.ResponseWriter, r *http.Request) {
	limit := int32(12)
	if raw := r.URL.Query().Get("limit"); raw != "" {
		if parsed, err := strconv.Atoi(raw); err == nil && parsed > 0 && parsed <= 50 {
			limit = int32(parsed)
		}
	}
	rows, err := h.svc.ListPublished(r.Context(), "post", limit)
	if err != nil {
		api.Internal(w, err)
		return
	}
	api.OK(w, rows)
}

func (h *Website) GetPublishedPost(w http.ResponseWriter, r *http.Request) {
	row, err := h.svc.GetPublishedBySlug(r.Context(), "post", chi.URLParam(r, "slug"))
	if errors.Is(err, pgx.ErrNoRows) {
		api.NotFound(w)
		return
	}
	if err != nil {
		writeClientError(w, err, "Konten website publik tidak valid")
		return
	}
	api.OK(w, row)
}

func (h *Website) ListPublishedAnnouncements(w http.ResponseWriter, r *http.Request) {
	limit := int32(10)
	if raw := r.URL.Query().Get("limit"); raw != "" {
		if parsed, err := strconv.Atoi(raw); err == nil && parsed > 0 && parsed <= 50 {
			limit = int32(parsed)
		}
	}
	rows, err := h.svc.ListPublished(r.Context(), "announcement", limit)
	if err != nil {
		api.Internal(w, err)
		return
	}
	api.OK(w, rows)
}

func (h *Website) GetPublishedAnnouncement(w http.ResponseWriter, r *http.Request) {
	row, err := h.svc.GetPublishedBySlug(r.Context(), "announcement", chi.URLParam(r, "slug"))
	if errors.Is(err, pgx.ErrNoRows) {
		api.NotFound(w)
		return
	}
	if err != nil {
		writeClientError(w, err, "Pengumuman publik tidak valid")
		return
	}
	api.OK(w, row)
}

func (h *Website) GetPublishedPage(w http.ResponseWriter, r *http.Request) {
	row, err := h.svc.GetPublishedBySlug(r.Context(), "page", chi.URLParam(r, "slug"))
	if errors.Is(err, pgx.ErrNoRows) {
		api.NotFound(w)
		return
	}
	if err != nil {
		writeClientError(w, err, "Halaman publik tidak valid")
		return
	}
	api.OK(w, row)
}
