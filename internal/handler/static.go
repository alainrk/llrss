package handler

import (
	"html/template"
	"llrss/internal/service"
	"net/http"

	"github.com/go-chi/chi/v5"
)

type StaticHandler struct {
	templates   *template.Template
	feedService service.FeedService
}

func NewStaticHandler(feedService service.FeedService) *StaticHandler {
	// Parse all templates
	tmpl, err := template.ParseGlob("templates/*.html")
	if err != nil {
		return nil
	}

	return &StaticHandler{
		templates:   tmpl,
		feedService: feedService,
	}
}

func (h *StaticHandler) RegisterRoutes(r chi.Router) {
	fileServer := http.FileServer(http.Dir("static"))
	r.Handle("/static/*", http.StripPrefix("/static/", fileServer))

	r.Get("/", h.handleHome)
}

func (h *StaticHandler) handleHome(w http.ResponseWriter, r *http.Request) {
	feeds, err := h.feedService.ListFeeds(r.Context())
	if err != nil {
		http.Error(w, "Failed to load feeds", http.StatusInternalServerError)
		return
	}

	data := map[string]interface{}{
		"Feeds": feeds,
	}

	h.templates.ExecuteTemplate(w, "home.html", data)
}
