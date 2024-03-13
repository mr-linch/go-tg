// Package contains simple echo bot, that demonstrates how to use handlers, filters and file uploads.
package main

import (
	"context"
	"fmt"
	"net/http"
	"strconv"

	"github.com/mr-linch/go-tg"
	"github.com/mr-linch/go-tg/examples"
	"github.com/mr-linch/go-tg/tgb"
)

type userDetailsCallbackData struct {
	UserID int
}

type userLocationCallbackData struct {
	UserID int
	Lat    float64
	Lng    float64
}

type postDetailsCallbackData struct {
	UserID int
	PostID int
}

type commentDetailsCallbackData struct {
	UserID    int
	PostID    int
	CommentID int
}

var (
	userDetailsCallbackDataFilter = tgb.NewCallbackDataFilter[userDetailsCallbackData](
		"user_details",
	)

	userLocationCallbackDataFilter = tgb.NewCallbackDataFilter[userLocationCallbackData](
		"user_location",
	)

	userListCallbackDataFilter = tgb.NewCallbackDataFilter[struct{}](
		"user_list",
	)

	postDetailsCallbackDataFilter = tgb.NewCallbackDataFilter[postDetailsCallbackData](
		"post_details",
	)

	commentDetailsCallbackDataFilter = tgb.NewCallbackDataFilter[commentDetailsCallbackData](
		"comment_details",
	)
)

func newUserListMessage(pm tg.ParseMode, users []User) *tgb.TextMessageCallBuilder {
	buttons := make([]tg.InlineKeyboardButton, 0, len(users))
	for _, user := range users {
		buttons = append(buttons, userDetailsCallbackDataFilter.MustButton(
			user.Name,
			userDetailsCallbackData{UserID: user.ID},
		))
	}

	return tgb.NewTextMessageCallBuilder(
		pm.Text(
			pm.Bold("👥 Users"),
			"",
			pm.Line("Total users: ", strconv.Itoa(len(users))),
			"",
			pm.Italic("Select user to view details:"),
		),
	).
		ParseMode(pm).
		ReplyMarkup(
			tg.NewInlineKeyboardMarkup(
				tg.NewButtonLayout(2, buttons...).Keyboard()...,
			),
		)
}

func newUserDetailsMessage(pm tg.ParseMode, user User, posts []Post) *tgb.TextMessageCallBuilder {
	buttons := make([]tg.InlineKeyboardButton, 0, len(posts)+1)

	for _, post := range posts {
		buttons = append(buttons, postDetailsCallbackDataFilter.MustButton(
			post.Title,
			postDetailsCallbackData{PostID: post.ID, UserID: user.ID},
		))
	}

	layout := tg.NewButtonLayout[tg.InlineKeyboardButton](2)

	layout.Row(userLocationCallbackDataFilter.MustButton("📍 Location", userLocationCallbackData{
		UserID: user.ID,
		Lat:    user.Address.Geo.Lat,
		Lng:    user.Address.Geo.Lng,
	}))

	layout.Add(buttons...)

	layout.Row(userListCallbackDataFilter.MustButton("🔙 Back", struct{}{}))

	buttons = append(buttons, userListCallbackDataFilter.MustButton("🔙 Back", struct{}{}))

	return tgb.NewTextMessageCallBuilder(
		pm.Text(
			pm.Bold("👤 User Details"),
			"",
			pm.Line("ID: ", strconv.Itoa(user.ID)),
			pm.Line("Name: ", user.Name),
			pm.Line("Username: ", user.Username),
			pm.Line("Email: ", user.Email),
			"",
			pm.Bold("Address:"),
			pm.Line("Street: ", user.Address.Street),
			pm.Line("Suite: ", user.Address.Suite),
			pm.Line("City: ", user.Address.City),
			pm.Line("Zipcode: ", user.Address.Zipcode),
			"",
			pm.Line("Phone: ", user.Phone),
			pm.Line("Website: ", user.Website),
			"",
			pm.Bold("Company:"),
			pm.Line("Name: ", user.Company.Name),
			pm.Line("Catch Phrase: ", user.Company.CatchPhrase),
			pm.Line("Bs: ", user.Company.Bs),
		),
	).
		ReplyMarkup(tg.NewInlineKeyboardMarkup(
			layout.Keyboard()...,
		)).
		ParseMode(pm)
}

func newPostDetails(pm tg.ParseMode, userID int, post Post, comments []Comment) *tgb.TextMessageCallBuilder {
	buttons := make([]tg.InlineKeyboardButton, 0, len(comments)+1)

	for _, comment := range comments {
		buttons = append(buttons, commentDetailsCallbackDataFilter.MustButton("💬 "+comment.Name, commentDetailsCallbackData{
			UserID:    userID,
			PostID:    post.ID,
			CommentID: comment.ID,
		}))
	}

	buttons = append(buttons, userDetailsCallbackDataFilter.MustButton("🔙 Back", userDetailsCallbackData{
		UserID: userID,
	}))

	return tgb.NewTextMessageCallBuilder(
		pm.Text(
			pm.Bold("📝 Post Details"),
			"",
			pm.Line(pm.Bold("ID: "), strconv.Itoa(post.ID)),
			pm.Line(pm.Bold("Title: "), post.Title),
			"",
			pm.Blockquote(post.Body),
		),
	).
		ParseMode(pm).
		ReplyMarkup(tg.NewInlineKeyboardMarkup(
			tg.NewButtonLayout(1, buttons...).Keyboard()...,
		))
}

func newCommentDetails(pm tg.ParseMode, userID int, postID int, comment Comment) *tgb.TextMessageCallBuilder {
	buttons := []tg.InlineKeyboardButton{
		postDetailsCallbackDataFilter.MustButton("🔙 Back to Post", postDetailsCallbackData{
			UserID: userID,
			PostID: postID,
		}),

		userDetailsCallbackDataFilter.MustButton("🔙 Back to User", userDetailsCallbackData{
			UserID: userID,
		}),
	}

	return tgb.NewTextMessageCallBuilder(
		pm.Text(
			pm.Bold("💬 Comment Details"),
			"",
			pm.Line(pm.Bold("ID: "), strconv.Itoa(comment.ID)),
			pm.Line(pm.Bold("Name: "), comment.Name),
			pm.Line(pm.Bold("Email: "), comment.Email),
			"",
			pm.Blockquote(comment.Body),
		),
	).
		ParseMode(pm).
		ReplyMarkup(tg.NewInlineKeyboardMarkup(
			tg.NewButtonLayout(1, buttons...).Keyboard()...,
		))
}

func main() {
	client := API{
		BaseURL: "https://jsonplaceholder.typicode.com",
		Client:  http.DefaultClient,
	}

	newUserListBuilder := func(ctx context.Context) (*tgb.TextMessageCallBuilder, error) {
		users, err := client.Users(ctx)
		if err != nil {
			return nil, fmt.Errorf("get users: %w", err)
		}

		return newUserListMessage(tg.HTML, users), nil
	}

	examples.Run(tgb.NewRouter().
		// start message and cbq handlers
		Message(func(ctx context.Context, msg *tgb.MessageUpdate) error {
			builder, err := newUserListBuilder(ctx)
			if err != nil {
				return err
			}

			return msg.Update.Reply(ctx, builder.AsSend(msg.Chat))
		}, tgb.Command("start")).
		CallbackQuery(func(ctx context.Context, cbq *tgb.CallbackQueryUpdate) error {
			_ = cbq.Answer().DoVoid(ctx)

			builder, err := newUserListBuilder(ctx)
			if err != nil {
				return err
			}

			return cbq.Update.Reply(ctx, builder.AsEditTextFromCBQ(cbq.CallbackQuery))
		}, userListCallbackDataFilter.Filter()).
		CallbackQuery(userDetailsCallbackDataFilter.Handler(func(ctx context.Context, cbq *tgb.CallbackQueryUpdate, cbd userDetailsCallbackData) error {
			_ = cbq.Answer().DoVoid(ctx)

			user, err := client.User(ctx, cbd.UserID)
			if err != nil {
				return fmt.Errorf("get user: %w", err)
			}

			posts, err := client.Posts(ctx, &PostsParams{
				UserID: cbd.UserID,
			})
			if err != nil {
				return fmt.Errorf("get posts: %w", err)
			}

			return cbq.Update.Reply(ctx,
				newUserDetailsMessage(tg.HTML, user, posts).
					AsEditTextFromCBQ(cbq.CallbackQuery),
			)
		}), userDetailsCallbackDataFilter.Filter()).
		CallbackQuery(userLocationCallbackDataFilter.Handler(func(ctx context.Context, cbq *tgb.CallbackQueryUpdate, cbd userLocationCallbackData) error {
			_ = cbq.Answer().DoVoid(ctx)

			return cbq.Update.Reply(ctx,
				tg.NewSendLocationCall(
					cbq.Message.Chat(),
					cbd.Lat,
					cbd.Lng,
				),
			)
		}), userLocationCallbackDataFilter.Filter()).
		CallbackQuery(postDetailsCallbackDataFilter.Handler(func(ctx context.Context, cbq *tgb.CallbackQueryUpdate, cbd postDetailsCallbackData) error {
			_ = cbq.Answer().DoVoid(ctx)

			post, err := client.Post(ctx, cbd.PostID)
			if err != nil {
				return fmt.Errorf("get post: %w", err)
			}

			comments, err := client.Comments(ctx, &CommentsParams{
				PostID: cbd.PostID,
			})
			if err != nil {
				return fmt.Errorf("get comments: %w", err)
			}

			return cbq.Update.Reply(ctx,
				newPostDetails(tg.HTML, cbd.UserID, post, comments).
					AsEditTextFromCBQ(cbq.CallbackQuery),
			)
		}), postDetailsCallbackDataFilter.Filter()).
		CallbackQuery(commentDetailsCallbackDataFilter.Handler(func(ctx context.Context, cbq *tgb.CallbackQueryUpdate, cbd commentDetailsCallbackData) error {
			_ = cbq.Answer().DoVoid(ctx)

			comment, err := client.Comment(ctx, cbd.CommentID)
			if err != nil {
				return fmt.Errorf("get comment: %w", err)
			}

			return cbq.Update.Reply(ctx,
				newCommentDetails(tg.HTML, cbd.UserID, cbd.PostID, comment).
					AsEditTextFromCBQ(cbq.CallbackQuery),
			)
		}), commentDetailsCallbackDataFilter.Filter()),
	)
}
