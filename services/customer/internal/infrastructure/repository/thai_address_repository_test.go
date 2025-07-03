package repository

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/saan-system/services/customer/internal/domain"
)

func TestThaiAddressRepository_GetAddressSuggestions(t *testing.T) {
	db := setupTestDB(t)
	repo := NewThaiAddressRepository(db)

	// Test data would be loaded here in real implementation
	// For now, just test the interface
	ctx := context.Background()
	
	suggestions, err := repo.GetAddressSuggestions(ctx, "หัวหมาก", 10)
	require.NoError(t, err)
	assert.IsType(t, []domain.AddressSuggestion{}, suggestions)
}
