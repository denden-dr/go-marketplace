//go:build integration

package api

import (
	"context"
	"encoding/json"
	"go-marketplace/internal/common"
	"go-marketplace/internal/core/cart"
	"go-marketplace/internal/core/merchant"
	"go-marketplace/internal/core/order"
	"go-marketplace/internal/core/payment"
	"go-marketplace/internal/core/product"
	userPkg "go-marketplace/internal/core/user"
	"go-marketplace/internal/domain"
	"go-marketplace/internal/testutil"
	"net/http"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
)

type PaymentApiTestSuite struct {
	testutil.ApiTestSuite
}

func TestPaymentApiTestSuite(t *testing.T) {
	suite.Run(t, new(PaymentApiTestSuite))
}

func (s *PaymentApiTestSuite) TestMidtransWebhook() {
	tests := []struct {
		name           string
		status         string
		expectedStatus int
		expectedOrder  domain.OrderStatus
	}{
		{
			name:           "Success",
			status:         "settlement",
			expectedStatus: http.StatusOK,
			expectedOrder:  domain.OrderStatusProcessing,
		},
		{
			name:           "Failure",
			status:         "deny",
			expectedStatus: http.StatusOK,
			expectedOrder:  domain.OrderStatusCancelled,
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			s.SetupTest()
			_, buyerToken := s.CreateSeedUser()
			merchantUser, merchantToken := s.CreateSeedUser()

			// 1. Setup: Merchant, Product, Address
			merchReq := merchant.MerchantRegisterRequest{Name: "Webhook Shop", TaxID: "W111"}
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
			`, uuid.New(), merchantUser.ID, "W-MERCH-W", 0, 0, "IDR", "active", time.Now(), time.Now())

			prodReq := product.ProductCreateRequest{
				StoreID: uuid.MustParse(merchID),
				Name:    "Webhook Item",
				Price:   decimal.NewFromInt(100),
				Stock:   10,
			}
			req = s.JSONRequest("POST", "/api/products", prodReq)
			req.Header.Set("Authorization", merchantToken)
			resp, _ = s.App.Test(req)
			s.Require().Equal(http.StatusCreated, resp.StatusCode)
			json.NewDecoder(resp.Body).Decode(&result)
			prodID := result.Data.(map[string]interface{})["id"].(string)

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

			// Add to Cart
			addReq := cart.AddToCartRequest{ProductID: uuid.MustParse(prodID), Quantity: 1}
			req = s.JSONRequest("POST", "/api/users/cart", addReq)
			req.Header.Set("Authorization", buyerToken)
			s.App.Test(req)

			// 2. Checkout with Midtrans
			s.MockPaymentProvider.On("CreateTransaction", mock.Anything, mock.Anything).Return("snap-123", nil)

			checkoutReq := order.CheckoutRequest{
				PaymentMethod: domain.PaymentMethodMidtrans,
				AddressID:     common.Ptr(uuid.MustParse(addrID)),
			}
			req = s.JSONRequest("POST", "/api/users/orders", checkoutReq)
			req.Header.Set("Authorization", buyerToken)
			resp, _ = s.App.Test(req)
			s.Require().Equal(http.StatusCreated, resp.StatusCode)

			json.NewDecoder(resp.Body).Decode(&result)
			payRes := result.Data.(map[string]interface{})
			paymentID := payRes["payment_id"].(string)
			orderID := payRes["order_ids"].([]interface{})[0].(string)

			// 3. Send Webhook
			webhookReq := payment.MidtransWebhookRequest{
				TransactionStatus: tt.status,
				OrderID:           paymentID,
			}
			req = s.JSONRequest("POST", "/api/payments/webhook", webhookReq)
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

			// 4. Verify Order Status
			req = s.JSONRequest("GET", "/api/users/orders/"+orderID, nil)
			req.Header.Set("Authorization", buyerToken)
			resp, _ = s.App.Test(req)
			s.Require().Equal(http.StatusOK, resp.StatusCode)

			json.NewDecoder(resp.Body).Decode(&result)
			orderData := result.Data.(map[string]interface{})
			s.Equal(string(tt.expectedOrder), orderData["status"])
		})
	}
}
