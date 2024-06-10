package service

import (
	"encoding/json"
	"net/http"
)

func (h *CNProxyHandler) Healthcheck(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	health := map[string]string{
		"status": "healthy",
	}
	json.NewEncoder(w).Encode(health)
}
