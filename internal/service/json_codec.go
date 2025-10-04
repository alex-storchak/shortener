package service

import (
	"encoding/json"
	"fmt"
	"io"

	"github.com/alex-storchak/shortener/internal/model"
)

type IJSONRequestDecoder interface {
	Decode(io.Reader) (model.ShortenRequest, error)
}

type JSONRequestDecoder struct{}

func (d JSONRequestDecoder) Decode(r io.Reader) (model.ShortenRequest, error) {
	var req model.ShortenRequest
	dec := json.NewDecoder(r)
	if err := dec.Decode(&req); err != nil {
		return model.ShortenRequest{}, fmt.Errorf("failed to decode request json: %w", err)
	}
	return req, nil
}

type IJSONBatchRequestDecoder interface {
	DecodeBatch(io.Reader) ([]model.BatchShortenRequestItem, error)
}

type JSONBatchRequestDecoder struct{}

func (d JSONBatchRequestDecoder) DecodeBatch(r io.Reader) ([]model.BatchShortenRequestItem, error) {
	var req []model.BatchShortenRequestItem
	dec := json.NewDecoder(r)
	if err := dec.Decode(&req); err != nil {
		return nil, fmt.Errorf("failed to decode batch request json: %w", err)
	}
	return req, nil
}

type IJSONEncoder interface {
	Encode(io.Writer, any) error
}

type JSONEncoder struct{}

func (e JSONEncoder) Encode(w io.Writer, v any) error {
	enc := json.NewEncoder(w)
	return enc.Encode(v)
}
