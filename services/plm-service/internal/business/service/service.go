package service

import (
	"context"
	"errors"
	"time"

	"erp-system/shared/utils"
	"github.com/erp-system/plm-service/internal/business/domain"
	"github.com/shopspring/decimal"
)

type BomLineInput struct {
	ComponentMaterialID string          `json:"component_material_id"`
	SequenceNumber      int             `json:"sequence_number"`
	QuantityRequired    decimal.Decimal `json:"quantity_required"`
	Uom                 string          `json:"uom"`
	ScrapPercentage     decimal.Decimal `json:"scrap_percentage"`
}

type BomExplosionGraph struct {
	BOMHeaderID string          `json:"bom_header_id"`
	MaterialID  string          `json:"material_id"`
	Version     string          `json:"version"`
	Components  []ExplosionNode `json:"components"`
}

type ExplosionNode struct {
	MaterialID       string          `json:"material_id"`
	Sku              string          `json:"sku"`
	Description      string          `json:"description"`
	QuantityRequired decimal.Decimal `json:"quantity_required"`
	Depth            int             `json:"depth"`
}

type MaterialService struct {
	matRepo   domain.MaterialMasterRepository
	publisher domain.EventPublisher
}

func NewMaterialService(matRepo domain.MaterialMasterRepository, publisher domain.EventPublisher) *MaterialService {
	return &MaterialService{
		matRepo:   matRepo,
		publisher: publisher,
	}
}

func (s *MaterialService) CreateMaterial(ctx context.Context, legalEntityId string, sku string, description string, uom string, pType domain.ProcurementType) (*domain.MaterialMaster, error) {
	m := &domain.MaterialMaster{
		ID:              utils.NewID("mat"),
		LegalEntityID:   legalEntityId,
		Sku:             sku,
		Description:     description,
		Uom:             uom,
		ProcurementType: pType,
		Status:          domain.MaterialStatusACTIVE,
		Version:         1,
		CreatedAt:       time.Now(),
		UpdatedAt:       time.Now(),
	}
	err := s.matRepo.Create(ctx, m)
	return m, err
}

func (s *MaterialService) UpdateTechnicalSpecs(ctx context.Context, materialId string, specs string) (*domain.MaterialMaster, error) {
	m, err := s.matRepo.GetByID(ctx, materialId)
	if err != nil {
		return nil, err
	}
	m.TechnicalSpecifications = specs
	m.Version++
	m.UpdatedAt = time.Now()
	err = s.matRepo.Update(ctx, m)
	return m, err
}

func (s *MaterialService) TransitionStatus(ctx context.Context, materialId string, newStatus domain.MaterialStatus) (*domain.MaterialMaster, error) {
	m, err := s.matRepo.GetByID(ctx, materialId)
	if err != nil {
		return nil, err
	}
	oldStatus := m.Status
	m.Status = newStatus
	m.Version++
	m.UpdatedAt = time.Now()
	err = s.matRepo.Update(ctx, m)
	if err != nil {
		return nil, err
	}

	if oldStatus != newStatus {
		if newStatus == domain.MaterialStatusOBSOLETE {
			_ = s.publisher.Publish(ctx, domain.TopicPlmMaterialObsoleted, m.ID, map[string]interface{}{
				"event_id":        utils.NewID("evt"),
				"legal_entity_id": m.LegalEntityID,
				"material_id":    m.ID,
				"sku":            m.Sku,
				"timestamp":       time.Now(),
			})
		}
	}

	return m, nil
}

type BomService struct {
	hdrRepo   domain.BomHeaderRepository
	lineRepo  domain.BomLineRepository
	matRepo   domain.MaterialMasterRepository
	publisher domain.EventPublisher
}

func NewBomService(hdrRepo domain.BomHeaderRepository, lineRepo domain.BomLineRepository, matRepo domain.MaterialMasterRepository, publisher domain.EventPublisher) *BomService {
	return &BomService{
		hdrRepo:   hdrRepo,
		lineRepo:  lineRepo,
		matRepo:   matRepo,
		publisher: publisher,
	}
}

func (s *BomService) EstablishBomHeader(ctx context.Context, legalEntityId string, materialId string, versionString string, lines []BomLineInput) (*domain.BomHeader, error) {
	// Verify target material exists
	_, err := s.matRepo.GetByID(ctx, materialId)
	if err != nil {
		return nil, errors.New("parent material not found")
	}

	bh := &domain.BomHeader{
		ID:            utils.NewID("bom"),
		LegalEntityID: legalEntityId,
		MaterialID:    materialId,
		VersionString: versionString,
		Status:        domain.BomStatusDRAFT,
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
	}

	err = s.hdrRepo.Create(ctx, bh)
	if err != nil {
		return nil, err
	}

	for _, l := range lines {
		bl := &domain.BomLine{
			ID:                 utils.NewID("bml"),
			BomHeaderID:        bh.ID,
			ComponentMaterialID: l.ComponentMaterialID,
			SequenceNumber:     l.SequenceNumber,
			QuantityRequired:   l.QuantityRequired,
			Uom:                l.Uom,
			ScrapPercentage:    l.ScrapPercentage,
			CreatedAt:          time.Now(),
			UpdatedAt:          time.Now(),
		}
		_ = s.lineRepo.Create(ctx, bl)
	}

	return bh, nil
}

func (s *BomService) ReleaseBom(ctx context.Context, bomHeaderId string) (*domain.BomHeader, error) {
	bh, err := s.hdrRepo.GetByID(ctx, bomHeaderId)
	if err != nil {
		return nil, err
	}
	bh.Status = domain.BomStatusRELEASED
	bh.UpdatedAt = time.Now()
	err = s.hdrRepo.Update(ctx, bh)
	if err != nil {
		return nil, err
	}

	// Fetch lines to publish components
	lines, _ := s.lineRepo.ListByHeaderID(ctx, bh.ID)
	components := make([]map[string]interface{}, 0)
	for _, l := range lines {
		components = append(components, map[string]interface{}{
			"component_material_id": l.ComponentMaterialID,
			"sequence_number":       l.SequenceNumber,
			"quantity_required":    l.QuantityRequired,
			"uom":                   l.Uom,
		})
	}

	_ = s.publisher.Publish(ctx, domain.TopicPlmBomReleased, bh.ID, map[string]interface{}{
		"event_id":        utils.NewID("evt"),
		"legal_entity_id": bh.LegalEntityID,
		"bom_header_id":   bh.ID,
		"material_id":    bh.MaterialID,
		"version_string":  bh.VersionString,
		"components":      components,
		"timestamp":       time.Now(),
	})

	return bh, nil
}

func (s *BomService) ExplodeBillOfMaterials(ctx context.Context, bomHeaderId string, maxDepth int) (*BomExplosionGraph, error) {
	bh, err := s.hdrRepo.GetByID(ctx, bomHeaderId)
	if err != nil {
		return nil, err
	}

	graph := &BomExplosionGraph{
		BOMHeaderID: bh.ID,
		MaterialID:  bh.MaterialID,
		Version:     bh.VersionString,
		Components:  make([]ExplosionNode, 0),
	}

	err = s.traverse(ctx, bh.ID, 1, maxDepth, graph)
	return graph, err
}

func (s *BomService) traverse(ctx context.Context, headerID string, currentDepth, maxDepth int, graph *BomExplosionGraph) error {
	if currentDepth > maxDepth {
		return nil
	}

	lines, err := s.lineRepo.ListByHeaderID(ctx, headerID)
	if err != nil {
		return err
	}

	for _, l := range lines {
		mat, err := s.matRepo.GetByID(ctx, l.ComponentMaterialID)
		sku := "UNKNOWN"
		desc := ""
		if err == nil {
			sku = mat.Sku
			desc = mat.Description
		}

		node := ExplosionNode{
			MaterialID:       l.ComponentMaterialID,
			Sku:              sku,
			Description:      desc,
			QuantityRequired: l.QuantityRequired,
			Depth:            currentDepth,
		}
		graph.Components = append(graph.Components, node)

		// Check if this component has its own BOM (recursive explosion)
		// For memory testing, we look for a BOM header registered for this material
		headers, _ := s.hdrRepo.List(ctx)
		for _, h := range headers {
			if h.MaterialID == l.ComponentMaterialID && h.Status == domain.BomStatusRELEASED {
				_ = s.traverse(ctx, h.ID, currentDepth+1, maxDepth, graph)
				break
			}
		}
	}

	return nil
}

type EngineeringChangeService struct {
	ecoRepo   domain.EngineeringChangeOrderRepository
	matRepo   domain.MaterialMasterRepository
	publisher domain.EventPublisher
}

func NewEngineeringChangeService(ecoRepo domain.EngineeringChangeOrderRepository, matRepo domain.MaterialMasterRepository, publisher domain.EventPublisher) *EngineeringChangeService {
	return &EngineeringChangeService{
		ecoRepo:   ecoRepo,
		matRepo:   matRepo,
		publisher: publisher,
	}
}

func (s *EngineeringChangeService) InitiateChangeRequest(ctx context.Context, legalEntityId string, materialId string, requesterHrId string, title string, description string) (*domain.EngineeringChangeOrder, error) {
	eco := &domain.EngineeringChangeOrder{
		ID:                utils.NewID("eco"),
		LegalEntityID:     legalEntityId,
		TargetMaterialID:  materialId,
		EcoNumber:         "ECO-" + utils.NewID("num")[:8],
		Title:             title,
		Description:       description,
		Status:            domain.EcoStatusDRAFT,
		RequestedByHrID:   requesterHrId,
		Version:           1,
		CreatedAt:         time.Now(),
		UpdatedAt:         time.Now(),
	}
	err := s.ecoRepo.Create(ctx, eco)
	return eco, err
}

func (s *EngineeringChangeService) ProcessApprovalAction(ctx context.Context, ecoId string, approverHrId string, action domain.EcoStatus) (*domain.EngineeringChangeOrder, error) {
	eco, err := s.ecoRepo.GetByID(ctx, ecoId)
	if err != nil {
		return nil, err
	}

	eco.ApprovedByHrID = &approverHrId
	eco.Status = action
	eco.Version++
	eco.UpdatedAt = time.Now()
	err = s.ecoRepo.Update(ctx, eco)
	if err != nil {
		return nil, err
	}

	if action == domain.EcoStatusAPPROVED {
		// Implement change: publish plm.eco.implemented event
		_ = s.publisher.Publish(ctx, domain.TopicPlmEcoImplemented, eco.ID, map[string]interface{}{
			"event_id":        utils.NewID("evt"),
			"legal_entity_id": eco.LegalEntityID,
			"eco_id":          eco.ID,
			"material_id":    eco.TargetMaterialID,
			"timestamp":       time.Now(),
		})
	}

	return eco, nil
}
