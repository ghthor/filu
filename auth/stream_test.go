package auth_test

import (
	"github.com/ghthor/filu/auth"
	"github.com/ghthor/gospec"

	. "github.com/ghthor/gospec"
)

func DescribeStream(c gospec.Context) {
	s := auth.NewStream()

	r := auth.NewRequest("test", "password")

	s.RequestAuthorization() <- r

	c.Specify("the result is an invalid password", func() {
		c.Assume((<-r.CreatedUser).Username, Equals, "test")

		r = auth.NewRequest("test", "invalid")
		s.RequestAuthorization() <- r

		result := <-r.InvalidPassword
		c.Expect(result.Username, Equals, "test")
	})

	c.Specify("the result is a user was created", func() {
		result := <-r.CreatedUser
		c.Expect(result.Username, Equals, "test")
	})

	c.Specify("the result is an authorized user", func() {
		c.Assume((<-r.CreatedUser).Username, Equals, "test")

		s.RequestAuthorization() <- r

		result := <-r.AuthorizedUser
		c.Expect(result.Username, Equals, "test")
	})
}
