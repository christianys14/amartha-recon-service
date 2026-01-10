package recon

import (
	"amartha-recon-service/configuration"
	"amartha-recon-service/infrastructure/repository/transaction"
	"context"
)

type (
	service struct {
		cfg        configuration.Configuration
		repository transaction.Repository
	}

	Service interface {
		Proceed(ctx context.Context, file *uploadFile) (err error)
	}
)

func NewService(
	cfg configuration.Configuration,
	repository transaction.Repository) Service {
	return &service{
		cfg:        cfg,
		repository: repository,
	}
}

func (s service) Proceed(ctx context.Context, file *uploadFile) (err error) {
	//TODO implement me
	panic("implement me")
}
