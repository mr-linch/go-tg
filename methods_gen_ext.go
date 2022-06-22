package tg

import "context"

//go:generate go run github.com/mr-linch/go-tg-gen@latest -methods-output methods_gen.go

// Me returns cached current bot info.
func (client *Client) Me(ctx context.Context) (User, error) {
	if client.me == nil {
		user, err := client.GetMe().Do(ctx)
		if err != nil {
			return User{}, err
		}
		client.me = &user
	}
	return *client.me, nil
}
