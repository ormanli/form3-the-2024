package simulator

// Service defines a contract for processing amounts.
type Service interface {
	Process(amount int) error
}
