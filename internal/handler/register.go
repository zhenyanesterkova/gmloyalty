package handler

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5/pgconn"

	"github.com/zhenyanesterkova/gmloyalty/internal/service/user"
)

func (rh *RepositorieHandler) Register(w http.ResponseWriter, r *http.Request) {
	log := rh.Logger.LogrusLog

	user := user.User{}
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&user); err != nil {
		log.Errorf("handler func Register(): error decode user data - %v", err)
		http.Error(w, TextServerError, http.StatusInternalServerError)
		return
	}

	err := rh.Repo.Register(r.Context(), user)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgerrcode.IsIntegrityConstraintViolation(pgErr.Code) {
			http.Error(w, TextLoginError, http.StatusBadRequest)
			return
		}
		log.Errorf("handler func Register(): error register user - %v", err)
		http.Error(w, TextServerError, http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}
