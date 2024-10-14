package store

import (
	"q/rpc"
	"sync"

	"github.com/rs/zerolog/log"
)

type CheckPointStore struct {
	signedCheckpoints map[int64]map[int64]*rpc.SignedCheckpoint // executeHeight -> id -> checkpoint
	quorumCheckPoints map[int64]*rpc.QuorumCheckpoint           // executeHeight -> checkpoint
	sync.Mutex
}

func NewCheckPointStore() *CheckPointStore {
	return &CheckPointStore{
		signedCheckpoints: make(map[int64]map[int64]*rpc.SignedCheckpoint),
		quorumCheckPoints: make(map[int64]*rpc.QuorumCheckpoint),
	}
}

func (s *CheckPointStore) AddSignedCheckpoint(checkPoint *rpc.SignedCheckpoint) {
	s.Lock()
	defer s.Unlock()
	if _, ok := s.signedCheckpoints[int64(checkPoint.Checkpoint.ExecuteHeight)]; !ok {
		s.signedCheckpoints[int64(checkPoint.Checkpoint.ExecuteHeight)] = make(map[int64]*rpc.SignedCheckpoint)
	}
	if _, ok := s.signedCheckpoints[int64(checkPoint.Checkpoint.ExecuteHeight)][int64(checkPoint.AuthorId)]; ok {
		log.Warn().Int64("executeHeight", int64(checkPoint.Checkpoint.ExecuteHeight)).Int64("authorId", int64(checkPoint.AuthorId)).Msg("PendingCheckPoint already exists")
		return
	}

	s.signedCheckpoints[int64(checkPoint.Checkpoint.ExecuteHeight)][int64(checkPoint.AuthorId)] = checkPoint
	log.Info().Int64("executeHeight", int64(checkPoint.Checkpoint.ExecuteHeight)).Int64("authorId", int64(checkPoint.AuthorId)).Msg("AddPendingCheckPoint")
}

func (s *CheckPointStore) AddQuorumCheckPoint(checkPoint *rpc.QuorumCheckpoint) {
	s.Lock()
	defer s.Unlock()
	if _, ok := s.quorumCheckPoints[int64(checkPoint.Checkpoint.ExecuteHeight)]; ok {
		log.Warn().Int64("executeHeight", int64(checkPoint.Checkpoint.ExecuteHeight)).Msg("QuorumCheckPoint already exists")
		return
	}

	s.quorumCheckPoints[int64(checkPoint.Checkpoint.ExecuteHeight)] = checkPoint
}
