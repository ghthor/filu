package net

import (
	"errors"
	"fmt"
	"log"

	"github.com/ghthor/filu/auth"
)

type AuthenticatedUser struct {
	Username string
}

var ErrInvalidLoginCredentials = errors.New("client provided invalid login credentials")

func AuthenticateFrom(conn Conn, authDB auth.Stream) (AuthenticatedUser, error) {
	eType, err := conn.NextType()
	if err != nil {
		return AuthenticatedUser{}, err
	}

	switch eType {
	default:
		// TODO: Log error to universal error log
		log.Println("unexpected EncodeType:", eType)

		protoError := ProtocolError(
			fmt.Sprintf("expected EncodeType(%v) got EncodeType(%v)", ET_USER_LOGIN_REQUEST, eType),
		)

		err := conn.Encode(protoError)
		if err != nil {
			return AuthenticatedUser{}, err
		}

		return AuthenticatedUser{}, protoError

	case ET_USER_LOGIN_REQUEST:
	}

	var r UserLoginRequest
	err = conn.Decode(&r)
	if err != nil {
		return AuthenticatedUser{}, err
	}

	authReq := auth.NewRequest(r.Name, r.Password)
	authDB.RequestAuthentication() <- authReq

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

		return AuthenticatedUser{}, ErrInvalidLoginCredentials
	}
}
