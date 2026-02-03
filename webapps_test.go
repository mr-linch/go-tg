package tg

import (
	"net/url"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestAuthWidget_Query(t *testing.T) {
	for _, test := range []struct {
		Widget AuthWidget
		Want   url.Values
	}{
		{
			Widget: AuthWidget{
				ID:        UserID(1),
				FirstName: "John",
				AuthDate:  UnixTime(1546300800),
				Hash:      "hash",
			},

			Want: url.Values{
				"id":         []string{"1"},
				"first_name": []string{"John"},
				"auth_date":  []string{"1546300800"},
				"hash":       []string{"hash"},
			},
		},
		{
			Widget: AuthWidget{
				ID:        UserID(1),
				FirstName: "John",
				LastName:  "Doe",
				Username:  "jdoe",
				PhotoURL:  "https://example.com/photo.jpg",
				AuthDate:  UnixTime(1546300800),
				Hash:      "hash",
			},

			Want: url.Values{
				"id":         []string{"1"},
				"first_name": []string{"John"},
				"last_name":  []string{"Doe"},
				"username":   []string{"jdoe"},
				"photo_url":  []string{"https://example.com/photo.jpg"},
				"auth_date":  []string{"1546300800"},
				"hash":       []string{"hash"},
			},
		},
	} {
		got := test.Widget.Query()

		assert.Equal(t, test.Want, got)
	}
}

func TestParseAuthWidgetQuery(t *testing.T) {
	for _, test := range []struct {
		Name     string
		Values   url.Values
		Excepted *AuthWidget
		Error    bool
	}{
		{
			Name: "AllValid",
			Values: url.Values{
				"id":         []string{"1"},
				"first_name": []string{"John"},
				"last_name":  []string{"Doe"},
				"username":   []string{"jdoe"},
				"photo_url":  []string{"https://example.com/photo.jpg"},
				"auth_date":  []string{"1546300800"},
				"hash":       []string{"hash"},
			},
			Excepted: &AuthWidget{
				ID:        UserID(1),
				FirstName: "John",
				LastName:  "Doe",
				Username:  "jdoe",
				PhotoURL:  "https://example.com/photo.jpg",
				AuthDate:  UnixTime(1546300800),
				Hash:      "hash",
			},
			Error: false,
		},
		{
			Name: "InvalidID",
			Values: url.Values{
				"id":         []string{"invalid"},
				"first_name": []string{"John"},
				"last_name":  []string{"Doe"},
				"username":   []string{"jdoe"},
				"photo_url":  []string{"https://example.com/photo.jpg"},
				"auth_date":  []string{"1546300800"},
				"hash":       []string{"hash"},
			},
			Error: true,
		},
		{
			Name: "InvalidAuthDate",
			Values: url.Values{
				"id":         []string{"1"},
				"first_name": []string{"John"},
				"last_name":  []string{"Doe"},
				"username":   []string{"jdoe"},
				"photo_url":  []string{"https://example.com/photo.jpg"},
				"auth_date":  []string{"invalid"},
				"hash":       []string{"hash"},
			},
			Error: true,
		},
	} {
		t.Run(test.Name, func(t *testing.T) {
			got, err := ParseAuthWidgetQuery(test.Values)
			if test.Error {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			assert.Equal(t, test.Excepted, got)
		})
	}
}

// https://dcd3-91-235-226-78.eu.ngrok.io/login-url?id=103980787&first_name=Sasha&username=MrLinch&photo_url=https%3A%2F%2Ft.me%2Fi%2Fuserpic%2F320%2Fq9a3ePyQ_J58XivHA6pL7UOLZWvphbLgBqh3OLhmtrs.jpg&auth_date=1656790495&hash=d64920549aa64c3f69577e217e77b253ca383bf0b9945266ab5e096739250d2d

func TestAuthWidget_Signature(t *testing.T) {
	w := AuthWidget{
		ID:        103980787,
		FirstName: "Sasha",
		Username:  "MrLinch",
		PhotoURL:  "https://t.me/i/userpic/320/q9a3ePyQ_J58XivHA6pL7UOLZWvphbLgBqh3OLhmtrs.jpg",
		AuthDate:  UnixTime(1656790495),
		Hash:      "d64920549aa64c3f69577e217e77b253ca383bf0b9945266ab5e096739250d2d",
	}

	signature := w.Signature("5433024556:AAF63JW91kEl7k8bhqBzu86niebek4ldogg")

	assert.Equal(t, w.Hash, signature)
}

func TestAuthWidget_Valid(t *testing.T) {
	token := "5433024556:AAF63JW91kEl7k8bhqBzu86niebek4ldogg"

	t.Run("Ok", func(t *testing.T) {
		w := AuthWidget{
			ID:        103980787,
			FirstName: "Sasha",
			Username:  "MrLinch",
			PhotoURL:  "https://t.me/i/userpic/320/q9a3ePyQ_J58XivHA6pL7UOLZWvphbLgBqh3OLhmtrs.jpg",
			AuthDate:  UnixTime(1656790495),
			Hash:      "d64920549aa64c3f69577e217e77b253ca383bf0b9945266ab5e096739250d2d",
		}

		assert.True(t, w.Valid(token))
	})

	t.Run("False", func(t *testing.T) {
		w := AuthWidget{
			ID:        103980786,
			FirstName: "Sasha",
			Username:  "MrLinch",
			PhotoURL:  "https://t.me/i/userpic/320/q9a3ePyQ_J58XivHA6pL7UOLZWvphbLgBqh3OLhmtrs.jpg",
			AuthDate:  UnixTime(1656790495),
			Hash:      "d64920549aa64c3f69577e217e77b253ca383bf0b9945266ab5e096739250d2d",
		}

		assert.False(t, w.Valid(token))
	})
}

func TestAuthWidget_Time(t *testing.T) {
	now := time.Now().Truncate(time.Second)

	w := AuthWidget{
		AuthDate: UnixTime(now.Unix()),
	}

	assert.Equal(t, now, w.AuthDate.Time())
}

func TestParseWebAppInitData(t *testing.T) {
	for _, test := range []struct {
		Name     string
		Values   url.Values
		Excepted *WebAppInitData
		Error    bool
	}{
		{
			Name: "AllValid",
			Values: url.Values{
				"query_id":       []string{"1"},
				"user":           []string{`{"id": 1}`},
				"receiver":       []string{`{"id": 2}`},
				"chat":           []string{`{"id": 3}`},
				"start_param":    []string{"start_param"},
				"can_send_after": []string{"10"},
				"auth_date":      []string{"1546300800"},
				"hash":           []string{"hash"},
			},
			Excepted: &WebAppInitData{
				QueryID:      "1",
				User:         &WebAppUser{ID: 1},
				Receiver:     &WebAppUser{ID: 2},
				Chat:         &WebAppChat{ID: 3},
				StartParam:   "start_param",
				CanSendAfter: 10,
				AuthDate:     UnixTime(1546300800),
				Hash:         "hash",
			},
		},
		{
			Name: "NoQueryID",
			Values: url.Values{
				"user":           []string{`{"id": 1}`},
				"receiver":       []string{`{"id": 2}`},
				"chat":           []string{`{"id": 3}`},
				"start_param":    []string{"start_param"},
				"can_send_after": []string{"10"},
				"auth_date":      []string{"1546300800"},
				"hash":           []string{"hash"},
			},
			Error: true,
		},
		{
			Name: "NoQueryID",
			Values: url.Values{
				"query_id":       []string{"1"},
				"user":           []string{`{"id": 1}`},
				"receiver":       []string{`{"id": 2}`},
				"chat":           []string{`{"id": 3}`},
				"start_param":    []string{"start_param"},
				"can_send_after": []string{"10"},
				"auth_date":      []string{"1546300800"},
			},
			Error: true,
		},
		{
			Name: "InvalidUser",
			Values: url.Values{
				"query_id":       []string{"1"},
				"user":           []string{`{"id": "asda"}`},
				"receiver":       []string{`{"id": 2}`},
				"chat":           []string{`{"id": 3}`},
				"start_param":    []string{"start_param"},
				"can_send_after": []string{"invalid"},
				"auth_date":      []string{"1546300800"},
				"hash":           []string{"hash"},
			},
			Error: true,
		},
		{
			Name: "InvalidReceiver",
			Values: url.Values{
				"query_id":       []string{"1"},
				"user":           []string{`{"id": 1}`},
				"receiver":       []string{`{"id": "asdv"}`},
				"chat":           []string{`{"id": 3}`},
				"start_param":    []string{"start_param"},
				"can_send_after": []string{"invalid"},
				"auth_date":      []string{"1546300800"},
				"hash":           []string{"hash"},
			},
			Error: true,
		},
		{
			Name: "InvalidChat",
			Values: url.Values{
				"query_id":       []string{"1"},
				"user":           []string{`{"id": 1}`},
				"receiver":       []string{`{"id": 2}`},
				"chat":           []string{`{"id": "asdv"}`},
				"start_param":    []string{"start_param"},
				"can_send_after": []string{"invalid"},
				"auth_date":      []string{"1546300800"},
				"hash":           []string{"hash"},
			},
			Error: true,
		},
		{
			Name: "InvalidCanSendAfter",
			Values: url.Values{
				"query_id":       []string{"1"},
				"user":           []string{`{"id": 1}`},
				"receiver":       []string{`{"id": 2}`},
				"chat":           []string{`{"id": 3}`},
				"start_param":    []string{"start_param"},
				"can_send_after": []string{"invalid"},
				"auth_date":      []string{"1546300800"},
				"hash":           []string{"hash"},
			},
			Error: true,
		},
		{
			Name: "InvalidAuthDate",
			Values: url.Values{
				"query_id":       []string{"1"},
				"user":           []string{`{"id": 1}`},
				"receiver":       []string{`{"id": 2}`},
				"chat":           []string{`{"id": 3}`},
				"start_param":    []string{"start_param"},
				"can_send_after": []string{"10"},
				"auth_date":      []string{"invalid"},
				"hash":           []string{"hash"},
			},
			Error: true,
		},
	} {
		t.Run(test.Name, func(t *testing.T) {
			if test.Excepted != nil {
				test.Excepted.raw = test.Values
			}

			got, err := ParseWebAppInitData(test.Values)

			if test.Error {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			assert.Equal(t, test.Excepted, got)
		})
	}
}

func TestWebAppInitData_Valid(t *testing.T) {
	const token = "5433024556:AAF63JW91kEl7k8bhqBzu86niebek4ldogg"

	t.Run("True", func(t *testing.T) {
		vs, err := url.ParseQuery("query_id=AAHznjIGAAAAAPOeMgZUHjBo&user=%7B%22id%22%3A103980787%2C%22first_name%22%3A%22Sasha%22%2C%22last_name%22%3A%22%22%2C%22username%22%3A%22MrLinch%22%2C%22language_code%22%3A%22uk%22%7D&auth_date=1656798871&hash=8c59e353f627a5c67d41f8a2e8f8c12d9e0fbec8ac44680d779ebed3c326a41a")
		assert.NoError(t, err)

		got, err := ParseWebAppInitData(vs)
		assert.NoError(t, err)
		assert.NotNil(t, got)

		assert.True(t, got.Valid(token))
	})
	t.Run("False", func(t *testing.T) {
		vs, err := url.ParseQuery("query_id=AAAHznjIGAAAAAPOeMgZUHjBo&user=%7B%22id%22%3A103980787%2C%22first_name%22%3A%22Sasha%22%2C%22last_name%22%3A%22%22%2C%22username%22%3A%22MrLinch%22%2C%22language_code%22%3A%22uk%22%7D&auth_date=1656798871&hash=8c59e353f627a5c67d41f8a2e8f8c12d9e0fbec8ac44680d779ebed3c326a41a")
		assert.NoError(t, err)

		got, err := ParseWebAppInitData(vs)
		assert.NoError(t, err)
		assert.NotNil(t, got)

		assert.False(t, got.Valid(token))
	})
}
