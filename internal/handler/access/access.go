package access

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"pg-access-bot/internal/config"
	"pg-access-bot/internal/handler/grant"
	"strings"
	"time"

	"github.com/go-playground/validator/v10"
)

var validate = validator.New()

type AccessRequest struct {
	UserIdentification string   `json:"user_identification" validate:"required"`
	Tables             []string `json:"tables" validate:"required,min=1,max=3,dive,required"`
	Actions            []string `json:"actions" validate:"required,min=1,max=4,dive,oneof=SELECT INSERT UPDATE DELETE"`
	ValidFor           int      `json:"valid_for" validate:"required,min=10,max=120"` // in minutes
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

	if err := validate.Struct(req); err != nil {
		var errs []string
		for _, e := range err.(validator.ValidationErrors) {
			errs = append(errs, fmt.Sprintf("field '%s' failed on the '%s' rule", e.Field(), e.Tag()))
		}
		http.Error(w, strings.Join(errs, ", "), http.StatusBadRequest)
		return
	}

	cfg := config.Load()
	dbConn, err := sql.Open("postgres", cfg.DBUrl)
	if err != nil {
		log.Printf("failed to connect to database: %s", err)
		http.Error(w, "failed to connect to database", http.StatusInternalServerError)
		return
	}
	defer dbConn.Close()

	duration := time.Duration(req.ValidFor) * time.Minute
	username, password, err := grant.Grant(r.Context(), dbConn, req.UserIdentification, duration, req.Tables, req.Actions)
	if err != nil {
		log.Printf("failed to grant acces: %s", err)
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
