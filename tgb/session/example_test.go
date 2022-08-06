package session_test

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/mr-linch/go-tg"
	"github.com/mr-linch/go-tg/examples"
	"github.com/mr-linch/go-tg/tgb"
	"github.com/mr-linch/go-tg/tgb/session"
)

func ExampleNewManager() {
	type Session struct {
		MessagesCount int
	}

	sessionManager := session.NewManager(Session{})

	targetCount := 5

	router := tgb.NewRouter().
		Use(sessionManager).
		Message(func(ctx context.Context, mu *tgb.MessageUpdate) error {
			session := sessionManager.Get(ctx)
			session.MessagesCount++

			left := targetCount - session.MessagesCount

			if left == 0 {
				return mu.Answer("🏆 You are done!").DoVoid(ctx)
			} else {
				return mu.Answer(fmt.Sprintf("Keep going, left %d", left)).DoVoid(ctx)
			}
		})

	examples.Run(router)
}

func ExampleManager_Filter() {
	type Step int8

	const (
		StepInit Step = iota
		StepName
		StepEmail
		StepPhone
	)

	type Session struct {
		Step Step

		Name  string
		Email string
		Phone string
	}

	sessionManager := session.NewManager(Session{})

	isStepEqual := func(step Step) tgb.Filter {
		return sessionManager.Filter(func(s *Session) bool {
			return s.Step == step
		})
	}

	router := tgb.NewRouter().
		Use(sessionManager).
		Message(func(ctx context.Context, mu *tgb.MessageUpdate) error {
			sessionManager.Get(ctx).Step = StepName

			return mu.Answer("(1/3) What is your name?").DoVoid(ctx)
		}, isStepEqual(StepInit)).
		Message(func(ctx context.Context, mu *tgb.MessageUpdate) error {
			session := sessionManager.Get(ctx)
			session.Name = mu.Text

			session.Step = StepEmail
			return mu.Answer("(2/3) What is your email?").DoVoid(ctx)
		}, isStepEqual(StepName)).
		Message(func(ctx context.Context, mu *tgb.MessageUpdate) error {
			session := sessionManager.Get(ctx)
			session.Email = mu.Text

			session.Step = StepPhone
			return mu.Answer("3/3) What is your phone?").DoVoid(ctx)
		}, isStepEqual(StepEmail)).
		Message(func(ctx context.Context, mu *tgb.MessageUpdate) error {
			session := sessionManager.Get(ctx)
			session.Phone = mu.Text

			v, err := json.MarshalIndent(session, "", "  ")
			if err != nil {
				return err
			}

			sessionManager.Reset(session)

			return mu.Answer(tg.HTML.Line(
				tg.HTML.Bold("Your session:"),
				tg.HTML.Code(string(v)),
			)).ParseMode(tg.HTML).DoVoid(ctx)
		}, isStepEqual(StepPhone))

	examples.Run(router)
}
