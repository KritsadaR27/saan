package application

import (
	"context"
	"fmt"

	"go.uber.org/zap"

	"github.com/saan-system/services/customer/internal/domain"
)

// ThaiAddressService implements domain.ThaiAddressService
type ThaiAddressService struct {
	thaiAddressRepo domain.ThaiAddressRepository
	logger          *zap.Logger
}

// NewThaiAddressService creates a new Thai address service
func NewThaiAddressService(
	thaiAddressRepo domain.ThaiAddressRepository,
	logger *zap.Logger,
) domain.ThaiAddressService {
	return &ThaiAddressService{
		thaiAddressRepo: thaiAddressRepo,
		logger:          logger,
	}
}

// GetAddressSuggestions returns address suggestions based on query
func (s *ThaiAddressService) GetAddressSuggestions(ctx context.Context, query string, limit int) ([]domain.AddressSuggestion, error) {
	s.logger.Debug("Getting address suggestions", 
		zap.String("query", query), 
		zap.Int("limit", limit))

	if query == "" {
		return []domain.AddressSuggestion{}, nil
	}

	suggestions, err := s.thaiAddressRepo.GetAddressSuggestions(ctx, query, limit)
	if err != nil {
		s.logger.Error("Failed to get address suggestions", 
			zap.String("query", query), 
			zap.Error(err))
		return nil, fmt.Errorf("failed to get address suggestions: %w", err)
	}

	s.logger.Debug("Successfully retrieved address suggestions", 
		zap.String("query", query), 
		zap.Int("count", len(suggestions)))

	return suggestions, nil
}

// GetBySubdistrict returns Thai addresses by subdistrict
func (s *ThaiAddressService) GetBySubdistrict(ctx context.Context, subdistrict string) ([]domain.ThaiAddress, error) {
	s.logger.Debug("Getting addresses by subdistrict", zap.String("subdistrict", subdistrict))

	addresses, err := s.thaiAddressRepo.GetBySubdistrict(ctx, subdistrict)
	if err != nil {
		s.logger.Error("Failed to get addresses by subdistrict", 
			zap.String("subdistrict", subdistrict), 
			zap.Error(err))
		return nil, fmt.Errorf("failed to get addresses by subdistrict: %w", err)
	}

	return addresses, nil
}

// GetProvinceDeliveryInfo returns delivery route info for a province
func (s *ThaiAddressService) GetProvinceDeliveryInfo(ctx context.Context, province string) (*domain.DeliveryRoute, error) {
	s.logger.Debug("Getting province delivery info", zap.String("province", province))

	route, err := s.thaiAddressRepo.GetProvinceDeliveryInfo(ctx, province)
	if err != nil {
		s.logger.Error("Failed to get province delivery info", 
			zap.String("province", province), 
			zap.Error(err))
		return nil, fmt.Errorf("failed to get province delivery info: %w", err)
	}

	return route, nil
}
