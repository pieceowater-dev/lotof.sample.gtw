package ctrl

import (
	"app/internal/core/graph/model"
	"app/internal/pkg/domainItem/svc"
	"context"
	"fmt"
)

// DomainItemController handles the operations related to domain items.
type DomainItemController struct {
	somethingsService *svc.DomainItemService // Service for fetching somethings.
}

// NewDomainItemController creates a new DomainItemController with the provided service.
func NewDomainItemController(service *svc.DomainItemService) *DomainItemController {
	return &DomainItemController{somethingsService: service}
}

// Somethings fetches a list of somethings using the service.
func (c DomainItemController) Somethings(ctx context.Context) ([]*model.Something, error) {
	somethings, err := c.somethingsService.Somethings(ctx)
	if err != nil {
		fmt.Println("Error fetching:", err)
		return nil, err
	}
	return somethings, nil
}
