//go:build integration

package api

import (
	"encoding/json"
	"go-marketplace/internal/common"
	"go-marketplace/internal/core/auth"
	"go-marketplace/internal/domain"
	"go-marketplace/internal/testutil"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
	"golang.org/x/oauth2"
)

type AuthApiTestSuite struct {
	testutil.ApiTestSuite
}

func TestAuthApiTestSuite(t *testing.T) {
	suite.Run(t, new(AuthApiTestSuite))
}

func (s *AuthApiTestSuite) TestRegister() {
	tests := []struct {
		name           string
		reqBody        auth.RegisterRequest
		setup          func() string // returns email to conflict with
		expectedStatus int
		expectedMsg    string
	}{
		{
			name: "Success",
			reqBody: auth.RegisterRequest{
				FullName: "New User",
				Email:    "newuser@example.com",
				Password: "password123",
				Username: "newuser",
			},
			expectedStatus: http.StatusCreated,
			expectedMsg:    "User registered successfully",
		},
		{
			name: "Conflict",
			setup: func() string {
				u, _ := s.CreateSeedUser()
				return u.Email
			},
			reqBody: auth.RegisterRequest{
				FullName: "Duplicate User",
				Password: "password123",
				Username: "duplicate",
			},
			expectedStatus: http.StatusConflict,
		},
		{
			name: "Validation_Failed",
			reqBody: auth.RegisterRequest{
				FullName: "",
				Email:    "invalid-email",
				Password: "123",
			},
			expectedStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			s.SetupTest() // Ensure clean DB for each subtest if needed, but suite usually does it per TestMethod.
			// Wait, s.SetupTest() is called by the suite before EACH TestMethod, NOT before EACH s.Run.
			// So if I want isolation between subtests, I must call it manually or handle it.
			// Given IntegrationSuite.SetupTest truncates tables, calling it here is safe.
			s.SetupTest()

			if tt.setup != nil {
				tt.reqBody.Email = tt.setup()
			}

			req := s.JSONRequest("POST", "/api/auth/register", tt.reqBody)
			resp, err := s.App.Test(req)
			s.Require().NoError(err)
			s.Equal(tt.expectedStatus, resp.StatusCode)

			if tt.expectedStatus < 400 {
				if tt.expectedMsg != "" {
					var result common.SuccessResponse
					json.NewDecoder(resp.Body).Decode(&result)
					s.Equal(tt.expectedMsg, result.Message)
				}
			} else {
				var pd common.ProblemDetails
				json.NewDecoder(resp.Body).Decode(&pd)
				s.Equal(tt.expectedStatus, pd.Status)
				s.NotEmpty(pd.Title)
				s.Contains(pd.Type, "/errors/")
				if tt.expectedStatus == http.StatusBadRequest && pd.Type == "/errors/validation-failed" {
					s.NotEmpty(pd.Errors)
				}
			}
		})
	}
}

func (s *AuthApiTestSuite) TestLogin() {
	tests := []struct {
		name           string
		setup          func() (string, string) // returns email, password
		expectedStatus int
	}{
		{
			name: "Success",
			setup: func() (string, string) {
				u, _ := s.CreateSeedUser()
				return u.Email, "password123"
			},
			expectedStatus: http.StatusOK,
		},
		{
			name: "InvalidCredentials",
			setup: func() (string, string) {
				u, _ := s.CreateSeedUser()
				return u.Email, "wrongpassword"
			},
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name: "UserNotFound",
			setup: func() (string, string) {
				return "nonexistent@example.com", "password123"
			},
			expectedStatus: http.StatusUnauthorized,
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			s.SetupTest()
			email, password := tt.setup()
			reqBody := auth.LoginRequest{
				Email:    email,
				Password: password,
			}

			req := s.JSONRequest("POST", "/api/auth/login", reqBody)
			resp, err := s.App.Test(req)
			s.Require().NoError(err)
			s.Equal(tt.expectedStatus, resp.StatusCode)

			if tt.expectedStatus == http.StatusOK {
				cookies := resp.Cookies()
				var hasAccess, hasRefresh bool
				for _, cookie := range cookies {
					if cookie.Name == "access_token" {
						hasAccess = true
					}
					if cookie.Name == "refresh_token" {
						hasRefresh = true
					}
				}
				s.True(hasAccess, "Should have access_token cookie")
				s.True(hasRefresh, "Should have refresh_token cookie")
			} else {
				var pd common.ProblemDetails
				json.NewDecoder(resp.Body).Decode(&pd)
				s.Equal(tt.expectedStatus, pd.Status)
				s.NotEmpty(pd.Title)
				s.Contains(pd.Type, "/errors/")
			}
		})
	}
}

func (s *AuthApiTestSuite) TestGetProfile() {
	tests := []struct {
		name           string
		setup          func() (*domain.User, string) // returns user, token
		expectedStatus int
		verify         func(user *domain.User, resp *http.Response)
	}{
		{
			name: "Success",
			setup: func() (*domain.User, string) {
				return s.CreateSeedUser()
			},
			expectedStatus: http.StatusOK,
			verify: func(u *domain.User, resp *http.Response) {
				var result common.SuccessResponse
				json.NewDecoder(resp.Body).Decode(&result)
				data := result.Data.(map[string]interface{})
				s.Equal(u.Email, data["email"])
			},
		},
		{
			name: "Unauthorized",
			setup: func() (*domain.User, string) {
				return nil, ""
			},
			expectedStatus: http.StatusUnauthorized,
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			s.SetupTest()
			u, token := tt.setup()
			req := s.JSONRequest("GET", "/api/users/me", nil)
			if token != "" {
				req.Header.Set("Authorization", token)
			}

			resp, err := s.App.Test(req)
			s.Require().NoError(err)
			s.Equal(tt.expectedStatus, resp.StatusCode)

			if tt.verify != nil {
				tt.verify(u, resp)
			}
		})
	}
}

func (s *AuthApiTestSuite) TestGoogleLogin() {
	s.SetupTest()
	s.MockGoogleClient.On("GetAuthURL", mock.Anything).Return("https://accounts.google.com/o/oauth2/auth?client_id=...")

	req := httptest.NewRequest("GET", "/api/auth/google/login", nil)
	resp, err := s.App.Test(req)
	s.Require().NoError(err)
	s.Equal(http.StatusFound, resp.StatusCode)

	cookies := resp.Cookies()
	var hasState bool
	for _, cookie := range cookies {
		if cookie.Name == "oauth_state" {
			hasState = true
			s.NotEmpty(cookie.Value)
		}
	}
	s.True(hasState, "Should have oauth_state cookie")
}

func (s *AuthApiTestSuite) TestGoogleCallback() {
	email := "google_test@example.com"
	sub := "google-sub-123"
	name := "Google Tester"

	tests := []struct {
		name           string
		setup          func() (string, string) // returns code, state
		mockSetup      func()
		expectedStatus int
	}{
		{
			name: "Success - New User",
			setup: func() (string, string) {
				return "valid-code", "valid-state"
			},
			mockSetup: func() {
				s.MockGoogleClient.On("ExchangeCode", mock.Anything, "valid-code").Return(&oauth2.Token{}, nil)
				s.MockGoogleClient.On("GetUserInfo", mock.Anything, mock.Anything).Return(&auth.GoogleUserInfo{
					Sub:   sub,
					Email: email,
					Name:  name,
				}, nil)
			},
			expectedStatus: http.StatusFound,
		},
		{
			name: "Invalid State",
			setup: func() (string, string) {
				return "valid-code", "wrong-state"
			},
			mockSetup: func() {
				// No calls expected
			},
			expectedStatus: http.StatusUnauthorized,
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			s.SetupTest()
			code, state := tt.setup()
			if tt.mockSetup != nil {
				tt.mockSetup()
			}

			req := httptest.NewRequest("GET", "/api/auth/google/callback?code="+code+"&state="+state, nil)
			// Add the expected state cookie
			req.AddCookie(&http.Cookie{
				Name:  "oauth_state",
				Value: "valid-state",
			})

			resp, err := s.App.Test(req)
			s.Require().NoError(err)
			s.Equal(tt.expectedStatus, resp.StatusCode)

			if tt.expectedStatus == http.StatusFound {
				// Verify user created
				var u domain.User
				err := s.DB.Get(&u, "SELECT * FROM users WHERE email = $1", email)
				s.Require().NoError(err)
				s.Equal(name, u.FullName)
				s.Equal(domain.AuthProviderGoogle, u.AuthProvider)
			}
		})
	}
}
