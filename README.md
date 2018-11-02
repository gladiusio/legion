
<p align="center">
  <h3 align="center">Legion</h3>
  <p align="center">Fast, modular, and extensible peer to peer library in Go</p>
  <p align="center">
    <a href="https://godoc.org/github.com/gladiusio/legion/"><img src="https://godoc.org/github.com/gladiusio/legion/network?status.svg"></a>
    <a href="https://travis-ci.com/gladiusio/legion"><img src="https://travis-ci.com/gladiusio/legion.svg?branch=master"></a>
    <a href="https://goreportcard.com/report/github.com/gladiusio/legion"><img src="https://goreportcard.com/badge/github.com/gladiusio/legion"></a>
  </p>
</p>

---

Legion is an easy to use, fast, and bare-bones peer to peer library designed to leave most of the network characteristics up to the
user through a simple yet powerful plugin system. It was written because here at [Gladius](https://gladius.io) we needed a peer
to peer overlay for our applications, and existing solutions were not a perfect fit for our use case.

We would also like to link to the Perlin Network's [Noise](https://github.com/perlin-network/noise), as their API inspired
a lot of our design philosophy and you should totally check it out.

## Overview

- User defined messages (allows you to build your own cryptography)
- Powerful plugin system with easy to use event context
- Single TCP stream opened to each peer
- User defined logging

## Background

TODO

## Usage

This should be considered a quick start guide, there are more examples in the
[examples folder](https://github.com/gladiusio/legion/tree/readme-cleanup/examples) and in the
[Gladius Network Gateway](https://github.com/gladiusio/gladius-network-gateway)

### Basic Usage
TODO
```golang

```

### Messaging
TODO
```golang

```

### Custom Logger
The internal logger is a generic type that can be overridden by the user as long
as the logger meets the interface below:
```go
type Generic interface {
	Debug(msg string, keyvals ...interface{})
	Info(msg string, keyvals ...interface{})
	Warn(msg string, keyvals ...interface{})
	Error(msg string, keyvals ...interface{})

	With(keyvals ...interface{}) Generic
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

A plugin is any struct that implements the plugin interface:

```go
type PluginInterface interface {
	NewMessage(ctx *MessageContext)
	PeerAdded(ctx *PeerContext)
	PeerPromotion(ctx *PeerContext)
	PeerDeleted(ctx *PeerContext)
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

func (m *MessagePlugin) NewMessage(ctx *MessageContext) {
	fmt.Println(mspecialData)
}
```

by doing this, you only need to implement the methods you need and still conform to the interface.
