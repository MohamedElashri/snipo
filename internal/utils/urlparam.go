package utils

import "net/http"

func URLParam(r *http.Request, key string) string {
	params, ok := r.Context().Value("path_params").(map[string]string)
	if !ok {
		return ""
	}
	return params[key]
}
