package tests

import (
	"io"
	"net/http"
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/suite"
)

type HealthSuite struct {
	suite.Suite
	client *http.Client
	port   string
}

func (s *HealthSuite) getURL(endpoint string) string {
	return "http://localhost:" + s.port + endpoint
}

func (s *HealthSuite) SetupSuite() {
	// start the server
	s.port = os.Getenv("PORT")
	if s.port == "" {
		s.Fail("got empty port from env")
	}

	s.client = &http.Client{}
}

func (s *HealthSuite) TearDownSuite() {
	// stop the server
}

func (s *HealthSuite) TestHealthz() {
	url := s.getURL("/api/healthz")
	req, err := http.NewRequest(http.MethodGet, url, http.NoBody)
	s.Require().NoError(err)

	resp, err := s.client.Do(req)
	s.Require().NoError(err)

	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	s.Require().NoError(err)

	bodyString := strings.TrimSpace(string(body))

	s.Equal(http.StatusOK, resp.StatusCode)
	s.Equal("application/json; charset=UTF-8", resp.Header.Get("Content-Type"))
	s.Equal(`{"status":"OK"}`, bodyString)
}

func (s *HealthSuite) TestHealth() {
	url := s.getURL("/health")
	req, err := http.NewRequest(http.MethodGet, url, http.NoBody)
	s.Require().NoError(err)

	resp, err := s.client.Do(req)
	s.Require().NoError(err)

	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	s.Require().NoError(err)

	bodyString := strings.TrimSpace(string(body))

	s.Equal(http.StatusOK, resp.StatusCode)
	s.Equal("application/json; charset=UTF-8", resp.Header.Get("Content-Type"))
	s.Equal(`{"status":"OK"}`, bodyString)
}

func TestHealthSuite(t *testing.T) {
	suite.Run(t, new(HealthSuite))
}
