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
	err := json.NewEncoder(w).Encode(health)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		h.Logger.Warn("Failed to encode healthcheck response", "error", err)
		return
	}
}
