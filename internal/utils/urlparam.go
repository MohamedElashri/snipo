package utils

import "net/http"

type contextKey string

const PathParamsKey contextKey = "path_params"

func URLParam(r *http.Request, key string) string {
	params, ok := r.Context().Value(PathParamsKey).(map[string]string)
	if !ok {
		return ""
	}
	return params[key]
}
