package order

import (
	"context"
	"fmt"
	"time"

	"go-marketplace/internal/core/cart"
	"go-marketplace/internal/core/merchant"
	"go-marketplace/internal/core/payment"
	"go-marketplace/internal/core/product"
	"go-marketplace/internal/core/user"
	"go-marketplace/internal/core/wallet"
	"go-marketplace/internal/domain"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/shopspring/decimal"
)

type OrderService interface {
	CreateUserCheckout(ctx context.Context, userID uuid.UUID, req CheckoutRequest) (*OrderPaymentResponse, error)
	CancelUserOrder(ctx context.Context, userID, orderID uuid.UUID) error
	AppealUserOrder(ctx context.Context, userID, orderID uuid.UUID, reason string) error
	MerchantUpdateStatus(ctx context.Context, merchantID, orderID uuid.UUID, status domain.OrderStatus) error
	MerchantCancelOrder(ctx context.Context, merchantID, orderID uuid.UUID) error
	GetOrder(ctx context.Context, id uuid.UUID) (*OrderResponse, error)
	HandlePaymentStatusChangeTX(ctx context.Context, tx *sqlx.Tx, paymentID uuid.UUID, status domain.PaymentStatus) error
}

type OrderPaymentResponse struct {
	PaymentID uuid.UUID       `json:"payment_id"`
	Amount    decimal.Decimal `json:"amount"`
	Orders    []uuid.UUID     `json:"order_ids"`
}

type orderService struct {
	orderRepo      OrderRepository
	cartRepo       cart.CartRepository
	productRepo    product.ProductRepository
	walletService  wallet.WalletService
	userRepo       user.UserRepository
	merchantRepo   merchant.MerchantRepository
	paymentService payment.PaymentService
}

func NewOrderService(
	orderRepo OrderRepository,
	cartRepo cart.CartRepository,
	productRepo product.ProductRepository,
	walletService wallet.WalletService,
	userRepo user.UserRepository,
	merchantRepo merchant.MerchantRepository,
	paymentService payment.PaymentService,
) OrderService {
	return &orderService{
		orderRepo:      orderRepo,
		cartRepo:       cartRepo,
		productRepo:    productRepo,
		walletService:  walletService,
		userRepo:       userRepo,
		merchantRepo:   merchantRepo,
		paymentService: paymentService,
	}
}

func (s *orderService) CreateUserCheckout(ctx context.Context, userID uuid.UUID, req CheckoutRequest) (*OrderPaymentResponse, error) {
	ctx, cancel := context.WithTimeout(ctx, 15*time.Second)
	defer cancel()

	tx, err := s.orderRepo.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer func() { _ = tx.Rollback() }()

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

	// 3. Prepare Payment Distributions (for escrow)
	distributions := []payment.PaymentDistribution{}
	for merchantID, items := range merchantItems {
		merchantAmount := decimal.Zero
		for _, item := range items {
			merchantAmount = merchantAmount.Add(item.Product.Price.Mul(decimal.NewFromInt(int64(item.Quantity))))
		}

		m, err := s.merchantRepo.GetByID(ctx, merchantID)
		if err != nil {
			return nil, err
		}
		distributions = append(distributions, payment.PaymentDistribution{RecipientID: m.UserID, Amount: merchantAmount})
	}

	// 4. Call Payment Service
	paymentReq := payment.CreatePaymentRequest{
		UserID:        userID,
		Amount:        totalAmount,
		Type:          domain.PaymentTypeOrder,
		Method:        req.PaymentMethod,
		ReferenceID:   uuid.New(), // Placeholder OrderID group or just a random ID for the payment session
		Distributions: distributions,
	}

	payRes, err := s.paymentService.CreatePaymentTX(ctx, tx, paymentReq)
	if err != nil {
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

	// 6. Create Orders and Items
	orderIDs := []uuid.UUID{}
	orderStatus := domain.OrderStatusPending
	if payRes.Status == domain.PaymentStatusSuccess {
		orderStatus = domain.OrderStatusProcessing
	}

	for merchantID, items := range merchantItems {
		orderID := uuid.New()
		orderAmount := decimal.Zero
		for _, item := range items {
			orderAmount = orderAmount.Add(item.Product.Price.Mul(decimal.NewFromInt(int64(item.Quantity))))
		}

		order := &domain.Order{
			ID:                    orderID,
			PaymentID:             payRes.PaymentID,
			MerchantID:            merchantID,
			UserID:                userID,
			Status:                orderStatus,
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
		PaymentID: payRes.PaymentID,
		Amount:    totalAmount,
		Orders:    orderIDs,
	}, nil
}

func (s *orderService) CancelUserOrder(ctx context.Context, userID, orderID uuid.UUID) error {
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	tx, err := s.orderRepo.Begin(ctx)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback() }()

	// Lock and fetch current state
	o, err := s.orderRepo.GetOrderByIDForUpdateTX(ctx, tx, orderID)
	if err != nil {
		return err
	}
	if o == nil {
		return domain.ErrOrderNotFound
	}

	order := NewOrder(o)
	if !order.IsAuthorized(userID) {
		return fmt.Errorf("unauthorized")
	}

	if !order.CanCancel() {
		return domain.ErrOrderNotCancellable
	}

	// Update Order Status
	if err := s.orderRepo.UpdateOrderStatusTX(ctx, tx, orderID, domain.OrderStatusCancelled); err != nil {
		return err
	}

	// 1. Deduct from Merchant Pending Balance (Escrow)
	m, err := s.merchantRepo.GetByID(ctx, o.MerchantID)
	if err != nil {
		return err
	}
	if m == nil {
		return domain.ErrMerchantNotFound
	}

	refundMerchantTxData := domain.WalletTransaction{
		ID:          uuid.New(),
		Amount:      o.TotalAmount,
		Direction:   domain.TransactionDirectionOut,
		Type:        domain.TransactionTypeRefund,
		Status:      domain.TransactionStatusSuccess,
		ReferenceID: "REF-ESC-" + uuid.New().String()[:8],
		Description: fmt.Sprintf("Escrow Refund (Cancellation) for Order %s", orderID),
		CreatedAt:   time.Now(),
	}
	if err := s.walletService.RefundFromPendingTX(ctx, tx, m.UserID, o.TotalAmount, refundMerchantTxData); err != nil {
		return err
	}

	// 2. Refund to User Wallet
	refundUserTxData := domain.WalletTransaction{
		ID:          uuid.New(),
		Amount:      o.TotalAmount,
		Direction:   domain.TransactionDirectionIn,
		Type:        domain.TransactionTypeRefund,
		Status:      domain.TransactionStatusSuccess,
		ReferenceID: "REF-" + uuid.New().String()[:8],
		Description: fmt.Sprintf("Refund for Order %s", orderID),
		CreatedAt:   time.Now(),
	}
	if err := s.walletService.AddBalanceTX(ctx, tx, userID, o.TotalAmount, refundUserTxData); err != nil {
		return err
	}

	// Return Stock
	items, err := s.orderRepo.GetOrderItems(ctx, orderID)
	if err != nil {
		return err
	}
	if len(items) > 0 {
		if err := s.productRepo.RestoreStockBatchTX(ctx, tx, items); err != nil {
			return err
		}
	}

	return tx.Commit()
}

func (s *orderService) AppealUserOrder(ctx context.Context, userID, orderID uuid.UUID, reason string) error {
	tx, err := s.orderRepo.Begin(ctx)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback() }()

	// Lock and fetch current state
	o, err := s.orderRepo.GetOrderByIDForUpdateTX(ctx, tx, orderID)
	if err != nil {
		return err
	}
	if o == nil {
		return domain.ErrOrderNotFound
	}

	order := NewOrder(o)
	if !order.IsAuthorized(userID) {
		return fmt.Errorf("unauthorized")
	}

	if !order.CanAppeal() {
		return domain.ErrOrderNotCancellable
	}

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

func (s *orderService) MerchantUpdateStatus(ctx context.Context, merchantID, orderID uuid.UUID, status domain.OrderStatus) error {
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	tx, err := s.orderRepo.Begin(ctx)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback() }()

	// Lock and fetch current state
	o, err := s.orderRepo.GetOrderByIDForUpdateTX(ctx, tx, orderID)
	if err != nil {
		return err
	}
	if o == nil {
		return domain.ErrOrderNotFound
	}

	order := NewOrder(o)
	if !order.IsMerchantAuthorized(merchantID) {
		return fmt.Errorf("unauthorized")
	}

	if err := order.ValidateStatusTransition(status); err != nil {
		return err
	}

	// Update Status
	if err := s.orderRepo.UpdateOrderStatusTX(ctx, tx, orderID, status); err != nil {
		return err
	}

	// If status is Delivered, settle the pending balance to merchant
	if status == domain.OrderStatusDelivered {
		m, err := s.merchantRepo.GetByID(ctx, o.MerchantID)
		if err != nil {
			return err
		}

		settleTxData := domain.WalletTransaction{
			ID:          uuid.New(),
			Amount:      o.TotalAmount,
			Direction:   domain.TransactionDirectionIn,
			Type:        domain.TransactionTypePayment,
			Status:      domain.TransactionStatusSuccess,
			ReferenceID: orderID.String(),
			Description: fmt.Sprintf("Settlement for Order %s", orderID),
			CreatedAt:   time.Now(),
		}
		if err := s.walletService.SettlePendingBalanceTX(ctx, tx, m.UserID, o.TotalAmount, settleTxData); err != nil {
			return err
		}
	}

	return tx.Commit()
}

func (s *orderService) MerchantCancelOrder(ctx context.Context, merchantID, orderID uuid.UUID) error {
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	tx, err := s.orderRepo.Begin(ctx)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback() }()

	// Lock and fetch current state
	o, err := s.orderRepo.GetOrderByIDForUpdateTX(ctx, tx, orderID)
	if err != nil {
		return err
	}
	if o == nil {
		return domain.ErrOrderNotFound
	}

	order := NewOrder(o)
	if !order.IsMerchantAuthorized(merchantID) {
		return fmt.Errorf("unauthorized")
	}

	if !order.CanMerchantCancel() {
		return domain.ErrOrderNotCancellable
	}

	if err := s.orderRepo.UpdateOrderStatusTX(ctx, tx, orderID, domain.OrderStatusCancelled); err != nil {
		return err
	}

	// 1. Deduct from Merchant Pending Balance (Escrow)
	m, err := s.merchantRepo.GetByID(ctx, o.MerchantID)
	if err != nil {
		return err
	}
	if m == nil {
		return domain.ErrMerchantNotFound
	}

	refundMerchantTxData := domain.WalletTransaction{
		ID:          uuid.New(),
		Amount:      o.TotalAmount,
		Direction:   domain.TransactionDirectionOut,
		Type:        domain.TransactionTypeRefund,
		Status:      domain.TransactionStatusSuccess,
		ReferenceID: "REF-ESC-M-" + uuid.New().String()[:8],
		Description: fmt.Sprintf("Merchant Escrow Refund for Order %s", orderID),
		CreatedAt:   time.Now(),
	}
	if err := s.walletService.RefundFromPendingTX(ctx, tx, m.UserID, o.TotalAmount, refundMerchantTxData); err != nil {
		return err
	}

	// 2. Refund to User Wallet
	refundUserTxData := domain.WalletTransaction{
		ID:          uuid.New(),
		Amount:      o.TotalAmount,
		Direction:   domain.TransactionDirectionIn,
		Type:        domain.TransactionTypeRefund,
		Status:      domain.TransactionStatusSuccess,
		ReferenceID: "REF-M-" + uuid.New().String()[:8],
		Description: fmt.Sprintf("Merchant Refund for Order %s", orderID),
		CreatedAt:   time.Now(),
	}
	if err := s.walletService.AddBalanceTX(ctx, tx, o.UserID, o.TotalAmount, refundUserTxData); err != nil {
		return err
	}

	// Return Stock
	items, err := s.orderRepo.GetOrderItems(ctx, orderID)
	if err != nil {
		return err
	}
	if len(items) > 0 {
		if err := s.productRepo.RestoreStockBatchTX(ctx, tx, items); err != nil {
			return err
		}
	}

	return tx.Commit()
}

func (s *orderService) GetOrder(ctx context.Context, id uuid.UUID) (*OrderResponse, error) {
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

func (s *orderService) HandlePaymentStatusChangeTX(ctx context.Context, tx *sqlx.Tx, paymentID uuid.UUID, status domain.PaymentStatus) error {
	switch status {
	case domain.PaymentStatusFailed, domain.PaymentStatusExpired:
		orders, err := s.orderRepo.GetOrdersByPaymentIDForUpdateTX(ctx, tx, paymentID)
		if err != nil {
			return err
		}

		var allItems []domain.OrderItem
		for _, o := range orders {
			// Update Order Status
			if err := s.orderRepo.UpdateOrderStatusTX(ctx, tx, o.ID, domain.OrderStatusCancelled); err != nil {
				return err
			}

			// Collect Items
			items, err := s.orderRepo.GetOrderItems(ctx, o.ID)
			if err != nil {
				return err
			}
			allItems = append(allItems, items...)
		}

		if len(allItems) > 0 {
			if err := s.productRepo.RestoreStockBatchTX(ctx, tx, allItems); err != nil {
				return err
			}
		}
	case domain.PaymentStatusSuccess:
		orders, err := s.orderRepo.GetOrdersByPaymentIDForUpdateTX(ctx, tx, paymentID)
		if err != nil {
			return err
		}
		for _, o := range orders {
			if err := s.orderRepo.UpdateOrderStatusTX(ctx, tx, o.ID, domain.OrderStatusProcessing); err != nil {
				return err
			}
		}
	}
	return nil
}
