package service

type IDGenerator interface {
	Generate() (string, error)
}

type ShortidIDGenerator struct {
	idGenerator IDGenerator
}

func NewShortidIDGenerator(generator IDGenerator) *ShortidIDGenerator {
	return &ShortidIDGenerator{
		idGenerator: generator,
	}
}

func (g *ShortidIDGenerator) Generate() (string, error) {
	return g.idGenerator.Generate()
}
