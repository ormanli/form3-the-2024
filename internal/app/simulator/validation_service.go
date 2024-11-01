package simulator

// ValidationService validates and processes amounts using an underlying service.
type ValidationService struct {
	service Service
}

// NewValidationService creates a new ValidationService with the given service.
func NewValidationService(service Service) *ValidationService {
	return &ValidationService{service: service}
}

// Process validates and processes the amount using the underlying service.
// It returns an error if the amount is invalid (i.e., less than 0).
func (v *ValidationService) Process(amount int) error {
	if amount < 0 {
		return ErrInvalidAmount
	}

	return v.service.Process(amount)
}
