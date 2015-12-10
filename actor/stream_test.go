package actor_test

import (
	"github.com/ghthor/filu"
	"github.com/ghthor/filu/actor"

	"github.com/ghthor/gospec"
	. "github.com/ghthor/gospec"
)

const logSize = 100

type logger struct {
	selectionRequests <-chan actor.SelectionRequest
	selectionResults  <-chan actor.SelectionResult

	getActorsRequests <-chan actor.GetActorsRequest
	existingActors    <-chan actor.ExistingActors
}

func (l *logger) ReadSelectionRequestsFrom(requests <-chan actor.SelectionRequest) actor.SelectionRequestSource {
	log := make(chan actor.SelectionRequest, logSize)
	l.selectionRequests = log

	out := make(chan actor.SelectionRequest)

	go func(out, log chan<- actor.SelectionRequest) {
		for r := range requests {
			log <- r
			out <- r
		}
		close(log)
		close(out)
	}(out, log)

	return out
}

func (l *logger) ReadGetActorsRequestsFrom(requests <-chan actor.GetActorsRequest) actor.GetActorsRequestSource {
	log := make(chan actor.GetActorsRequest, logSize)
	l.getActorsRequests = log

	out := make(chan actor.GetActorsRequest)

	go func(out, log chan<- actor.GetActorsRequest) {
		for r := range requests {
			log <- r
			out <- r
		}
		close(log)
		close(out)
	}(out, log)

	return out
}

func (l *logger) ReadSelectionResultsFrom(results <-chan actor.SelectionResult) actor.SelectionResultSource {
	log := make(chan actor.SelectionResult, logSize)
	l.selectionResults = log

	out := make(chan actor.SelectionResult)

	go func(out, log chan<- actor.SelectionResult) {
		for r := range results {
			log <- r
			out <- r
		}
		close(log)
		close(out)
	}(out, log)

	return out
}

func (l *logger) ReadExistingActorsFrom(results <-chan actor.ExistingActors) actor.ExistingActorsSource {
	log := make(chan actor.ExistingActors, logSize)
	l.existingActors = log

	out := make(chan actor.ExistingActors)

	go func(out, log chan<- actor.ExistingActors) {
		for r := range results {
			log <- r
			out <- r
		}
		close(log)
		close(out)
	}(out, log)

	return out
}

type closeVerifier struct {
	selectionStreamClosed <-chan bool
	getActorsStreamClosed <-chan bool
}

func (c *closeVerifier) ReadSelectionResultsFrom(results <-chan actor.SelectionResult) actor.SelectionResultSource {
	closed := make(chan bool)
	c.selectionStreamClosed = closed

	out := make(chan actor.SelectionResult)

	go func(out chan<- actor.SelectionResult, closed chan<- bool) {
		for r := range results {
			out <- r
		}

		close(out)
		closed <- true
	}(out, closed)

	return out
}

func (c *closeVerifier) ReadExistingActorsFrom(results <-chan actor.ExistingActors) actor.ExistingActorsSource {
	closed := make(chan bool)
	c.getActorsStreamClosed = closed

	out := make(chan actor.ExistingActors)

	go func(out chan<- actor.ExistingActors, closed chan<- bool) {
		for r := range results {
			out <- r
		}

		close(out)
		closed <- true
	}(out, closed)

	return out
}

func makeActors(username string, actors ...string) []filu.Actor {
	out := make([]filu.Actor, 0, len(actors))

	for _, a := range actors {
		out = append(out, filu.Actor{
			Username: username,
			Name:     a,
		})
	}

	return out
}

func DescribeStream(c gospec.Context) {
	requestStream := make(chan actor.SelectionRequest)
	getActorsStream := make(chan actor.GetActorsRequest)

	log := &logger{}

	getActorsProc := actor.NewGetActorsRequestProcessor()

	closeVerifier := &closeVerifier{}

	actor.SelectionRequestSource(requestStream).
		WriteTo(log).
		WriteToProcessor(actor.NewSelectionProcessor()).
		WriteTo(getActorsProc).
		WriteTo(log).
		WriteTo(closeVerifier).
		End()

	actor.GetActorsRequestSource(getActorsStream).
		WriteTo(log).
		WriteToProcessor(getActorsProc).
		WriteTo(log).
		WriteTo(closeVerifier).
		End()

	defer func() {
		close(requestStream)
		c.Expect(<-closeVerifier.selectionStreamClosed, IsTrue)
	}()

	defer func() {
		close(getActorsStream)
		c.Expect(<-closeVerifier.getActorsStreamClosed, IsTrue)
	}()

	c.Specify("a selection stream", func() {
		a := filu.Actor{
			Username: "user",
			Name:     "actor name",
		}

		r := actor.NewSelectionRequest(a)
		requestStream <- r
		c.Assume((<-r.CreatedActor).Actor, Equals, a)
		c.Assume((<-r.SelectedActor).Actor, Equals, filu.Actor{})

		c.Specify("will create an actor", func() {
			a.Name = "another actor"
			r = actor.NewSelectionRequest(a)
			requestStream <- r
			c.Expect((<-r.CreatedActor).Actor, Equals, a)

			c.Specify("and close result channels", func() {
				c.Expect((<-r.SelectedActor).Actor, Equals, filu.Actor{})
			})
		})

		c.Specify("will select an actor", func() {
			r = actor.NewSelectionRequest(a)
			requestStream <- r
			c.Expect((<-r.SelectedActor).Actor, Equals, a)

			c.Specify("and close result channels", func() {
				c.Expect((<-r.CreatedActor).Actor, Equals, filu.Actor{})
			})
		})

		c.Specify("will log a request", func() {
			c.Expect((<-log.selectionRequests).Actor, Equals, a)
		})

		c.Specify("will log a result", func() {
			c.Expect((<-log.selectionResults).(actor.CreatedActor).Actor, Equals, a)
		})
	})

	c.Specify("a get actors stream", func() {
		database := map[string][]filu.Actor{
			"jim":  makeActors("jim", "raynor", "samuel", "samwise"),
			"mary": makeActors("mary", "little bits", "lamb"),
			"sam":  []filu.Actor{},
		}

		allActors := []filu.Actor{}
		for _, actors := range database {
			allActors = append(allActors, actors...)
		}

		for _, a := range allActors {
			r := actor.NewSelectionRequest(a)
			requestStream <- r
			c.Assume((<-r.CreatedActor).Actor, Equals, a)
			c.Assume((<-r.SelectedActor).Actor, Equals, filu.Actor{})
		}

		requests := map[string]actor.GetActorsRequest{
			"jim":  actor.NewGetActorsRequest("jim"),
			"mary": actor.NewGetActorsRequest("mary"),
			"sam":  actor.NewGetActorsRequest("sam"),
		}

		requestedOrder := make([]string, 0, len(requests))
		for username, r := range requests {
			getActorsStream <- r
			requestedOrder = append(requestedOrder, username)
		}

		c.Specify("will return a list of actors for a username", func() {
			for username, r := range requests {
				c.Expect(<-r.Actors, ContainsAll, database[username])
			}
		})

		c.Specify("will log a request", func() {
			for _, username := range requestedOrder {
				c.Expect((<-log.getActorsRequests).Username, Equals, username)
			}
		})

		c.Specify("will log a result", func() {
			for _, username := range requestedOrder {
				c.Expect((<-log.existingActors).Actors, ContainsAll, database[username])
			}
		})
	})
}
