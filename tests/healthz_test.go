package tests

import (
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestHealthz(t *testing.T) {
	port := "8081"
	endpoint := "/api/healthz"
	url := "http://localhost:" + port + endpoint
	req, _ := http.NewRequest(http.MethodGet, url, http.NoBody)

	client := &http.Client{}
	resp, _ := client.Do(req) //nolint:bodyclose // we close it in defer

	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			t.Errorf("got error on closing body %s", err)
		}
	}(resp.Body)

	body, _ := io.ReadAll(resp.Body)
	bodyString := strings.TrimSpace(string(body))

	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.Equal(t, "application/json; charset=UTF-8", resp.Header.Get("Content-Type"))
	assert.Equal(t, `{"status":"OK"}`, bodyString)
}

func TestHealth(t *testing.T) {
	port := "8081"
	endpoint := "/health"
	url := "http://localhost:" + port + endpoint
	req, _ := http.NewRequest(http.MethodGet, url, http.NoBody)

	client := &http.Client{}
	resp, _ := client.Do(req) //nolint:bodyclose // we close it in defer

	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			t.Errorf("got error on closing body %s", err)
		}
	}(resp.Body)

	body, _ := io.ReadAll(resp.Body)
	bodyString := strings.TrimSpace(string(body))

	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.Equal(t, "application/json; charset=UTF-8", resp.Header.Get("Content-Type"))
	assert.Equal(t, `{"status":"OK"}`, bodyString)
}
