package repositories

import (
	"crm-service/internal/models"

	"gorm.io/gorm"
)

type ContactRepository struct {
	db *gorm.DB
}

func NewContactRepository(db *gorm.DB) *ContactRepository {
	return &ContactRepository{db: db}
}

func (r *ContactRepository) Create(contact *models.Contact) error {
	return r.db.Create(contact).Error
}

func (r *ContactRepository) GetByID(id uint) (*models.Contact, error) {
	var contact models.Contact
	err := r.db.First(&contact, id).Error
	if err != nil {
		return nil, err
	}
	return &contact, nil
}

func (r *ContactRepository) GetAll(page, limit int) ([]models.Contact, int64, error) {
	var contacts []models.Contact
	var total int64

	offset := (page - 1) * limit

	err := r.db.Model(&models.Contact{}).Count(&total).Error
	if err != nil {
		return nil, 0, err
	}

	err = r.db.Offset(offset).Limit(limit).Find(&contacts).Error
	if err != nil {
		return nil, 0, err
	}

	return contacts, total, nil
}

func (r *ContactRepository) Update(contact *models.Contact) error {
	return r.db.Save(contact).Error
}

func (r *ContactRepository) Delete(id uint) error {
	return r.db.Delete(&models.Contact{}, id).Error
}

func (r *ContactRepository) Search(query string, page, limit int) ([]models.Contact, int64, error) {
	var contacts []models.Contact
	var total int64

	offset := (page - 1) * limit
	searchQuery := "%" + query + "%"

	err := r.db.Model(&models.Contact{}).
		Where("first_name ILIKE ? OR last_name ILIKE ? OR email ILIKE ? OR company ILIKE ?",
			searchQuery, searchQuery, searchQuery, searchQuery).
		Count(&total).Error
	if err != nil {
		return nil, 0, err
	}

	err = r.db.Where("first_name ILIKE ? OR last_name ILIKE ? OR email ILIKE ? OR company ILIKE ?",
		searchQuery, searchQuery, searchQuery, searchQuery).
		Offset(offset).Limit(limit).Find(&contacts).Error
	if err != nil {
		return nil, 0, err
	}

	return contacts, total, nil
}
