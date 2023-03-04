package accrualservice

import (
	"encoding/json"
	"io"
	"log"
	"net/http"
	"net/url"
	"path"
	"strconv"

	"github.com/bbt-t/ya-go-d/internal/config"
	"github.com/bbt-t/ya-go-d/internal/entity"
)

type AccrualSystem interface {
	GetOrderUpdates(order entity.Order) (entity.Order, int, error)
}

func NewAccrualSystem(cfg config.Config) AccrualSystem {
	return newExAccrualSystem(cfg)
}

type exAccrualSystem struct {
	baseURL string
}

func newExAccrualSystem(cfg config.Config) *exAccrualSystem {
	return &exAccrualSystem{baseURL: cfg.AccrualAddress}
}

func (s *exAccrualSystem) GetOrderUpdates(order entity.Order) (entity.Order, int, error) {
	reqURL, err := url.Parse(s.baseURL)
	if err != nil {
		log.Fatalln("Wrong accrual system URL:", err)
	}

	reqURL.Path = path.Join("/api/orders/", strconv.Itoa(order.Number))

	r, errGet := http.Get(reqURL.String())
	if errGet != nil {
		log.Printf("Can't get order updates from external API: %+v\n", err)
		return order, 0, errGet
	}

	body, errBody := io.ReadAll(r.Body)
	defer r.Body.Close()

	if errBody != nil {
		log.Printf("Can't read response body: %+v\n", err)
		return order, 0, errBody
	}

	switch r.StatusCode {
	case http.StatusNoContent:
		return order, 0, nil
	case http.StatusTooManyRequests:
		retryAfter, err := strconv.Atoi(r.Header.Get("Retry-After"))
		if err != nil {
			return order, 0, err
		}
		return order, retryAfter, err
	}

	if err = json.Unmarshal(body, &order); err != nil {
		log.Println(err)
		return order, 0, err
	}
	return order, 0, nil
}
