package auth_test

import (
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

		resultLog := newResultLog()
		resultStore := newResultStore()

		s := auth.NewStream(
			auth.NewRequestStream(requestLog, requestStore),
			auth.NewResultStream(resultLog, resultStore))

		defer func() {
			close(s.RequestAuthentication())
		}()

		r := auth.NewRequest("test", "password")
		s.RequestAuthentication() <- r

		c.Specify("will return", func() {
			c.Specify("an invalid password", func() {
				c.Assume((<-r.CreatedUser).Username, Equals, "test")

				r = auth.NewRequest("test", "invalid")
				s.RequestAuthentication() <- r

				result := <-r.InvalidPassword
				c.Expect(result.Username, Equals, "test")
				c.Expect(result.HappenedAt().After(r.HappenedAt()), IsTrue)
			})

			c.Specify("a created user", func() {
				result := <-r.CreatedUser
				c.Expect(result.Username, Equals, "test")
				c.Expect(result.HappenedAt().After(r.HappenedAt()), IsTrue)
			})

			c.Specify("an authenicated  user", func() {
				c.Assume((<-r.CreatedUser).Username, Equals, "test")

				s.RequestAuthentication() <- r

				result := <-r.AuthenticatedUser
				c.Expect(result.Username, Equals, "test")
				c.Expect(result.HappenedAt().After(r.HappenedAt()), IsTrue)
			})
		})

		c.Specify("will log the request", func() {
			c.Expect((<-requestLog.loggedRequest).Username, Equals, "test")
		})

		c.Specify("will store the request", func() {
			c.Assume((<-r.CreatedUser).Username, Equals, "test")
			c.Expect(requestStore.allRequests[0].Username, Equals, "test")
		})

		c.Specify("will log the result", func() {
			c.Expect((<-resultLog.loggedResult).(auth.CreatedUser).Username, Equals, "test")
		})

		c.Specify("will store the result", func() {
			c.Assume((<-r.CreatedUser).Username, Equals, "test")
			c.Expect(resultStore.allResults[0].(auth.CreatedUser).Username, Equals, "test")
		})
	})
}
