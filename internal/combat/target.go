package combat

import (
	"math"
	"slices"
)

type TargetSystem struct {
	actors []*Actor
}

func NewTargetSystem() *TargetSystem {
	return &TargetSystem{}
}

func (ts *TargetSystem) Register(a *Actor) {
	ts.actors = append(ts.actors, a)
}

func (ts *TargetSystem) Unregister(a *Actor) {
	idx := -1
	for i, act := range ts.actors {
		if act.ActorID == a.ActorID {
			idx = i
			break
		}
	}
	if idx >= 0 {
		ts.actors = slices.Delete(ts.actors, idx, idx+1)
	}
}

func (ts *TargetSystem) GetClosestEnemy(source *Actor, sourceX, sourceY float64) *Actor {
	var closest *Actor
	closestDist := math.MaxFloat64
	for _, a := range ts.actors {
		if a.Team == source.Team || !a.Alive {
			continue
		}
		dist := math.Abs(sourceX) + math.Abs(sourceY)
		if dist < closestDist {
			closestDist = dist
			closest = a
		}
	}
	return closest
}
