package service

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/erp-system/scm-service/internal/business/domain"
	"github.com/shopspring/decimal"
)

type ProductManagementService struct {
	repo      domain.ProductRepository
	catRepo   domain.ProductCategoryRepository
	locRepo   domain.LocationRepository
	publisher domain.EventPublisher
}

func NewProductManagementService(repo domain.ProductRepository, catRepo domain.ProductCategoryRepository, locRepo domain.LocationRepository, publisher domain.EventPublisher) *ProductManagementService {
	return &ProductManagementService{
		repo:      repo,
		catRepo:   catRepo,
		locRepo:   locRepo,
		publisher: publisher,
	}
}

func (s *ProductManagementService) ListProducts(ctx context.Context) ([]domain.Product, error) {
	return s.repo.List(ctx)
}

func (s *ProductManagementService) CreateProduct(ctx context.Context, code, name, desc, pType, uom string, cost, price decimal.Decimal, categoryID *string) (*domain.Product, error) {
	id := fmt.Sprintf("prod_%d", time.Now().UnixNano())

	p := &domain.Product{
		ID:              id,
		ProductCode:     code,
		ProductName:     name,
		Description:     desc,
		ProductType:     pType,
		CategoryID:      categoryID,
		UnitOfMeasure:   uom,
		StandardCost:    cost,
		ListPrice:       price,
		IsActive:        true,
		CreatedAt:       time.Now(),
		UpdatedAt:       time.Now(),
	}

	err := s.repo.Create(ctx, p)
	if err != nil {
		return nil, err
	}

	if err := s.publisher.Publish(ctx, domain.TopicScmProductCreated, p.ID, domain.ProductCreatedEvent{
		ProductID:   p.ID,
		ProductCode: p.ProductCode,
		ProductName: p.ProductName,
		ProductType: p.ProductType,
		Timestamp:   time.Now(),
	}); err != nil {
		log.Printf("ERROR: failed to publish event %s: %v", domain.TopicScmProductCreated, err)
	}

	return p, nil
}

func (s *ProductManagementService) GetProduct(ctx context.Context, id string) (*domain.Product, error) {
	return s.repo.GetByID(ctx, id)
}

func (s *ProductManagementService) UpdateProduct(ctx context.Context, id, code, name, desc, pType, uom string, cost, price decimal.Decimal, isActive bool, categoryID *string) (*domain.Product, error) {
	p, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	p.ProductCode = code
	p.ProductName = name
	p.Description = desc
	p.ProductType = pType
	p.CategoryID = categoryID
	p.UnitOfMeasure = uom
	p.StandardCost = cost
	p.ListPrice = price
	p.IsActive = isActive
	p.UpdatedAt = time.Now()

	err = s.repo.Update(ctx, p)
	if err != nil {
		return nil, err
	}

	if err := s.publisher.Publish(ctx, domain.TopicScmProductUpdated, p.ID, domain.ProductUpdatedEvent{
		ProductID:   p.ID,
		ProductCode: p.ProductCode,
		ProductName: p.ProductName,
		IsActive:    p.IsActive,
		Timestamp:   time.Now(),
	}); err != nil {
		log.Printf("ERROR: failed to publish event %s: %v", domain.TopicScmProductUpdated, err)
	}

	return p, nil
}

func (s *ProductManagementService) DeleteProduct(ctx context.Context, id string) error {
	err := s.repo.Delete(ctx, id)
	if err != nil {
		return err
	}

	if err := s.publisher.Publish(ctx, domain.TopicScmProductDiscontinued, id, domain.ProductDiscontinuedEvent{
		ProductID: id,
		Timestamp: time.Now(),
	}); err != nil {
		log.Printf("ERROR: failed to publish event %s: %v", domain.TopicScmProductDiscontinued, err)
	}

	return nil
}

// Product Categories CRUD

func (s *ProductManagementService) ListCategories(ctx context.Context) ([]domain.ProductCategory, error) {
	return s.catRepo.List(ctx)
}

func (s *ProductManagementService) CreateCategory(ctx context.Context, code, name, desc string) (*domain.ProductCategory, error) {
	id := fmt.Sprintf("cat_%d", time.Now().UnixNano())
	pc := &domain.ProductCategory{
		ID:          id,
		Code:        code,
		Name:        name,
		Description: desc,
	}

	err := s.catRepo.Create(ctx, pc)
	if err != nil {
		return nil, err
	}
	return pc, nil
}

func (s *ProductManagementService) GetCategory(ctx context.Context, id string) (*domain.ProductCategory, error) {
	return s.catRepo.GetByID(ctx, id)
}

func (s *ProductManagementService) UpdateCategory(ctx context.Context, id, code, name, desc string) (*domain.ProductCategory, error) {
	pc, err := s.catRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	pc.Code = code
	pc.Name = name
	pc.Description = desc

	err = s.catRepo.Update(ctx, pc)
	if err != nil {
		return nil, err
	}
	return pc, nil
}

func (s *ProductManagementService) DeleteCategory(ctx context.Context, id string) error {
	return s.catRepo.Delete(ctx, id)
}

// Locations CRUD
func (s *ProductManagementService) ListLocations(ctx context.Context) ([]domain.Location, error) {
	return s.locRepo.List(ctx)
}

func (s *ProductManagementService) CreateLocation(ctx context.Context, code, name, locType string) (*domain.Location, error) {
	id := fmt.Sprintf("loc_%d", time.Now().UnixNano())
	loc := &domain.Location{
		ID:           id,
		LocationCode: code,
		LocationName: name,
		LocationType: locType,
		IsActive:     true,
	}
	err := s.locRepo.Create(ctx, loc)
	if err != nil {
		return nil, err
	}
	return loc, nil
}

func (s *ProductManagementService) GetLocation(ctx context.Context, id string) (*domain.Location, error) {
	return s.locRepo.GetByID(ctx, id)
}

func (s *ProductManagementService) UpdateLocation(ctx context.Context, id, code, name, locType string, isActive bool) (*domain.Location, error) {
	loc, err := s.locRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	loc.LocationCode = code
	loc.LocationName = name
	loc.LocationType = locType
	loc.IsActive = isActive
	err = s.locRepo.Update(ctx, loc)
	if err != nil {
		return nil, err
	}
	return loc, nil
}

func (s *ProductManagementService) DeleteLocation(ctx context.Context, id string) error {
	return s.locRepo.Delete(ctx, id)
}

