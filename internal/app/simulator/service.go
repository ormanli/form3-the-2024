package simulator

type Service interface {
	Process(amount int) error
}
