package goFE

import "github.com/google/uuid"

type Component interface {
	Render() string
	GetID() uuid.UUID
	GetChildren() []Component
	GetKill(id uuid.UUID) chan bool
	AddKill(id uuid.UUID)
	KillAll()
	InitEventListeners()
}
