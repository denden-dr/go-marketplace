//go:build integration

package api

import (
	"encoding/json"
	"go-marketplace/internal/common"
	"go-marketplace/internal/testutil"
	"net/http"
	"testing"

	"github.com/stretchr/testify/suite"
)

type WalletApiTestSuite struct {
	testutil.ApiTestSuite
}

func TestWalletApiTestSuite(t *testing.T) {
	suite.Run(t, new(WalletApiTestSuite))
}

func (s *WalletApiTestSuite) TestWalletEndpoints() {
	tests := []struct {
		name           string
		method         string
		path           string
		setup          func() string // returns token
		expectedStatus int
		verify         func(resp *http.Response)
	}{
		{
			name:   "Create_Success",
			method: "POST",
			path:   "/api/wallets/",
			setup: func() string {
				_, token := s.CreateSeedUser()
				return token
			},
			expectedStatus: http.StatusCreated,
		},
		{
			name:   "Get_Success",
			method: "GET",
			path:   "/api/wallets/",
			setup: func() string {
				_, token := s.CreateSeedUser()
				// Create wallet first
				req := s.JSONRequest("POST", "/api/wallets/", nil)
				req.Header.Set("Authorization", token)
				s.App.Test(req)
				return token
			},
			expectedStatus: http.StatusOK,
			verify: func(resp *http.Response) {
				var result common.SuccessResponse
				json.NewDecoder(resp.Body).Decode(&result)
				walletData := result.Data.(map[string]interface{})
				s.Equal("active", walletData["status"])
			},
		},
		{
			name:           "Unauthorized",
			method:         "GET",
			path:           "/api/wallets/",
			setup:          func() string { return "" },
			expectedStatus: http.StatusUnauthorized,
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			s.SetupTest()
			token := tt.setup()
			req := s.JSONRequest(tt.method, tt.path, nil)
			if token != "" {
				req.Header.Set("Authorization", token)
			}

			resp, err := s.App.Test(req)
			s.Require().NoError(err)
			s.Equal(tt.expectedStatus, resp.StatusCode)

			if tt.expectedStatus >= 400 {
				var pd common.ProblemDetails
				json.NewDecoder(resp.Body).Decode(&pd)
				s.Equal(tt.expectedStatus, pd.Status)
				s.NotEmpty(pd.Title)
				s.Contains(pd.Type, "/errors/")
				if tt.expectedStatus == http.StatusBadRequest && pd.Type == "/errors/validation-failed" {
					s.NotEmpty(pd.Errors)
				}
			}

			if tt.verify != nil {
				tt.verify(resp)
			}
		})
	}
}
