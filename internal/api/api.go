package api

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"

	"github.com/XimilalaXiang/ReferenceAnswerRSS/internal/auth"
	"github.com/XimilalaXiang/ReferenceAnswerRSS/internal/rss"
	"github.com/XimilalaXiang/ReferenceAnswerRSS/internal/store"
	"github.com/XimilalaXiang/ReferenceAnswerRSS/internal/sync"
)

type Handler struct {
	store   *store.Store
	auth    *auth.Auth
	sync    *sync.Service
	baseURL string
}

func NewHandler(s *store.Store, a *auth.Auth, sy *sync.Service, baseURL string) *Handler {
	return &Handler{store: s, auth: a, sync: sy, baseURL: baseURL}
}

func (h *Handler) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("POST /api/login", h.handleLogin)

	mux.Handle("GET /api/articles", h.auth.Middleware(http.HandlerFunc(h.handleListArticles)))
	mux.Handle("GET /api/articles/{id}", h.auth.Middleware(http.HandlerFunc(h.handleGetArticle)))
	mux.Handle("GET /api/settings", h.auth.Middleware(http.HandlerFunc(h.handleGetSettings)))
	mux.Handle("POST /api/sync", h.auth.Middleware(http.HandlerFunc(h.handleTriggerSync)))
	mux.Handle("GET /api/sync/status", h.auth.Middleware(http.HandlerFunc(h.handleSyncStatus)))
	mux.Handle("POST /api/feed-token/regenerate", h.auth.Middleware(http.HandlerFunc(h.handleRegenerateFeedToken)))

	mux.HandleFunc("GET /feed.xml", h.handleRSSFeed)
	mux.HandleFunc("GET /feed.atom", h.handleAtomFeed)
}

func (h *Handler) handleLogin(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		jsonError(w, "invalid request body", http.StatusBadRequest)
		return
	}

	user, err := h.store.ValidateUser(req.Username, req.Password)
	if err != nil {
		jsonError(w, "invalid credentials", http.StatusUnauthorized)
		return
	}

	token, err := h.auth.GenerateToken(user.ID, user.Username)
	if err != nil {
		jsonError(w, "failed to generate token", http.StatusInternalServerError)
		return
	}

	jsonResponse(w, map[string]interface{}{
		"token":     token,
		"username":  user.Username,
		"feedToken": user.FeedToken,
	})
}

func (h *Handler) handleListArticles(w http.ResponseWriter, r *http.Request) {
	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	if page < 1 {
		page = 1
	}
	pageSize, _ := strconv.Atoi(r.URL.Query().Get("pageSize"))
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}
	keyword := r.URL.Query().Get("keyword")

	articles, total, err := h.store.ListArticles(page, pageSize, keyword)
	if err != nil {
		jsonError(w, "failed to list articles", http.StatusInternalServerError)
		return
	}

	jsonResponse(w, map[string]interface{}{
		"articles": articles,
		"total":    total,
		"page":     page,
		"pageSize": pageSize,
	})
}

func (h *Handler) handleGetArticle(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	article, err := h.store.GetArticle(id)
	if err != nil {
		jsonError(w, "article not found", http.StatusNotFound)
		return
	}
	jsonResponse(w, article)
}

func (h *Handler) handleGetSettings(w http.ResponseWriter, r *http.Request) {
	userIDStr := r.Header.Get("X-User-ID")
	userID, _ := strconv.Atoi(userIDStr)
	user, err := h.store.GetUserByID(userID)
	if err != nil {
		jsonError(w, "user not found", http.StatusNotFound)
		return
	}

	status := h.sync.Status()
	jsonResponse(w, map[string]interface{}{
		"feedToken":  user.FeedToken,
		"feedURL":    h.baseURL + "/feed.xml?token=" + user.FeedToken,
		"atomURL":    h.baseURL + "/feed.atom?token=" + user.FeedToken,
		"syncStatus": status,
	})
}

func (h *Handler) handleTriggerSync(w http.ResponseWriter, r *http.Request) {
	go h.sync.RunOnce()
	jsonResponse(w, map[string]string{"message": "sync triggered"})
}

func (h *Handler) handleSyncStatus(w http.ResponseWriter, r *http.Request) {
	jsonResponse(w, h.sync.Status())
}

func (h *Handler) handleRegenerateFeedToken(w http.ResponseWriter, r *http.Request) {
	userIDStr := r.Header.Get("X-User-ID")
	userID, _ := strconv.Atoi(userIDStr)
	token, err := h.store.RegenerateFeedToken(userID)
	if err != nil {
		jsonError(w, "failed to regenerate token", http.StatusInternalServerError)
		return
	}
	jsonResponse(w, map[string]string{
		"feedToken": token,
		"feedURL":   h.baseURL + "/feed.xml?token=" + token,
		"atomURL":   h.baseURL + "/feed.atom?token=" + token,
	})
}

func (h *Handler) handleRSSFeed(w http.ResponseWriter, r *http.Request) {
	token := r.URL.Query().Get("token")
	if token == "" || !h.store.ValidateFeedToken(token) {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	articles, err := h.store.GetAllArticlesForFeed()
	if err != nil {
		log.Printf("[rss] failed to get articles: %v", err)
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	data, err := rss.GenerateRSS(articles, h.baseURL)
	if err != nil {
		log.Printf("[rss] failed to generate RSS: %v", err)
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/rss+xml; charset=utf-8")
	w.Write(data)
}

func (h *Handler) handleAtomFeed(w http.ResponseWriter, r *http.Request) {
	token := r.URL.Query().Get("token")
	if token == "" || !h.store.ValidateFeedToken(token) {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	articles, err := h.store.GetAllArticlesForFeed()
	if err != nil {
		log.Printf("[atom] failed to get articles: %v", err)
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	data, err := rss.GenerateAtom(articles, h.baseURL)
	if err != nil {
		log.Printf("[atom] failed to generate Atom: %v", err)
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/atom+xml; charset=utf-8")
	w.Write(data)
}

func jsonResponse(w http.ResponseWriter, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(data)
}

func jsonError(w http.ResponseWriter, message string, code int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(map[string]string{"error": message})
}
