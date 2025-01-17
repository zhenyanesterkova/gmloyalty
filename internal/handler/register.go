package handler

import (
	"encoding/json"
	"net/http"

	"github.com/sirupsen/logrus"

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

	log.WithFields(logrus.Fields{
		"login":    user.Login,
		"password": user.Password,
	}).Debug("register user")

	w.WriteHeader(http.StatusOK)
}
