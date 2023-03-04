package accrualservice

import (
	"encoding/json"
	"io"
	"net/http"
	"net/url"
	"path"
	"strconv"

	"github.com/bbt-t/ya-go-d/internal/config"
	"github.com/bbt-t/ya-go-d/internal/entity"
	"github.com/bbt-t/ya-go-d/pkg"
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
		pkg.Log.Fatal(err)
	}

	reqURL.Path = path.Join("/api/orders/", strconv.Itoa(order.Number))

	r, errGet := http.Get(reqURL.String())
	if errGet != nil {
		pkg.Log.Info(err.Error())
		return order, 0, errGet
	}

	body, errBody := io.ReadAll(r.Body)
	defer r.Body.Close()

	if errBody != nil {
		pkg.Log.Info(err.Error())
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
		pkg.Log.Warn(err.Error())
		return order, 0, err
	}
	return order, 0, nil
}
