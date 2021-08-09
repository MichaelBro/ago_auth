package authenticator

import (
	"bytes"
	"context"
	"github.com/go-chi/chi"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestAuthenticatorHTTPMux(t *testing.T) {
	mux := http.NewServeMux()
	authenticatorMd := Authenticator(func(ctx context.Context) (*string, error) {
		id := "192.0.2.1"
		return &id, nil
	}, func(ctx context.Context, id *string) (interface{}, error) {
		return "USER_AUTH", nil
	})
	mux.Handle(
		"/get",
		authenticatorMd(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
			profile, err := Authentication(request.Context())
			if err != nil {
				if err == ErrNoAuthentication {
					writer.WriteHeader(http.StatusUnauthorized)
					return
				}
				t.Fatal(err)
			}
			data := profile.(string)

			if data == "USER_AUTH" && request.RemoteAddr != "192.0.2.1" {
				writer.WriteHeader(http.StatusUnauthorized)
				return
			}

			_, err = writer.Write([]byte(data))
			if err != nil {
				t.Fatal(err)
			}
		})),
	)

	type args struct {
		method string
		path   string
		addr   string
	}

	tests := []struct {
		name        string
		args        args
		want_text   []byte
		want_status int
	}{
		{name: "GET", args: args{method: "GET", path: "/get", addr: "192.0.2.1"}, want_status: 200, want_text: []byte("USER_AUTH")},
		{name: "POST", args: args{method: "POST", path: "/get", addr: "192.0.2.1"}, want_status: 200, want_text: []byte("USER_AUTH")},
		{name: "POST_404", args: args{method: "POST", path: "/post", addr: "192.0.2.1"}, want_status: 404, want_text: []byte("404 page not found\n")},
		{name: "POST_401", args: args{method: "POST", path: "/get", addr: "127.0.0.1"}, want_status: 401, want_text: []byte{}},
	}

	for _, tt := range tests {
		request := httptest.NewRequest(tt.args.method, tt.args.path, nil)
		request.RemoteAddr = tt.args.addr
		response := httptest.NewRecorder()
		mux.ServeHTTP(response, request)
		gotCode := response.Code
		if tt.want_status != gotCode {
			t.Errorf("%s: got %d, want_status %d", tt.name, gotCode, tt.want_status)
		}
		gotBytes := response.Body.Bytes()
		if !bytes.Equal(tt.want_text, gotBytes) {
			t.Errorf("%s: got %s, want %s", tt.name, gotBytes, tt.want_text)
		}
	}
}

func TestAuthenticatorChi(t *testing.T) {
	router := chi.NewRouter()
	authenticatorMd := Authenticator(func(ctx context.Context) (*string, error) {
		id := "192.0.2.1"
		return &id, nil
	}, func(ctx context.Context, id *string) (interface{}, error) {
		return "USER_AUTH", nil
	})
	router.With(authenticatorMd).Get(
		"/get",
		func(writer http.ResponseWriter, request *http.Request) {
			profile, err := Authentication(request.Context())
			if err != nil {
				if err == ErrNoAuthentication {
					writer.WriteHeader(http.StatusUnauthorized)
					return
				}
				t.Fatal(err)
			}
			data := profile.(string)

			if data == "USER_AUTH" && request.RemoteAddr != "192.0.2.1" {
				writer.WriteHeader(http.StatusUnauthorized)
				return
			}

			_, err = writer.Write([]byte(data))
			if err != nil {
				t.Fatal(err)
			}
		},
	)

	type args struct {
		method string
		path   string
		addr   string
	}

	tests := []struct {
		name       string
		args       args
		wantStatus int
		wantText   []byte
	}{
		{name: "GET", args: args{method: "GET", path: "/get", addr: "192.0.2.1"}, wantStatus: 200, wantText: []byte("USER_AUTH")},
		{name: "POST", args: args{method: "POST", path: "/get", addr: "192.0.2.1"}, wantStatus: 405, wantText: []byte{}},
		{name: "POST_404", args: args{method: "POST", path: "/post", addr: "192.0.2.1"}, wantStatus: 404, wantText: []byte("404 page not found\n")},
		{name: "POST_401", args: args{method: "GET", path: "/get", addr: "127.0.0.1"}, wantStatus: 401, wantText: []byte{}},
	}

	for _, tt := range tests {
		request := httptest.NewRequest(tt.args.method, tt.args.path, nil)
		request.RemoteAddr = tt.args.addr
		response := httptest.NewRecorder()
		router.ServeHTTP(response, request)
		gotCode := response.Code
		if tt.wantStatus != gotCode {
			t.Errorf("%s: got %d, want_status %d", tt.name, gotCode, tt.wantStatus)
		}
		gotBytes := response.Body.Bytes()
		if !bytes.Equal(tt.wantText, gotBytes) {
			t.Errorf("%s: got %s, want %s", tt.name, gotBytes, tt.wantText)
		}
	}
}
