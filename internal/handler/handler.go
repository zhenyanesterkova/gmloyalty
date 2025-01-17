package handler

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/zhenyanesterkova/gmloyalty/internal/middleware"
	"github.com/zhenyanesterkova/gmloyalty/internal/service/logger"
)

const (
	TextServerError = "Something went wrong... Server error"
)

type Repositorie interface {
	Ping() error
}

type RepositorieHandler struct {
	Repo    Repositorie
	Logger  logger.LogrusLogger
	hashKey *string
}

func NewRepositorieHandler(
	rep Repositorie,
	log logger.LogrusLogger,
	key *string,
) *RepositorieHandler {
	return &RepositorieHandler{
		Repo:    rep,
		Logger:  log,
		hashKey: key,
	}
}

func (rh *RepositorieHandler) InitChiRouter(router *chi.Mux) {
	mdlWare := middleware.NewMiddlewareStruct(rh.Logger, rh.hashKey)
	router.Use(mdlWare.ResetRespDataStruct)
	router.Use(mdlWare.RequestLogger)
	router.Use(mdlWare.GZipMiddleware)
	router.Route("/", func(r chi.Router) {
		r.Get("/ping", rh.Ping)
	})
}

func (rh *RepositorieHandler) Ping(w http.ResponseWriter, r *http.Request) {
	log := rh.Logger.LogrusLog

	err := rh.Repo.Ping()

	if err != nil {
		log.Errorf("failed ping storage: %v", err)
		http.Error(w, TextServerError, http.StatusInternalServerError)
		return
	}
}
