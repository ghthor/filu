package auth_test

import (
	"bytes"
	"encoding/gob"
	"fmt"

	"github.com/ghthor/filu/auth"
	"github.com/ghthor/gospec"

	. "github.com/ghthor/gospec"
)

type requestLog struct {
	loggedRequest chan auth.Request
	requests      chan auth.Request
}

func newRequestLog() requestLog {
	return requestLog{
		loggedRequest: make(chan auth.Request, 20),
		requests:      make(chan auth.Request, 20),
	}
}

func (l requestLog) Write(r auth.Request) {
	l.loggedRequest <- r
	l.requests <- r
}

func (l requestLog) Read() <-chan auth.Request {
	return l.requests
}

type requestStore struct {
	allRequests []auth.Request
	requests    chan auth.Request
}

func newRequestStore() *requestStore {
	return &requestStore{
		allRequests: make([]auth.Request, 0, 10),
		requests:    make(chan auth.Request, 10),
	}
}

func (s *requestStore) Write(r auth.Request) {
	s.allRequests = append(s.allRequests, r)
	s.requests <- r
}

func (s requestStore) Read() <-chan auth.Request {
	return s.requests
}

type resultLog struct {
	loggedResult chan auth.Result
	results      chan auth.Result
}

func newResultLog() resultLog {
	return resultLog{
		loggedResult: make(chan auth.Result, 20),
		results:      make(chan auth.Result, 20),
	}
}

func (s resultLog) Write(r auth.Result) {
	s.loggedResult <- r
	s.results <- r
}

func (s resultLog) Read() <-chan auth.Result {
	return s.results
}

type resultStore struct {
	allResults []auth.Result
	results    chan auth.Result
}

func newResultStore() *resultStore {
	return &resultStore{
		allResults: make([]auth.Result, 0, 10),
		results:    make(chan auth.Result, 10),
	}
}

func (s *resultStore) Write(r auth.Result) {
	s.allResults = append(s.allResults, r)
	s.results <- r
}

func (s resultStore) Read() <-chan auth.Result {
	return s.results
}

func DescribeStream(c gospec.Context) {
	c.Specify("a stream", func() {
		requestLog := newRequestLog()
		requestStore := newRequestStore()
		requestStream := auth.NewRequestStream(requestLog, requestStore)

		var buf bytes.Buffer
		enc := gob.NewEncoder(&buf)

		requests := []struct {
			Username, Password string
		}{
			{"test", "password"},
			{"mary", "listallwords"},
			{"test", "password"},
			{"mary", "invalid"},
		}

		for _, r := range requests {
			c.Assume(enc.Encode(auth.NewRequest(r.Username, r.Password)), IsNil)
		}

		processor, err := auth.NewProcessor(&buf)
		c.Assume(err, IsNil)

		resultLog := newResultLog()
		resultStore := newResultStore()
		resultStream := auth.NewResultStream(resultLog, resultStore)

		s := auth.NewStream(
			requestStream,
			processor,
			resultStream)

		defer func() {
			close(s.RequestAuthentication())
		}()

		runRequest := func(username, password string) auth.Request {
			r := auth.NewRequest(username, password)
			s.RequestAuthentication() <- r
			return r
		}

		c.Specify("will fast forward", func() {
			r := runRequest("test", "password")
			c.Expect((<-r.AuthenticatedUser).Username, Equals, "test")

			r = runRequest("mary", "invalid")
			c.Expect((<-r.InvalidPassword).Username, Equals, "mary")

			r = runRequest("created", "user")
			c.Expect((<-r.CreatedUser).Username, Equals, "created")
		})

		c.Specify("will return", func() {
			c.Specify("an invalid password", func() {
				r := runRequest("mary", "invalid")
				select {
				case <-r.CreatedUser:
					panic(fmt.Sprintf("error: race condition reading result to request {%s}", r.Username))
				case <-r.AuthenticatedUser:
					panic(fmt.Sprintf("error: race condition reading result to request {%s}", r.Username))
				case result := <-r.InvalidPassword:
					c.Expect(result.Username, Equals, "mary")
					c.Expect(result.HappenedAt().After(r.HappenedAt()), IsTrue)
				}
			})

			c.Specify("a created user", func() {
				r := runRequest("created", "user")
				select {
				case result := <-r.CreatedUser:
					c.Expect(result.Username, Equals, "created")
					c.Expect(result.HappenedAt().After(r.HappenedAt()), IsTrue)
				case <-r.AuthenticatedUser:
					panic(fmt.Sprintf("error: race condition reading result to request {%s}", r.Username))
				case <-r.InvalidPassword:
					panic(fmt.Sprintf("error: race condition reading result to request {%s}", r.Username))
				}
			})

			c.Specify("an authenticated user", func() {
				r := runRequest("test", "password")
				select {
				case <-r.CreatedUser:
					panic(fmt.Sprintf("error: race condition reading result to request {%s}", r.Username))
				case result := <-r.AuthenticatedUser:
					c.Expect(result.Username, Equals, "test")
					c.Expect(result.HappenedAt().After(r.HappenedAt()), IsTrue)
				case <-r.InvalidPassword:
					panic(fmt.Sprintf("error: race condition reading result to request {%s}", r.Username))
				}
			})
		})

		c.Specify("will log and store", func() {
			r := runRequest("test", "password")

			c.Specify("the request", func() {
				c.Assume((<-r.AuthenticatedUser).Username, Equals, "test")
				c.Expect((<-requestLog.loggedRequest).Username, Equals, "test")
				c.Expect(requestStore.allRequests[0].Username, Equals, "test")
			})

			c.Specify("the result", func() {
				c.Assume((<-r.AuthenticatedUser).Username, Equals, "test")
				c.Expect((<-resultLog.loggedResult).(auth.AuthenticatedUser).Username, Equals, "test")
				c.Expect(resultStore.allResults[0].(auth.AuthenticatedUser).Username, Equals, "test")
			})
		})
	})
}
