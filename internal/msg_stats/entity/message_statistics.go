package entity

type Status struct {
	SuccessCount uint32
	ErrorCount   uint32
}

type MsgStats struct {
	Count  uint32
	Status Status
}
