package app

import (
	"context"

	wailsruntime "github.com/wailsapp/wails/v2/pkg/runtime"
)

type EventEmitter struct {
	ctx context.Context
}

func NewEventEmitter(ctx context.Context) *EventEmitter {
	return &EventEmitter{ctx: ctx}
}

func (e *EventEmitter) Emit(eventName string, data interface{}) {
	wailsruntime.EventsEmit(e.ctx, eventName, data)
}

func (e *EventEmitter) SetContext(ctx context.Context) {
	e.ctx = ctx
}
