//go:build integration

package api

import (
	"context"
	"encoding/json"
	"go-marketplace/internal/common"
	"go-marketplace/internal/core/cart"
	"go-marketplace/internal/core/merchant"
	"go-marketplace/internal/core/order"
	"go-marketplace/internal/core/product"
	userPkg "go-marketplace/internal/core/user"
	"go-marketplace/internal/domain"
	"go-marketplace/internal/testutil"
	"net/http"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/suite"
)

type OrderApiTestSuite struct {
	testutil.ApiTestSuite
}

func TestOrderApiTestSuite(t *testing.T) {
	suite.Run(t, new(OrderApiTestSuite))
}

func (s *OrderApiTestSuite) TestCheckout() {
	tests := []struct {
		name           string
		paymentMethod  domain.PaymentMethod
		setup          func() (string, uuid.UUID) // returns buyerToken, addrID
		expectedStatus int
	}{
		{
			name:          "Wallet_Success",
			paymentMethod: domain.PaymentMethodWallet,
			setup: func() (string, uuid.UUID) {
				buyer, buyerToken := s.CreateSeedUser()
				merchantUser, merchantToken := s.CreateSeedUser()

				// Register Merchant
				merchReq := merchant.MerchantRegisterRequest{Name: "Order Shop", TaxID: "111"}
				req := s.JSONRequest("POST", "/api/auth/register-merchant", merchReq)
				req.Header.Set("Authorization", merchantToken)
				resp, _ := s.App.Test(req)
				s.Require().Equal(http.StatusCreated, resp.StatusCode)

				var result common.SuccessResponse
				json.NewDecoder(resp.Body).Decode(&result)
				merchID := result.Data.(map[string]interface{})["id"].(string)

				// Merchant wallet
				s.DB.ExecContext(context.Background(), `
					INSERT INTO wallets (id, user_id, wallet_number, balance, pending_balance, currency, status, created_at, updated_at)
					VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
				`, uuid.New(), merchantUser.ID, "W-MERCH", 0, 0, "IDR", "active", time.Now(), time.Now())

				// Create Product
				prodReq := product.ProductCreateRequest{
					StoreID: uuid.MustParse(merchID),
					Name:    "Order Item",
					Price:   decimal.NewFromInt(100),
					Stock:   100,
				}
				req = s.JSONRequest("POST", "/api/products", prodReq)
				req.Header.Set("Authorization", merchantToken)
				resp, _ = s.App.Test(req)
				s.Require().Equal(http.StatusCreated, resp.StatusCode)
				json.NewDecoder(resp.Body).Decode(&result)
				prodID := result.Data.(map[string]interface{})["id"].(string)

				// Create Address
				addrReq := userPkg.AddressRequest{
					Tag: domain.AddressTagHome, RecipientName: "Recipient", PhoneNumber: "123",
					StreetAddress: "Street", City: "City", Province: "Province", PostalCode: "12345",
				}
				req = s.JSONRequest("POST", "/api/users/addresses", addrReq)
				req.Header.Set("Authorization", buyerToken)
				resp, _ = s.App.Test(req)
				s.Require().Equal(http.StatusCreated, resp.StatusCode)
				json.NewDecoder(resp.Body).Decode(&result)
				addrID := result.Data.(map[string]interface{})["id"].(string)

				// Buyer Wallet Balance
				s.DB.ExecContext(context.Background(), `
					INSERT INTO wallets (id, user_id, wallet_number, balance, pending_balance, currency, status, created_at, updated_at)
					VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
				`, uuid.New(), buyer.ID, "W-BUYER", 1000, 0, "IDR", "active", time.Now(), time.Now())

				// Add to Cart
				addReq := cart.AddToCartRequest{ProductID: uuid.MustParse(prodID), Quantity: 2}
				req = s.JSONRequest("POST", "/api/users/cart", addReq)
				req.Header.Set("Authorization", buyerToken)
				s.App.Test(req)

				return buyerToken, uuid.MustParse(addrID)
			},
			expectedStatus: http.StatusCreated,
		},
		{
			name:          "Wallet_InsufficientBalance",
			paymentMethod: domain.PaymentMethodWallet,
			setup: func() (string, uuid.UUID) {
				buyer, buyerToken := s.CreateSeedUser()
				merchantUser, merchantToken := s.CreateSeedUser()

				// Register Merchant
				merchReq := merchant.MerchantRegisterRequest{Name: "Order Shop", TaxID: "111"}
				req := s.JSONRequest("POST", "/api/auth/register-merchant", merchReq)
				req.Header.Set("Authorization", merchantToken)
				resp, _ := s.App.Test(req)
				s.Require().Equal(http.StatusCreated, resp.StatusCode)

				var result common.SuccessResponse
				json.NewDecoder(resp.Body).Decode(&result)
				merchID := result.Data.(map[string]interface{})["id"].(string)

				// Merchant wallet
				s.DB.ExecContext(context.Background(), `
					INSERT INTO wallets (id, user_id, wallet_number, balance, pending_balance, currency, status, created_at, updated_at)
					VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
				`, uuid.New(), merchantUser.ID, "W-MERCH-FLOW", 0, 0, "IDR", "active", time.Now(), time.Now())

				// Create Product
				prodReq := product.ProductCreateRequest{
					StoreID: uuid.MustParse(merchID),
					Name:    "Order Item",
					Price:   decimal.NewFromInt(100),
					Stock:   100,
				}
				req = s.JSONRequest("POST", "/api/products", prodReq)
				req.Header.Set("Authorization", merchantToken)
				resp, _ = s.App.Test(req)
				s.Require().Equal(http.StatusCreated, resp.StatusCode)
				json.NewDecoder(resp.Body).Decode(&result)
				prodID := result.Data.(map[string]interface{})["id"].(string)

				// Create Address
				addrReq := userPkg.AddressRequest{
					Tag: domain.AddressTagHome, RecipientName: "Recipient", PhoneNumber: "123",
					StreetAddress: "Street", City: "City", Province: "Province", PostalCode: "12345",
				}
				req = s.JSONRequest("POST", "/api/users/addresses", addrReq)
				req.Header.Set("Authorization", buyerToken)
				resp, _ = s.App.Test(req)
				s.Require().Equal(http.StatusCreated, resp.StatusCode)
				json.NewDecoder(resp.Body).Decode(&result)
				addrID := result.Data.(map[string]interface{})["id"].(string)

				// Buyer Wallet Balance (0 balance)
				s.DB.ExecContext(context.Background(), `
					INSERT INTO wallets (id, user_id, wallet_number, balance, pending_balance, currency, status, created_at, updated_at)
					VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
				`, uuid.New(), buyer.ID, "W-BUYER-ZERO", 0, 0, "IDR", "active", time.Now(), time.Now())

				// Add to Cart
				addReq := cart.AddToCartRequest{ProductID: uuid.MustParse(prodID), Quantity: 2}
				req = s.JSONRequest("POST", "/api/users/cart", addReq)
				req.Header.Set("Authorization", buyerToken)
				s.App.Test(req)

				return buyerToken, uuid.MustParse(addrID)
			},
			expectedStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			s.SetupTest()
			token, addrID := tt.setup()

			checkoutReq := order.CheckoutRequest{
				PaymentMethod: tt.paymentMethod,
				AddressID:     common.Ptr(addrID),
			}
			req := s.JSONRequest("POST", "/api/users/orders", checkoutReq)
			req.Header.Set("Authorization", token)

			resp, err := s.App.Test(req)
			s.Require().NoError(err)
			s.Equal(tt.expectedStatus, resp.StatusCode)

			if tt.expectedStatus == http.StatusCreated {
				var result common.SuccessResponse
				json.NewDecoder(resp.Body).Decode(&result)
				payRes := result.Data.(map[string]interface{})
				orders := payRes["order_ids"].([]interface{})
				s.Len(orders, 1)
			} else if tt.expectedStatus >= 400 {
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
