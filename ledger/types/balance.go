package types

type Balance struct {
	Amounts map[uint32]int64
}

func NewBalance() *Balance {
	return &Balance{
		Amounts: make(map[uint32]int64),
	}
}
