package handler

import (
	"github.com/grysha11/camagru-backend/internal/config"
	"github.com/grysha11/camagru-backend/internal/api"
	"context"
)

type Handler struct {
	Cfg	*config.Config
}

func (h *Handler) GetHealthz(ctx context.Context, r api.GetHealthzRequestObject) (api.GetHealthzResponseObject, error) {
	return api.GetHealthz200JSONResponse{
		Status: "ok",
	}, nil
}
