package handler

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"

	"github.com/zhenyanesterkova/gmloyalty/internal/service/user"
)

func (rh *RepositorieHandler) Login(w http.ResponseWriter, r *http.Request) {
	log := rh.Logger.LogrusLog

	userData := user.User{}
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&userData); err != nil {
		var syntaxErr *json.SyntaxError
		if errors.As(err, &syntaxErr) {
			http.Error(w, TextInvalidFormatError, http.StatusBadRequest)
			return
		}
		http.Error(w, TextServerError, http.StatusInternalServerError)
		return
	}

	if userData.Login == "" || userData.Password == "" {
		http.Error(w, TextInvalidFormatError, http.StatusBadRequest)
		return
	}

	userID, err := rh.Repo.Login(userData)
	if err != nil {
		var pgErr *pgconn.PgError
		switch {
		case errors.As(err, &pgErr) && pgerrcode.IsNoData(pgErr.Code):
		case errors.Is(err, pgx.ErrNoRows):
			http.Error(w, "No user", http.StatusBadRequest)
			return
		case errors.Is(err, user.ErrBadPass):
			log.Debug(err)
			http.Error(w, "Invalid username/password", http.StatusUnauthorized)
			return
		}
		log.Errorf("failed login user: %v", err)
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
