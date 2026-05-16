//go:build integration

package api

import (
	"encoding/json"
	"go-marketplace/internal/common"
	"go-marketplace/internal/core/cart"
	"go-marketplace/internal/core/product"
	"go-marketplace/internal/testutil"
	"net/http"
	"testing"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/suite"
)

type CartApiTestSuite struct {
	testutil.ApiTestSuite
}

func TestCartApiTestSuite(t *testing.T) {
	suite.Run(t, new(CartApiTestSuite))
}

func (s *CartApiTestSuite) TestCartEndpoints() {
	tests := []struct {
		name           string
		method         string
		setup          func() (string, string) // returns token, prodID
		body           func(prodID string) interface{}
		expectedStatus int
		verify         func(resp *http.Response)
	}{
		{
			name:   "Add_To_Cart_Success",
			method: "POST",
			setup: func() (string, string) {
				_, token, merchID := s.CreateSeedMerchant()

				prodReq := product.ProductCreateRequest{
					StoreID: uuid.MustParse(merchID),
					Name:    "Item",
					Price:   decimal.NewFromInt(100),
					Stock:   10,
				}
				req := s.JSONRequest("POST", "/api/products", prodReq)
				req.Header.Set("Authorization", token)
				resp, _ := s.App.Test(req)
				var result common.SuccessResponse
				json.NewDecoder(resp.Body).Decode(&result)
				prodID := result.Data.(map[string]interface{})["id"].(string)
				return token, prodID
			},
			body: func(prodID string) interface{} {
				return cart.AddToCartRequest{
					ProductID: uuid.MustParse(prodID),
					Quantity:  2,
				}
			},
			expectedStatus: http.StatusCreated,
		},
		{
			name:   "Get_Cart_Success",
			method: "GET",
			setup: func() (string, string) {
				_, token, merchID := s.CreateSeedMerchant()

				prodReq := product.ProductCreateRequest{
					StoreID: uuid.MustParse(merchID),
					Name:    "Item",
					Price:   decimal.NewFromInt(100),
					Stock:   10,
				}
				req := s.JSONRequest("POST", "/api/products", prodReq)
				req.Header.Set("Authorization", token)
				resp, _ := s.App.Test(req)
				var result common.SuccessResponse
				json.NewDecoder(resp.Body).Decode(&result)
				prodID := result.Data.(map[string]interface{})["id"].(string)

				addReq := cart.AddToCartRequest{ProductID: uuid.MustParse(prodID), Quantity: 1}
				req = s.JSONRequest("POST", "/api/users/cart", addReq)
				req.Header.Set("Authorization", token)
				s.App.Test(req)
				return token, ""
			},
			expectedStatus: http.StatusOK,
			verify: func(resp *http.Response) {
				var result common.SuccessResponse
				json.NewDecoder(resp.Body).Decode(&result)
				cartData := result.Data.(map[string]interface{})
				items := cartData["items"].([]interface{})
				s.Len(items, 1)
			},
		},
		{
			name:   "Update_Cart_Item_Success",
			method: "PUT",
			setup: func() (string, string) {
				_, token, merchID := s.CreateSeedMerchant()

				prodReq := product.ProductCreateRequest{
					StoreID: uuid.MustParse(merchID),
					Name:    "Item",
					Price:   decimal.NewFromInt(100),
					Stock:   10,
				}
				req := s.JSONRequest("POST", "/api/products", prodReq)
				req.Header.Set("Authorization", token)
				resp, _ := s.App.Test(req)
				var result common.SuccessResponse
				json.NewDecoder(resp.Body).Decode(&result)
				prodID := result.Data.(map[string]interface{})["id"].(string)

				addReq := cart.AddToCartRequest{ProductID: uuid.MustParse(prodID), Quantity: 1}
				req = s.JSONRequest("POST", "/api/users/cart", addReq)
				req.Header.Set("Authorization", token)
				s.App.Test(req)
				return token, prodID
			},
			body: func(prodID string) interface{} {
				return cart.UpdateCartItemRequest{Quantity: 5}
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:   "Remove_From_Cart_Success",
			method: "DELETE",
			setup: func() (string, string) {
				_, token, merchID := s.CreateSeedMerchant()

				prodReq := product.ProductCreateRequest{
					StoreID: uuid.MustParse(merchID),
					Name:    "Item",
					Price:   decimal.NewFromInt(100),
					Stock:   10,
				}
				req := s.JSONRequest("POST", "/api/products", prodReq)
				req.Header.Set("Authorization", token)
				resp, _ := s.App.Test(req)
				var result common.SuccessResponse
				json.NewDecoder(resp.Body).Decode(&result)
				prodID := result.Data.(map[string]interface{})["id"].(string)

				addReq := cart.AddToCartRequest{ProductID: uuid.MustParse(prodID), Quantity: 1}
				req = s.JSONRequest("POST", "/api/users/cart", addReq)
				req.Header.Set("Authorization", token)
				s.App.Test(req)
				return token, prodID
			},
			expectedStatus: http.StatusOK,
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			s.SetupTest()
			token, prodID := tt.setup()
			path := "/api/users/cart"
			if prodID != "" && (tt.method == "PUT" || tt.method == "DELETE") {
				path += "/" + prodID
			}

			var body interface{}
			if tt.body != nil {
				body = tt.body(prodID)
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
