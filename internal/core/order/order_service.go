package order

import (
	"context"
	"fmt"
	"time"

	"go-marketplace/internal/core/cart"
	"go-marketplace/internal/core/product"
	"go-marketplace/internal/core/user"
	"go-marketplace/internal/core/wallet"
	"go-marketplace/internal/domain"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/shopspring/decimal"
)

type OrderServiceInterface interface {
	CreateUserCheckout(ctx context.Context, userID uuid.UUID, req CheckoutRequest) (*OrderPaymentResponse, error)
	CancelUserOrder(ctx context.Context, userID, orderID uuid.UUID) error
	AppealUserOrder(ctx context.Context, userID, orderID uuid.UUID, reason string) error
	MerchantUpdateStatus(ctx context.Context, merchantID, orderID uuid.UUID, status domain.OrderStatus) error
	MerchantCancelOrder(ctx context.Context, merchantID, orderID uuid.UUID) error
	GetOrder(ctx context.Context, id uuid.UUID) (*OrderResponse, error)
}

type OrderRepository interface {
	Begin(ctx context.Context) (*sqlx.Tx, error)
	CreateOrderPaymentTX(ctx context.Context, tx *sqlx.Tx, p *domain.OrderPayment) error
	CreateOrderTX(ctx context.Context, tx *sqlx.Tx, o *domain.Order) error
	CreateOrderItemTX(ctx context.Context, tx *sqlx.Tx, item *domain.OrderItem) error
	GetOrderByID(ctx context.Context, id uuid.UUID) (*domain.Order, error)
	UpdateOrderStatus(ctx context.Context, id uuid.UUID, status domain.OrderStatus) error
	UpdateOrderStatusTX(ctx context.Context, tx *sqlx.Tx, id uuid.UUID, status domain.OrderStatus) error
	CreateAppeal(ctx context.Context, appeal *domain.CancellationAppeal) error
	UpdateOrderAppealTX(ctx context.Context, tx *sqlx.Tx, orderID uuid.UUID, isAppealed bool) error
	GetOrderItems(ctx context.Context, orderID uuid.UUID) ([]domain.OrderItem, error)
}

type OrderPaymentResponse struct {
	PaymentID uuid.UUID       `json:"payment_id"`
	Amount    decimal.Decimal `json:"amount"`
	Orders    []uuid.UUID     `json:"order_ids"`
}

type OrderService struct {
	orderRepo   OrderRepository
	cartRepo    cart.CartRepository
	productRepo product.ProductRepository
	walletRepo  wallet.WalletRepository
	userRepo    user.UserRepository
}

func NewOrderService(orderRepo OrderRepository, cartRepo cart.CartRepository, productRepo product.ProductRepository, walletRepo wallet.WalletRepository, userRepo user.UserRepository) *OrderService {
	return &OrderService{
		orderRepo:   orderRepo,
		cartRepo:    cartRepo,
		productRepo: productRepo,
		walletRepo:  walletRepo,
		userRepo:    userRepo,
	}
}

func (s *OrderService) CreateUserCheckout(ctx context.Context, userID uuid.UUID, req CheckoutRequest) (*OrderPaymentResponse, error) {
	tx, err := s.orderRepo.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	// 1. Get Cart Items
	cartItems, err := s.cartRepo.GetCartByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}
	if len(cartItems) == 0 {
		return nil, fmt.Errorf("cart is empty")
	}

	// 2. Lock and Validate Stock, Calculate Total
	totalAmount := decimal.Zero
	merchantItems := make(map[uuid.UUID][]domain.CartItem)

	for _, ci := range cartItems {
		p, err := s.productRepo.GetByIDForUpdateTX(ctx, tx, ci.ProductID)
		if err != nil {
			return nil, err
		}
		if p == nil {
			return nil, domain.ErrProductNotFound
		}
		if p.Stock < ci.Quantity {
			return nil, domain.ErrInsufficientStock
		}

		itemTotal := p.Price.Mul(decimal.NewFromInt(int64(ci.Quantity)))
		totalAmount = totalAmount.Add(itemTotal)

		merchantItems[p.StoreID] = append(merchantItems[p.StoreID], ci)
	}

	// 3. Deduct Wallet (Mock External Payment handled by just successful deduction for now)
	w, err := s.walletRepo.GetWalletByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}
	if w == nil {
		return nil, domain.ErrWalletNotFound
	}

	walletTxData := domain.WalletTransaction{
		ID:          uuid.New(),
		WalletID:    w.ID,
		Amount:      totalAmount,
		Direction:   domain.TransactionDirectionOut,
		Type:        domain.TransactionTypePayment,
		Status:      domain.TransactionStatusSuccess,
		ReferenceID: "PAY-" + uuid.New().String()[:8],
		Description: "Order Payment",
		CreatedAt:   time.Now(),
	}

	if err := s.walletRepo.DeductBalanceTX(ctx, tx, w.ID, totalAmount, walletTxData); err != nil {
		return nil, err
	}

	// 4. Resolve Shipping Address snapshot
	var addrRecipientName, addrPhone, addrStreet, addrCity, addrProvince, addrPostal string

	if req.AddressID != nil {
		addr, err := s.userRepo.GetAddressByID(ctx, *req.AddressID)
		if err != nil {
			return nil, err
		}
		if addr == nil || addr.UserID != userID {
			return nil, fmt.Errorf("invalid address id")
		}
		addrRecipientName = addr.RecipientName
		addrPhone = addr.PhoneNumber
		addrStreet = addr.StreetAddress
		addrCity = addr.City
		addrProvince = addr.Province
		addrPostal = addr.PostalCode
	} else if req.ShippingRecipientName != "" {
		// Mandatory fields for custom address
		if req.ShippingPhoneNumber == "" || req.ShippingStreetAddress == "" || req.ShippingCity == "" || req.ShippingProvince == "" || req.ShippingPostalCode == "" {
			return nil, fmt.Errorf("incomplete custom shipping address")
		}
		addrRecipientName = req.ShippingRecipientName
		addrPhone = req.ShippingPhoneNumber
		addrStreet = req.ShippingStreetAddress
		addrCity = req.ShippingCity
		addrProvince = req.ShippingProvince
		addrPostal = req.ShippingPostalCode
	} else {
		// Try to get default address
		addresses, err := s.userRepo.GetAddressesByUserID(ctx, userID)
		if err != nil {
			return nil, err
		}
		for _, a := range addresses {
			if a.IsDefault {
				addrRecipientName = a.RecipientName
				addrPhone = a.PhoneNumber
				addrStreet = a.StreetAddress
				addrCity = a.City
				addrProvince = a.Province
				addrPostal = a.PostalCode
				break
			}
		}
		if addrRecipientName == "" {
			return nil, fmt.Errorf("shipping address is required: please provide address_id or full shipping details")
		}
	}

	// 5. Create OrderPayment
	payment := &domain.OrderPayment{
		ID:            uuid.New(),
		UserID:        userID,
		Amount:        totalAmount,
		PaymentMethod: req.PaymentMethod,
		Status:        "success",
		CreatedAt:     time.Now(),
	}
	if err := s.orderRepo.CreateOrderPaymentTX(ctx, tx, payment); err != nil {
		return nil, err
	}

	// 5. Create Orders and Items
	orderIDs := []uuid.UUID{}
	for merchantID, items := range merchantItems {
		orderID := uuid.New()
		orderAmount := decimal.Zero
		for _, item := range items {
			orderAmount = orderAmount.Add(item.Product.Price.Mul(decimal.NewFromInt(int64(item.Quantity))))
		}

		order := &domain.Order{
			ID:                    orderID,
			PaymentID:             payment.ID,
			MerchantID:            merchantID,
			UserID:                userID,
			Status:                domain.OrderStatusProcessing, // Direct to Processing since payment is success
			TotalAmount:           orderAmount,
			ShippingRecipientName: addrRecipientName,
			ShippingPhoneNumber:   addrPhone,
			ShippingStreetAddress: addrStreet,
			ShippingCity:          addrCity,
			ShippingProvince:      addrProvince,
			ShippingPostalCode:    addrPostal,
			IsAppealed:            false,
			CreatedAt:             time.Now(),
			UpdatedAt:             time.Now(),
		}
		if err := s.orderRepo.CreateOrderTX(ctx, tx, order); err != nil {
			return nil, err
		}

		for _, ci := range items {
			orderItem := &domain.OrderItem{
				ID:        uuid.New(),
				OrderID:   orderID,
				ProductID: ci.ProductID,
				Quantity:  ci.Quantity,
				Price:     ci.Product.Price,
				CreatedAt: time.Now(),
			}
			if err := s.orderRepo.CreateOrderItemTX(ctx, tx, orderItem); err != nil {
				return nil, err
			}

			// Update Product Stock
			newStock := ci.Product.Stock - ci.Quantity
			if err := s.productRepo.UpdateStockTX(ctx, tx, ci.ProductID, newStock); err != nil {
				return nil, err
			}
		}
		orderIDs = append(orderIDs, orderID)
	}

	// 6. Clear Cart
	if err := s.cartRepo.ClearCartTX(ctx, tx, userID); err != nil {
		return nil, err
	}

	if err := tx.Commit(); err != nil {
		return nil, err
	}

	return &OrderPaymentResponse{
		PaymentID: payment.ID,
		Amount:    totalAmount,
		Orders:    orderIDs,
	}, nil
}

func (s *OrderService) CancelUserOrder(ctx context.Context, userID, orderID uuid.UUID) error {
	o, err := s.orderRepo.GetOrderByID(ctx, orderID)
	if err != nil {
		return err
	}
	if o == nil {
		return domain.ErrOrderNotFound
	}

	if o.UserID != userID {
		return fmt.Errorf("unauthorized")
	}

	// Business Rule: Permitted ONLY if status is Processing or Packaging AND time_elapsed < 1 hour
	if (o.Status != domain.OrderStatusProcessing && o.Status != domain.OrderStatusPackaging) || time.Since(o.CreatedAt) > time.Hour {
		return domain.ErrOrderNotCancellable
	}

	// Start Refund Transaction
	tx, err := s.orderRepo.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// Update Order Status
	if err := s.orderRepo.UpdateOrderStatusTX(ctx, tx, orderID, domain.OrderStatusCancelled); err != nil {
		return err
	}

	// Refund to Wallet
	w, err := s.walletRepo.GetWalletByUserID(ctx, userID)
	if err != nil {
		return err
	}
	refundTxData := domain.WalletTransaction{
		ID:          uuid.New(),
		WalletID:    w.ID,
		Amount:      o.TotalAmount,
		Direction:   domain.TransactionDirectionIn,
		Type:        domain.TransactionTypeRefund,
		Status:      domain.TransactionStatusSuccess,
		ReferenceID: "REF-" + uuid.New().String()[:8],
		Description: fmt.Sprintf("Refund for Order %s", orderID),
		CreatedAt:   time.Now(),
	}
	if err := s.walletRepo.AddBalanceTX(ctx, tx, w.ID, o.TotalAmount, refundTxData); err != nil {
		return err
	}

	// Return Stock
	items, err := s.orderRepo.GetOrderItems(ctx, orderID)
	if err != nil {
		return err
	}
	for _, item := range items {
		p, err := s.productRepo.GetByIDForUpdateTX(ctx, tx, item.ProductID)
		if err != nil {
			return err
		}
		if err := s.productRepo.UpdateStockTX(ctx, tx, item.ProductID, p.Stock+item.Quantity); err != nil {
			return err
		}
	}

	return tx.Commit()
}

func (s *OrderService) AppealUserOrder(ctx context.Context, userID, orderID uuid.UUID, reason string) error {
	o, err := s.orderRepo.GetOrderByID(ctx, orderID)
	if err != nil {
		return err
	}
	if o == nil {
		return domain.ErrOrderNotFound
	}
	if o.UserID != userID {
		return fmt.Errorf("unauthorized")
	}

	// Business Rule: Permitted ONLY if status is Packaging AND time_elapsed > 1 hour
	if o.Status != domain.OrderStatusPackaging || time.Since(o.CreatedAt) <= time.Hour {
		return domain.ErrOrderNotCancellable
	}

	tx, err := s.orderRepo.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	appeal := &domain.CancellationAppeal{
		ID:        uuid.New(),
		OrderID:   orderID,
		Reason:    reason,
		Status:    "pending",
		CreatedAt: time.Now(),
	}
	if err := s.orderRepo.CreateAppeal(ctx, appeal); err != nil {
		return err
	}

	if err := s.orderRepo.UpdateOrderAppealTX(ctx, tx, orderID, true); err != nil {
		return err
	}

	return tx.Commit()
}

func (s *OrderService) MerchantUpdateStatus(ctx context.Context, merchantID, orderID uuid.UUID, status domain.OrderStatus) error {
	o, err := s.orderRepo.GetOrderByID(ctx, orderID)
	if err != nil {
		return err
	}
	if o == nil {
		return domain.ErrOrderNotFound
	}
	if o.MerchantID != merchantID {
		return fmt.Errorf("unauthorized")
	}

	// Merchant status restrictions
	switch status {
	case domain.OrderStatusPackaging:
		if o.Status != domain.OrderStatusProcessing {
			return domain.ErrInvalidStatusTransition
		}
	case domain.OrderStatusShipping:
		// Rule: NOT Permitted if time_elapsed < 1 hour since creation/processing
		if time.Since(o.CreatedAt) < time.Hour {
			return domain.ErrMerchantShipmentTooEarly
		}
		if o.Status != domain.OrderStatusPackaging {
			return domain.ErrInvalidStatusTransition
		}
	case domain.OrderStatusDelivered:
		if o.Status != domain.OrderStatusShipping {
			return domain.ErrInvalidStatusTransition
		}
	default:
		return domain.ErrInvalidStatusTransition
	}

	return s.orderRepo.UpdateOrderStatus(ctx, orderID, status)
}

func (s *OrderService) MerchantCancelOrder(ctx context.Context, merchantID, orderID uuid.UUID) error {
	o, err := s.orderRepo.GetOrderByID(ctx, orderID)
	if err != nil {
		return err
	}
	if o == nil {
		return domain.ErrOrderNotFound
	}
	if o.MerchantID != merchantID {
		return fmt.Errorf("unauthorized")
	}

	// Merchant Cancellation: Permitted anytime before Shipping
	if o.Status == domain.OrderStatusShipping || o.Status == domain.OrderStatusDelivered || o.Status == domain.OrderStatusCancelled {
		return domain.ErrOrderNotCancellable
	}

	tx, err := s.orderRepo.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	if err := s.orderRepo.UpdateOrderStatusTX(ctx, tx, orderID, domain.OrderStatusCancelled); err != nil {
		return err
	}

	// Refund to Wallet
	w, err := s.walletRepo.GetWalletByUserID(ctx, o.UserID)
	if err != nil {
		return err
	}
	refundTxData := domain.WalletTransaction{
		ID:          uuid.New(),
		WalletID:    w.ID,
		Amount:      o.TotalAmount,
		Direction:   domain.TransactionDirectionIn,
		Type:        domain.TransactionTypeRefund,
		Status:      domain.TransactionStatusSuccess,
		ReferenceID: "REF-M-" + uuid.New().String()[:8],
		Description: fmt.Sprintf("Merchant Refund for Order %s", orderID),
		CreatedAt:   time.Now(),
	}
	if err := s.walletRepo.AddBalanceTX(ctx, tx, w.ID, o.TotalAmount, refundTxData); err != nil {
		return err
	}

	// Return Stock
	items, err := s.orderRepo.GetOrderItems(ctx, orderID)
	if err != nil {
		return err
	}
	for _, item := range items {
		p, err := s.productRepo.GetByIDForUpdateTX(ctx, tx, item.ProductID)
		if err != nil {
			return err
		}
		if err := s.productRepo.UpdateStockTX(ctx, tx, item.ProductID, p.Stock+item.Quantity); err != nil {
			return err
		}
	}

	return tx.Commit()
}

func (s *OrderService) GetOrder(ctx context.Context, id uuid.UUID) (*OrderResponse, error) {
	o, err := s.orderRepo.GetOrderByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if o == nil {
		return nil, domain.ErrOrderNotFound
	}

	items, err := s.orderRepo.GetOrderItems(ctx, id)
	if err != nil {
		return nil, err
	}

	itemResponses := []OrderItemResponse{}
	for _, item := range items {
		itemResponses = append(itemResponses, OrderItemResponse{
			ID:        item.ID,
			ProductID: item.ProductID,
			Quantity:  item.Quantity,
			Price:     item.Price,
		})
	}

	return &OrderResponse{
		ID:                    o.ID,
		PaymentID:             o.PaymentID,
		MerchantID:            o.MerchantID,
		Status:                o.Status,
		TotalAmount:           o.TotalAmount,
		ShippingRecipientName: o.ShippingRecipientName,
		ShippingPhoneNumber:   o.ShippingPhoneNumber,
		ShippingStreetAddress: o.ShippingStreetAddress,
		ShippingCity:          o.ShippingCity,
		ShippingProvince:      o.ShippingProvince,
		ShippingPostalCode:    o.ShippingPostalCode,
		IsAppealed:            o.IsAppealed,
		Items:                 itemResponses,
		CreatedAt:             o.CreatedAt,
		UpdatedAt:             o.UpdatedAt,
	}, nil
}
