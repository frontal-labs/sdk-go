package handlers

import "errors"

// IsAuthError reports whether err is an API authentication or authorization error.
func IsAuthError(err error) bool {
	apiError, ok := apiErrorFrom(err)
	return ok && (apiError.StatusCode == 401 || apiError.StatusCode == 403)
}

// IsRateLimitError reports whether err is an API rate limit error.
func IsRateLimitError(err error) bool {
	apiError, ok := apiErrorFrom(err)
	return ok && apiError.StatusCode == 429
}

// IsValidationError reports whether err is an API request validation error.
func IsValidationError(err error) bool {
	apiError, ok := apiErrorFrom(err)
	return ok && (apiError.StatusCode == 400 || apiError.StatusCode == 422)
}

// IsServerError reports whether err is an API 5xx error.
func IsServerError(err error) bool {
	apiError, ok := apiErrorFrom(err)
	return ok && apiError.StatusCode >= 500 && apiError.StatusCode < 600
}

func apiErrorFrom(err error) (*APIError, bool) {
	var apiError *APIError
	if !errors.As(err, &apiError) || apiError == nil {
		return nil, false
	}
	return apiError, true
}
