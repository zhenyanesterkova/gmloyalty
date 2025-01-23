package handler

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"

	"github.com/jackc/pgx/v5"

	"github.com/zhenyanesterkova/gmloyalty/internal/helper"
	"github.com/zhenyanesterkova/gmloyalty/internal/middleware"
	"github.com/zhenyanesterkova/gmloyalty/internal/service/order"
)

func (rh *RepositorieHandler) Orders(w http.ResponseWriter, r *http.Request) {
	log := rh.Logger.LogrusLog
	userID, ok := r.Context().Value(middleware.UserIDContextKey).(int)
	if !ok {
		http.Error(w, "No auth", http.StatusUnauthorized)
		return
	}

	orderNumBytes, err := io.ReadAll(r.Body)
	if err != nil {
		log.Errorf("failed reading order number from body of request to add order: %v", err)
		http.Error(w, TextServerError, http.StatusInternalServerError)
		return
	}

	if r.Header.Get("Content-Type") != "text/plain" ||
		len(orderNumBytes) == 0 {
		http.Error(w, "content-type must be text/plain and lenght of order number should not be zero", http.StatusBadRequest)
		return
	}

	orderNum := string(orderNumBytes)
	validNumber := helper.LuhnCheck(orderNum)
	if !validNumber {
		http.Error(w, "incorrect order number format", http.StatusUnprocessableEntity)
		return
	}

	orderData, err := rh.Repo.GetOrderByOrderNum(orderNum)
	if err != nil {
		if !errors.Is(err, pgx.ErrNoRows) {
			log.Errorf("failed get order from orders: %v", err)
			http.Error(w, TextServerError, http.StatusInternalServerError)
			return
		}

		orderData.Status = order.StatusNew
		orderData.Number = orderNum
		orderData.UserID = userID

		err := rh.Repo.AddOrder(orderData)
		if err != nil {
			log.Errorf("failed add order to orders: %v", err)
			http.Error(w, TextServerError, http.StatusInternalServerError)
			return
		}

		rh.pool.Queue <- orderData

		w.WriteHeader(http.StatusAccepted)
		return
	}

	if orderData.Status == order.StatusNew ||
		orderData.Status == order.StatusProcessing {
		rh.pool.Queue <- orderData
	}

	if orderData.UserID != userID {
		http.Error(w, TextConflictUserIDError, http.StatusConflict)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (rh *RepositorieHandler) GetOrderList(w http.ResponseWriter, r *http.Request) {
	log := rh.Logger.LogrusLog
	userID, ok := r.Context().Value(middleware.UserIDContextKey).(int)
	if !ok {
		http.Error(w, "No auth", http.StatusUnauthorized)
		return
	}

	orderList, err := rh.Repo.GetOrderList(userID)
	if err != nil {
		log.Errorf("failed get order list from DB: %v", err)
		http.Error(w, TextServerError, http.StatusInternalServerError)
		return
	}

	if len(orderList) == 0 {
		w.WriteHeader(http.StatusNoContent)
		return
	}

	w.Header().Set("Content-Type", "application/json")

	enc := json.NewEncoder(w)
	if err := enc.Encode(orderList); err != nil {
		log.Errorf("error encode order list in get order list handler - %v", err)
		http.Error(w, TextServerError, http.StatusInternalServerError)
		return
	}
}
