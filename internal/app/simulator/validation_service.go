package simulator

type ValidationService struct {
	service Service
}

func NewValidationService(service Service) *ValidationService {
	return &ValidationService{service: service}
}

func (v *ValidationService) Process(amount int) error {
	if amount < 0 {
		return ErrInvalidAmount
	}

	return v.service.Process(amount)
}
