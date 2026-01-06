package ring

import (
	"errors"
	"time"

	"google.golang.org/protobuf/types/known/timestamppb"

	"github.com/smartcontractkit/chainlink-common/pkg/workflows/ring/pb"
)

func IsInSteadyState(state *pb.RoutingState) bool {
	if state == nil {
		return false
	}
	_, ok := state.State.(*pb.RoutingState_RoutableShards)
	return ok
}

func NextStateFromSteady(currentID uint64, currentShards, wantShards uint32, now time.Time, timeToSync time.Duration) *pb.RoutingState {
	if currentShards == wantShards {
		return &pb.RoutingState{
			Id:    currentID,
			State: &pb.RoutingState_RoutableShards{RoutableShards: currentShards},
		}
	}

	return &pb.RoutingState{
		Id: currentID + 1,
		State: &pb.RoutingState_Transition{
			Transition: &pb.Transition{
				WantShards:       wantShards,
				LastStableCount:  currentShards,
				ChangesSafeAfter: timestamppb.New(now.Add(timeToSync)),
			},
		},
	}
}

func NextStateFromTransition(currentID uint64, transition *pb.Transition, now time.Time) *pb.RoutingState {
	safeAfter := transition.ChangesSafeAfter.AsTime()

	if now.Before(safeAfter) {
		return &pb.RoutingState{
			Id: currentID,
			State: &pb.RoutingState_Transition{
				Transition: transition,
			},
		}
	}

	return &pb.RoutingState{
		Id: currentID + 1,
		State: &pb.RoutingState_RoutableShards{
			RoutableShards: transition.WantShards,
		},
	}
}

func NextState(current *pb.RoutingState, wantShards uint32, now time.Time, timeToSync time.Duration) (*pb.RoutingState, error) {
	if current == nil {
		return nil, errors.New("current state is nil")
	}

	switch s := current.State.(type) {
	case *pb.RoutingState_RoutableShards:
		return NextStateFromSteady(current.Id, s.RoutableShards, wantShards, now, timeToSync), nil

	case *pb.RoutingState_Transition:
		return NextStateFromTransition(current.Id, s.Transition, now), nil

	// coverage:ignore
	default:
		return nil, errors.New("unknown state type")
	}
}
