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

type closeVerifier struct {
	selectionStreamClosed <-chan bool
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

func DescribeStream(c gospec.Context) {
	requestStream := make(chan actor.SelectionRequest)

	log := &logger{}


	closeVerifier := &closeVerifier{}

	actor.SelectionRequestSource(requestStream).
		WriteTo(log).
		WriteToProcessor(actor.NewSelectionProcessor()).
		WriteTo(log).
		WriteTo(closeVerifier).
		End()

	defer func() {
		close(requestStream)
		c.Expect(<-closeVerifier.selectionStreamClosed, IsTrue)
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
}
