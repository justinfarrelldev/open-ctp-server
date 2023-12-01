# Open CTP Server

## What is this?

This is the Open Call to Power Server project, or Open CTP Server for short.

This is a fan-developed backend for the Call to Power series of games which intends to bring multiplayer to the franchise in a meaningful way. 

## Goals

There are multiple goals this project intends to tackle: 
- Create a proper backend for the Call to Power games series
- Support fan-made frontends (such as the Apolyton Edition and the Spyroviper Edition of the games, or future fan remakes) with no dependencies required
    - The server should be accessible and completely usable using cURL
- Allow for a vanilla+ experience for the games
    - Obvious bugs should not be preserved
    - Non-vanilla changes (such as tweaking unit health or adding extra units) should NOT be present
        - Feel free to fork this repo and modify it to your heart's content!

## Why Go?

To accomplish these goals with a small budget, Go was chosen. I chose Go in particular because Go is extremely fast, has very low memory requirements, is able to handle concurrency extremely well and is extremely easy to learn. 

Go also has an excellent gRPC connector as well as great Protobuf support which should enable the following languages to interface with the server (many other languages are supported as well by third-party tools): 
- C++
- C#
- Java
- Kotlin
- Objective-C
- PHP
- Python
- Ruby
- Dart
- Go

This means that interfacing with this server should be extremely easy if a client is written in any of the above languages. See the Protocol Buffers documentation for more details on how this is possible, and see the "Tooling" section for more info on how this technology is implemented in this repository.

## Tooling

There is too much to cover in this section here, so please see `TOOLING.md`.

## Contributing

Please see `CONTRIBUTING.md` for information on how to contribute to this effort. All skill levels are welcome!
