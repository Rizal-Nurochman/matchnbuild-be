package gateway

import (
	"context"

	"github.com/Rizal-Nurochman/matchnbuild/config"
	"github.com/midtrans/midtrans-go/coreapi"
	"github.com/midtrans/midtrans-go/snap"
)

type MidtransGateway interface {
	CreateSnapTransaction(ctx context.Context, req *snap.Request) (*snap.Response, error)
	CheckTransaction(ctx context.Context, midtransOrderID string) (*coreapi.TransactionStatusResponse, error)
}

type midtransGateway struct {
	config config.MidtransConfig
}

func NewMidtransGateway(cfg config.MidtransConfig) MidtransGateway {
	return &midtransGateway{config: cfg}
}

func (g *midtransGateway) CreateSnapTransaction(ctx context.Context, req *snap.Request) (*snap.Response, error) {
	snapClient, _ := config.NewMidtransClients(g.config)
	snapClient.Options.SetContext(ctx)
	response, err := snapClient.CreateTransaction(req)
	if err != nil {
		return nil, err
	}
	return response, nil
}

func (g *midtransGateway) CheckTransaction(ctx context.Context, midtransOrderID string) (*coreapi.TransactionStatusResponse, error) {
	_, coreClient := config.NewMidtransClients(g.config)
	coreClient.Options.SetContext(ctx)
	response, err := coreClient.CheckTransaction(midtransOrderID)
	if err != nil {
		return nil, err
	}
	return response, nil
}
