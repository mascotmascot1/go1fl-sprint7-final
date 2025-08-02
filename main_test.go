package main

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strconv"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCafeNegative(t *testing.T) {
	handler := http.HandlerFunc(mainHandle)

	requests := []struct {
		requestURI string
		status     int
		message    string
	}{
		{"/cafe", http.StatusBadRequest, "unknown city"},
		{"/cafe?city=omsk", http.StatusBadRequest, "unknown city"},
		{"/cafe?city=tula&count=na", http.StatusBadRequest, "incorrect count"},
	}
	for _, v := range requests {
		response := httptest.NewRecorder()
		req := httptest.NewRequest("GET", v.requestURI, nil)
		handler.ServeHTTP(response, req)

		assert.Equal(t, v.status, response.Code)
		assert.Equal(t, v.message, strings.TrimSpace(response.Body.String()))
	}
}

func TestCafeWhenOk(t *testing.T) {
	handler := http.HandlerFunc(mainHandle)

	requests := []string{
		"/cafe?count=2&city=moscow",
		"/cafe?city=tula",
		"/cafe?city=moscow&search=ложка",
	}
	for _, v := range requests {
		response := httptest.NewRecorder()
		req := httptest.NewRequest("GET", v, nil)

		handler.ServeHTTP(response, req)

		assert.Equal(t, http.StatusOK, response.Code)
	}
}

func TestCafeCount(t *testing.T) {
	const (
		basePath = "/cafe"
		city     = "moscow"
	)

	cafeCount := len(cafeList[city])
	handler := http.HandlerFunc(mainHandle)

	requests := []struct {
		count, want int
	}{
		{0, min(0, cafeCount)},
		{1, min(1, cafeCount)},
		{2, min(2, cafeCount)},
		{100, min(100, cafeCount)},
	}

	for _, r := range requests {
		q := url.Values{}
		q.Add("city", city)
		q.Add("count", strconv.Itoa(r.count))
		requestURI := fmt.Sprintf("%s?%s", basePath, q.Encode())

		request := httptest.NewRequest(http.MethodGet, requestURI, nil)
		response := httptest.NewRecorder()
		handler.ServeHTTP(response, request)

		require.Equal(t, http.StatusOK, response.Code)

		cafesRaw := strings.TrimSpace(response.Body.String())
		actualCount := 0
		if cafesRaw != "" {
			actualCount = len(strings.Split(cafesRaw, ","))
		}
		assert.Equal(t, r.want, actualCount)
	}
}

func TestCafeSearch(t *testing.T) {
	const (
		basePath = "/cafe"
		city     = "moscow"
	)

	handler := http.HandlerFunc(mainHandle)
	requests := []struct {
		search    string
		wantCount int
	}{
		{"фасоль", 0},
		{"кофе", 2},
		{"вилка", 1},
	}

	for _, r := range requests {
		q := url.Values{}
		q.Add("city", city)
		q.Add("search", r.search)

		requestURI := fmt.Sprintf("%s?%s", basePath, q.Encode())

		request := httptest.NewRequest(http.MethodGet, requestURI, nil)
		response := httptest.NewRecorder()
		handler.ServeHTTP(response, request)

		require.Equal(t, http.StatusOK, response.Code)

		cafesRaw := strings.TrimSpace(response.Body.String())

		if cafesRaw == "" {
			assert.Equal(t, r.wantCount, 0)
			continue
		}

		cafes := strings.Split(cafesRaw, ",")
		assert.Equal(t, r.wantCount, len(cafes))

		for _, cafe := range cafes {
			assert.True(t, strings.Contains(strings.ToLower(cafe), strings.ToLower(r.search)))
		}
	}
}
