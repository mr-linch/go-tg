// Package contains simple echo bot, that demonstrates how to use handlers, filters and file uploads.
package main

import (
	"context"
	"fmt"
	"regexp"
	"strconv"

	"github.com/mr-linch/go-tg"
	"github.com/mr-linch/go-tg/_examples/runner"
	"github.com/mr-linch/go-tg/tgb"
	"github.com/mr-linch/go-tg/tgb/session"
)

type SessionStep int8

const (
	SessionStepInit = iota
	SessionStepName
	SessionStepAge
	SessionStepGender
)

var genders = []string{
	"Male",
	"Female",
	"Attack Helicopter",
	"Other",
}

type Session struct {
	Step SessionStep

	Name   string
	Age    int
	Gender string
}

func main() {
	// create session manager with default session value
	sessionManager := session.NewManager(Session{
		Step: SessionStepInit,
	}, session.WithStore(
		session.NewStoreFile("sessions"),
	))

	isSessionStep := func(state SessionStep) tgb.Filter {
		return sessionManager.Filter(func(session *Session) bool {
			return session.Step == state
		})
	}

	isDigit := tgb.Regexp(regexp.MustCompile(`^\d+$`))

	runner.Run(tgb.NewRouter().
		Use(sessionManager).
		Message(func(ctx context.Context, msg *tgb.MessageUpdate) error {
			// handle /start command
			sessionManager.Get(ctx).Step = SessionStepName
			return msg.Update.Reply(ctx, msg.Answer("Hi, what is your name?"))
		}, tgb.Command("start")).
		Message(func(ctx context.Context, mu *tgb.MessageUpdate) error {
			// handle no command with SessionStepInitial
			return mu.Update.Reply(ctx, mu.Answer("Press /start to fill the form"))
		}, isSessionStep(SessionStepInit)).
		Message(func(ctx context.Context, msg *tgb.MessageUpdate) error {
			// handle name input
			sess := sessionManager.Get(ctx)

			sess.Name = msg.Text
			sess.Step = SessionStepAge

			return msg.Update.Reply(ctx, msg.Answer("What is your age?"))
		}, isSessionStep(SessionStepName)).
		Message(func(ctx context.Context, msg *tgb.MessageUpdate) error {
			// handle no digit input when state is SessionStepAge
			return msg.Update.Reply(ctx, msg.Answer("Please, send me just number"))
		}, isSessionStep(SessionStepAge), tgb.Not(isDigit)).
		Message(func(ctx context.Context, msg *tgb.MessageUpdate) error {
			// handle correct age input
			age, err := strconv.Atoi(msg.Text)
			if err != nil {
				return fmt.Errorf("parse age: %w", err)
			}

			sess := sessionManager.Get(ctx)
			sess.Age = age
			sess.Step = SessionStepGender

			kb := tg.NewReplyKeyboard().Resize()
			for _, gender := range genders {
				kb.Text(gender)
			}

			return msg.Update.Reply(ctx, msg.Answer("What is your gender?").ReplyMarkup(kb.Adjust(1)))
		}, isSessionStep(SessionStepAge), isDigit).
		Message(func(ctx context.Context, mu *tgb.MessageUpdate) error {
			// handle gender input and display results
			sess := sessionManager.Get(ctx)

			sess.Gender = mu.Text

			answer := mu.Answer(tg.HTML.Text(
				tg.HTML.Line(tg.HTML.Underline(tg.HTML.Text("Your profile:"))),
				tg.HTML.Line(tg.HTML.Bold("├ Your name:"), tg.HTML.Code(sess.Name)),
				tg.HTML.Line(tg.HTML.Bold("├ Your age:"), tg.HTML.Code(strconv.Itoa(sess.Age))),
				tg.HTML.Line(tg.HTML.Bold("└ Your gender:"), tg.HTML.Code(sess.Gender)),
				"",
				tg.HTML.Line(tg.HTML.Italic("press /start to fill again")),
			)).ReplyMarkup(tg.NewReplyKeyboardRemove()).ParseMode(tg.HTML)

			sessionManager.Reset(sess)

			return mu.Update.Reply(ctx, answer)
		}, isSessionStep(SessionStepGender), tgb.TextIn(genders)).
		Message(func(ctx context.Context, msg *tgb.MessageUpdate) error {
			return msg.Update.Reply(ctx, msg.Answer("Please, choose one of the buttons below 👇"))
		}, isSessionStep(SessionStepGender), tgb.Not(tgb.TextIn(genders))),
	)
}
