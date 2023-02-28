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
	return NewExAccrualSystem(cfg)
}

type ExAccrualSystem struct {
	BaseURL string
}

func NewExAccrualSystem(cfg config.Config) *ExAccrualSystem {
	return &ExAccrualSystem{BaseURL: cfg.AccrualAddress}
}

func (s *ExAccrualSystem) GetOrderUpdates(order entity.Order) (entity.Order, int, error) {
	var sleep int

	reqURL, err := url.Parse(s.BaseURL)
	if err != nil {
		log.Fatalln("Wrong accrual system URL:", err)
	}

	reqURL.Path = path.Join("/api/orders/", strconv.Itoa(order.Number))

	r, err := http.Get(reqURL.String())
	if err != nil {
		log.Println("Can't get order updates from external API:", err)
		return order, sleep, err
	}

	body, err := io.ReadAll(r.Body)
	defer r.Body.Close()

	if err != nil {
		log.Println("Can't read response body:", err)
		return order, sleep, err
	}
	if r.StatusCode == http.StatusNoContent {
		return order, sleep, nil
	}
	if r.StatusCode == http.StatusTooManyRequests {
		retryAfter, err := strconv.Atoi(r.Header.Get("Retry-After"))
		if err != nil {
			return order, sleep, err
		}
		return order, retryAfter, err
	}
	if err = json.Unmarshal(body, &order); err != nil {
		log.Println(err)
		return order, sleep, err
	}
	return order, sleep, nil
}
