# Open CTP Server

## What is this?

This is the Open Call to Power Server project, or Open CTP Server for short.

This is a fan-developed server for the Call to Power series of games which intends to bring multiplayer to the franchise in a meaningful way (eventually supporting both Call to Power and Call to Power 2).

This project is neither made nor sponsored by Activision. See `LICENSE.md` for more details.

## Goals

There are multiple goals this project intends to tackle: 
- Create a proper backend for the Call to Power games series
    - Both Call to Power 1 and Call to Power 2
- Support fan-made frontends (such as the Apolyton Edition and the Spyroviper Edition of the games, or future fan remakes) with minimal requirements
    - The server should be accessible and completely usable using cURL or gRPC
- Allow for a vanilla+ experience for the games
    - Obvious bugs should not be preserved
    - Non-vanilla changes (such as tweaking unit health or adding extra units) should NOT be present
        - Feel free to fork this repo and modify it to your heart's content!

## Why Go?

To accomplish these goals with a small footprint, Go was chosen. I chose Go in particular because Go is extremely fast, has very low memory requirements, is able to handle concurrency extremely well. It is also extremely easy to learn, which is important since the programmer fans of this series are diverse in both preferred language and preferred stack. 

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

This means that interfacing with this server should be relatively straightforward if a client is written in any of the above languages. See the Protocol Buffers documentation for more details on how this is possible, and see the "Tooling" section for more info on how this technology is implemented in this repository.

If you are trying to make a client and are struggling with the protobuffer implementation, feel free to reach out to Ninjaboy on Discord!

## Tooling

See `TOOLING.md`.

## Contributing

See `CONTRIBUTING.md` for information on how to contribute to this effort. All skill levels are welcome!
