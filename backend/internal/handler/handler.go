package handler

import (
	"log"
	"github.com/grysha11/camagru-backend/internal/config"
	"net/http"
	"encoding/json"
)

type Handler struct {
	Cfg	*config.Config
}

func (*Handler) Healthz(w http.ResponseWriter, r *http.Request) {
	type response struct {
		Status	string	`json:"status"`
	}

	payload := &response{
		Status: "ok",
	}
	w.Header().Set("Content-Type", "application/json")
	data, err := json.Marshal(payload)
	if err != nil {
		log.Printf("Error marshaling JSON: %v", err)
		w.WriteHeader(500)
		return
	}

	w.WriteHeader(200)
	w.Write(data)
}
