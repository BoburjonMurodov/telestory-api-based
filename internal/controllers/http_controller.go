package controllers

import (
	"fmt"
	"net/http"
)

type HTTPController struct{}

func NewHTTPController() *HTTPController {
	return &HTTPController{}
}

func (c *HTTPController) SetupRoutes() {
	http.HandleFunc("/health", c.HealthCheck)
}

func (c *HTTPController) HealthCheck(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	fmt.Fprintln(w, "OK")
}
