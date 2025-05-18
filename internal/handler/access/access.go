package access

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"pg-access-bot/internal/config"
	"pg-access-bot/internal/handler/grant"
	"time"
)

type AccessRequest struct {
	UserIdentification string   `json:"user_identification"`
	Tables             []string `json:"tables"`
	Permissions        []string `json:"permissions"`
	ValidFor           int      `json:"valid_for"`
}

type AccessResponse struct {
	Username   string `json:"username"`
	Password   string `json:"password"`
	ValidUntil string `json:"valid_until"`
}

func RequestAccess(w http.ResponseWriter, r *http.Request) {
	var req AccessRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	cfg := config.Load()
	dbConn, err := sql.Open("postgres", cfg.DBUrl)
	if err != nil {
		http.Error(w, "failed to connect to database", http.StatusInternalServerError)
		return
	}
	defer dbConn.Close()

	duration := time.Duration(req.ValidFor) * time.Minute
	username, password, err := grant.Grant(r.Context(), dbConn, req.UserIdentification, duration)
	if err != nil {
		http.Error(w, "failed to grant access", http.StatusInternalServerError)
		return
	}

	resp := &AccessResponse{
		Username:   username,
		Password:   password,
		ValidUntil: time.Now().Add(duration).Format(time.RFC3339),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}
