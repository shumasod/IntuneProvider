package client_test

import (
	"testing"

	"github.com/shumasod/IntuneProvider/internal/client"
)

func TestIsNotFound_withNotFoundMessage(t *testing.T) {
	cases := []struct {
		err  error
		want bool
	}{
		{nil, false},
		{errorf("Graph API returned status 404: not found"), true},
		{errorf("Graph API returned status 403: forbidden"), false},
		{errorf("something else entirely"), false},
		{errorf("error 404 happened"), true},
	}
	for _, tc := range cases {
		got := client.IsNotFound(tc.err)
		if got != tc.want {
			t.Errorf("IsNotFound(%v) = %v, want %v", tc.err, got, tc.want)
		}
	}
}

func TestNewClient_invalidCredentials_returnsError(t *testing.T) {
	// 不正な認証情報ではクライアント作成は成功するが、
	// 実際のAPI呼び出し時にエラーになる（azidentity は遅延検証）
	_, err := client.NewClient("invalid-tenant", "invalid-client", "invalid-secret")
	if err != nil {
		t.Errorf("NewClient() with invalid creds should not error at construction: %v", err)
	}
}

func TestNewClient_emptyCredentials(t *testing.T) {
	// 空文字の場合は azidentity がエラーを返す
	_, err := client.NewClient("", "", "")
	if err == nil {
		t.Error("NewClient() with empty credentials should return error")
	}
}

// errorf は error インターフェースを実装するシンプルなヘルパー
type testError struct{ msg string }

func (e *testError) Error() string { return e.msg }

func errorf(msg string) error { return &testError{msg: msg} }
