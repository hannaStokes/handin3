First open one terminal, which will act as the server, by running server.go
Next, open up any number of additional terminals, depending on how many clients you want to use. In these you run client.go or client.go -name <name> if you want to name the client
Whenever a new client is run, and it automatically subscribes to the server, a message is sent to all other clients.
To write from one client to the other clients, simply enter the message you'd like to send in the terminal.
If you want to disconnect a client, close the terminal running it. There is no other way to disconnect (ie. no command or special message you can send).
