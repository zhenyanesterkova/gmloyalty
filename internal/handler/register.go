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
		var syntaxErr *json.SyntaxError
		if errors.As(err, &syntaxErr) {
			http.Error(w, TextInvalidFormatError, http.StatusBadRequest)
			return
		}
		log.Errorf("failed decode user: %v", err)
		http.Error(w, TextServerError, http.StatusInternalServerError)
		return
	}

	if user.Login == "" || user.Password == "" {
		http.Error(w, TextInvalidFormatError, http.StatusBadRequest)
		return
	}

	userID, err := rh.Repo.Register(r.Context(), user)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgerrcode.IsIntegrityConstraintViolation(pgErr.Code) {
			http.Error(w, TextLoginError, http.StatusConflict)
			return
		}
		log.Errorf("handler func Register(): error register user - %v", err)
		http.Error(w, TextServerError, http.StatusInternalServerError)
		return
	}

	tokenJWT, err := rh.jwtSess.Create(userID)
	if err != nil {
		log.Errorf("failed create token JWT: %v", err)
		http.Error(w, TextServerError, http.StatusInternalServerError)
		return
	}

	w.Header().Add("Authorization", tokenJWT)

	w.WriteHeader(http.StatusOK)
}
