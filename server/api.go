package server

import "net/http"

func (h *handler) allRulesets(w http.ResponseWriter, r *http.Request) {
	l, err := h.store.All(r.Context())
	if err != nil {
		h.writeError(w, err, http.StatusInternalServerError)
		return
	}

	h.encodeJSON(w, l, http.StatusOK)
}
