package engine

import (
	"github.com/ghthor/engine/rpg2d/coord"
	"github.com/ghthor/engine/time"
	"github.com/ghthor/gospec"
	. "github.com/ghthor/gospec"
)

func DescribePlayerJson(c gospec.Context) {
	player := PlayerJson{
		Facing: coord.North.String(),
		Cell:   coord.Cell{0, 0},
	}

	c.Specify("can identify there has been no change if", func() {
		c.Specify("when the player is standing still", func() {
			c.Expect(player.IsDifferentFrom(player), IsFalse)
		})

		c.Specify("when the player is moving", func() {
			player.PathActions = []coord.PathActionJson{
				coord.PathAction{
					time.NewSpan(10, 30),
					player.Cell,
					player.Cell.Neighbor(coord.North),
				}.Json(),
			}
			c.Expect(player.IsDifferentFrom(player), IsFalse)
		})
	})

	c.Specify("can identify a change has occurred", func() {
		playerChanged := player

		c.Specify("when the player's facing has changed", func() {
			playerChanged.Facing = coord.South.String()
			c.Expect(player.IsDifferentFrom(playerChanged), IsTrue)
		})

		c.Specify("when the player has started moving", func() {
			playerChanged.PathActions = []coord.PathActionJson{
				coord.PathAction{
					time.NewSpan(10, 30),
					player.Cell,
					player.Cell.Neighbor(coord.North),
				}.Json(),
			}

			c.Expect(player.IsDifferentFrom(playerChanged), IsTrue)
		})

		c.Specify("when the player has stopped moving", func() {
			playerChanged.Cell = player.Cell.Neighbor(coord.North)
			player.PathActions = []coord.PathActionJson{
				coord.PathAction{
					time.NewSpan(10, 30),
					player.Cell,
					player.Cell.Neighbor(coord.North),
				}.Json(),
			}

			c.Expect(player.IsDifferentFrom(playerChanged), IsTrue)
		})

		// This case is concerned when movement is "chained" together.
		c.Specify("when the player has finished a movement and continued moving", func() {
			// TODO Specify "and continued moving in the same direction"
			// TODO Specify "and continued moving in a different direction"
			player.PathActions = []coord.PathActionJson{
				coord.PathAction{
					time.NewSpan(10, 30),
					player.Cell,
					player.Cell.Neighbor(coord.North),
				}.Json(),
			}

			playerChanged.Cell = player.Cell.Neighbor(coord.North)
			playerChanged.PathActions = []coord.PathActionJson{
				coord.PathAction{
					time.NewSpan(30, 50),
					playerChanged.Cell,
					playerChanged.Cell.Neighbor(coord.West),
				}.Json(),
			}

			c.Expect(player.IsDifferentFrom(playerChanged), IsTrue)
		})
	})
}
