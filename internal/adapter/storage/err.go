package storage

import "errors"

var (
	ErrNotFound         = errors.New("not found")
	ErrExists           = errors.New("already exists")
	ErrNoEnoughBalance  = errors.New("not enough money on balance")
	ErrNumAlreadyLoaded = errors.New("already loaded order number")
	ErrWrongUser        = errors.New("already loaded by another user")
	ErrEmptyQueue       = errors.New("queue is empty")
	ErrSearchType       = errors.New("received wrong search type")
)
