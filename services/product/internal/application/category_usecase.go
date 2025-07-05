package application

import (
	"context"
	"fmt"

	"product/internal/domain/entity"
	"product/internal/domain/repository"

	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
)

// CategoryUsecase handles category business logic
type CategoryUsecase struct {
	categoryRepo repository.CategoryRepository
	logger       *logrus.Logger
}

// NewCategoryUsecase creates a new category usecase
func NewCategoryUsecase(categoryRepo repository.CategoryRepository, logger *logrus.Logger) *CategoryUsecase {
	return &CategoryUsecase{
		categoryRepo: categoryRepo,
		logger:       logger,
	}
}

// CreateCategoryRequest represents the request to create a category
type CreateCategoryRequest struct {
	Name        string     `json:"name" validate:"required"`
	Description string     `json:"description"`
	ParentID    *uuid.UUID `json:"parent_id"`
	SortOrder   int        `json:"sort_order"`
}

// UpdateCategoryRequest represents the request to update a category
type UpdateCategoryRequest struct {
	Name        *string    `json:"name"`
	Description *string    `json:"description"`
	ParentID    *uuid.UUID `json:"parent_id"`
	SortOrder   *int       `json:"sort_order"`
	IsActive    *bool      `json:"is_active"`
}

// CreateCategory creates a new category
func (uc *CategoryUsecase) CreateCategory(ctx context.Context, req *CreateCategoryRequest) (*entity.Category, error) {
	// Validate request
	if req.Name == "" {
		return nil, fmt.Errorf("category name is required")
	}

	// Validate parent category exists if provided
	if req.ParentID != nil {
		parent, err := uc.categoryRepo.GetByID(ctx, *req.ParentID)
		if err != nil {
			return nil, fmt.Errorf("failed to validate parent category: %w", err)
		}
		if parent == nil {
			return nil, fmt.Errorf("parent category not found")
		}
	}

	// Create new category
	category, err := entity.NewCategory(req.Name)
	if err != nil {
		return nil, fmt.Errorf("failed to create category: %w", err)
	}

	// Set optional fields
	category.Description = req.Description
	category.ParentID = req.ParentID
	category.SortOrder = req.SortOrder

	// Save to database
	if err := uc.categoryRepo.Create(ctx, category); err != nil {
		return nil, fmt.Errorf("failed to save category: %w", err)
	}

	uc.logger.WithField("category_id", category.ID).Info("Category created successfully")
	return category, nil
}

// GetCategory retrieves a category by ID
func (uc *CategoryUsecase) GetCategory(ctx context.Context, id uuid.UUID) (*entity.Category, error) {
	category, err := uc.categoryRepo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get category: %w", err)
	}

	if category == nil {
		return nil, fmt.Errorf("category not found")
	}

	return category, nil
}

// UpdateCategory updates an existing category
func (uc *CategoryUsecase) UpdateCategory(ctx context.Context, id uuid.UUID, req *UpdateCategoryRequest) (*entity.Category, error) {
	// Get existing category
	category, err := uc.categoryRepo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get category: %w", err)
	}

	if category == nil {
		return nil, fmt.Errorf("category not found")
	}

	// Check if category can be modified
	if !category.CanBeModified() {
		return nil, fmt.Errorf("category cannot be modified due to master data protection")
	}

	// Validate parent category if provided
	if req.ParentID != nil {
		if *req.ParentID == id {
			return nil, fmt.Errorf("category cannot be its own parent")
		}

		parent, err := uc.categoryRepo.GetByID(ctx, *req.ParentID)
		if err != nil {
			return nil, fmt.Errorf("failed to validate parent category: %w", err)
		}
		if parent == nil {
			return nil, fmt.Errorf("parent category not found")
		}
	}

	// Update fields if provided
	if req.Name != nil {
		if err := category.UpdateName(*req.Name); err != nil {
			return nil, fmt.Errorf("failed to update name: %w", err)
		}
	}

	if req.Description != nil {
		category.SetDescription(*req.Description)
	}

	if req.ParentID != nil {
		if *req.ParentID == uuid.Nil {
			category.RemoveParent()
		} else {
			category.SetParent(*req.ParentID)
		}
	}

	if req.SortOrder != nil {
		category.SetSortOrder(*req.SortOrder)
	}

	if req.IsActive != nil {
		if *req.IsActive {
			category.Activate()
		} else {
			category.Deactivate()
		}
	}

	// Save changes
	if err := uc.categoryRepo.Update(ctx, category); err != nil {
		return nil, fmt.Errorf("failed to update category: %w", err)
	}

	uc.logger.WithField("category_id", category.ID).Info("Category updated successfully")
	return category, nil
}

// DeleteCategory deletes a category
func (uc *CategoryUsecase) DeleteCategory(ctx context.Context, id uuid.UUID) error {
	// Get existing category
	category, err := uc.categoryRepo.GetByID(ctx, id)
	if err != nil {
		return fmt.Errorf("failed to get category: %w", err)
	}

	if category == nil {
		return fmt.Errorf("category not found")
	}

	// Check if category can be modified
	if !category.CanBeModified() {
		return fmt.Errorf("category cannot be deleted due to master data protection")
	}

	// Check if category has children
	children, err := uc.categoryRepo.GetChildren(ctx, id)
	if err != nil {
		return fmt.Errorf("failed to check child categories: %w", err)
	}
	if len(children) > 0 {
		return fmt.Errorf("cannot delete category with child categories")
	}

	// Delete category
	if err := uc.categoryRepo.Delete(ctx, id); err != nil {
		return fmt.Errorf("failed to delete category: %w", err)
	}

	uc.logger.WithField("category_id", id).Info("Category deleted successfully")
	return nil
}

// ListCategories lists categories with filtering
func (uc *CategoryUsecase) ListCategories(ctx context.Context, filter repository.CategoryFilter) ([]*entity.Category, error) {
	categories, err := uc.categoryRepo.List(ctx, filter)
	if err != nil {
		return nil, fmt.Errorf("failed to list categories: %w", err)
	}

	return categories, nil
}

// GetRootCategories gets all root categories (no parent)
func (uc *CategoryUsecase) GetRootCategories(ctx context.Context) ([]*entity.Category, error) {
	categories, err := uc.categoryRepo.GetRoot(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get root categories: %w", err)
	}

	return categories, nil
}

// GetChildCategories gets child categories of a parent
func (uc *CategoryUsecase) GetChildCategories(ctx context.Context, parentID uuid.UUID) ([]*entity.Category, error) {
	categories, err := uc.categoryRepo.GetChildren(ctx, parentID)
	if err != nil {
		return nil, fmt.Errorf("failed to get child categories: %w", err)
	}

	return categories, nil
}

// SetManualOverride sets manual override for a category
func (uc *CategoryUsecase) SetManualOverride(ctx context.Context, id uuid.UUID, override bool) error {
	// Get existing category
	category, err := uc.categoryRepo.GetByID(ctx, id)
	if err != nil {
		return fmt.Errorf("failed to get category: %w", err)
	}

	if category == nil {
		return fmt.Errorf("category not found")
	}

	// Set manual override
	category.SetManualOverride(override)

	// Save changes
	if err := uc.categoryRepo.Update(ctx, category); err != nil {
		return fmt.Errorf("failed to update category: %w", err)
	}

	uc.logger.WithFields(logrus.Fields{
		"category_id": id,
		"override":    override,
	}).Info("Category manual override updated")

	return nil
}
