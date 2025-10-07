package service

import (
	"encoding/json"
	"fmt"
	"io"

	"github.com/alex-storchak/shortener/internal/model"
)

type JSONShortenRequestDecoder struct{}

func (d JSONShortenRequestDecoder) Decode(r io.Reader) (model.ShortenRequest, error) {
	var req model.ShortenRequest
	dec := json.NewDecoder(r)
	if err := dec.Decode(&req); err != nil {
		return model.ShortenRequest{}, fmt.Errorf("failed to decode request json: %w", err)
	}
	return req, nil
}

type JSONShortenBatchRequestDecoder struct{}

func (d JSONShortenBatchRequestDecoder) Decode(r io.Reader) ([]model.BatchShortenRequestItem, error) {
	var req []model.BatchShortenRequestItem
	dec := json.NewDecoder(r)
	if err := dec.Decode(&req); err != nil {
		return nil, fmt.Errorf("failed to decode shorten batch request json: %w", err)
	}
	return req, nil
}

type JSONDeleteBatchRequestDecoder struct{}

func (d JSONDeleteBatchRequestDecoder) Decode(r io.Reader) ([]string, error) {
	var req []string
	dec := json.NewDecoder(r)
	if err := dec.Decode(&req); err != nil {
		return nil, fmt.Errorf("failed to decode delete batch request json: %w", err)
	}
	return req, nil
}

type Encoder interface {
	Encode(io.Writer, any) error
}

type JSONEncoder struct{}

func (e JSONEncoder) Encode(w io.Writer, v any) error {
	enc := json.NewEncoder(w)
	return enc.Encode(v)
}
