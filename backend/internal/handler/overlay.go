package handler

import (
	"context"

	"github.com/grysha11/camagru-backend/internal/api"
)

func (h *Handler) ListOverlays(ctx context.Context, r api.ListOverlaysRequestObject) (api.ListOverlaysResponseObject, error) {
	entries := h.Cfg.Overlays.List()

	resp := make([]api.OverlayResponse, 0, len(entries))
	for _, e := range entries {
		resp = append(resp, api.OverlayResponse{
			Id:  e.ID,
			Url: "/api/overlays/" + e.ID,
		})
	}

	return api.ListOverlays200JSONResponse(resp), nil
}
