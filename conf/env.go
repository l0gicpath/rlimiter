package conf

type Env int

const (
	DevelopmentEnv Env = iota
	ProductionEnv
	TestEnv
)

func (e Env) String() string {
	return [...]string{"Development", "Production", "Test"}[e]
}
