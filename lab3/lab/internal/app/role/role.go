package role

type Role int

const (
	Guest     Role = iota // 0
	Organiser             // 1
	Manager               // 2
)
