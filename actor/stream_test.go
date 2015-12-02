package actor_test

import (
	"github.com/ghthor/filu"
	"github.com/ghthor/filu/actor"

	"github.com/ghthor/gospec"
	. "github.com/ghthor/gospec"
)

type requestLogger struct {
	requests <-chan actor.SelectionRequest
}

func (l *requestLogger) ReadFrom(requests <-chan actor.SelectionRequest) actor.SelectionRequestSource {
	log := make(chan actor.SelectionRequest, 10)
	l.requests = log

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

type resultLogger struct {
	results <-chan actor.SelectionResult
}

func (l *resultLogger) ReadFrom(results <-chan actor.SelectionResult) actor.SelectionResultSource {
	log := make(chan actor.SelectionResult, 10)
	l.results = log

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

type closeVerifier struct {
	wasClosed <-chan bool
}

func (c *closeVerifier) ReadFrom(results <-chan actor.SelectionResult) actor.SelectionResultSource {
	closed := make(chan bool)
	c.wasClosed = closed

	out := make(chan actor.SelectionResult)

	go func(out chan<- actor.SelectionResult, closed chan<- bool) {
		for r := range results {
			out <- r
		}

		closed <- true

		close(out)
	}(out, closed)

	return out
}

func DescribeStream(c gospec.Context) {
	requestStream := make(chan actor.SelectionRequest)

	verifyStream := &closeVerifier{}

	requestLog := &requestLogger{}
	resultLog := &resultLogger{}

	actor.SelectionRequestSource(requestStream).
		WriteTo(requestLog).
		WriteToProcessor(actor.NewSelectionProcessor()).
		WriteTo(resultLog).
		WriteTo(verifyStream).
		End()

	defer func() {
		close(requestStream)
		c.Expect(<-verifyStream.wasClosed, IsTrue)
	}()

	c.Specify("a stream", func() {
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
			c.Expect((<-requestLog.requests).Actor, Equals, a)
		})

		c.Specify("will log a result", func() {
			c.Expect((<-resultLog.results).(actor.CreatedActor).Actor, Equals, a)
		})
	})
}
