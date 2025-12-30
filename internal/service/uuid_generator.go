package service

import "github.com/google/uuid"

// UUIDGenerator gera IDs usando UUID v4
type UUIDGenerator struct{}

// NewUUIDGenerator cria uma nova instância
func NewUUIDGenerator() *UUIDGenerator {
	return &UUIDGenerator{}
}

// Generate gera um novo UUID
func (g *UUIDGenerator) Generate() string {
	return uuid.New().String()
}
