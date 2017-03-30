package cluster

// Strategy is a type that indicates the type of load balance.
type Strategy uint8

// StrategyFunc is a func that implements the load balance: assign resources to participants
// and return the final decision.
type StrategyFunc func(participants []Participant, resources []Resource) (decision Decision)

const (
	// StrategyRoundRobin is a round-robin.
	StrategyRoundRobin Strategy = 1

	// StrategyWeightedRoundRobin is based upon StrategyRoundRobin while taking consideration of the weight.
	StrategyWeightedRoundRobin Strategy = 2

	StrategySticky = 3
)

var (
	strategies = map[Strategy]StrategyFunc{
		StrategyRoundRobin: assignRoundRobin,
	}
)

// GetStrategyFunc returns the func according to the specified strategy.
// IMPORTANT the returned func might be nil: it is caller's job to check.
func GetStrategyFunc(s Strategy) StrategyFunc {
	return strategies[s]
}
