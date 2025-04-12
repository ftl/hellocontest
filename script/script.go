package script

import (
	"context"
	"fmt"
	"log"
	"os/exec"
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
			app.ShowInfo("[SCREENSHOT]\n\n%s\n\nin %v", description, delay)
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

func TriggerScreenshot() Step {
	return TriggerScreenshotWithDelay(0)
}

func TriggerScreenshotWithDelay(delay time.Duration) Step {
	return func(_ context.Context, _ *app.Controller, _ func(func())) time.Duration {
		// TODO: evaluate ctx.Done() and stop the flameshot process
		cmd := exec.Command("flameshot", "gui")
		if delay > 0 {
			cmd.Args = append(cmd.Args, "--delay", fmt.Sprintf("%d", delay.Milliseconds()))
		}

		err := cmd.Run()
		if err != nil {
			log.Printf("Screenshot failed: %v", err)
		} else {
			log.Println("Screenshot successful")
		}
		return 0
	}
}
