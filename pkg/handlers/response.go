package handlers

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/frontal-labs/sdk-go/pkg/headers"
	"github.com/frontal-labs/sdk-go/pkg/utils"
)

const (
	maxErrorBodyBytes       = 1 << 20
	maxFallbackMessageBytes = 4096
)

// MaxJSONResponseBytes bounds the amount of JSON response data decoded into memory.
const MaxJSONResponseBytes = 32 << 20

// HandleResponse consumes and closes a response body, decoding JSON into out when provided.
// If out implements io.Writer, the successful response body is copied to it unchanged.
func HandleResponse(response *http.Response, out any) error {
	return HandleResponseLimit(response, out, MaxJSONResponseBytes)
}

// HandleResponseLimit consumes a response and limits decoded JSON to maxBytes.
// Successful responses copied to an io.Writer are streamed without this limit.
func HandleResponseLimit(response *http.Response, out any, maxBytes int64) error {
	if response == nil {
		return errors.New("frontal: response is nil")
	}
	if response.Body == nil {
		return errors.New("frontal: response body is nil")
	}
	defer func() { _ = response.Body.Close() }()
	if maxBytes <= 0 || maxBytes == int64(^uint64(0)>>1) {
		return errors.New("frontal: JSON response limit must be positive and below the maximum int64 value")
	}
	if response.StatusCode < http.StatusOK || response.StatusCode >= http.StatusMultipleChoices && response.StatusCode != http.StatusNotModified {
		return decodeAPIError(response)
	}
	if writer, ok := out.(io.Writer); ok {
		if _, err := io.Copy(writer, response.Body); err != nil {
			return fmt.Errorf("frontal: read response body: %w", err)
		}
		return nil
	}
	if out == nil || response.StatusCode == http.StatusNoContent || response.StatusCode == http.StatusNotModified {
		if _, err := io.Copy(io.Discard, response.Body); err != nil {
			return fmt.Errorf("frontal: read response body: %w", err)
		}
		return nil
	}
	limitedBody := &io.LimitedReader{R: response.Body, N: maxBytes + 1}
	decoder := json.NewDecoder(limitedBody)
	if err := decoder.Decode(out); err != nil {
		if limitedBody.N == 0 {
			return fmt.Errorf("frontal: JSON response exceeds %d bytes", maxBytes)
		}
		return fmt.Errorf("frontal: decode response body: %w", err)
	}
	var trailing any
	decodeErr := decoder.Decode(&trailing)
	if limitedBody.N == 0 {
		return fmt.Errorf("frontal: JSON response exceeds %d bytes", maxBytes)
	}
	if !errors.Is(decodeErr, io.EOF) {
		if decodeErr == nil {
			return errors.New("frontal: response contains multiple JSON values")
		}
		return fmt.Errorf("frontal: decode trailing response data: %w", decodeErr)
	}
	return nil
}

func decodeAPIError(response *http.Response) error {
	body, err := io.ReadAll(io.LimitReader(response.Body, maxErrorBodyBytes+1))
	if err != nil {
		return fmt.Errorf("frontal: read API error response: %w", err)
	}
	truncated := len(body) > maxErrorBodyBytes
	if truncated {
		body = body[:maxErrorBodyBytes]
	}

	apiError := &APIError{
		StatusCode: response.StatusCode,
		RequestID:  response.Header.Get(headers.RequestID),
		Retryable:  retryableStatus(response.StatusCode),
	}
	if delay, ok := utils.RetryAfter(response.Header.Get("Retry-After"), time.Now()); ok {
		apiError.RetryAfter = delay
	}
	if !decodeErrorPayload(body, apiError) && len(body) > 0 {
		message := strings.TrimSpace(string(bytes.TrimSpace(body)))
		if len(message) > maxFallbackMessageBytes {
			message = message[:maxFallbackMessageBytes]
		}
		apiError.Message = message
	}
	if apiError.RequestID == "" {
		apiError.RequestID = requestIDFromPayload(body)
	}
	if apiError.RequestID == "" && response.Request != nil {
		apiError.RequestID = response.Request.Header.Get(headers.RequestID)
	}
	apiError.Retryable = retryableStatus(response.StatusCode)
	if apiError.Message == "" && apiError.Code == "" {
		apiError.Message = http.StatusText(response.StatusCode)
	}
	if truncated && apiError.Message == "" {
		apiError.Message = "API error response exceeded the 1 MiB limit"
	}
	return apiError
}

func requestIDFromPayload(body []byte) string {
	var envelope struct {
		Error     json.RawMessage `json:"error"`
		RequestID string          `json:"requestId"`
	}
	if err := json.Unmarshal(body, &envelope); err != nil {
		return ""
	}
	if len(envelope.Error) > 0 && !bytes.Equal(envelope.Error, []byte("null")) && envelope.Error[0] != '"' {
		var nested struct {
			Snake string `json:"request_id"`
			Camel string `json:"requestId"`
		}
		if err := json.Unmarshal(envelope.Error, &nested); err == nil {
			if nested.Snake != "" {
				return nested.Snake
			}
			return nested.Camel
		}
	}
	return envelope.RequestID
}

func retryableStatus(status int) bool {
	switch status {
	case http.StatusRequestTimeout, http.StatusTooEarly, http.StatusTooManyRequests,
		http.StatusInternalServerError, http.StatusBadGateway, http.StatusServiceUnavailable, http.StatusGatewayTimeout:
		return true
	default:
		return false
	}
}

func decodeErrorPayload(body []byte, target *APIError) bool {
	var envelope struct {
		Error json.RawMessage `json:"error"`
	}
	if err := json.Unmarshal(body, &envelope); err == nil && len(envelope.Error) > 0 && !bytes.Equal(envelope.Error, []byte("null")) {
		if envelope.Error[0] == '"' {
			return json.Unmarshal(envelope.Error, &target.Message) == nil
		}
		if err := json.Unmarshal(envelope.Error, target); err == nil {
			return target.Message != "" || target.Code != "" || target.Type != ""
		}
	}
	if err := json.Unmarshal(body, target); err != nil {
		return false
	}
	return target.Message != "" || target.Code != "" || target.Type != ""
}
