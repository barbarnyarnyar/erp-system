package domain

import (
	"time"
)

type OpportunityStageHistory struct {
	ID            string           `json:"id"`
	OpportunityID string           `json:"opportunity_id"`
	Stage         OpportunityStage `json:"stage"`
	ChangedAt     time.Time        `json:"changed_at"`
	ChangedBy     string           `json:"changed_by"`
}
