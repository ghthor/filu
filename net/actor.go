package net

import (
	"fmt"
	"log"

	"github.com/ghthor/filu"
	"github.com/ghthor/filu/actor"
)

func getActors(actorsDB chan<- actor.GetActorsRequest, username string) ActorsList {
	r := actor.NewGetActorsRequest(username)
	actorsDB <- r

	actors := <-r.Actors
	names := make([]string, 0, len(actors))
	for _, a := range actors {
		names = append(names, a.Name)
	}

	return names
}

func SendActors(conn Conn, actorDB chan<- actor.GetActorsRequest, user AuthenticatedUser) error {
	return conn.Encode(getActors(actorDB, user.Username))
}

func SelectActorFrom(conn Conn, actorDB chan<- actor.SelectionRequest, user AuthenticatedUser) (filu.Actor, error) {
	eType, err := conn.NextType()
	if err != nil {
		return filu.Actor{}, err
	}

	switch eType {
	default:
		// TODO: Log error to universal error log
		log.Println("unexpected EncodeType:", eType)

		protoError := ProtocolError(
			fmt.Sprintf("expected EncodeType(%v) got EncodeType(%v)", ET_SELECT_ACTOR, eType),
		)

		err := conn.Encode(protoError)
		if err != nil {
			return filu.Actor{}, err
		}

		return filu.Actor{}, protoError

	case ET_SELECT_ACTOR:
	}

	var r SelectActorRequest
	err = conn.Decode(&r)
	if err != nil {
		return filu.Actor{}, err
	}

	selectReq := actor.NewSelectionRequest(filu.Actor{
		Username: user.Username,
		Name:     r.Name,
	})
	actorDB <- selectReq

	select {
	case result := <-selectReq.CreatedActor:
		return result.Actor, conn.Encode(CreateActorSuccess{
			Actor: result.Actor,
		})

	case result := <-selectReq.SelectedActor:
		return result.Actor, conn.Encode(SelectActorSuccess{
			Actor: result.Actor,
		})
	}
}
