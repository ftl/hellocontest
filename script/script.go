package script

import (
	"context"
	"log"
	"time"

	"github.com/ftl/hellocontest/core/app"
)

type Script struct {
	clock    *Clock
	sections []*Section

	currentSection int
}

type Section struct {
	enter       Condition
	steps       []Step
	alternative Step

	waitUntil   time.Time
	currentStep int
}

type Runtime struct {
	Clock *Clock
	App   *app.Controller
	UI    func(func())
}

type Step func(ctx context.Context, r *Runtime) time.Duration

type Condition func(ctx context.Context, r *Runtime) (bool, time.Duration)

func (s *Script) Now() time.Time {
	if s.clock == nil {
		return time.Now()
	}
	return s.clock.Now()
}

func (s *Script) Step(ctx context.Context, app *app.Controller, ui func(func())) bool {
	if s.currentSection >= len(s.sections) {
		return false
	}
	if s.clock == nil {
		s.clock = &Clock{}
	}

	section := s.sections[s.currentSection]
	runtime := &Runtime{
		Clock: s.clock,
		App:   app,
		UI:    ui,
	}

	cont := section.Step(ctx, runtime)
	if !cont {
		s.currentSection += 1
	}

	return true
}

func (s *Section) Step(ctx context.Context, r *Runtime) bool {
	if s.currentStep == 0 {
		return s.checkEntryCondition(ctx, r)
	}

	// TODO: execute the alternative if checkEntryCondition returns false

	return s.executeStep(ctx, r)
}

func (s *Section) checkEntryCondition(ctx context.Context, r *Runtime) bool {
	s.currentStep += 1
	if s.enter == nil {
		return true
	}

	now := time.Now()
	enter, delay := s.enter(ctx, r)
	if delay != 0 {
		s.waitUntil = now.Add(delay)
	}

	return enter
}

func (s *Section) executeStep(ctx context.Context, r *Runtime) bool {
	now := time.Now()
	if now.Before(s.waitUntil) {
		return true
	}

	if s.currentStep > len(s.steps) {
		return false
	}

	step := s.steps[s.currentStep-1]

	nextDuration := step(ctx, r)
	s.waitUntil = now.Add(nextDuration)
	s.currentStep += 1

	return true
}

func Ask(description string, delay time.Duration) Condition {
	return func(_ context.Context, r *Runtime) (bool, time.Duration) {
		enter := make(chan bool, 1)
		defer close(enter)
		r.UI(func() {
			enter <- r.App.ShowQuestion("%s\n\nin %v", description, delay)
		})
		doEnter := <-enter
		return doEnter, delay
	}
}

func Describe(description string, delay time.Duration) Step {
	return func(_ context.Context, r *Runtime) time.Duration {
		ready := make(chan struct{})
		r.UI(func() {
			r.App.ShowInfo("%s\n\nin %v", description, delay)
			close(ready)
		})
		<-ready
		return delay
	}
}

func Wait(duration time.Duration) Step {
	return func(context.Context, *Runtime) time.Duration {
		log.Printf("[WAITING FOR %v]", duration)
		return duration
	}
}

func SetTimebase(timebase string) Step {
	return func(ctx context.Context, r *Runtime) time.Duration {
		r.Clock.SetFromRFC3339(timebase)
		return 0
	}
}

func AddTimebaseOffset(offset time.Duration) Step {
	return func(ctx context.Context, r *Runtime) time.Duration {
		r.Clock.Add(offset)
		return 0
	}
}

func ResetTimebase() Step {
	return func(ctx context.Context, r *Runtime) time.Duration {
		r.Clock.Reset()
		return 0
	}
}
