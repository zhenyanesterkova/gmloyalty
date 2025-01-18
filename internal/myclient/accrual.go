package myclient

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"

	"github.com/zhenyanesterkova/gmloyalty/internal/service/order"
)

var (
	ErrNoContent       = errors.New("order is not registered in payment system")
	ErrTooManyRequests = errors.New("too many requests to payment system")
	ErrServer          = errors.New("accrual server error")
)

type AccrualStruct struct {
	client  *http.Client
	address string
}

type respStruct struct {
	Status  string  `json:"status"`
	Number  int64   `json:"order"`
	Accrual float64 `json:"accrual"`
}

func Accrual(address string) *AccrualStruct {
	return &AccrualStruct{
		address: address,
	}
}

func (acc AccrualStruct) GetCalculatingPoints(orderNum int64) (order.Order, error) {
	url := fmt.Sprintf("http://%s/api/orders/%d/", acc.address, orderNum)

	req, err := http.NewRequest(http.MethodGet, url, http.NoBody)
	if err != nil {
		return order.Order{}, fmt.Errorf("failed create request to accrual - %w", err)
	}

	resp, err := acc.client.Do(req)
	if err != nil {
		return order.Order{}, fmt.Errorf("failed do request - %w", err)
	}
	defer func() {
		errBodyClose := resp.Body.Close()
		if errBodyClose != nil {
			log.Fatalf("failed close accrual resp body - %v", errBodyClose)
		}
	}()

	if resp.StatusCode == http.StatusNoContent {
		return order.Order{}, ErrNoContent
	}
	if resp.StatusCode == http.StatusTooManyRequests {
		return order.Order{}, ErrTooManyRequests
	}
	if resp.StatusCode == http.StatusInternalServerError {
		return order.Order{}, ErrServer
	}

	orderData := respStruct{}
	decoder := json.NewDecoder(resp.Body)
	if err := decoder.Decode(&orderData); err != nil {
		return order.Order{}, fmt.Errorf("failed decode response from accrual - %w", err)
	}

	return order.Order{
		Number:  orderData.Number,
		Status:  orderData.Status,
		Accrual: orderData.Accrual,
	}, nil
}
