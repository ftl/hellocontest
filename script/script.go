package script

import (
	"context"
	"log"
	"time"

	"github.com/ftl/hellocontest/core/app"
)

type Step func(ctx context.Context, app *app.Controller, ui func(func())) time.Duration

type Script struct {
	steps []Step

	waitUntil time.Time
	nextStep  int
}

func (s *Script) Step(ctx context.Context, app *app.Controller, ui func(func())) bool {
	if s.nextStep >= len(s.steps) {
		return false
	}

	now := time.Now()
	if now.Before(s.waitUntil) {
		return true
	}

	nextDuration := s.steps[s.nextStep](ctx, app, ui)
	s.waitUntil = now.Add(nextDuration)
	s.nextStep += 1

	return true
}

func Describe(description string, delay time.Duration) Step {
	return func(_ context.Context, app *app.Controller, ui func(func())) time.Duration {
		ready := make(chan struct{})
		ui(func() {
			app.ShowInfo("%s\n\nin %v", description, delay)
			close(ready)
		})
		<-ready
		return delay
	}
}

func Wait(duration time.Duration) Step {
	return func(context.Context, *app.Controller, func(func())) time.Duration {
		log.Printf("[WAITING FOR %v]", duration)
		return duration
	}
}
