//go:build integration

package api

import (
	"go-marketplace/internal/common"
	"go-marketplace/internal/core/cart"
	"go-marketplace/internal/core/order"
	"go-marketplace/internal/core/payment"
	"go-marketplace/internal/core/product"
	"go-marketplace/internal/core/user"
	"go-marketplace/internal/domain"
	"go-marketplace/internal/testutil"
	"net/http"
	"testing"

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

			// 1. Setup: Merchant, Product, Address
			_, merchantToken, merchID := s.CreateSeedMerchant()

			prodReq := product.ProductCreateRequest{
				StoreID: uuid.MustParse(merchID),
				Name:    "Webhook Item",
				Price:   decimal.NewFromInt(100),
				Stock:   10,
			}
			req := s.JSONRequest("POST", "/api/products", prodReq)
			req.Header.Set("Authorization", merchantToken)
			resp, _ := s.App.Test(req)
			s.Require().Equal(http.StatusCreated, resp.StatusCode)
			result := s.DecodeSuccess(resp)
			prodID := result.Data.(map[string]interface{})["id"].(string)

			addrReq := user.AddressRequest{
				Tag: domain.AddressTagHome, RecipientName: "Recipient", PhoneNumber: "123",
				StreetAddress: "Street", City: "City", Province: "Province", PostalCode: "12345",
			}
			req = s.JSONRequest("POST", "/api/users/addresses", addrReq)
			req.Header.Set("Authorization", buyerToken)
			resp, _ = s.App.Test(req)
			s.Require().Equal(http.StatusCreated, resp.StatusCode)
			addrResult := s.DecodeSuccess(resp)
			addrID := addrResult.Data.(map[string]interface{})["id"].(string)

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

			payResult := s.DecodeSuccess(resp)
			payRes := payResult.Data.(map[string]interface{})
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
				s.DecodeResponse(resp, &pd)
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

			orderResult := s.DecodeSuccess(resp)
			orderData := orderResult.Data.(map[string]interface{})
			s.Equal(string(tt.expectedOrder), orderData["status"])
		})
	}
}
