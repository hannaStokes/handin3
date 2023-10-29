# handin3

# System Requirements

R1: The service ChittyChat has server-side streaming. The client asks for a subscription to the server stream. The server returns a stream consisting of chat messages it receives from other clients. 

R2: GRPC_ARG_MAX_SEND_MESSAGE_LENGTH ?

R3: There is an endpoint where clients can "publish" aka send messages, and then the server returns an accept message. The messgae is "broadcasted" via the server's stream that the clients are subscribed to. This is the broadcasting part of our program.

R4: Because of this requirement our implementation considers a message receival as an event that increments the lamport time

R5: via the subscribe endpoint

R6: via the subscribe endpoint

R7: to allow for chat clients to drop out and for this to be visible to the other clients we changed the Subscribe method in the server. The for loop waiting for messages from the stream was changed to a go routine followed by the method stream.Context().Done().

R8: to implement this we need to send a message before stream.Context().Done() to the stream.

# Technical Requirements


