package conf

// Env type represents the environment in which the service is
// running inside.
type Env int

const (
	DevelopmentEnv Env = iota
	ProductionEnv
	TestEnv
)

func (e Env) String() string {
	return [...]string{"Development", "Production", "Test"}[e]
}
