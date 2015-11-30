package filu

// An Actor is the ID of a unique entity in a world. The
// uniqueness of a Username+ActorName combination allows
// an Actor to be the key in a key/value database
// containing game specific statistics and values associated
// with a specific Actor.
type Actor struct {
	// The Name of the User that owns the Actor.
	Username string

	// The Name used to represent an Actor in a world.
	Name string
}
