package repositories

import (
	"crm-service/internal/models"

	"gorm.io/gorm"
)

type ServiceRepository struct {
	db *gorm.DB
}

func NewServiceRepository(db *gorm.DB) *ServiceRepository {
	return &ServiceRepository{db: db}
}

func (r *ServiceRepository) Create(service *models.Service) error {
	return r.db.Create(service).Error
}

func (r *ServiceRepository) GetByID(id uint) (*models.Service, error) {
	var service models.Service
	err := r.db.First(&service, id).Error
	if err != nil {
		return nil, err
	}
	return &service, nil
}

func (r *ServiceRepository) GetAll(page, limit int) ([]models.Service, int64, error) {
	var services []models.Service
	var total int64

	offset := (page - 1) * limit

	err := r.db.Model(&models.Service{}).Count(&total).Error
	if err != nil {
		return nil, 0, err
	}

	err = r.db.Offset(offset).Limit(limit).Find(&services).Error
	if err != nil {
		return nil, 0, err
	}

	return services, total, nil
}

func (r *ServiceRepository) Update(service *models.Service) error {
	return r.db.Save(service).Error
}

func (r *ServiceRepository) Delete(id uint) error {
	return r.db.Delete(&models.Service{}, id).Error
}

func (r *ServiceRepository) GetActiveServices() ([]models.Service, error) {
	var services []models.Service
	err := r.db.Where("is_active = ?", true).Find(&services).Error
	return services, err
}

// ServiceRequestRepository methods
type ServiceRequestRepository struct {
	db *gorm.DB
}

func NewServiceRequestRepository(db *gorm.DB) *ServiceRequestRepository {
	return &ServiceRequestRepository{db: db}
}

func (r *ServiceRequestRepository) Create(request *models.ServiceRequest) error {
	return r.db.Create(request).Error
}

func (r *ServiceRequestRepository) GetByID(id uint) (*models.ServiceRequest, error) {
	var request models.ServiceRequest
	err := r.db.Preload("Contact").Preload("Service").First(&request, id).Error
	if err != nil {
		return nil, err
	}
	return &request, nil
}

func (r *ServiceRequestRepository) GetAll(page, limit int) ([]models.ServiceRequest, int64, error) {
	var requests []models.ServiceRequest
	var total int64

	offset := (page - 1) * limit

	err := r.db.Model(&models.ServiceRequest{}).Count(&total).Error
	if err != nil {
		return nil, 0, err
	}

	err = r.db.Preload("Contact").Preload("Service").
		Offset(offset).Limit(limit).Find(&requests).Error
	if err != nil {
		return nil, 0, err
	}

	return requests, total, nil
}

func (r *ServiceRequestRepository) Update(request *models.ServiceRequest) error {
	return r.db.Save(request).Error
}

func (r *ServiceRequestRepository) GetByContactID(contactID uint, page, limit int) ([]models.ServiceRequest, int64, error) {
	var requests []models.ServiceRequest
	var total int64

	offset := (page - 1) * limit

	err := r.db.Model(&models.ServiceRequest{}).Where("contact_id = ?", contactID).Count(&total).Error
	if err != nil {
		return nil, 0, err
	}

	err = r.db.Preload("Contact").Preload("Service").
		Where("contact_id = ?", contactID).
		Offset(offset).Limit(limit).Find(&requests).Error
	if err != nil {
		return nil, 0, err
	}

	return requests, total, nil
}
