package handlers

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/frontal-labs/sdk-go/headers"
	"github.com/frontal-labs/sdk-go/models"
)

const maxErrorBodyBytes = 1 << 20
const maxFallbackMessageBytes = 4096

// MaxJSONResponseBytes bounds the amount of JSON response data decoded into memory.
const MaxJSONResponseBytes = 32 << 20

// HandleResponse consumes and closes a response body, decoding JSON into out when provided.
// If out implements io.Writer, the successful response body is copied to it unchanged.
func HandleResponse(response *http.Response, out any) error {
	if response == nil {
		return errors.New("frontal: response is nil")
	}
	if response.Body == nil {
		return errors.New("frontal: response body is nil")
	}
	defer response.Body.Close()
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
	limitedBody := &io.LimitedReader{R: response.Body, N: MaxJSONResponseBytes + 1}
	decoder := json.NewDecoder(limitedBody)
	if err := decoder.Decode(out); err != nil {
		if limitedBody.N == 0 {
			return fmt.Errorf("frontal: JSON response exceeds %d bytes", MaxJSONResponseBytes)
		}
		return fmt.Errorf("frontal: decode response body: %w", err)
	}
	var trailing any
	decodeErr := decoder.Decode(&trailing)
	if limitedBody.N == 0 {
		return fmt.Errorf("frontal: JSON response exceeds %d bytes", MaxJSONResponseBytes)
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

	apiError := &models.APIError{
		StatusCode: response.StatusCode,
		RequestID:  response.Header.Get(headers.RequestID),
	}
	if !decodeErrorPayload(body, apiError) && len(body) > 0 {
		message := strings.TrimSpace(string(bytes.TrimSpace(body)))
		if len(message) > maxFallbackMessageBytes {
			message = message[:maxFallbackMessageBytes]
		}
		apiError.Message = message
	}
	if apiError.Message == "" && apiError.Code == "" {
		apiError.Message = http.StatusText(response.StatusCode)
	}
	if truncated && apiError.Message == "" {
		apiError.Message = "API error response exceeded the 1 MiB limit"
	}
	return apiError
}

func decodeErrorPayload(body []byte, target *models.APIError) bool {
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
