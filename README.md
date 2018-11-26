
<p align="center">
  <h3 align="center">Legion</h3>
  <p align="center">Fast, modular, and extensible peer to peer library in Go</p>
  <p align="center">
    <a href="https://godoc.org/github.com/gladiusio/legion/"><img src="https://godoc.org/github.com/gladiusio/legion/network?status.svg"></a>
    <a href="https://travis-ci.com/gladiusio/legion"><img src="https://travis-ci.com/gladiusio/legion.svg?branch=master"></a>
    <a href="https://goreportcard.com/report/github.com/gladiusio/legion"><img src="https://goreportcard.com/badge/github.com/gladiusio/legion"></a>
	<a href="https://opensource.org/licenses/MIT"><img src="https://img.shields.io/badge/License-MIT-yellow.svg"></a>
  </p>
</p>

---

Legion is an easy to use, fast, and bare-bones peer to peer library designed to leave most of the network characteristics up to the
user through a simple yet powerful plugin system. It was written because here at [Gladius](https://gladius.io) we needed a peer
to peer overlay for our applications, and existing solutions were not a perfect fit for our use case.

We would also like to link to the Perlin Network's [Noise](https://github.com/perlin-network/noise), as their API inspired
a lot of our design philosophy and you should totally check it out.

## Overview

- User defined messages, allowing you to build your own cryptography and message validation system
- Powerful plugin system with easy to use event context
- Single TCP connection opened to each peer, where messages are sent over multiplexed streams
- User defined logging via a [logging interface](#custom-logger)

## Background

### Basic Concepts
There are two types of peers in a Legion network, promoted and unpromoted. This allows a user to connect to a peer and message it without having it generally sendable via a call to Broadcast, so that you can have some authentication or validation of a remote peer without having to code special logic in your plugins to handle that scenario. 

Messages are extremely customizable and simple, we have no cryptography by default and don't enforce any specific message body requirements. We also 
allow a user to set your own message validator to check an incoming message, which means you can add in your own crypto, compression, or really any other criteria you want.

## Usage

This should be considered a quick start guide, there are more examples in the
[examples folder](https://github.com/gladiusio/legion/tree/readme-cleanup/examples) and in the
[Gladius Network Gateway](https://github.com/gladiusio/gladius-network-gateway)

### Basic Usage
Here we create a legion object with a default config, wait until it's listening, add a peer, and promote it so it is generally sendable.
```golang
func main(){
    // Build a basic config
    conf := legion.SimpleConfig("localhost", 7947)
    
    // Build a new legion object from the config
    l := legion.New(conf)

    // Listen in a new goroutine
    go l.Listen()
    // Wait until the network is listening
    l.Started()

    // Dial a peer and add it to the non-messagable peers
    err := l.AddPeer(utils.LegionAddressFromString("localhost:7946"))
    if err != nil {
        panic(err)
    }

    // Make that peer sendable by promoting it
     err := l.PromotePeer(utils.LegionAddressFromString("localhost:7946"))
    if err != nil {
        panic(err)
    }

    // Block forever
    select {}
}
```

### Messaging
There are several ways to send messages to peers in the network:
```golang
func main(){
    // ... build our legion object

    // Dial a peer and add it to the non-messagable peers
    err := l.AddPeer(utils.LegionAddressFromString("localhost:7946"))
    if err != nil {
        panic(err)
    }

    // Will send to all promoted peers
    l.Broadcast(message.New(config.BindAddress,"ping", []byte("ping"))

    // Send to a specific peer
    l.Broadcast(message.New(config.BindAddress,"ping", []byte("ping"), 
        utils.LegionAddressFromString("localhost:7946"))
    
    // Broadcast to a random 5 promoted peers
    l.BroadcastRandom(message.New(config.BindAddress,"ping", []byte("ping"), 5)

    // Block forever
    select {}
}
```
You can also set your own message validation function by setting the config value `MessageValidator` with:
```golang
config.MessageValidator = func(m *message.Message) bool {
    // Do your validation here and return true if valid and false if not
    return true
}
```
Any message that is not valid will not be passed to plugins.

### Custom Logger
The internal logger is a generic type that can be overridden by the user as long
as your logger meets the requirements below:
```go
// GenericLogger is the logger interface that legion uses, you can plug in
// your own logger as long as your logger implements this interface.
type GenericLogger interface {
	// Base log types
	Debug() GenericContext
	Info() GenericContext
	Warn() GenericContext
	Error() GenericContext

	// Add context like logger.With(NewContext().Field("test", "val"))
	With(ctx GenericContext) GenericLogger
}

// GenericContext provides a way to add fields to a log event
type GenericContext interface {
	Field(key string, val interface{}) GenericContext

	// Actually log the built up log line with the message
	Log(msg string)
}
```
You can register a new logger by calling

```go
logger.SetLogger(YourLogger)
```
By default we use [zerolog](https://github.com/rs/zerolog), you can see our implemented logger [here](https://github.com/gladiusio/legion/blob/master/logger/zerolog_logger.go).
If you want to edit the underlying zerolog instance, you can call:

```go
l := logger.GetLogger() // Get the wrapper
zerologger := l.(logger.ZeroLogger).Logger // Get the actual Zerolog instance (can change things like the formatting, output location, etc)
```

### Plugins

#### Writing your own

A plugin is any struct that implements the plugin interface:

```go
type PluginInterface interface {
	NewMessage(ctx *MessageContext)
	PeerAdded(ctx *PeerContext)
	PeerPromotion(ctx *PeerContext)
	PeerDisconnect(ctx *PeerContext)
	Startup(ctx *NetworkContext)
	Close(ctx *NetworkContext)
}
```

If you don't need all of these methods, you can use our handy GenericPlugin as an
[anonymous field](http://golangtutorials.blogspot.com/2011/06/anonymous-fields-in-structs-like-object.html)
in your struct, like this:

```go
type MyPlugin struct {
	network.GenericPlugin
	specialData string
}

func (m *MyPlugin) NewMessage(ctx *MessageContext) {
	fmt.Println(mspecialData)
}
```

by doing this, you only need to implement the methods you need and still conform to the interface.

#### Included plugins

Legion includes a few plugins to make your life easier, you can add them like this:

```go
l := legion.New(conf)
// Register our simple discovery plugin which sends lists of known peers to newly connected peers, as 
// well as informing it's peers of the new connection
l.RegisterPlugin(new(simpledisc.Plugin))

```