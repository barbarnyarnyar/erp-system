package service

import (
	"context"
	"errors"
	"testing"

	"github.com/erp-system/scm-service/internal/business/domain"
	"github.com/erp-system/scm-service/internal/data/memory"
	"github.com/shopspring/decimal"
)

type MockProductRepo struct {
	domain.ProductRepository
	createErr error
	getErr    error
	updateErr error
	deleteErr error
}

func (m *MockProductRepo) Create(ctx context.Context, p *domain.Product) error {
	if m.createErr != nil {
		return m.createErr
	}
	return m.ProductRepository.Create(ctx, p)
}

func (m *MockProductRepo) GetByID(ctx context.Context, id string) (*domain.Product, error) {
	if m.getErr != nil {
		return nil, m.getErr
	}
	return m.ProductRepository.GetByID(ctx, id)
}

func (m *MockProductRepo) Update(ctx context.Context, p *domain.Product) error {
	if m.updateErr != nil {
		return m.updateErr
	}
	return m.ProductRepository.Update(ctx, p)
}

func (m *MockProductRepo) Delete(ctx context.Context, id string) error {
	if m.deleteErr != nil {
		return m.deleteErr
	}
	return m.ProductRepository.Delete(ctx, id)
}

type MockProductCategoryRepo struct {
	domain.ProductCategoryRepository
	createErr error
	getErr    error
	updateErr error
}

func (m *MockProductCategoryRepo) Create(ctx context.Context, pc *domain.ProductCategory) error {
	if m.createErr != nil {
		return m.createErr
	}
	return m.ProductCategoryRepository.Create(ctx, pc)
}

func (m *MockProductCategoryRepo) GetByID(ctx context.Context, id string) (*domain.ProductCategory, error) {
	if m.getErr != nil {
		return nil, m.getErr
	}
	return m.ProductCategoryRepository.GetByID(ctx, id)
}

func (m *MockProductCategoryRepo) Update(ctx context.Context, pc *domain.ProductCategory) error {
	if m.updateErr != nil {
		return m.updateErr
	}
	return m.ProductCategoryRepository.Update(ctx, pc)
}

type MockLocationRepo struct {
	domain.LocationRepository
	createErr error
	getErr    error
	updateErr error
}

func (m *MockLocationRepo) Create(ctx context.Context, loc *domain.Location) error {
	if m.createErr != nil {
		return m.createErr
	}
	return m.LocationRepository.Create(ctx, loc)
}

func (m *MockLocationRepo) GetByID(ctx context.Context, id string) (*domain.Location, error) {
	if m.getErr != nil {
		return nil, m.getErr
	}
	return m.LocationRepository.GetByID(ctx, id)
}

func (m *MockLocationRepo) Update(ctx context.Context, loc *domain.Location) error {
	if m.updateErr != nil {
		return m.updateErr
	}
	return m.LocationRepository.Update(ctx, loc)
}

type MockPublisher struct {
	PublishFunc func(ctx context.Context, topic string, key string, event interface{}) error
}

func (m *MockPublisher) Publish(ctx context.Context, topic string, key string, event interface{}) error {
	if m.PublishFunc != nil {
		return m.PublishFunc(ctx, topic, key, event)
	}
	return nil
}

func TestProductManagementService_Products(t *testing.T) {
	ctx := context.Background()

	t.Run("Create and List Products", func(t *testing.T) {
		repo := memory.NewMemoryProductRepo()
		catRepo := memory.NewMemoryProductCategoryRepo()
		locRepo := memory.NewMemoryLocationRepo()
		pub := &MockPublisher{}
		svc := NewProductManagementService(repo, catRepo, locRepo, pub)

		categoryID := "cat_1"
		p, err := svc.CreateProduct(ctx, "P001", "Product 1", "Desc", "RAW", "EA", decimal.NewFromFloat(10.0), decimal.NewFromFloat(20.0), &categoryID)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if p.ID == "" {
			t.Error("expected generated ID")
		}

		list, err := svc.ListProducts(ctx)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(list) != 1 {
			t.Errorf("expected 1 product, got %d", len(list))
		}
	})

	t.Run("Create Product Repo Error", func(t *testing.T) {
		repo := &MockProductRepo{
			ProductRepository: memory.NewMemoryProductRepo(),
			createErr:         errors.New("db create error"),
		}
		svc := NewProductManagementService(repo, nil, nil, &MockPublisher{})
		_, err := svc.CreateProduct(ctx, "P001", "Product 1", "Desc", "RAW", "EA", decimal.NewFromFloat(10.0), decimal.NewFromFloat(20.0), nil)
		if err == nil {
			t.Error("expected error, got nil")
		}
	})

	t.Run("Create Product Publisher Error (logs but succeeds)", func(t *testing.T) {
		repo := memory.NewMemoryProductRepo()
		pub := &MockPublisher{
			PublishFunc: func(ctx context.Context, topic string, key string, event interface{}) error {
				return errors.New("pub error")
			},
		}
		svc := NewProductManagementService(repo, nil, nil, pub)
		_, err := svc.CreateProduct(ctx, "P001", "Product 1", "Desc", "RAW", "EA", decimal.NewFromFloat(10.0), decimal.NewFromFloat(20.0), nil)
		if err != nil {
			t.Fatalf("expected success even if publisher fails, got: %v", err)
		}
	})

	t.Run("Get Product", func(t *testing.T) {
		repo := memory.NewMemoryProductRepo()
		svc := NewProductManagementService(repo, nil, nil, &MockPublisher{})
		p, _ := svc.CreateProduct(ctx, "P001", "Product 1", "Desc", "RAW", "EA", decimal.NewFromFloat(10.0), decimal.NewFromFloat(20.0), nil)

		got, err := svc.GetProduct(ctx, p.ID)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if got.ProductCode != "P001" {
			t.Errorf("expected code P001, got %s", got.ProductCode)
		}

		_, err = svc.GetProduct(ctx, "nonexistent")
		if err == nil {
			t.Error("expected error for nonexistent product, got nil")
		}
	})

	t.Run("Update Product Success", func(t *testing.T) {
		repo := memory.NewMemoryProductRepo()
		svc := NewProductManagementService(repo, nil, nil, &MockPublisher{})
		p, _ := svc.CreateProduct(ctx, "P001", "Product 1", "Desc", "RAW", "EA", decimal.NewFromFloat(10.0), decimal.NewFromFloat(20.0), nil)

		updated, err := svc.UpdateProduct(ctx, p.ID, "P001-Updated", "Product 1 Updated", "Desc Updated", "FINISHED", "KG", decimal.NewFromFloat(12.0), decimal.NewFromFloat(24.0), false, nil)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if updated.ProductCode != "P001-Updated" || updated.IsActive != false {
			t.Errorf("unexpected updated values: %+v", updated)
		}
	})

	t.Run("Update Product Get Error", func(t *testing.T) {
		repo := &MockProductRepo{
			ProductRepository: memory.NewMemoryProductRepo(),
			getErr:            errors.New("not found"),
		}
		svc := NewProductManagementService(repo, nil, nil, &MockPublisher{})
		_, err := svc.UpdateProduct(ctx, "nonexistent", "P001", "Product 1", "Desc", "RAW", "EA", decimal.NewFromFloat(10.0), decimal.NewFromFloat(20.0), true, nil)
		if err == nil {
			t.Error("expected error, got nil")
		}
	})

	t.Run("Update Product Update Error", func(t *testing.T) {
		repo := memory.NewMemoryProductRepo()
		mockRepo := &MockProductRepo{
			ProductRepository: repo,
			updateErr:         errors.New("db update error"),
		}
		svc := NewProductManagementService(mockRepo, nil, nil, &MockPublisher{})
		// Seed
		seedSvc := NewProductManagementService(repo, nil, nil, &MockPublisher{})
		p, _ := seedSvc.CreateProduct(ctx, "P001", "Product 1", "Desc", "RAW", "EA", decimal.NewFromFloat(10.0), decimal.NewFromFloat(20.0), nil)

		_, err := svc.UpdateProduct(ctx, p.ID, "P001", "Product 1", "Desc", "RAW", "EA", decimal.NewFromFloat(10.0), decimal.NewFromFloat(20.0), true, nil)
		if err == nil {
			t.Error("expected error, got nil")
		}
	})

	t.Run("Delete Product Success", func(t *testing.T) {
		repo := memory.NewMemoryProductRepo()
		svc := NewProductManagementService(repo, nil, nil, &MockPublisher{})
		p, _ := svc.CreateProduct(ctx, "P001", "Product 1", "Desc", "RAW", "EA", decimal.NewFromFloat(10.0), decimal.NewFromFloat(20.0), nil)

		err := svc.DeleteProduct(ctx, p.ID)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		_, err = svc.GetProduct(ctx, p.ID)
		if err == nil {
			t.Error("expected error for deleted product")
		}
	})

	t.Run("Delete Product Error", func(t *testing.T) {
		repo := &MockProductRepo{
			ProductRepository: memory.NewMemoryProductRepo(),
			deleteErr:         errors.New("db delete error"),
		}
		svc := NewProductManagementService(repo, nil, nil, &MockPublisher{})
		err := svc.DeleteProduct(ctx, "some-id")
		if err == nil {
			t.Error("expected error, got nil")
		}
	})
}

func TestProductManagementService_Categories(t *testing.T) {
	ctx := context.Background()

	t.Run("Create and List Categories", func(t *testing.T) {
		repo := memory.NewMemoryProductRepo()
		catRepo := memory.NewMemoryProductCategoryRepo()
		locRepo := memory.NewMemoryLocationRepo()
		svc := NewProductManagementService(repo, catRepo, locRepo, &MockPublisher{})

		pc, err := svc.CreateCategory(ctx, "C001", "Category 1", "Desc")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if pc.ID == "" {
			t.Error("expected generated ID")
		}

		list, err := svc.ListCategories(ctx)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(list) != 1 {
			t.Errorf("expected 1 category, got %d", len(list))
		}
	})

	t.Run("Create Category Error", func(t *testing.T) {
		catRepo := &MockProductCategoryRepo{
			ProductCategoryRepository: memory.NewMemoryProductCategoryRepo(),
			createErr:                 errors.New("db create error"),
		}
		svc := NewProductManagementService(nil, catRepo, nil, &MockPublisher{})
		_, err := svc.CreateCategory(ctx, "C001", "Category 1", "Desc")
		if err == nil {
			t.Error("expected error, got nil")
		}
	})

	t.Run("Get Category", func(t *testing.T) {
		catRepo := memory.NewMemoryProductCategoryRepo()
		svc := NewProductManagementService(nil, catRepo, nil, &MockPublisher{})
		pc, _ := svc.CreateCategory(ctx, "C001", "Category 1", "Desc")

		got, err := svc.GetCategory(ctx, pc.ID)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if got.Name != "Category 1" {
			t.Errorf("expected Category 1, got %s", got.Name)
		}

		_, err = svc.GetCategory(ctx, "nonexistent")
		if err == nil {
			t.Error("expected error, got nil")
		}
	})

	t.Run("Update Category Success", func(t *testing.T) {
		catRepo := memory.NewMemoryProductCategoryRepo()
		svc := NewProductManagementService(nil, catRepo, nil, &MockPublisher{})
		pc, _ := svc.CreateCategory(ctx, "C001", "Category 1", "Desc")

		updated, err := svc.UpdateCategory(ctx, pc.ID, "C001-Updated", "Category 1 Updated", "Desc Updated")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if updated.Name != "Category 1 Updated" {
			t.Errorf("expected Category 1 Updated, got %s", updated.Name)
		}
	})

	t.Run("Update Category Get Error", func(t *testing.T) {
		catRepo := &MockProductCategoryRepo{
			ProductCategoryRepository: memory.NewMemoryProductCategoryRepo(),
			getErr:                    errors.New("not found"),
		}
		svc := NewProductManagementService(nil, catRepo, nil, &MockPublisher{})
		_, err := svc.UpdateCategory(ctx, "nonexistent", "C001", "Cat 1", "")
		if err == nil {
			t.Error("expected error, got nil")
		}
	})

	t.Run("Update Category Update Error", func(t *testing.T) {
		catRepo := memory.NewMemoryProductCategoryRepo()
		mockCatRepo := &MockProductCategoryRepo{
			ProductCategoryRepository: catRepo,
			updateErr:                 errors.New("db update error"),
		}
		svc := NewProductManagementService(nil, mockCatRepo, nil, &MockPublisher{})
		// Seed
		seedSvc := NewProductManagementService(nil, catRepo, nil, &MockPublisher{})
		pc, _ := seedSvc.CreateCategory(ctx, "C001", "Category 1", "Desc")

		_, err := svc.UpdateCategory(ctx, pc.ID, "C001", "Category 1", "Desc")
		if err == nil {
			t.Error("expected error, got nil")
		}
	})

	t.Run("Delete Category", func(t *testing.T) {
		catRepo := memory.NewMemoryProductCategoryRepo()
		svc := NewProductManagementService(nil, catRepo, nil, &MockPublisher{})
		pc, _ := svc.CreateCategory(ctx, "C001", "Category 1", "Desc")

		err := svc.DeleteCategory(ctx, pc.ID)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		_, err = svc.GetCategory(ctx, pc.ID)
		if err == nil {
			t.Error("expected error for deleted category")
		}
	})
}

func TestProductManagementService_Locations(t *testing.T) {
	ctx := context.Background()

	t.Run("Create and List Locations", func(t *testing.T) {
		locRepo := memory.NewMemoryLocationRepo()
		svc := NewProductManagementService(nil, nil, locRepo, &MockPublisher{})

		loc, err := svc.CreateLocation(ctx, "L001", "Location 1", "WAREHOUSE")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if loc.ID == "" {
			t.Error("expected generated ID")
		}

		list, err := svc.ListLocations(ctx)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(list) != 1 {
			t.Errorf("expected 1 location, got %d", len(list))
		}
	})

	t.Run("Create Location Error", func(t *testing.T) {
		locRepo := &MockLocationRepo{
			LocationRepository: memory.NewMemoryLocationRepo(),
			createErr:          errors.New("db create error"),
		}
		svc := NewProductManagementService(nil, nil, locRepo, &MockPublisher{})
		_, err := svc.CreateLocation(ctx, "L001", "Location 1", "WAREHOUSE")
		if err == nil {
			t.Error("expected error, got nil")
		}
	})

	t.Run("Get Location", func(t *testing.T) {
		locRepo := memory.NewMemoryLocationRepo()
		svc := NewProductManagementService(nil, nil, locRepo, &MockPublisher{})
		loc, _ := svc.CreateLocation(ctx, "L001", "Location 1", "WAREHOUSE")

		got, err := svc.GetLocation(ctx, loc.ID)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if got.LocationName != "Location 1" {
			t.Errorf("expected Location 1, got %s", got.LocationName)
		}

		_, err = svc.GetLocation(ctx, "nonexistent")
		if err == nil {
			t.Error("expected error, got nil")
		}
	})

	t.Run("Update Location Success", func(t *testing.T) {
		locRepo := memory.NewMemoryLocationRepo()
		svc := NewProductManagementService(nil, nil, locRepo, &MockPublisher{})
		loc, _ := svc.CreateLocation(ctx, "L001", "Location 1", "WAREHOUSE")

		updated, err := svc.UpdateLocation(ctx, loc.ID, "L001-Updated", "Location 1 Updated", "STORE", false)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if updated.LocationName != "Location 1 Updated" || updated.IsActive != false {
			t.Errorf("unexpected updated values: %+v", updated)
		}
	})

	t.Run("Update Location Get Error", func(t *testing.T) {
		locRepo := &MockLocationRepo{
			LocationRepository: memory.NewMemoryLocationRepo(),
			getErr:             errors.New("not found"),
		}
		svc := NewProductManagementService(nil, nil, locRepo, &MockPublisher{})
		_, err := svc.UpdateLocation(ctx, "nonexistent", "L001", "Loc 1", "WAREHOUSE", true)
		if err == nil {
			t.Error("expected error, got nil")
		}
	})

	t.Run("Update Location Update Error", func(t *testing.T) {
		locRepo := memory.NewMemoryLocationRepo()
		mockLocRepo := &MockLocationRepo{
			LocationRepository: locRepo,
			updateErr:          errors.New("db update error"),
		}
		svc := NewProductManagementService(nil, nil, mockLocRepo, &MockPublisher{})
		// Seed
		seedSvc := NewProductManagementService(nil, nil, locRepo, &MockPublisher{})
		loc, _ := seedSvc.CreateLocation(ctx, "L001", "Location 1", "WAREHOUSE")

		_, err := svc.UpdateLocation(ctx, loc.ID, "L001", "Location 1", "WAREHOUSE", true)
		if err == nil {
			t.Error("expected error, got nil")
		}
	})

	t.Run("Delete Location", func(t *testing.T) {
		locRepo := memory.NewMemoryLocationRepo()
		svc := NewProductManagementService(nil, nil, locRepo, &MockPublisher{})
		loc, _ := svc.CreateLocation(ctx, "L001", "Location 1", "WAREHOUSE")

		err := svc.DeleteLocation(ctx, loc.ID)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		_, err = svc.GetLocation(ctx, loc.ID)
		if err == nil {
			t.Error("expected error for deleted location")
		}
	})
}
