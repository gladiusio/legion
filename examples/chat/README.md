# Legion chat example
A simple chat example between two peers showing how to handle new peers, message processing and sending, as well startup events.

## How to run

Start the first peer with: 

```bash
go run main.go -bindaddress localhost:6000
```

Start the second peer with:

```bash
go run main.go -bindaddress localhost:6001 -remote localhost:6000
```

This will connect both peers together and allow you to send messages between the two. Note that this example does not include peer discovery, so you will have to manually connect peers.