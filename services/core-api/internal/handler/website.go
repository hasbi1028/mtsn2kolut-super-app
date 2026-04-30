package handler

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5"
	"mtsn2kolut-super-app/backend/internal/api"
	"mtsn2kolut-super-app/backend/internal/service"
)

type Website struct{ svc *service.Website }

func NewWebsite(svc *service.Website) *Website { return &Website{svc: svc} }

func websiteActorUsername(r *http.Request) string {
	if claims, ok := api.ClaimsFromContext(r.Context()); ok {
		if sub, ok := claims["sub"].(string); ok {
			return sub
		}
	}
	return ""
}

func (h *Website) List(w http.ResponseWriter, r *http.Request) {
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
	var body struct {
		Kind          string `json:"kind"`
		Title         string `json:"title"`
		Slug          string `json:"slug"`
		Excerpt       string `json:"excerpt"`
		ContentHTML   string `json:"content_html"`
		CoverImageURL string `json:"cover_image_url"`
		Status        string `json:"status"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		api.BadRequest(w, "invalid json")
		return
	}
	row, err := h.svc.Create(r.Context(), service.SaveWebsiteContentInput{
		Kind:          body.Kind,
		Title:         body.Title,
		Slug:          body.Slug,
		Excerpt:       body.Excerpt,
		ContentHTML:   body.ContentHTML,
		CoverImageURL: body.CoverImageURL,
		Status:        body.Status,
		ActorUsername: websiteActorUsername(r),
	})
	if err != nil {
		api.BadRequest(w, err.Error())
		return
	}
	api.Created(w, row)
}

func (h *Website) Update(w http.ResponseWriter, r *http.Request) {
	id, err := parseUUID(chi.URLParam(r, "id"))
	if err != nil {
		api.BadRequest(w, "invalid id")
		return
	}
	var body struct {
		Kind          string `json:"kind"`
		Title         string `json:"title"`
		Slug          string `json:"slug"`
		Excerpt       string `json:"excerpt"`
		ContentHTML   string `json:"content_html"`
		CoverImageURL string `json:"cover_image_url"`
		Status        string `json:"status"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		api.BadRequest(w, "invalid json")
		return
	}
	row, err := h.svc.Update(r.Context(), service.SaveWebsiteContentInput{
		ID:            id,
		Kind:          body.Kind,
		Title:         body.Title,
		Slug:          body.Slug,
		Excerpt:       body.Excerpt,
		ContentHTML:   body.ContentHTML,
		CoverImageURL: body.CoverImageURL,
		Status:        body.Status,
		ActorUsername: websiteActorUsername(r),
	})
	if errors.Is(err, pgx.ErrNoRows) {
		api.NotFound(w)
		return
	}
	if err != nil {
		api.BadRequest(w, err.Error())
		return
	}
	api.OK(w, row)
}

func (h *Website) Delete(w http.ResponseWriter, r *http.Request) {
	id, err := parseUUID(chi.URLParam(r, "id"))
	if err != nil {
		api.BadRequest(w, "invalid id")
		return
	}
	if err := h.svc.Delete(r.Context(), id); err != nil {
		api.Internal(w, err)
		return
	}
	api.NoContent(w)
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
		api.BadRequest(w, err.Error())
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
		api.BadRequest(w, err.Error())
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
		api.BadRequest(w, err.Error())
		return
	}
	api.OK(w, row)
}
