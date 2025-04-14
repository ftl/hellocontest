package script

import (
	"context"
	"log"
	"time"

	"github.com/ftl/hellocontest/core/app"
)

type Script struct {
	sections []*Section

	currentSection int
}

type Section struct {
	enter Condition
	steps []Step

	waitUntil   time.Time
	currentStep int
}

type Step func(ctx context.Context, app *app.Controller, ui func(func())) time.Duration

type Condition func(ctx context.Context, app *app.Controller, ui func(func())) (bool, time.Duration)

func (s *Script) Step(ctx context.Context, app *app.Controller, ui func(func())) bool {
	if s.currentSection >= len(s.sections) {
		return false
	}

	section := s.sections[s.currentSection]

	cont := section.Step(ctx, app, ui)
	if !cont {
		s.currentSection += 1
	}

	return true
}

func (s *Section) Step(ctx context.Context, app *app.Controller, ui func(func())) bool {
	if s.currentStep == 0 {
		return s.checkEntryCondition(ctx, app, ui)
	}

	return s.executeStep(ctx, app, ui)
}

func (s *Section) checkEntryCondition(ctx context.Context, app *app.Controller, ui func(func())) bool {
	s.currentStep += 1
	if s.enter == nil {
		return true
	}

	now := time.Now()
	enter, delay := s.enter(ctx, app, ui)
	if delay != 0 {
		s.waitUntil = now.Add(delay)
	}

	return enter
}

func (s *Section) executeStep(ctx context.Context, app *app.Controller, ui func(func())) bool {
	if s.currentStep > len(s.steps) {
		return false
	}

	now := time.Now()
	if now.Before(s.waitUntil) {
		return true
	}

	step := s.steps[s.currentStep-1]

	nextDuration := step(ctx, app, ui)
	s.waitUntil = now.Add(nextDuration)
	s.currentStep += 1

	return true
}

func Ask(description string, delay time.Duration) Condition {
	return func(_ context.Context, app *app.Controller, ui func(func())) (bool, time.Duration) {
		enter := make(chan bool, 1)
		defer close(enter)
		ui(func() {
			enter <- app.ShowQuestion("%s\n\nin %v", description, delay)
		})
		doEnter := <-enter
		return doEnter, delay
	}
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
