package svc

import (
	"context"

	"app/internal/core/graph/model"
	genericpb "app/internal/core/grpc/generated/generic/utils"
	pb "app/internal/core/grpc/generated/lotof.sample.svc/domainItem"

	"google.golang.org/grpc"
)

// DomainItemService is the template's example client -- delete/rename this
// whole module when bootstrapping a real domain, or copy its shape for a
// real entity's client.
type DomainItemService struct {
	client pb.SampleDomainItemServiceClient
}

// NewDomainItemService wraps the shared pooled connection to the svc --
// every module's client is constructed the same way, from the one
// connection dialed in app.go (see grpcpool.NewPooledClient).
func NewDomainItemService(conn grpc.ClientConnInterface) *DomainItemService {
	return &DomainItemService{client: pb.NewSampleDomainItemServiceClient(conn)}
}

func entityFromPb(item *pb.DomainItem) *model.DomainItem {
	if item == nil {
		return nil
	}
	return &model.DomainItem{
		ID:     item.Id,
		Name:   item.Name,
		Status: model.DomainItemStatus(item.Status.String()),
	}
}

func (s *DomainItemService) CreateDomainItem(ctx context.Context, name string) (*model.DomainItem, error) {
	res, err := s.client.CreateDomainItem(ctx, &pb.CreateDomainItemRequest{Name: name})
	if err != nil {
		return nil, err
	}
	return entityFromPb(res), nil
}

func (s *DomainItemService) DomainItem(ctx context.Context, id string) (*model.DomainItem, error) {
	res, err := s.client.GetDomainItem(ctx, &pb.GetDomainItemRequest{Id: id})
	if err != nil {
		return nil, err
	}
	return entityFromPb(res), nil
}

func (s *DomainItemService) DomainItems(ctx context.Context, page, length int32) (*model.DomainItemList, error) {
	res, err := s.client.ListDomainItems(ctx, &pb.ListDomainItemsRequest{
		Pagination: &genericpb.Pagination{Page: page, Length: length},
	})
	if err != nil {
		return nil, err
	}
	rows := make([]*model.DomainItem, len(res.DomainItems))
	for i, item := range res.DomainItems {
		rows[i] = entityFromPb(item)
	}
	return &model.DomainItemList{
		Rows: rows,
		Info: &model.PaginationInfo{Count: int(res.PaginationInfo.GetCount())},
	}, nil
}

func (s *DomainItemService) UpdateDomainItem(ctx context.Context, id, name string, status pb.DomainItemStatus) (*model.DomainItem, error) {
	res, err := s.client.UpdateDomainItem(ctx, &pb.UpdateDomainItemRequest{Id: id, Name: name, Status: status})
	if err != nil {
		return nil, err
	}
	return entityFromPb(res), nil
}

func (s *DomainItemService) DeleteDomainItem(ctx context.Context, id string) (bool, error) {
	res, err := s.client.DeleteDomainItem(ctx, &pb.DeleteDomainItemRequest{Id: id})
	if err != nil {
		return false, err
	}
	return res.GetSuccess(), nil
}
