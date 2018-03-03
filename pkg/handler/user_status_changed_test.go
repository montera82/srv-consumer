package handler

import (
	"context"
	"errors"
	"testing"

	"github.com/rafaeljesus/srv-consumer/pkg"
	"github.com/rafaeljesus/srv-consumer/pkg/mock"
	"github.com/rafaeljesus/srv-consumer/pkg/platform/message"
)

func TestUserStatusChanged(t *testing.T) {
	t.Parallel()

	tests := []struct {
		scenario string
		function func(*testing.T, *mock.UserStore, *mock.Acknowledger)
	}{
		{
			scenario: "When valid payload is supplied, then should successfully save user",
			function: testShouldSuccessfullyChangeUserStatus,
		},
		{
			scenario: "When invalid payload is supplied, then should fail to unmarshal body",
			function: testStatusChangeHandlerShouldFailToUnmarshalBody,
		},
		{
			scenario: "When Not found user is supplied, then should fail to save",
			function: testStatusChangeHandlerNotFoundError,
		},

		{
			scenario: "When unexpected error occurs, should be handled properly",
			function: testStatusChangeHandlerUnexpectedSaveError,
		},
	}

	for _, test := range tests {
		t.Run(test.scenario, func(t *testing.T) {
			store := new(mock.UserStore)
			acker := new(mock.Acknowledger)
			test.function(t, store, acker)

		})
	}
}

func testShouldSuccessfullyChangeUserStatus(t *testing.T, store *mock.UserStore, acker *mock.Acknowledger) {
	store.SaveFunc = func(user *pkg.User) error { return nil }
	acker.AckFunc = func(multiple bool) error { return nil }

	body := []byte(`{
		"email": "foo@mail.com",
		"username": "foo",
        "status": "active"
	}`)

	msg := message.New(acker, body)
	handler := NewUserStatusChanged(store)
	err := handler.Handle(context.Background(), msg)

	if err != nil {
		t.Fatalf("expected error to be nil, but got %v", err)
	}
	if !store.SaveInvoked {
		t.Fatal("expected store.save() to be called")
	}

	if !acker.AckInvoked {
		t.Fatal("expected message.ack() to be called")
	}
}

func testStatusChangeHandlerShouldFailToUnmarshalBody(t *testing.T, store *mock.UserStore, acker *mock.Acknowledger) {
	store.SaveFunc = func(user *pkg.User) error { return nil }
	acker.AckFunc = func(multiple bool) error { return nil }
	body := []byte(`INVALID`)

	msg := message.New(acker, body)
	h := NewUserStatusChanged(store)
	err := h.Handle(context.Background(), msg)
	if err == nil {
		t.Fatalf("expected to return err but got nil")
	}
	if store.SaveInvoked {
		t.Fatal("expected store.save() to not be called")
	}
	if !acker.AckInvoked {
		t.Fatal("expected message.Ack() to be called")
	}
}

func testStatusChangeHandlerNotFoundError(t *testing.T, store *mock.UserStore, acker *mock.Acknowledger) {
	store.SaveFunc = func(user *pkg.User) error { return pkg.ErrNotFound }
	acker.AckFunc = func(multiple bool) error { return nil }
	body := []byte(`{
		"email": "foo@mail.com",
		"username": "foo",
		"status": "active"
	}`)

	msg := message.New(acker, body)
	h := NewUserStatusChanged(store)
	err := h.Handle(context.Background(), msg)
	if err != pkg.ErrNotFound {
		t.Fatalf("expected to return err but got %v", err)
	}
	if !store.SaveInvoked {
		t.Fatal("expected store.Save() to not be called")
	}
	if !acker.AckInvoked {
		t.Fatal("expected message.Ack() to be called")
	}
}

func testStatusChangeHandlerUnexpectedSaveError(t *testing.T, store *mock.UserStore, acker *mock.Acknowledger) {
	store.SaveFunc = func(user *pkg.User) error { return errors.New("unexpected error") }
	acker.AckFunc = func(multiple bool) error { return nil }
	body := []byte(`{
		"email": "foo@mail.com",
		"username": "foo",
		"status": "active"
	}`)

	msg := message.New(acker, body)
	h := NewUserStatusChanged(store)
	err := h.Handle(context.Background(), msg)
	if err == nil {
		t.Fatalf("expected to return err but got nil")
	}
	if !store.SaveInvoked {
		t.Fatal("expected store.Save() to not be called")
	}
	if !acker.AckInvoked {
		t.Fatal("expected message.Ack() to be called")
	}
}
