package combat

type Team int

const (
	TeamPlayer Team = iota
	TeamAlly
	TeamEnemy
	TeamNeutral
)

func (t Team) IsHostile(other Team) bool {
	switch t {
	case TeamPlayer, TeamAlly:
		return other == TeamEnemy
	case TeamEnemy:
		return other == TeamPlayer || other == TeamAlly
	default:
		return false
	}
}
