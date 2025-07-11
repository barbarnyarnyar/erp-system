package repositories

import (
	"crm-service/internal/models"

	"gorm.io/gorm"
)

type SupportTicketRepository struct {
	db *gorm.DB
}

func NewSupportTicketRepository(db *gorm.DB) *SupportTicketRepository {
	return &SupportTicketRepository{db: db}
}

func (r *SupportTicketRepository) Create(ticket *models.SupportTicket) error {
	return r.db.Create(ticket).Error
}

func (r *SupportTicketRepository) GetByID(id uint) (*models.SupportTicket, error) {
	var ticket models.SupportTicket
	err := r.db.Preload("Contact").First(&ticket, id).Error
	if err != nil {
		return nil, err
	}
	return &ticket, nil
}

func (r *SupportTicketRepository) GetByTicketNumber(ticketNumber string) (*models.SupportTicket, error) {
	var ticket models.SupportTicket
	err := r.db.Preload("Contact").Where("ticket_number = ?", ticketNumber).First(&ticket).Error
	if err != nil {
		return nil, err
	}
	return &ticket, nil
}

func (r *SupportTicketRepository) GetAll(page, limit int) ([]models.SupportTicket, int64, error) {
	var tickets []models.SupportTicket
	var total int64

	offset := (page - 1) * limit

	err := r.db.Model(&models.SupportTicket{}).Count(&total).Error
	if err != nil {
		return nil, 0, err
	}

	err = r.db.Preload("Contact").
		Offset(offset).Limit(limit).Find(&tickets).Error
	if err != nil {
		return nil, 0, err
	}

	return tickets, total, nil
}

func (r *SupportTicketRepository) Update(ticket *models.SupportTicket) error {
	return r.db.Save(ticket).Error
}

func (r *SupportTicketRepository) GetByStatus(status string, page, limit int) ([]models.SupportTicket, int64, error) {
	var tickets []models.SupportTicket
	var total int64

	offset := (page - 1) * limit

	err := r.db.Model(&models.SupportTicket{}).Where("status = ?", status).Count(&total).Error
	if err != nil {
		return nil, 0, err
	}

	err = r.db.Preload("Contact").
		Where("status = ?", status).
		Offset(offset).Limit(limit).Find(&tickets).Error
	if err != nil {
		return nil, 0, err
	}

	return tickets, total, nil
}

func (r *SupportTicketRepository) GetByContactID(contactID uint, page, limit int) ([]models.SupportTicket, int64, error) {
	var tickets []models.SupportTicket
	var total int64

	offset := (page - 1) * limit

	err := r.db.Model(&models.SupportTicket{}).Where("contact_id = ?", contactID).Count(&total).Error
	if err != nil {
		return nil, 0, err
	}

	err = r.db.Preload("Contact").
		Where("contact_id = ?", contactID).
		Offset(offset).Limit(limit).Find(&tickets).Error
	if err != nil {
		return nil, 0, err
	}

	return tickets, total, nil
}

// SupportResponseRepository methods
type SupportResponseRepository struct {
	db *gorm.DB
}

func NewSupportResponseRepository(db *gorm.DB) *SupportResponseRepository {
	return &SupportResponseRepository{db: db}
}

func (r *SupportResponseRepository) Create(response *models.SupportResponse) error {
	return r.db.Create(response).Error
}

func (r *SupportResponseRepository) GetByTicketID(ticketID uint) ([]models.SupportResponse, error) {
	var responses []models.SupportResponse
	err := r.db.Where("ticket_id = ?", ticketID).Find(&responses).Error
	return responses, err
}

func (r *SupportResponseRepository) GetByID(id uint) (*models.SupportResponse, error) {
	var response models.SupportResponse
	err := r.db.Preload("Ticket").First(&response, id).Error
	if err != nil {
		return nil, err
	}
	return &response, nil
}
