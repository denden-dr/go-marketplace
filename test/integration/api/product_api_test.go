//go:build integration

package api

import (
	"encoding/json"
	"go-marketplace/internal/common"
	"go-marketplace/internal/core/merchant"
	"go-marketplace/internal/core/product"
	"go-marketplace/internal/domain"
	"go-marketplace/internal/testutil"
	"net/http"
	"testing"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/suite"
)

type ProductApiTestSuite struct {
	testutil.ApiTestSuite
}

func TestProductApiTestSuite(t *testing.T) {
	suite.Run(t, new(ProductApiTestSuite))
}

func (s *ProductApiTestSuite) TestProductEndpoints() {
	tests := []struct {
		name           string
		method         string
		setup          func() (string, string, string) // returns token, merchID, prodID
		body           func(merchID string) interface{}
		query          string
		expectedStatus int
		verify         func(resp *http.Response)
	}{
		{
			name:   "Create_Product_Success",
			method: "POST",
			setup: func() (string, string, string) {
				u, token := s.CreateSeedUser()
				// Register merchant first
				merchReq := merchant.MerchantRegisterRequest{Name: "Shop", TaxID: "123"}
				req := s.JSONRequest("POST", "/api/auth/register-merchant", merchReq)
				req.Header.Set("Authorization", token)
				resp, _ := s.App.Test(req)
				var result common.SuccessResponse
				json.NewDecoder(resp.Body).Decode(&result)
				merchID := result.Data.(map[string]interface{})["id"].(string)
				// Get new token with merchant role
				u.Role = domain.RoleMerchant
				token = s.GetAuthHeader(u)
				return token, merchID, ""
			},
			body: func(merchID string) interface{} {
				return product.ProductCreateRequest{
					StoreID: uuid.MustParse(merchID),
					Name:    "Test Product",
					Price:   decimal.NewFromInt(100),
					Stock:   10,
				}
			},
			expectedStatus: http.StatusCreated,
		},
		{
			name:   "Search_Product_Success",
			method: "GET",
			setup: func() (string, string, string) {
				u, token := s.CreateSeedUser()
				// Setup merchant and product
				merchReq := merchant.MerchantRegisterRequest{Name: "Shop", TaxID: "123"}
				req := s.JSONRequest("POST", "/api/auth/register-merchant", merchReq)
				req.Header.Set("Authorization", token)
				resp, _ := s.App.Test(req)
				var result common.SuccessResponse
				json.NewDecoder(resp.Body).Decode(&result)
				merchID := result.Data.(map[string]interface{})["id"].(string)

				// Get new token with merchant role
				u.Role = domain.RoleMerchant
				token = s.GetAuthHeader(u)

				prodReq := product.ProductCreateRequest{
					StoreID: uuid.MustParse(merchID),
					Name:    "Searchable Item",
					Price:   decimal.NewFromInt(100),
					Stock:   10,
				}
				req = s.JSONRequest("POST", "/api/products", prodReq)
				req.Header.Set("Authorization", token)
				s.App.Test(req)
				return "", "", ""
			},
			query:          "q=Searchable",
			expectedStatus: http.StatusOK,
			verify: func(resp *http.Response) {
				var result common.SuccessResponse
				json.NewDecoder(resp.Body).Decode(&result)
				products := result.Data.([]interface{})
				s.NotEmpty(products)
			},
		},
		{
			name:   "Update_Product_Success",
			method: "PUT",
			setup: func() (string, string, string) {
				u, token := s.CreateSeedUser()
				merchReq := merchant.MerchantRegisterRequest{Name: "Shop", TaxID: "123"}
				req := s.JSONRequest("POST", "/api/auth/register-merchant", merchReq)
				req.Header.Set("Authorization", token)
				resp, _ := s.App.Test(req)
				var result common.SuccessResponse
				json.NewDecoder(resp.Body).Decode(&result)
				merchID := result.Data.(map[string]interface{})["id"].(string)

				// Get new token with merchant role
				u.Role = domain.RoleMerchant
				token = s.GetAuthHeader(u)

				prodReq := product.ProductCreateRequest{
					StoreID: uuid.MustParse(merchID),
					Name:    "Old Name",
					Price:   decimal.NewFromInt(100),
					Stock:   10,
				}
				req = s.JSONRequest("POST", "/api/products", prodReq)
				req.Header.Set("Authorization", token)
				resp, _ = s.App.Test(req)
				json.NewDecoder(resp.Body).Decode(&result)
				prodID := result.Data.(map[string]interface{})["id"].(string)
				return token, merchID, prodID
			},
			body: func(merchID string) interface{} {
				return product.ProductUpdateRequest{
					Name:  "New Name",
					Price: decimal.NewFromInt(150),
					Stock: 20,
				}
			},
			expectedStatus: http.StatusOK,
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			s.SetupTest()
			token, merchID, prodID := tt.setup()
			path := "/api/products"
			if prodID != "" && tt.method == "PUT" {
				path += "/" + prodID
			} else if tt.method == "GET" && tt.query != "" {
				path += "/search?" + tt.query
			}

			var body interface{}
			if tt.body != nil {
				body = tt.body(merchID)
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
