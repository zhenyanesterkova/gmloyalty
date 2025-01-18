package handler

import (
	"context"
	"net/http"

	"github.com/go-chi/chi/v5"

	"github.com/zhenyanesterkova/gmloyalty/internal/config"
	"github.com/zhenyanesterkova/gmloyalty/internal/middleware"
	"github.com/zhenyanesterkova/gmloyalty/internal/myclient"
	"github.com/zhenyanesterkova/gmloyalty/internal/service/logger"
	"github.com/zhenyanesterkova/gmloyalty/internal/service/session"
	"github.com/zhenyanesterkova/gmloyalty/internal/service/user"
)

const (
	TextServerError        = "Something went wrong... Server error"
	TextLoginError         = "User with this login already exists"
	TextInvalidFormatError = "Invalid request format"
)

type Repositorie interface {
	Ping() error
	Register(ctx context.Context, user user.User) (int, error)
	Login(userData user.User) (int, error)
}

type RepositorieHandler struct {
	Repo    Repositorie
	Logger  logger.LogrusLogger
	accrual *myclient.AccrualStruct
	jwtSess *session.SessionsJWT
}

func NewRepositorieHandler(
	rep Repositorie,
	log logger.LogrusLogger,
	cfgJWT config.JWTConfig,
	accrualAddress string,
) *RepositorieHandler {
	jwtSession := session.NewSessionsJWT(cfgJWT)
	acc := myclient.Accrual(accrualAddress)
	return &RepositorieHandler{
		Repo:    rep,
		Logger:  log,
		jwtSess: jwtSession,
		accrual: acc,
	}
}

func (rh *RepositorieHandler) InitChiRouter(router *chi.Mux) {
	mdlWare := middleware.NewMiddlewareStruct(rh.Logger, rh.jwtSess)
	router.Use(mdlWare.ResetRespDataStruct)
	router.Use(mdlWare.RequestLogger)
	router.Use(mdlWare.Auth)
	router.Use(mdlWare.GZipMiddleware)
	router.Route("/", func(r chi.Router) {
		r.Get("/ping", rh.Ping)
		r.Route("/api/user/", func(r chi.Router) {
			r.Post("/register", rh.Register)
			r.Post("/login", rh.Login)
		})
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
