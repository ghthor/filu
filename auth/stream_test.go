package auth_test

import (
	"github.com/ghthor/filu/auth"
	"github.com/ghthor/gospec"

	. "github.com/ghthor/gospec"
)

type logResultStream struct {
	loggedResult chan auth.Result
	results      chan auth.Result
}

func newLogResultStream() logResultStream {
	return logResultStream{
		loggedResult: make(chan auth.Result, 20),
		results:      make(chan auth.Result, 20),
	}
}

func (s logResultStream) Write(r auth.Result) {
	s.loggedResult <- r
	s.results <- r
}

func (s logResultStream) Read() <-chan auth.Result {
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
		log := newLogResultStream()
		store := newResultStore()

		s := auth.NewStream(log, store)
		defer func() {
			close(s.RequestAuthorization())
		}()

		r := auth.NewRequest("test", "password")
		s.RequestAuthorization() <- r

		c.Specify("will return", func() {
			c.Specify("an invalid password", func() {
				c.Assume((<-r.CreatedUser).Username, Equals, "test")

				r = auth.NewRequest("test", "invalid")
				s.RequestAuthorization() <- r

				result := <-r.InvalidPassword
				c.Expect(result.Username, Equals, "test")
			})

			c.Specify("a created user", func() {
				result := <-r.CreatedUser
				c.Expect(result.Username, Equals, "test")
			})

			c.Specify("an authorized user", func() {
				c.Assume((<-r.CreatedUser).Username, Equals, "test")

				s.RequestAuthorization() <- r

				result := <-r.AuthorizedUser
				c.Expect(result.Username, Equals, "test")
			})
		})

		c.Specify("will log the result", func() {
			c.Expect((<-log.loggedResult).(auth.CreatedUser).Username, Equals, "test")
		})

		c.Specify("will store the result", func() {
			c.Expect(store.allResults[0].(auth.CreatedUser).Username, Equals, "test")
		})
	})
}
