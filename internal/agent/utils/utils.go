package utils

import (
	"errors"
	"fmt"
	"reflect"
	"runtime"
	"strconv"
	"strings"
)

var (
	ErrNumberIsNotNumber = errors.New(`is not number`)
	ErrFieldNotFound     = errors.New(`not found`)
)

func convertReflectValueToString(value reflect.Value) (string, error) {
	switch value.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return strconv.Itoa(int(value.Int())), nil
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return strconv.Itoa(int(value.Uint())), nil
	case reflect.Float32, reflect.Float64:
		return strconv.FormatFloat(value.Float(), 'g', -1, 64), nil
	case reflect.Bool:
		if value.Bool() {
			return `1`, nil
		} else {
			return `0`, nil
		}
	case reflect.String:
		s := value.String()
		if _, err := strconv.ParseFloat(s, 64); err != nil {
			return ``, fmt.Errorf(`"%v" %w: %w`, value, ErrNumberIsNotNumber, err)
		}
		return s, nil
	case reflect.Ptr:
		return convertReflectValueToString(value.Elem())
	default:
		return ``, fmt.Errorf(`"%v" %w`, value, ErrNumberIsNotNumber)
	}
}

func GetRuntimeMetrics(fields ...string) (result map[string]string, errNotNumber error, errNotFound error) {
	result = map[string]string{}
	memStats := &runtime.MemStats{}
	runtime.ReadMemStats(memStats)
	v := reflect.ValueOf(*memStats)
	var (
		r   string
		err error
	)
	var notNumberFields, notFoundFields []string

	for _, i := range fields {
		value := v.FieldByName(i)
		if value.IsValid() {
			r, err = convertReflectValueToString(value)
			if err == nil {
				result[i] = r
			} else {
				notNumberFields = append(notNumberFields, i)
			}
		} else {
			notFoundFields = append(notFoundFields, i)
		}
	}
	if len(notNumberFields) != 0 {
		errNotNumber = fmt.Errorf(`%s: %w`, `"`+strings.Join(notNumberFields, `", "`)+`"`, ErrNumberIsNotNumber)
	}
	if len(notFoundFields) != 0 {
		errNotFound = fmt.Errorf(`%s: %w`, `"`+strings.Join(notFoundFields, `", "`)+`"`, ErrFieldNotFound)
	}
	return
}
