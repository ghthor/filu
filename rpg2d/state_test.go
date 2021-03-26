package rpg2d_test

import (
	"bytes"
	"encoding/gob"
	"encoding/json"

	"github.com/ghthor/filu/rpg2d"
	"github.com/ghthor/filu/rpg2d/coord"
	"github.com/ghthor/filu/rpg2d/entity"
	"github.com/ghthor/filu/rpg2d/entity/entitytest"
	"github.com/ghthor/filu/rpg2d/quad"
	"github.com/ghthor/filu/rpg2d/rpg2dtest"
	"github.com/ghthor/filu/rpg2d/worldterrain"
	"github.com/ghthor/filu/sim/stime"

	"github.com/ghthor/gospec"
	. "github.com/ghthor/gospec"
)

func init() {
	gob.Register(entitytest.MockEntityState{})
}

func DescribeWorldState(c gospec.Context) {
	quadTree, err := quad.New(coord.Bounds{
		coord.Cell{-4, 4},
		coord.Cell{3, -3},
	}, 20, nil)
	c.Assume(err, IsNil)

	terrain, err := worldterrain.NewMap(quadTree.Bounds(), `
DDDDDDDD
DGGGGGGD
DGGRRGGD
DGRRRRGD
DGRRRRGD
DGGRRGGD
DGGGGGGD
DDDDDDDD
`)
	c.Assume(err, IsNil)

	world := rpg2d.NewWorld(stime.Time(0), quadTree, terrain)

	mockEntity := entitytest.MockEntity{EntityId: 0}
	world.Insert(mockEntity)

	worldState := world.ToState()

	c.Assume(worldState.Time, Equals, stime.Time(0))
	c.Assume(len(worldState.Entities), Equals, 1)

	c.Specify("a world state", func() {
		c.Specify("can be encoded as json", func() {
			jsonBytes, err := json.Marshal(worldState)
			c.Expect(err, IsNil)
			c.Expect(string(jsonBytes), Equals, `{"time":0,"bounds":{"tl":{"x":-4,"y":4},"br":{"x":3,"y":-3}},"entities":[{"id":0,"name":"MockEntity0","cell":{"x":0,"y":0}}],"entitiesRemoved":[],"entitiesNew":[],"entitiesChanged":[],"entitiesUnchanged":[{"id":0,"name":"MockEntity0","cell":{"x":0,"y":0}}],"terrainMap":{"bounds":{"tl":{"x":-4,"y":4},"br":{"x":3,"y":-3}},"terrain":"\nDDDDDDDD\nDGGGGGGD\nDGGRRGGD\nDGRRRRGD\nDGRRRRGD\nDGGRRGGD\nDGGGGGGD\nDDDDDDDD\n"}}`)
		})

		func() {
			buf := bytes.NewBuffer(make([]byte, 0, 1024))
			enc := gob.NewEncoder(buf)

			c.Specify("can be encoded as a gob object", func() {
				c.Expect(enc.Encode(worldState), IsNil)
			})

			c.Specify("can be decoded from a gob object", func() {
				dec := gob.NewDecoder(buf)
				c.Assume(enc.Encode(worldState), IsNil)

				state := rpg2d.WorldState{}
				c.Expect(dec.Decode(&state), IsNil)
				c.Expect(state, rpg2dtest.StateEquals, worldState)
			})
		}()

		c.Specify("can be cloned and modified", func() {
			world.Insert(entitytest.MockEntity{EntityId: 1})
			world.Insert(entitytest.MockEntity{EntityId: 2})
			world.Insert(entitytest.MockEntity{EntityId: 3})
			world.Insert(entitytest.MockEntity{EntityId: 4})

			worldState = world.ToState()
			clone := worldState.Clone()

			// Modify the clone
			clone.Entities = append(clone.Entities[:2], clone.Entities[3:]...)

			// Check that the modification didn't effect the original
			for i, e := range worldState.Entities {
				e, isMockEntity := e.(entitytest.MockEntityState)
				c.Assume(isMockEntity, IsTrue)
				c.Expect(e.EntityId(), Equals, entity.Id(i))
			}
		})

		c.Specify("can be culled by a bounding rectangle", func() {
			toBeCulled := []entity.State{
				entitytest.MockEntity{EntityCell: coord.Cell{-3, 3}}.ToState(),
				entitytest.MockEntity{EntityCell: coord.Cell{3, 3}}.ToState(),
				entitytest.MockEntity{EntityCell: coord.Cell{3, -3}}.ToState(),
				entitytest.MockEntity{EntityCell: coord.Cell{-3, -3}}.ToState(),
			}

			wontBeCulled := []entity.State{
				entitytest.MockEntity{EntityCell: coord.Cell{-2, 2}}.ToState(),
				entitytest.MockEntity{EntityCell: coord.Cell{1, 2}}.ToState(),
				entitytest.MockEntity{EntityCell: coord.Cell{1, -1}}.ToState(),
				entitytest.MockEntity{EntityCell: coord.Cell{-2, -1}}.ToState(),
			}

			worldState.Entities = append(worldState.Entities[:0], wontBeCulled...)
			worldState.Entities = append(worldState.Entities, toBeCulled...)

			worldState = worldState.Cull(coord.Bounds{
				coord.Cell{-2, 2},
				coord.Cell{1, -1},
			})

			c.Expect(worldState.Bounds, Equals, coord.Bounds{
				coord.Cell{-2, 2},
				coord.Cell{1, -1},
			})

			c.Expect(worldState.Entities, Not(ContainsAll), toBeCulled)
			c.Expect(worldState.Entities, ContainsAll, wontBeCulled)
			c.Expect(worldState.TerrainMap.String(), Equals, `
GRRG
RRRR
RRRR
GRRG
`)
		})

		north := coord.Bounds{
			coord.Cell{-2, 4},
			coord.Cell{1, 1},
		}

		northEast := coord.Bounds{
			coord.Cell{0, 4},
			coord.Cell{3, 1},
		}

		east := coord.Bounds{
			coord.Cell{0, 2},
			coord.Cell{3, -1},
		}

		southEast := coord.Bounds{
			coord.Cell{0, 0},
			coord.Cell{3, -3},
		}

		south := coord.Bounds{
			coord.Cell{-2, 0},
			coord.Cell{1, -3},
		}

		southWest := coord.Bounds{
			coord.Cell{-4, 0},
			coord.Cell{-1, -3},
		}

		west := coord.Bounds{
			coord.Cell{-4, 2},
			coord.Cell{-1, -1},
		}

		northWest := coord.Bounds{
			coord.Cell{-4, 4},
			coord.Cell{-1, 1},
		}

		c.Specify("can calculate the differences with a previous worldState state", func() {
			c.Specify("when there are no differences", func() {
				c.Expect(len(worldState.Diff(worldState).Entities), Equals, 0)
			})

			c.Specify("when an entity has changed state", func() {
				clone := worldState.Clone()
				entity := clone.Entities[0].(entitytest.MockEntityState)

				// This is a state change
				entity.Cell = coord.Cell{-1, 0}
				clone.Entities[0] = entity

				c.Expect(len(worldState.Diff(clone).Entities), Equals, 1)
			})

			c.Specify("when there is a new entity", func() {
				clone := worldState.Clone()
				clone.Entities = append(clone.Entities, entitytest.MockEntity{EntityId: 1}.ToState())
				c.Expect(len(worldState.Diff(clone).Entities), Equals, 1)
			})

			c.Specify("when an entity doesn't exist anymore", func() {
				clone := worldState.Clone()
				clone.Entities = clone.Entities[:0]
				c.Expect(len(worldState.Diff(clone).Removed), Equals, 1)
			})

			c.Specify("when the viewport has changed", func() {
				mockEntity0 := mockEntity
				mockEntity1 := entitytest.MockEntity{
					EntityId:   1,
					EntityCell: coord.Cell{-1, 0},
				}

				mockEntity2 := entitytest.MockEntity{
					EntityId:   2,
					EntityCell: coord.Cell{-1, 1},
				}

				mockEntity3 := entitytest.MockEntity{
					EntityId:   3,
					EntityCell: coord.Cell{0, 1},
				}

				world.Insert(mockEntity0)
				world.Insert(mockEntity1)
				world.Insert(mockEntity2)
				world.Insert(mockEntity3)
				worldState = world.ToState()

				initialState := worldState.Cull(coord.Bounds{
					coord.Cell{-2, 2},
					coord.Cell{1, -1},
				})

				var nextState rpg2d.WorldState

				expectDiffEquals := func(expectedDiff rpg2d.WorldStateDiff) {
					diff := initialState.Diff(nextState)
					c.Expect(diff, rpg2dtest.StateEquals, expectedDiff)
				}

				c.Specify("north", func() {
					nextState = worldState.Cull(north)

					c.Assume(north.Contains(mockEntity0.EntityCell), IsFalse)
					c.Assume(north.Contains(mockEntity1.EntityCell), IsFalse)
					c.Assume(north.Contains(mockEntity2.EntityCell), IsTrue)
					c.Assume(north.Contains(mockEntity3.EntityCell), IsTrue)

					expectDiffEquals(rpg2d.WorldStateDiff{
						Bounds: north,
						Removed: entity.StateSlice{
							mockEntity0.ToState(),
							mockEntity1.ToState(),
						},

						TerrainMapSlices: worldterrain.NewStateSlices(
							north, worldterrain.MapStateSlice{
								coord.Bounds{
									coord.Cell{-2, 4},
									coord.Cell{1, 3},
								},
								"\nDDDD\nGGGG\n",
							}),
					})
				})

				c.Specify("north & east", func() {
					nextState = worldState.Cull(northEast)

					c.Assume(northEast.Contains(mockEntity0.EntityCell), IsFalse)
					c.Assume(northEast.Contains(mockEntity1.EntityCell), IsFalse)
					c.Assume(northEast.Contains(mockEntity2.EntityCell), IsFalse)
					c.Assume(northEast.Contains(mockEntity3.EntityCell), IsTrue)

					expectDiffEquals(rpg2d.WorldStateDiff{
						Bounds: northEast,
						Removed: entity.StateSlice{
							mockEntity0.ToState(),
							mockEntity1.ToState(),
							mockEntity2.ToState(),
						},
						TerrainMapSlices: worldterrain.NewStateSlices(
							northEast,
							[]worldterrain.MapStateSlice{{
								Bounds: coord.Bounds{
									coord.Cell{0, 4},
									coord.Cell{1, 3},
								},
								Terrain: "\nDD\nGG\n",
							}, {
								Bounds: coord.Bounds{
									coord.Cell{2, 4},
									coord.Cell{3, 3},
								},
								Terrain: "\nDD\nGD\n",
							}, {
								Bounds: coord.Bounds{
									coord.Cell{2, 2},
									coord.Cell{3, 1},
								},
								Terrain: "\nGD\nGD\n",
							}}...),
					})
				})

				c.Specify("east", func() {
					nextState = worldState.Cull(east)

					c.Assume(east.Contains(mockEntity0.EntityCell), IsTrue)
					c.Assume(east.Contains(mockEntity1.EntityCell), IsFalse)
					c.Assume(east.Contains(mockEntity2.EntityCell), IsFalse)
					c.Assume(east.Contains(mockEntity3.EntityCell), IsTrue)

					expectDiffEquals(rpg2d.WorldStateDiff{
						Bounds: east,
						Removed: entity.StateSlice{
							mockEntity1.ToState(),
							mockEntity2.ToState(),
						},
						TerrainMapSlices: worldterrain.NewStateSlices(
							east, worldterrain.MapStateSlice{
								coord.Bounds{
									coord.Cell{2, 2},
									coord.Cell{3, -1},
								},
								"\nGD\nGD\nGD\nGD\n",
							}),
					})
				})

				c.Specify("south & east", func() {
					nextState = worldState.Cull(southEast)

					c.Assume(southEast.Contains(mockEntity0.EntityCell), IsTrue)
					c.Assume(southEast.Contains(mockEntity1.EntityCell), IsFalse)
					c.Assume(southEast.Contains(mockEntity2.EntityCell), IsFalse)
					c.Assume(southEast.Contains(mockEntity3.EntityCell), IsFalse)

					expectDiffEquals(rpg2d.WorldStateDiff{
						Bounds: southEast,
						Removed: entity.StateSlice{
							mockEntity1.ToState(),
							mockEntity2.ToState(),
							mockEntity3.ToState(),
						},
						TerrainMapSlices: worldterrain.NewStateSlices(
							southEast,
							[]worldterrain.MapStateSlice{{
								Bounds: coord.Bounds{
									coord.Cell{2, 0},
									coord.Cell{3, -1},
								},
								Terrain: "\nGD\nGD\n",
							}, {
								Bounds: coord.Bounds{
									coord.Cell{2, -2},
									coord.Cell{3, -3},
								},
								Terrain: "\nGD\nDD\n",
							}, {
								Bounds: coord.Bounds{
									coord.Cell{0, -2},
									coord.Cell{1, -3},
								},
								Terrain: "\nGG\nDD\n",
							}}...),
					})
				})

				c.Specify("south", func() {
					nextState = worldState.Cull(south)

					c.Assume(south.Contains(mockEntity0.EntityCell), IsTrue)
					c.Assume(south.Contains(mockEntity1.EntityCell), IsTrue)
					c.Assume(south.Contains(mockEntity2.EntityCell), IsFalse)
					c.Assume(south.Contains(mockEntity3.EntityCell), IsFalse)

					expectDiffEquals(rpg2d.WorldStateDiff{
						Bounds: south,
						Removed: entity.StateSlice{
							mockEntity2.ToState(),
							mockEntity3.ToState(),
						},
						TerrainMapSlices: worldterrain.NewStateSlices(
							south, worldterrain.MapStateSlice{
								coord.Bounds{
									coord.Cell{-2, -2},
									coord.Cell{1, -3},
								},
								"\nGGGG\nDDDD\n",
							}),
					})
				})

				c.Specify("south & west", func() {
					nextState = worldState.Cull(southWest)

					c.Assume(southWest.Contains(mockEntity0.EntityCell), IsFalse)
					c.Assume(southWest.Contains(mockEntity1.EntityCell), IsTrue)
					c.Assume(southWest.Contains(mockEntity2.EntityCell), IsFalse)
					c.Assume(southWest.Contains(mockEntity3.EntityCell), IsFalse)

					expectDiffEquals(rpg2d.WorldStateDiff{
						Bounds: southWest,
						Removed: entity.StateSlice{
							mockEntity0.ToState(),
							mockEntity2.ToState(),
							mockEntity3.ToState(),
						},
						TerrainMapSlices: worldterrain.NewStateSlices(
							southWest,
							[]worldterrain.MapStateSlice{{
								Bounds: coord.Bounds{
									coord.Cell{-2, -2},
									coord.Cell{-1, -3},
								},
								Terrain: "\nGG\nDD\n",
							}, {
								Bounds: coord.Bounds{
									coord.Cell{-4, -2},
									coord.Cell{-3, -3},
								},
								Terrain: "\nDG\nDD\n",
							}, {
								Bounds: coord.Bounds{
									coord.Cell{-4, 0},
									coord.Cell{-3, -1},
								},
								Terrain: "\nDG\nDG\n",
							}}...),
					})
				})

				c.Specify("west", func() {
					nextState = worldState.Cull(west)

					c.Assume(west.Contains(mockEntity0.EntityCell), IsFalse)
					c.Assume(west.Contains(mockEntity1.EntityCell), IsTrue)
					c.Assume(west.Contains(mockEntity2.EntityCell), IsTrue)
					c.Assume(west.Contains(mockEntity3.EntityCell), IsFalse)

					expectDiffEquals(rpg2d.WorldStateDiff{
						Bounds: west,
						Removed: entity.StateSlice{
							mockEntity0.ToState(),
							mockEntity3.ToState(),
						},
						TerrainMapSlices: worldterrain.NewStateSlices(
							west, worldterrain.MapStateSlice{
								coord.Bounds{
									coord.Cell{-4, 2},
									coord.Cell{-3, -1},
								},
								"\nDG\nDG\nDG\nDG\n",
							}),
					})
				})

				c.Specify("north & west", func() {
					nextState = worldState.Cull(northWest)

					c.Assume(northWest.Contains(mockEntity0.EntityCell), IsFalse)
					c.Assume(northWest.Contains(mockEntity1.EntityCell), IsFalse)
					c.Assume(northWest.Contains(mockEntity2.EntityCell), IsTrue)
					c.Assume(northWest.Contains(mockEntity3.EntityCell), IsFalse)

					expectDiffEquals(rpg2d.WorldStateDiff{
						Bounds: northWest,
						Removed: entity.StateSlice{
							mockEntity0.ToState(),
							mockEntity1.ToState(),
							mockEntity3.ToState(),
						},
						TerrainMapSlices: worldterrain.NewStateSlices(
							northWest,
							[]worldterrain.MapStateSlice{{
								Bounds: coord.Bounds{
									coord.Cell{-4, 2},
									coord.Cell{-3, 1},
								},
								Terrain: "\nDG\nDG\n",
							}, {
								Bounds: coord.Bounds{
									coord.Cell{-4, 4},
									coord.Cell{-3, 3},
								},
								Terrain: "\nDD\nDG\n",
							}, {
								Bounds: coord.Bounds{
									coord.Cell{-2, 4},
									coord.Cell{-1, 3},
								},
								Terrain: "\nDD\nGG\n",
							}}...),
					})
				})
			})

			c.Specify("when the viewport hasn't changed", func() {
				clone := worldState.Clone()
				c.Expect(worldState.Diff(clone).TerrainMapSlices, IsNil)
			})
		})

		c.Specify("can be updated with a world state diff", func() {
			c.Specify("that updates the time", func() {
				nextState := world.ToState()
				nextState.Time++

				worldState.Apply(worldState.Diff(nextState))

				c.Expect(worldState, rpg2dtest.StateEquals, nextState)
			})

			c.Specify("that contains a new entity", func() {
				world.Insert(entitytest.MockEntity{EntityId: 1})
				nextState := world.ToState()
				diff := worldState.Diff(nextState)

				worldState.Apply(diff)
				c.Expect(worldState, rpg2dtest.StateEquals, nextState)
			})

			c.Specify("that removes an entity", func() {
				world.Remove(mockEntity)
				nextState := world.ToState()
				diff := worldState.Diff(nextState)

				worldState.Apply(diff)
				c.Expect(worldState, rpg2dtest.StateEquals, nextState)
			})

			c.Specify("where the diff's bounds", func() {
				c.Specify("overlap with the state's bounds to the", func() {
					mockEntity0 := mockEntity
					mockEntity1 := entitytest.MockEntity{
						EntityId:   1,
						EntityCell: coord.Cell{-1, 0},
					}

					mockEntity2 := entitytest.MockEntity{
						EntityId:   2,
						EntityCell: coord.Cell{-1, 1},
					}

					mockEntity3 := entitytest.MockEntity{
						EntityId:   3,
						EntityCell: coord.Cell{0, 1},
					}

					world.Insert(mockEntity0)
					world.Insert(mockEntity1)
					world.Insert(mockEntity2)
					world.Insert(mockEntity3)
					worldState = world.ToState()

					initialState := worldState.Cull(coord.Bounds{
						coord.Cell{-2, 2},
						coord.Cell{1, -1},
					}).Clone()

					var nextState rpg2d.WorldState

					expectDiffApplied := func() {
						initialState.Apply(initialState.Diff(nextState))
						c.Expect(initialState, rpg2dtest.StateEquals, nextState)
					}

					c.Specify("north", func() {
						nextState = worldState.Cull(north)
						expectDiffApplied()
					})

					c.Specify("north & east", func() {
						nextState = worldState.Cull(northEast)
						expectDiffApplied()
					})

					c.Specify("east", func() {
						nextState = worldState.Cull(east)
						expectDiffApplied()
					})

					c.Specify("south & east", func() {
						nextState = worldState.Cull(southEast)
						expectDiffApplied()
					})

					c.Specify("south", func() {
						nextState = worldState.Cull(south)
						expectDiffApplied()
					})

					c.Specify("south & west", func() {
						nextState = worldState.Cull(southWest)
						expectDiffApplied()
					})

					c.Specify("west", func() {
						nextState = worldState.Cull(west)
						expectDiffApplied()
					})

					c.Specify("north & west", func() {
						nextState = worldState.Cull(northWest)
						expectDiffApplied()
					})
				})

				c.Specify("does NOT overlap the state's bounds", func() {
					mockEntity0 := entitytest.MockEntity{
						EntityId:   0,
						EntityCell: coord.Cell{-3, 3},
					}

					mockEntity1 := entitytest.MockEntity{
						EntityId:   1,
						EntityCell: coord.Cell{-4, 3},
					}

					mockEntity2 := entitytest.MockEntity{
						EntityId:   2,
						EntityCell: coord.Cell{3, 3},
					}

					world.Insert(mockEntity0)
					world.Insert(mockEntity1)
					world.Insert(mockEntity2)
					worldState = world.ToState()

					initialState := worldState.Cull(coord.Bounds{
						coord.Cell{-4, 4},
						coord.Cell{-2, 2},
					}).Clone()

					mockEntity0.EntityCell = coord.Cell{2, 3}
					world.Insert(mockEntity0)

					nextState := world.ToState()
					nextState = nextState.Cull(coord.Bounds{
						coord.Cell{1, 4},
						coord.Cell{3, 2},
					})

					initialState.Apply(initialState.Diff(nextState))
					c.Expect(initialState, rpg2dtest.StateEquals, nextState)
				})
			})
		})
	})
}
