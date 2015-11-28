## Streams

### Auth Stream
1. PreAuthLog
2. AuthProcessor
	- A materialized view of the PreAuthLog.
3. PostAuthLog
	- A materialized log of the PreAuthLog via AuthProcessor.

### World Stream

1. InputLog
	- Estimated to have the Highest write/sec of all logs
2. InputProcessor
	- A materialized view of the InputLog.
3. WorldLog
	- A materialized log of the InputLog via InputProcessor.
4. WorldProcessor
	- A materialized view of the WorldLog
	- Calculates the state change made by an Input Event
5. WorldStateLog
	- A materialized log of the WorldLog via WorldProcessor

### Examples

#### Auth Stream from CreateReq to CreateSuccessResp

	Client make
	CreateReq and sends
	  v
	(socket)
	  v
	ClientHandler makes
	CreateEvent and inserts
	  v
	PreAuthLog -> LongTermStore
	  v
	AuthProcessor makes
	NewActorEvent and inserts
	  v
	PostAuthLog -> LongTermStore
	  v
	ClientHandler transforms into
	ActorHandler and makes
	CreateSuccessResp and sends
	  v
	(socket)
	  v
	Client


#### Stream from input to rendering.

	Actor makes
	ActorInput and sends
	  v
	(socket)
	  v
	ActorHandler makes
	InputEvent and inserts
	  v
	InputLog -> LongTermStore
	  v
	InputProcessor makes
	WorldEvent and inserts
	  v
	WorldLog -> LongTermStore
	  v
	WorldProcessor calculates how
	the state of the world changes
	  v
	WorldStateLog -> LongTermStore
	  v
	ActorHandler
	- culls changes
	- diffs changes
	WorldStateDiff
	  v
	(socket)
	  v
	Actor
	- merges changes
	RenderEvent
	  v
	Renderer
