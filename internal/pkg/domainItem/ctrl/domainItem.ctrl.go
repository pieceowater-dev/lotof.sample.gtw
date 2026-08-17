package ctrl

import (
	"context"

	"app/internal/core/graph/model"
	pb "app/internal/core/grpc/generated/lotof.sample.svc/domainItem"
	"app/internal/pkg/domainItem/svc"
)

// DomainItemController handles the operations related to domain items.
type DomainItemController struct {
	svc *svc.DomainItemService
}

func NewDomainItemController(service *svc.DomainItemService) *DomainItemController {
	return &DomainItemController{svc: service}
}

func (c *DomainItemController) CreateDomainItem(ctx context.Context, name string) (*model.DomainItem, error) {
	return c.svc.CreateDomainItem(ctx, name)
}

func (c *DomainItemController) DomainItem(ctx context.Context, id string) (*model.DomainItem, error) {
	return c.svc.DomainItem(ctx, id)
}

func (c *DomainItemController) DomainItems(ctx context.Context, filter *model.DefaultFilterInput) (*model.DomainItemList, error) {
	page, length := int32(1), int32(20)
	if filter != nil && filter.Pagination != nil {
		page, length = BuildPagination(filter.Pagination)
	}
	return c.svc.DomainItems(ctx, page, length)
}

// BuildPagination converts the GraphQL pagination input into (page, length)
// -- copy this into your first real module's ctrl package once domainItem
// is gone (see e.g. lotof.issues.gtw's board/ctrl.BuildPagination, the same
// helper, for a module with more than one entity needing it).
func BuildPagination(p *model.DefaultFilterPaginationInput) (page, length int32) {
	page, length = 1, 20
	if p.Page != nil {
		page = int32(*p.Page)
	}
	if p.Length != nil {
		length = PaginationLengthToInt32(*p.Length)
	}
	return page, length
}

func PaginationLengthToInt32(l model.FilterPaginationLengthEnum) int32 {
	switch l {
	case model.FilterPaginationLengthEnumTen:
		return 10
	case model.FilterPaginationLengthEnumFifteen:
		return 15
	case model.FilterPaginationLengthEnumTwenty:
		return 20
	case model.FilterPaginationLengthEnumTwentyFive:
		return 25
	case model.FilterPaginationLengthEnumThirty:
		return 30
	case model.FilterPaginationLengthEnumFifty:
		return 50
	case model.FilterPaginationLengthEnumOneHundred:
		return 100
	default:
		return 20
	}
}

func (c *DomainItemController) UpdateDomainItem(ctx context.Context, id, name string, status model.DomainItemStatus) (*model.DomainItem, error) {
	return c.svc.UpdateDomainItem(ctx, id, name, pb.DomainItemStatus(pb.DomainItemStatus_value[string(status)]))
}

func (c *DomainItemController) DeleteDomainItem(ctx context.Context, id string) (bool, error) {
	return c.svc.DeleteDomainItem(ctx, id)
}
