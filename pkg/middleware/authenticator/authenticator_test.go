package authenticator

import (
	"bytes"
	"context"
	"errors"
	"github.com/go-chi/chi"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)
var ErrNoId = errors.New("No id")
func TestAuthenticatorHTTPMux(t *testing.T) {
	mux := http.NewServeMux()
	authenticatorMd := Authenticator(
		func(ctx context.Context)(*string, error){
			var id string
			p_id := ctx.Value("id")
			if p_id == ""{
				return &id, ErrNoId
			}
			id = p_id.(string)
			return &id, nil
		}, func(ctx context.Context, id *string) (interface{}, error) {
			if strings.Compare(*id, "0.0.0.0")==0{
				return "USERAUTH", nil
			}
			return "", ErrNoAuthentication
		})
	mux.Handle(
		"/get",
		authenticatorMd(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
			profile, err := Authentication(request.Context())
			if err != nil {
				t.Fatal(err)
			}
			data := profile.(string)

			_, err = writer.Write([]byte(data))
			if err != nil {
				t.Fatal(err)
			}
		})),
	)

	type args struct {
		method string
		path   string
		addr string
	}

	tests := []struct {
		name string
		args args
		want_text []byte
		want_status int
	}{
		{name: "OK", args: args{method: "GET", path: "/get", addr: "0.0.0.0"}, want_text: []byte("USERAUTH"), want_status: http.StatusOK},
		{name: "FAIL_1", args: args{method: "GET", path: "/get", addr: "0.0.0.1"}, want_text: []byte(""), want_status: http.StatusUnauthorized},
		{name: "FAIL_2", args: args{method: "GET", path: "/get", }, want_text: []byte(""), want_status: http.StatusUnauthorized},

	}

	for _, tt := range tests {
		request := httptest.NewRequest(tt.args.method, tt.args.path, nil)
		response := httptest.NewRecorder()
		ctx := context.WithValue(request.Context(), "id", tt.args.addr)
		request = request.WithContext(ctx)
		mux.ServeHTTP(response, request)
		got_text := response.Body.Bytes()
		got_status := response.Code
		if(got_status != tt.want_status){
			t.Errorf("got %d, want %d", got_status, tt.want_status)
		}
		if !bytes.Equal(tt.want_text, got_text) {
			t.Errorf("got %s, want %s", got_text, tt.want_text)
		}
	}
}

func TestAuthenticatorChi(t *testing.T) {
	router := chi.NewRouter()
	authenticatorMd := Authenticator(func(ctx context.Context)(*string, error){
		var id string
		p_id := ctx.Value("id")
		if p_id == ""{
			return &id, ErrNoId
		}
		id = p_id.(string)
		return &id, nil
	}, func(ctx context.Context, id *string) (interface{}, error) {
		if strings.Compare(*id, "0.0.0.0")==0{
			return "USERAUTH", nil
		}
		return "", ErrNoAuthentication
	})
	router.With(authenticatorMd).Get(
		"/get",
		func(writer http.ResponseWriter, request *http.Request) {
			profile, err := Authentication(request.Context())
			if err != nil {
				t.Fatal(err)
			}
			data := profile.(string)

			_, err = writer.Write([]byte(data))
			if err != nil {
				t.Fatal(err)
			}
		},
	)
	type args struct {
		method string
		path   string
		addr string
	}

	tests := []struct {
		name string
		args args
		want_text []byte
		want_status int
	}{
		{name: "OK", args: args{method: "GET", path: "/get", addr: "0.0.0.0"}, want_text: []byte("USERAUTH"), want_status: http.StatusOK},
		{name: "FAIL_1", args: args{method: "GET", path: "/get", addr: "0.0.0.1"}, want_text: []byte(""), want_status: http.StatusUnauthorized},
		{name: "FAIL_2", args: args{method: "GET", path: "/get", }, want_text: []byte(""), want_status: http.StatusUnauthorized},

	}

	for _, tt := range tests {
		request := httptest.NewRequest(tt.args.method, tt.args.path, nil)
		response := httptest.NewRecorder()
		ctx := context.WithValue(request.Context(), "id", tt.args.addr)
		request = request.WithContext(ctx)
		router.ServeHTTP(response, request)
		got_text := response.Body.Bytes()
		got_status := response.Code
		if(got_status != tt.want_status){
			t.Errorf("got %d, want %d", got_status, tt.want_status)
		}
		if !bytes.Equal(tt.want_text, got_text) {
			t.Errorf("got %s, want %s", got_text, tt.want_text)
		}
	}
}