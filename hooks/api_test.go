package hooks

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestSuccessHandler(t *testing.T) {
	handler := http.HandlerFunc(successHandler)
	req, _ := http.NewRequest("POST", "/", nil)
	rr := httptest.NewRecorder()
	handler(rr, req)
	if rr.Code != http.StatusOK {
		t.Errorf("Expected success 200 but received: %v", rr.Code)
	}
}

func TestAuthMiddlewareSuccess(t *testing.T) {
	handler := func(http.ResponseWriter, *http.Request) *AppError {
		return nil
	}
	key := "TESTKEY"
	req, _ := http.NewRequest("GET", "/", nil)
	req.Header.Set("X-API-KEY", key)
	authHandler := authMiddleware(handler, key)
	rr := httptest.NewRecorder()
	err := authHandler(rr, req)
	if err != nil {
		t.Errorf("Expected auth to succeed, instead got error: %v", err)
	}
}

func TestAuthMiddlewareFail(t *testing.T) {
	handler := func(http.ResponseWriter, *http.Request) *AppError {
		return nil
	}
	key := "TESTKEY"
	req, _ := http.NewRequest("GET", "/", nil)
	req.Header.Set("X-API-KEY", "INCORRECTKEY")
	authHandler := authMiddleware(handler, key)
	rr := httptest.NewRecorder()
	err := authHandler(rr, req)
	if err == nil {
		t.Errorf("Expected auth error, instead got nil")
	}
}
