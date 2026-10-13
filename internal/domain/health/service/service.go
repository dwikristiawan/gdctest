package service

type HealthResponse struct {
	Status string `json:"status"`
}

type HealthServiceInterface interface {
	Liveness() (*HealthResponse, error)
}

type HealthService struct{}

func NewHealthService() HealthServiceInterface {
	return &HealthService{}
}

func (s *HealthService) Liveness() (*HealthResponse, error) {
	return &HealthResponse{Status: "ok"}, nil
}
