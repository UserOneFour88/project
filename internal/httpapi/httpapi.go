package httpapi

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"golang.org/x/crypto/bcrypt"

	"pipelineapp/internal/auth"
	"pipelineapp/internal/store"
)

type API struct {
	repo *store.Repo
	jwt  *auth.Service
}

func New(repo *store.Repo, jwt *auth.Service) *API {
	return &API{repo: repo, jwt: jwt}
}

func (a *API) Router() http.Handler {
	r := chi.NewRouter()

	r.Get("/health", func(w http.ResponseWriter, _ *http.Request) {
		writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
	})

	r.Route("/auth", func(r chi.Router) {
		r.Post("/register", a.handleRegister)
		r.Post("/login", a.handleLogin)
		r.Post("/refresh", a.handleRefresh)
		r.Post("/logout", a.handleLogout)
	})

	r.Group(func(r chi.Router) {
		r.Use(a.authMiddleware)
		r.Get("/me", a.handleMe)
		r.Get("/rooms", a.handleListRooms)
		r.Post("/rooms", a.handleCreateRoom)
		r.Get("/rooms/{roomID}/messages", a.handleListMessages)
		r.Post("/rooms/{roomID}/messages", a.handleCreateMessage)
	})

	return r
}

type ctxKey int

const ctxUser ctxKey = 1

type authedUser struct {
	ID       int64
	Username string
}

func (a *API) authMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		h := r.Header.Get("Authorization")
		if h == "" || !strings.HasPrefix(h, "Bearer ") {
			writeErr(w, http.StatusUnauthorized, "missing bearer token")
			return
		}
		raw := strings.TrimSpace(strings.TrimPrefix(h, "Bearer "))
		claims, err := a.jwt.ParseAccess(raw)
		if err != nil {
			writeErr(w, http.StatusUnauthorized, "invalid access token")
			return
		}
		u := authedUser{ID: claims.UserID, Username: claims.Username}
		ctx := context.WithValue(r.Context(), ctxUser, u)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func userFromCtx(r *http.Request) (authedUser, bool) {
	u, ok := r.Context().Value(ctxUser).(authedUser)
	return u, ok
}

type registerReq struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

func (a *API) handleRegister(w http.ResponseWriter, r *http.Request) {
	var req registerReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeErr(w, http.StatusBadRequest, "invalid json")
		return
	}
	req.Username = strings.TrimSpace(req.Username)
	if req.Username == "" || len(req.Username) > 40 {
		writeErr(w, http.StatusBadRequest, "invalid username")
		return
	}
	if len(req.Password) < 6 {
		writeErr(w, http.StatusBadRequest, "password too short")
		return
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		writeErr(w, http.StatusInternalServerError, "hash error")
		return
	}
	u, err := a.repo.CreateUser(r.Context(), req.Username, string(hash))
	if err != nil {
		writeErr(w, http.StatusBadRequest, "user already exists or invalid")
		return
	}
	writeJSON(w, http.StatusCreated, map[string]any{"id": u.ID, "username": u.Username})
}

type loginReq struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

func (a *API) handleLogin(w http.ResponseWriter, r *http.Request) {
	var req loginReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeErr(w, http.StatusBadRequest, "invalid json")
		return
	}
	req.Username = strings.TrimSpace(req.Username)
	u, err := a.repo.GetUserByUsername(r.Context(), req.Username)
	if err != nil {
		writeErr(w, http.StatusUnauthorized, "invalid credentials")
		return
	}
	if bcrypt.CompareHashAndPassword([]byte(u.PasswordHash), []byte(req.Password)) != nil {
		writeErr(w, http.StatusUnauthorized, "invalid credentials")
		return
	}

	access, _, err := a.jwt.IssueAccess(u.ID, u.Username)
	if err != nil {
		writeErr(w, http.StatusInternalServerError, "token error")
		return
	}
	refresh, refreshExp, err := a.jwt.IssueRefresh(u.ID)
	if err != nil {
		writeErr(w, http.StatusInternalServerError, "token error")
		return
	}
	if err := a.repo.StoreRefreshToken(r.Context(), u.ID, refresh, refreshExp); err != nil {
		writeErr(w, http.StatusInternalServerError, "token store error")
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"access_token":  access,
		"refresh_token": refresh,
	})
}

type refreshReq struct {
	RefreshToken string `json:"refresh_token"`
}

func (a *API) handleRefresh(w http.ResponseWriter, r *http.Request) {
	var req refreshReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeErr(w, http.StatusBadRequest, "invalid json")
		return
	}
	raw := strings.TrimSpace(req.RefreshToken)
	if raw == "" {
		writeErr(w, http.StatusBadRequest, "missing refresh_token")
		return
	}
	claims, err := a.jwt.ParseRefresh(raw)
	if err != nil {
		writeErr(w, http.StatusUnauthorized, "invalid refresh token")
		return
	}
	if err := a.repo.ValidateRefreshToken(r.Context(), claims.UserID, raw); err != nil {
		writeErr(w, http.StatusUnauthorized, "invalid refresh token")
		return
	}

	// rotate: revoke old + issue new
	_ = a.repo.RevokeRefreshToken(r.Context(), claims.UserID, raw)

	u, err := a.repo.GetUserByID(r.Context(), claims.UserID)
	if err != nil {
		writeErr(w, http.StatusUnauthorized, "invalid refresh token")
		return
	}

	access, _, err := a.jwt.IssueAccess(claims.UserID, u.Username)
	if err != nil {
		writeErr(w, http.StatusInternalServerError, "token error")
		return
	}

	newRefresh, refreshExp, err := a.jwt.IssueRefresh(claims.UserID)
	if err != nil {
		writeErr(w, http.StatusInternalServerError, "token error")
		return
	}
	if err := a.repo.StoreRefreshToken(r.Context(), claims.UserID, newRefresh, refreshExp); err != nil {
		writeErr(w, http.StatusInternalServerError, "token store error")
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"access_token":  access,
		"refresh_token": newRefresh,
	})
}

type logoutReq struct {
	RefreshToken string `json:"refresh_token"`
}

func (a *API) handleLogout(w http.ResponseWriter, r *http.Request) {
	var req logoutReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeErr(w, http.StatusBadRequest, "invalid json")
		return
	}
	raw := strings.TrimSpace(req.RefreshToken)
	if raw == "" {
		writeErr(w, http.StatusBadRequest, "missing refresh_token")
		return
	}
	claims, err := a.jwt.ParseRefresh(raw)
	if err != nil {
		writeErr(w, http.StatusUnauthorized, "invalid refresh token")
		return
	}
	_ = a.repo.RevokeRefreshToken(r.Context(), claims.UserID, raw)
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func (a *API) handleMe(w http.ResponseWriter, r *http.Request) {
	u, ok := userFromCtx(r)
	if !ok {
		writeErr(w, http.StatusUnauthorized, "unauthorized")
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"id": u.ID, "username": u.Username})
}

type createRoomReq struct {
	Name string `json:"name"`
}

func (a *API) handleCreateRoom(w http.ResponseWriter, r *http.Request) {
	var req createRoomReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeErr(w, http.StatusBadRequest, "invalid json")
		return
	}
	req.Name = strings.TrimSpace(req.Name)
	if req.Name == "" || len(req.Name) > 60 {
		writeErr(w, http.StatusBadRequest, "invalid room name")
		return
	}
	rm, err := a.repo.CreateRoom(r.Context(), req.Name)
	if err != nil {
		writeErr(w, http.StatusBadRequest, "room create failed")
		return
	}
	writeJSON(w, http.StatusCreated, rm)
}

func (a *API) handleListRooms(w http.ResponseWriter, r *http.Request) {
	rooms, err := a.repo.ListRooms(r.Context())
	if err != nil {
		writeErr(w, http.StatusInternalServerError, "list rooms failed")
		return
	}
	writeJSON(w, http.StatusOK, rooms)
}

type createMessageReq struct {
	Text string `json:"text"`
}

func (a *API) handleCreateMessage(w http.ResponseWriter, r *http.Request) {
	u, ok := userFromCtx(r)
	if !ok {
		writeErr(w, http.StatusUnauthorized, "unauthorized")
		return
	}
	roomID, err := strconv.ParseInt(chi.URLParam(r, "roomID"), 10, 64)
	if err != nil || roomID <= 0 {
		writeErr(w, http.StatusBadRequest, "invalid roomID")
		return
	}
	var req createMessageReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeErr(w, http.StatusBadRequest, "invalid json")
		return
	}
	req.Text = strings.TrimSpace(req.Text)
	if req.Text == "" || len(req.Text) > 1000 {
		writeErr(w, http.StatusBadRequest, "invalid text")
		return
	}
	m, err := a.repo.CreateMessage(r.Context(), roomID, u.ID, req.Text)
	if err != nil {
		writeErr(w, http.StatusBadRequest, "create message failed")
		return
	}
	writeJSON(w, http.StatusCreated, m)
}

func (a *API) handleListMessages(w http.ResponseWriter, r *http.Request) {
	roomID, err := strconv.ParseInt(chi.URLParam(r, "roomID"), 10, 64)
	if err != nil || roomID <= 0 {
		writeErr(w, http.StatusBadRequest, "invalid roomID")
		return
	}
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	msgs, err := a.repo.ListMessages(r.Context(), roomID, limit)
	if err != nil {
		writeErr(w, http.StatusInternalServerError, "list messages failed")
		return
	}
	writeJSON(w, http.StatusOK, msgs)
}

func writeErr(w http.ResponseWriter, code int, msg string) {
	writeJSON(w, code, map[string]string{"error": msg})
}

func writeJSON(w http.ResponseWriter, code int, v any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(code)
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	_ = enc.Encode(v)
}

// avoid unused imports if build tags change
var _ = errors.New
var _ = time.Now

