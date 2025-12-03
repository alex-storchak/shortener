package codec

import (
	"fmt"
	"io"
	"net/http"

	"github.com/mailru/easyjson"
)

// EasyJSONEncode кодирует тип, реализующий easyjson.Marshaler, напрямую в ResponseWriter.
func EasyJSONEncode(w http.ResponseWriter, status int, v easyjson.Marshaler) error {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if _, err := easyjson.MarshalToWriter(v, w); err != nil {
		return fmt.Errorf("easyjson encode: %w", err)
	}
	return nil
}

// EasyJSONDecode считывает весь body (нужно для easyjson) и распаковывает в v.
// v должен быть указателем на тип, реализующий easyjson.Unmarshaler.
func EasyJSONDecode(r *http.Request, v easyjson.Unmarshaler) error {
	data, err := io.ReadAll(r.Body)
	if err != nil {
		return fmt.Errorf("read body: %w", err)
	}
	if cErr := r.Body.Close(); cErr != nil {
		return fmt.Errorf("close body: %w", cErr)
	}

	if err := easyjson.Unmarshal(data, v); err != nil {
		return fmt.Errorf("easyjson decode: %w", err)
	}
	return nil
}
