package user

import (
	"context"

	"go-shop-yourself/internal/domain"

	"time"

	"github.com/google/uuid"
)

type UserServiceInterface interface {
	GetUserByID(ctx context.Context, id uuid.UUID) (*UserResponse, error)
	// Addresses
	AddAddress(ctx context.Context, userID uuid.UUID, req *AddressRequest) (*AddressResponse, error)
	ListAddresses(ctx context.Context, userID uuid.UUID) ([]AddressResponse, error)
	UpdateAddress(ctx context.Context, userID, addressID uuid.UUID, req *AddressRequest) (*AddressResponse, error)
	DeleteAddress(ctx context.Context, userID, addressID uuid.UUID) error
}

type UserRepository interface {
	CreateUser(ctx context.Context, u *domain.User) error
	GetUserByEmail(ctx context.Context, email string) (*domain.User, error)
	GetUserByID(ctx context.Context, id uuid.UUID) (*domain.User, error)
	GetUserByProviderID(ctx context.Context, provider string, providerID string) (*domain.User, error)

	// Addresses
	CreateAddress(ctx context.Context, addr *domain.UserAddress) error
	GetAddressesByUserID(ctx context.Context, userID uuid.UUID) ([]domain.UserAddress, error)
	GetAddressByID(ctx context.Context, addressID uuid.UUID) (*domain.UserAddress, error)
	UpdateAddress(ctx context.Context, addr *domain.UserAddress) error
	DeleteAddress(ctx context.Context, addressID uuid.UUID) error
	UnsetDefaultAddresses(ctx context.Context, userID uuid.UUID) error
}

type UserService struct {
	userRepo UserRepository
}

func NewUserService(userRepo UserRepository) *UserService {
	return &UserService{userRepo: userRepo}
}

func (s *UserService) GetUserByID(ctx context.Context, id uuid.UUID) (*UserResponse, error) {
	user, err := s.userRepo.GetUserByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, domain.ErrUserNotFound
	}

	return &UserResponse{
		ID:        user.ID,
		FullName:  user.FullName,
		Username:  user.Username,
		Email:     user.Email,
		CreatedAt: user.CreatedAt,
	}, nil
}

func (s *UserService) AddAddress(ctx context.Context, userID uuid.UUID, req *AddressRequest) (*AddressResponse, error) {
	if req.IsDefault {
		_ = s.userRepo.UnsetDefaultAddresses(ctx, userID)
	}

	addr := &domain.UserAddress{
		ID:            uuid.New(),
		UserID:        userID,
		Tag:           req.Tag,
		RecipientName: req.RecipientName,
		PhoneNumber:   req.PhoneNumber,
		StreetAddress: req.StreetAddress,
		City:          req.City,
		Province:      req.Province,
		PostalCode:    req.PostalCode,
		IsDefault:     req.IsDefault,
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
	}

	if err := s.userRepo.CreateAddress(ctx, addr); err != nil {
		return nil, err
	}

	return &AddressResponse{
		ID:            addr.ID,
		Tag:           addr.Tag,
		RecipientName: addr.RecipientName,
		PhoneNumber:   addr.PhoneNumber,
		StreetAddress: addr.StreetAddress,
		City:          addr.City,
		Province:      addr.Province,
		PostalCode:    addr.PostalCode,
		IsDefault:     addr.IsDefault,
		CreatedAt:     addr.CreatedAt,
	}, nil
}

func (s *UserService) ListAddresses(ctx context.Context, userID uuid.UUID) ([]AddressResponse, error) {
	addresses, err := s.userRepo.GetAddressesByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}

	res := make([]AddressResponse, 0, len(addresses))
	for _, a := range addresses {
		res = append(res, AddressResponse{
			ID:            a.ID,
			Tag:           a.Tag,
			RecipientName: a.RecipientName,
			PhoneNumber:   a.PhoneNumber,
			StreetAddress: a.StreetAddress,
			City:          a.City,
			Province:      a.Province,
			PostalCode:    a.PostalCode,
			IsDefault:     a.IsDefault,
			CreatedAt:     a.CreatedAt,
		})
	}
	return res, nil
}

func (s *UserService) UpdateAddress(ctx context.Context, userID, addressID uuid.UUID, req *AddressRequest) (*AddressResponse, error) {
	addr, err := s.userRepo.GetAddressByID(ctx, addressID)
	if err != nil {
		return nil, err
	}
	if addr == nil || addr.UserID != userID {
		return nil, domain.ErrForbidden // Or some generic error
	}

	if req.IsDefault && !addr.IsDefault {
		_ = s.userRepo.UnsetDefaultAddresses(ctx, userID)
	}

	addr.Tag = req.Tag
	addr.RecipientName = req.RecipientName
	addr.PhoneNumber = req.PhoneNumber
	addr.StreetAddress = req.StreetAddress
	addr.City = req.City
	addr.Province = req.Province
	addr.PostalCode = req.PostalCode
	addr.IsDefault = req.IsDefault
	addr.UpdatedAt = time.Now()

	if err := s.userRepo.UpdateAddress(ctx, addr); err != nil {
		return nil, err
	}

	return &AddressResponse{
		ID:            addr.ID,
		Tag:           addr.Tag,
		RecipientName: addr.RecipientName,
		PhoneNumber:   addr.PhoneNumber,
		StreetAddress: addr.StreetAddress,
		City:          addr.City,
		Province:      addr.Province,
		PostalCode:    addr.PostalCode,
		IsDefault:     addr.IsDefault,
		CreatedAt:     addr.CreatedAt,
	}, nil
}

func (s *UserService) DeleteAddress(ctx context.Context, userID, addressID uuid.UUID) error {
	addr, err := s.userRepo.GetAddressByID(ctx, addressID)
	if err != nil {
		return err
	}
	if addr == nil || addr.UserID != userID {
		return domain.ErrForbidden
	}

	return s.userRepo.DeleteAddress(ctx, addressID)
}
