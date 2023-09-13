package goFE

import "github.com/google/uuid"

type Component interface {
	Render() string
	GetID() uuid.UUID
	GetChildren() []Component
	GetKill() chan bool
	InitEventListeners()
}
