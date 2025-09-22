package repository

import "go.uber.org/zap"

type UUIDManager struct {
	current uint64
	logger  *zap.Logger
}

func NewUUIDManager(logger *zap.Logger) *UUIDManager {
	return &UUIDManager{
		current: 0,
		logger:  logger,
	}
}

func (um *UUIDManager) next() uint64 {
	um.current++
	return um.current
}

func (um *UUIDManager) init(uuid uint64) {
	um.current = uuid
}
