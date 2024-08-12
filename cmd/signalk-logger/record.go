package main

import (
	"time"
)

type Record struct {
	Values     map[string]float64
	Timestamps map[string]time.Time
}

func NewRecord() *Record {
	return &Record{
		Values: make(map[string]float64),
	}
}

func (r *Record) AddValue(ts time.Time, key string, value float64) {
	r.Values[key] = value
}

func (r *Record) HasRequiredFields(requiredFields []string) bool {
	for _, requiredField := range requiredFields {
		if _, ok := r.Values[requiredField]; !ok {
			return false
		}
	}

	return true
}

func (r *Record) Clear() {
	r.Values = make(map[string]float64)
}
