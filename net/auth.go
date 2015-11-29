package net

import (
	"log"

	"github.com/ghthor/filu/auth"
)

type AuthenticatedUser struct {
	Username string
}

func AuthenticateFrom(conn Conn, stream auth.Stream) (AuthenticatedUser, error) {
readNextType:
	eType, err := conn.NextType()
	if err != nil {
		return AuthenticatedUser{}, err
	}

	switch eType {
	default:
		log.Println("unexpected encoded type: ", eType)
		// TODO: Notify client their protocol usage is incorrect
		// TODO: Log error to universal error log
		goto readNextType

	case ET_USER_LOGIN_REQUEST:
	}

	var r UserLoginRequest
	err = conn.Decode(&r)
	if err != nil {
		return AuthenticatedUser{}, err
	}

	authReq := auth.NewRequest(r.Name, r.Password)
	stream.RequestAuthentication() <- authReq

	select {
	case user := <-authReq.CreatedUser:
		err = conn.Encode(UserCreateSuccess{user.Username})
		if err != nil {
			return AuthenticatedUser{}, err
		}

		return AuthenticatedUser{user.Username}, nil

	case user := <-authReq.AuthenticatedUser:
		err = conn.Encode(UserLoginSuccess{user.Username})
		if err != nil {
			return AuthenticatedUser{}, err
		}

		return AuthenticatedUser{user.Username}, nil

	case invalid := <-authReq.InvalidPassword:
		err := conn.Encode(UserLoginFailure{invalid.Username})
		if err != nil {
			return AuthenticatedUser{}, err
		}
	}

	goto readNextType
}
