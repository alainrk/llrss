package handler

import (
	"encoding/json"
	"llrss/internal/models"
	"llrss/internal/service"
	"net/http"

	"github.com/go-chi/chi/v5"
)

type FeedHandler struct {
	feedService service.FeedService
}

func NewFeedHandler(feedService service.FeedService) *FeedHandler {
	return &FeedHandler{
		feedService: feedService,
	}
}

func (h *FeedHandler) RegisterRoutes(r chi.Router) {
	r.Get("/feeds", h.ListFeeds)
	r.Post("/feeds", h.AddFeed)
	r.Get("/feeds/{id}", h.GetFeed)
	r.Delete("/feeds/{id}", h.DeleteFeed)
	r.Put("/feeds/{id}", h.UpdateFeed)
}

func (h *FeedHandler) ListFeeds(w http.ResponseWriter, r *http.Request) {
	feeds, err := h.feedService.ListFeeds(r.Context())
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	err = json.NewEncoder(w).Encode(feeds)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func (h *FeedHandler) AddFeed(w http.ResponseWriter, r *http.Request) {
	var req struct {
		URL string `json:"url"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	ID, err := h.feedService.AddFeed(r.Context(), req.URL)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	w.Write([]byte(ID))
}

func (h *FeedHandler) GetFeed(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	feed, err := h.feedService.GetFeed(r.Context(), id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	err = json.NewEncoder(w).Encode(feed)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func (h *FeedHandler) DeleteFeed(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if err := h.feedService.DeleteFeed(r.Context(), id); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *FeedHandler) UpdateFeed(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	var feed models.Feed
	if err := json.NewDecoder(r.Body).Decode(&feed); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	feed.ID = id
	if err := h.feedService.UpdateFeed(r.Context(), &feed); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	err := json.NewEncoder(w).Encode(feed)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}
