package user

import (
	"context"

	"go-marketplace/internal/domain"

	"time"

	"github.com/google/uuid"
)

type UserService interface {
	GetUserByID(ctx context.Context, id uuid.UUID) (*UserResponse, error)
	// Addresses
	AddAddress(ctx context.Context, userID uuid.UUID, req *AddressRequest) (*AddressResponse, error)
	ListAddresses(ctx context.Context, userID uuid.UUID) ([]AddressResponse, error)
	UpdateAddress(ctx context.Context, userID, addressID uuid.UUID, req *AddressRequest) (*AddressResponse, error)
	DeleteAddress(ctx context.Context, userID, addressID uuid.UUID) error
}

type userService struct {
	userRepo UserRepository
}

func NewUserService(userRepo UserRepository) UserService {
	return &userService{userRepo: userRepo}
}

func (s *userService) GetUserByID(ctx context.Context, id uuid.UUID) (*UserResponse, error) {
	user, err := s.userRepo.GetUserByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, domain.ErrUserNotFound
	}

	return &UserResponse{
		ID:           user.ID,
		FullName:     user.FullName,
		Username:     user.Username,
		Email:        user.Email,
		AuthProvider: user.AuthProvider,
		ProviderID:   user.ProviderID,
		CreatedAt:    user.CreatedAt,
	}, nil
}

func (s *userService) AddAddress(ctx context.Context, userID uuid.UUID, req *AddressRequest) (*AddressResponse, error) {
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

func (s *userService) ListAddresses(ctx context.Context, userID uuid.UUID) ([]AddressResponse, error) {
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

func (s *userService) UpdateAddress(ctx context.Context, userID, addressID uuid.UUID, req *AddressRequest) (*AddressResponse, error) {
	m, err := s.userRepo.GetAddressByID(ctx, addressID)
	if err != nil {
		return nil, err
	}
	if m == nil {
		return nil, domain.ErrForbidden // Or AddressNotFound
	}

	addr := NewAddress(m)
	if !addr.IsOwnedBy(userID) {
		return nil, domain.ErrForbidden
	}

	if req.IsDefault && !m.IsDefault {
		_ = s.userRepo.UnsetDefaultAddresses(ctx, userID)
	}

	addr.Update(req)

	if err := s.userRepo.UpdateAddress(ctx, m); err != nil {
		return nil, err
	}

	return &AddressResponse{
		ID:            m.ID,
		Tag:           m.Tag,
		RecipientName: m.RecipientName,
		PhoneNumber:   m.PhoneNumber,
		StreetAddress: m.StreetAddress,
		City:          m.City,
		Province:      m.Province,
		PostalCode:    m.PostalCode,
		IsDefault:     m.IsDefault,
		CreatedAt:     m.CreatedAt,
	}, nil
}

func (s *userService) DeleteAddress(ctx context.Context, userID, addressID uuid.UUID) error {
	m, err := s.userRepo.GetAddressByID(ctx, addressID)
	if err != nil {
		return err
	}
	if m == nil {
		return domain.ErrForbidden
	}

	addr := NewAddress(m)
	if !addr.IsOwnedBy(userID) {
		return domain.ErrForbidden
	}

	return s.userRepo.DeleteAddress(ctx, addressID)
}
