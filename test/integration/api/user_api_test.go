//go:build integration

package api

import (
	"encoding/json"
	"go-marketplace/internal/common"
	"go-marketplace/internal/core/user"
	"go-marketplace/internal/domain"
	"go-marketplace/internal/testutil"
	"net/http"
	"testing"

	"github.com/stretchr/testify/suite"
)

type UserApiTestSuite struct {
	testutil.ApiTestSuite
}

func TestUserApiTestSuite(t *testing.T) {
	suite.Run(t, new(UserApiTestSuite))
}

func (s *UserApiTestSuite) TestAddressEndpoints() {
	tests := []struct {
		name           string
		method         string
		setup          func() (string, string) // returns token, addrID (if any)
		body           func() interface{}      // returns request body
		expectedStatus int
		verify         func(resp *http.Response)
	}{
		{
			name:   "Add_Address_Success",
			method: "POST",
			setup: func() (string, string) {
				_, token := s.CreateSeedUser()
				return token, ""
			},
			body: func() interface{} {
				return user.AddressRequest{
					Tag:           domain.AddressTagHome,
					RecipientName: "John Doe",
					PhoneNumber:   "08123456789",
					StreetAddress: "Main Street 123",
					City:          "Jakarta",
					Province:      "DKI Jakarta",
					PostalCode:    "12345",
					IsDefault:     true,
				}
			},
			expectedStatus: http.StatusCreated,
		},
		{
			name:   "List_Addresses_Success",
			method: "GET",
			setup: func() (string, string) {
				_, token := s.CreateSeedUser()
				// Add one address
				addReq := user.AddressRequest{
					Tag: domain.AddressTagHome, RecipientName: "John", PhoneNumber: "123",
					StreetAddress: "Street", City: "City", Province: "Province", PostalCode: "123",
				}
				req := s.JSONRequest("POST", "/api/users/addresses", addReq)
				req.Header.Set("Authorization", token)
				s.App.Test(req)
				return token, ""
			},
			expectedStatus: http.StatusOK,
			verify: func(resp *http.Response) {
				var result common.SuccessResponse
				json.NewDecoder(resp.Body).Decode(&result)
				addresses := result.Data.([]interface{})
				s.Len(addresses, 1)
			},
		},
		{
			name:   "Update_Address_Success",
			method: "PUT",
			setup: func() (string, string) {
				_, token := s.CreateSeedUser()
				addReq := user.AddressRequest{
					Tag: domain.AddressTagHome, RecipientName: "John", PhoneNumber: "123",
					StreetAddress: "Street", City: "City", Province: "Province", PostalCode: "123",
				}
				req := s.JSONRequest("POST", "/api/users/addresses", addReq)
				req.Header.Set("Authorization", token)
				resp, _ := s.App.Test(req)
				var result common.SuccessResponse
				json.NewDecoder(resp.Body).Decode(&result)
				addrID := result.Data.(map[string]interface{})["id"].(string)
				return token, addrID
			},
			body: func() interface{} {
				return user.AddressRequest{
					Tag:           domain.AddressTagWork,
					RecipientName: "Updated Name",
					PhoneNumber:   "08123456789",
					StreetAddress: "Main Street 123",
					City:          "Jakarta",
					Province:      "DKI Jakarta",
					PostalCode:    "12345",
				}
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:   "Delete_Address_Success",
			method: "DELETE",
			setup: func() (string, string) {
				_, token := s.CreateSeedUser()
				addReq := user.AddressRequest{
					Tag: domain.AddressTagHome, RecipientName: "John", PhoneNumber: "123",
					StreetAddress: "Street", City: "City", Province: "Province", PostalCode: "123",
				}
				req := s.JSONRequest("POST", "/api/users/addresses", addReq)
				req.Header.Set("Authorization", token)
				resp, _ := s.App.Test(req)
				var result common.SuccessResponse
				json.NewDecoder(resp.Body).Decode(&result)
				addrID := result.Data.(map[string]interface{})["id"].(string)
				return token, addrID
			},
			expectedStatus: http.StatusOK,
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			s.SetupTest()
			token, addrID := tt.setup()
			path := "/api/users/addresses"
			if addrID != "" && (tt.method == "PUT" || tt.method == "DELETE") {
				path += "/" + addrID
			}

			var body interface{}
			if tt.body != nil {
				body = tt.body()
			}

			req := s.JSONRequest(tt.method, path, body)
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
			}

			if tt.verify != nil {
				tt.verify(resp)
			}
		})
	}
}
