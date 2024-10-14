package store

import (
	"q/rpc"

	"github.com/rs/zerolog/log"
)

type CheckPointStore struct {
	PendingCheckPoints map[int64]map[int64]*rpc.SignedCheckpoint // executeHeight -> id -> checkpoint
	QuorumCheckPoints  map[int64]*rpc.QuorumCheckpoint           // executeHeight -> checkpoint
}

func NewCheckPointStore() *CheckPointStore {
	return &CheckPointStore{
		PendingCheckPoints: make(map[int64]map[int64]*rpc.SignedCheckpoint),
		QuorumCheckPoints:  make(map[int64]*rpc.QuorumCheckpoint),
	}
}

func (s *CheckPointStore) AddPendingCheckPoint(checkPoint *rpc.SignedCheckpoint) {
	if _, ok := s.PendingCheckPoints[int64(checkPoint.Checkpoint.ExecuteHeight)]; !ok {
		s.PendingCheckPoints[int64(checkPoint.Checkpoint.ExecuteHeight)] = make(map[int64]*rpc.SignedCheckpoint)
	}
	if _, ok := s.PendingCheckPoints[int64(checkPoint.Checkpoint.ExecuteHeight)][int64(checkPoint.AuthorId)]; ok {
		log.Warn().Int64("executeHeight", int64(checkPoint.Checkpoint.ExecuteHeight)).Int64("authorId", int64(checkPoint.AuthorId)).Msg("PendingCheckPoint already exists")
		return
	}

	s.PendingCheckPoints[int64(checkPoint.Checkpoint.ExecuteHeight)][int64(checkPoint.AuthorId)] = checkPoint
	log.Info().Int64("executeHeight", int64(checkPoint.Checkpoint.ExecuteHeight)).Int64("authorId", int64(checkPoint.AuthorId)).Msg("AddPendingCheckPoint")
}

func (s *CheckPointStore) AddQuorumCheckPoint(checkPoint *rpc.QuorumCheckpoint) {
	if _, ok := s.QuorumCheckPoints[int64(checkPoint.Checkpoint.ExecuteHeight)]; ok {
		log.Warn().Int64("executeHeight", int64(checkPoint.Checkpoint.ExecuteHeight)).Msg("QuorumCheckPoint already exists")
		return
	}

	s.QuorumCheckPoints[int64(checkPoint.Checkpoint.ExecuteHeight)] = checkPoint
}
