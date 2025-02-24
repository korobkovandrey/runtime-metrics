package model

import "errors"

var (
	ErrMetricNotFound     = errors.New("metric not found")
	ErrMetricAlreadyExist = errors.New("metric already exist")
	ErrTypeIsNotValid     = errors.New("type is not valid")
	ErrValueIsNotValid    = errors.New("value is not valid")
)
