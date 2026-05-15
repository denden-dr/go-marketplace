package testutil

import (
	"bytes"
	"context"
	"encoding/json"
	"go-marketplace/internal/common"
	"go-marketplace/internal/core/auth"
	"go-marketplace/internal/core/cart"
	"go-marketplace/internal/core/health"
	"go-marketplace/internal/core/merchant"
	"go-marketplace/internal/core/order"
	"go-marketplace/internal/core/payment"
	"go-marketplace/internal/core/product"
	"go-marketplace/internal/core/user"
	"go-marketplace/internal/core/wallet"
	"go-marketplace/internal/domain"
	"go-marketplace/internal/middleware"
	"go-marketplace/internal/server"
	"net/http"
	"net/http/httptest"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/requestid"
	"github.com/google/uuid"
	"github.com/stretchr/testify/mock"
	"golang.org/x/crypto/bcrypt"
)

type ApiTestSuite struct {
	IntegrationSuite
	App                 *fiber.App
	JwtSecret           string
	MockPaymentProvider *payment.MockPaymentProvider
	MockGoogleClient    *auth.MockGoogleClient
}

func (s *ApiTestSuite) SetupSuite() {
	s.IntegrationSuite.SetupSuite()
	s.JwtSecret = "test-secret"
}

func (s *ApiTestSuite) SetupTest() {
	s.IntegrationSuite.SetupTest()

	// Initialize Fiber app
	s.App = fiber.New(fiber.Config{
		Immutable:    true,
		ErrorHandler: common.ErrorHandler,
	})

	// Apply global middlewares
	s.App.Use(requestid.New())
	s.App.Use(middleware.Logger())

	// Initialize Layers
	userRepo := user.NewUserRepository(s.DB)
	merchantRepo := merchant.NewMerchantRepository(s.DB)
	productRepo := product.NewProductRepository(s.DB)
	walletRepo := wallet.NewWalletRepository(s.DB)
	sessionRepo := auth.NewSessionRepository(s.DB)
	verificationRepo := auth.NewVerificationRepository(s.DB)
	cartRepo := cart.NewCartRepository(s.DB)
	orderRepo := order.NewOrderRepository(s.DB)
	paymentRepo := payment.NewPaymentRepository(s.DB)

	// Mocks
	mailService := auth.NewMockMailService(s.T())
	mailService.On("SendVerificationCode", mock.Anything, mock.Anything, mock.Anything).Return(nil).Maybe()

	s.MockPaymentProvider = payment.NewMockPaymentProvider(s.T())
	s.MockGoogleClient = auth.NewMockGoogleClient(s.T())

	authService := auth.NewAuthService(userRepo, sessionRepo, verificationRepo, mailService, s.MockGoogleClient, s.JwtSecret)
	userService := user.NewUserService(userRepo)
	merchantService := merchant.NewMerchantService(merchantRepo, userRepo, walletRepo)
	productService := product.NewProductService(productRepo, merchantRepo)
	walletService := wallet.NewWalletService(walletRepo)
	cartService := cart.NewCartService(cartRepo, productRepo)

	orderManager := order.NewOrderManager(orderRepo, productRepo)
	paymentService := payment.NewPaymentService(paymentRepo, walletService, s.MockPaymentProvider, orderManager, s.DB)
	orderService := order.NewOrderService(orderRepo, cartRepo, productRepo, walletService, userRepo, merchantRepo, paymentService)

	authHandler := auth.NewAuthHandler(authService, s.MockGoogleClient, "http://test-marketplace.local/login-success")
	userHandler := user.NewUserHandler(userService)
	merchantHandler := merchant.NewMerchantHandler(merchantService)
	productHandler := product.NewProductHandler(productService)
	walletHandler := wallet.NewWalletHandler(walletService)
	cartHandler := cart.NewCartHandler(cartService)
	orderHandler := order.NewOrderHandler(orderService, merchantRepo)
	paymentHandler := payment.NewPaymentHandler(paymentService)
	healthHandler := health.NewHealthHandler(s.DB)

	server.SetupRoutes(
		s.App,
		authHandler,
		userHandler,
		merchantHandler,
		productHandler,
		walletHandler,
		cartHandler,
		orderHandler,
		paymentHandler,
		healthHandler,
		s.JwtSecret,
		"test",
	)
}

func (s *ApiTestSuite) JSONRequest(method, url string, body interface{}) *http.Request {
	var buf bytes.Buffer
	if body != nil {
		err := json.NewEncoder(&buf).Encode(body)
		s.Require().NoError(err)
	}

	req := httptest.NewRequest(method, url, &buf)
	req.Header.Set("Content-Type", "application/json")
	return req
}

func (s *ApiTestSuite) GetAuthHeader(user *domain.User) string {
	token, err := auth.GenerateAccessToken(user.ID, s.JwtSecret)
	s.Require().NoError(err)
	return "Bearer " + token
}

func (s *ApiTestSuite) CreateSeedUser() (*domain.User, string) {
	pass := "password123"
	hashedPass, _ := bcrypt.GenerateFromPassword([]byte(pass), bcrypt.DefaultCost)
	hashedPassStr := string(hashedPass)

	u := &domain.User{
		ID:           uuid.New(),
		FullName:     "Test User",
		Username:     "testuser_" + uuid.New().String()[:8],
		Email:        "test_" + uuid.New().String()[:8] + "@example.com",
		Password:     &hashedPassStr,
		AuthProvider: domain.AuthProviderLocal,
		IsVerified:   true,
		CreatedAt:    time.Now(),
	}

	_, err := s.DB.NamedExecContext(context.Background(), `
		INSERT INTO users (id, full_name, username, email, password, auth_provider, is_verified, created_at)
		VALUES (:id, :full_name, :username, :email, :password, :auth_provider, :is_verified, :created_at)
	`, u)
	s.Require().NoError(err)

	token, err := auth.GenerateAccessToken(u.ID, s.JwtSecret)
	s.Require().NoError(err)

	return u, "Bearer " + token
}
