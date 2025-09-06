package repository

import "go.uber.org/zap"

type UUIDManager struct {
	current uint64
	logger  *zap.Logger
}

func NewUUIDManager(logger *zap.Logger) *UUIDManager {
	logger = logger.With(
		zap.String("component", "UUID manager"),
	)
	return &UUIDManager{
		current: 0,
		logger:  logger,
	}
}

func (um *UUIDManager) next() uint64 {
	um.current++
	um.logger.Debug("Next UUID for record", zap.Uint64("UUID", um.current))
	return um.current
}

func (um *UUIDManager) init(uuid uint64) {
	um.logger.Debug("Initializing current UUID", zap.Uint64("UUID", uuid))
	um.current = uuid
}
