package service

type IDGenerator interface {
	Generate() (string, error)
}

type ShortIDGenerator struct {
	idGenerator IDGenerator
}

func NewShortIDGenerator(generator IDGenerator) *ShortIDGenerator {
	return &ShortIDGenerator{
		idGenerator: generator,
	}
}

func (g *ShortIDGenerator) Generate() (string, error) {
	return g.idGenerator.Generate()
}
