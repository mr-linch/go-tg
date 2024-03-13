package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
)

type API struct {
	BaseURL string
	Client  *http.Client
}

type User struct {
	ID       int    `json:"id"`
	Name     string `json:"name"`
	Username string `json:"username"`
	Email    string `json:"email"`
	Address  struct {
		Street  string `json:"street"`
		Suite   string `json:"suite"`
		City    string `json:"city"`
		Zipcode string `json:"zipcode"`
		Geo     struct {
			Lat float64 `json:"lat,string"`
			Lng float64 `json:"lng,string"`
		} `json:"geo"`
	} `json:"address"`
	Phone   string `json:"phone"`
	Website string `json:"website"`
	Company struct {
		Name        string `json:"name"`
		CatchPhrase string `json:"catchPhrase"`
		Bs          string `json:"bs"`
	} `json:"company"`
}

type Post struct {
	UserID int    `json:"userId"`
	ID     int    `json:"id"`
	Title  string `json:"title"`
	Body   string `json:"body"`
}

type Comment struct {
	PostID int    `json:"postId"`
	ID     int    `json:"id"`
	Name   string `json:"name"`
	Email  string `json:"email"`
	Body   string `json:"body"`
}

func (a *API) request(ctx context.Context, path string, params url.Values, dst any) error {
	endpoint, err := url.JoinPath(a.BaseURL, path)
	if err != nil {
		return fmt.Errorf("build endpoint: %w", err)
	}

	if len(params) > 0 {
		endpoinAsURL, err := url.Parse(endpoint)
		if err != nil {
			return fmt.Errorf("parse endpoint: %w", err)
		}

		endpoinAsURL.RawQuery = params.Encode()

		endpoint = endpoinAsURL.String()
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, http.NoBody)
	if err != nil {
		return fmt.Errorf("build http request: %w", err)
	}

	res, err := a.Client.Do(req)
	if err != nil {
		return fmt.Errorf("execute http request: %w", err)
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		return fmt.Errorf("http status: %d", res.StatusCode)
	}

	if err := json.NewDecoder(res.Body).Decode(dst); err != nil {
		return fmt.Errorf("unmarshal: %w", err)
	}

	return nil
}

func (a *API) Users(ctx context.Context) (users []User, err error) {
	err = a.request(ctx, "/users", nil, &users)
	return
}

func (a *API) User(ctx context.Context, id int) (user User, err error) {
	err = a.request(ctx, fmt.Sprintf("/users/%d", id), nil, &user)
	return
}

type PostsParams struct {
	UserID int `json:"userId"`
}

func (a *API) Posts(ctx context.Context, params *PostsParams) (posts []Post, err error) {
	vs := url.Values{}

	if params != nil {
		if params.UserID != 0 {
			vs.Set("userId", strconv.Itoa(params.UserID))
		}
	}

	err = a.request(ctx, "/posts", vs, &posts)

	return
}

func (a *API) Post(ctx context.Context, id int) (post Post, err error) {
	err = a.request(ctx, fmt.Sprintf("/posts/%d", id), nil, &post)
	return
}

type CommentsParams struct {
	PostID int `json:"postId"`
}

func (a *API) Comments(ctx context.Context, params *CommentsParams) (comments []Comment, err error) {
	vs := url.Values{}

	if params != nil {
		if params.PostID != 0 {
			vs.Set("postId", strconv.Itoa(params.PostID))
		}
	}

	err = a.request(ctx, "/comments", vs, &comments)

	return
}

func (a *API) Comment(ctx context.Context, id int) (comment Comment, err error) {
	err = a.request(ctx, fmt.Sprintf("/comments/%d", id), nil, &comment)
	return
}
