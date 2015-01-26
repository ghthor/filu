### Encoding

This encoding is a inspired from the
[socket.io-spec](https://github.com/LearnBoost/socket.io-spec/blob/master/README.md#encoding).

Messages have to be encoded before they're sent. The structure of a message is
as follows:

    [message type] ':' [message id ('+')] ':' ([message message]) ':' ([message payload])

The message type is a single digit integer.

The message id is an incremental integer, required for ACKs (can be ommitted).
If the message id is followed by a `+`, the ACK is not handled by socket.io,
but by the user instead. NOTE: This is currently Unimplemented.

The last 2 sections are conditional based on the `[message type]`.  This allows
the encoding to be extremely flexible.

### (`0`) Disconnect

Signals disconnection.

Example:

- Disconnect the whole socket

      0:::

### (`1`) Connect - Unused

This might be useful when this encoding is used over UDP.

### (`2`) Heartbeat - Currently Unimplemented

Sends a heartbeat. Heartbeats must be sent within the interval negotiated with
the server. It's up to the client to decide the padding (for example, if the
heartbeat timeout negotiated with the server is 20s, the client might want to
send a heartbeat evert 15s).

### (`3`) Message

This is the generic form of JSON Message and applies no restrictions on the
last field of the message. When sending a message the short string is something
consitent that can be switched upon and the long message is something that
could be useful when debugging or human readable version of the message.

    '3:' [message id ('+')] ':' [message short] ':' [message long]

Examples:

- Server response to successful login

    3:1:loginSuccess:

- Global chat broadcast

    3:1:chatBroadcast:Server will be shutting down in 10mins

### (`4`) JSON Message

 The json datatype could be part of the json object like this.

    { "type": "user", "obj": {...} }

I'm choose not to do this and break compatibility with socket.io becuase this
scheme causes the json to be unmarshaled twice. First to determine what
structure `obj` is going to have and the second time to unmarshal `obj` into
it's native type.

The [message json datatype] can used to determine what type of object the json
can be Unmarshaled into and avoid 2 step unmarshaling.

If the [json] errors when parsing the socket will send an error to the sender.

    '4:' [message id ('+')] ':' [message json datatype] ':' [json]

A JSON encoded message.

    4:1:createCharacter:{"name":"thundercleese"}

### (`5`) Event - Not Used

### (`6`) ACK

    '6::' [message id] ':' [data]

An acknowledgment contains the message id as the the first data field. The
second [data] field can used for more complex acknowledgement.

Example 1: simple acknowledgement

    6::4:

Example 2: complex acknowledgement

    6::4:["A","B"]

### (`7`) Error

    '7::' [reason] ':' [advice]

For example, if a connection to a sub-socket is unauthorized. This message has
the same structure as a message with the only difference being the [message
type].

### (`8`) Noop

No operation. Used for example to close a poll after the polling duration times
out.
