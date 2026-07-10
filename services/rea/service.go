package rea

import "context"

type MockService struct {
}

func (s MockService) Stop(ctx context.Context) {

}

func (s MockService) Run(ctx context.Context) {

}

func NewService() MockService {
	return MockService{}
}
