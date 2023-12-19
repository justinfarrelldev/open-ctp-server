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
- Players will be able to choose a different username every game (same as base Call to Power multiplayer did)
    - Players will be able to password-protect their login to each game (so that others cannot join as them)
        - These will be temporary and will only last the duration of the game
        - No actual account will ever be required as this would mean a lot of work client-side for the Apolyton and Spyroviper Editions
- Additional game modes
    - Long-turn (with customizable turn timers) similar to Freeciv
    - Simultaneous Turns (where everyone moves during one turn)
    - Hybrid Turns (Simultaneous Turns except for when nations are at war, where they take turns in-sequence)
    - Email game play will not be supported (as this gamemode has since become obsolete)

## Getting Started

### Requirements

- Environment with Bash
- Go Language Tools
- (Optional but highly recommended) Recommended Extensions

To get started, download the Go language tools and install them into an environment with Bash. Bash is necessary to run some of the scripts and is generally a good idea (though development is definitely possible without Bash).

If you are on Windows, one quick way to get Bash is by downloading and running Git Bash (part of Git for Windows): https://git-scm.com/download/win

If you are using VS Code, you should also get some extension recommendations. I highly recommend you install them, the spell checker is very nice and has saved many headaches!

### Getting Dependencies

First, run `go install`. This will install all Go module dependencies.

To get all of the binary dependencies as well, simply run: 
```
go generate
```

This will trigger `scripts/deps.sh` which will grab the necessary dependencies and store them in a folder in the project, `bin`.

From here, I **highly** recommend adding the `bin` folder to your PATH variable so that you can run them from your command line. Alternatively, you can run them from the command line directly when you need them with, for example, `./bin/task [...args]` (where `./bin/task` is the executable you want to run, and `[...args]` are the arguments you want to supply).

To add this folder to your PATH, try running the following command from the root of the project (replacing .bashrc with .zshrc if you are using Zsh, etc.). This should append to your PATH at the bottom of your .bashrc file: 
```
echo 'export PATH=$PATH:'$(pwd)/bin >> ~/.bashrc
```

While adding this, also add the following command (for protoc-gen-go, the Go plugin for protoc): 
```
echo 'export PATH=$PATH:'$(go env GOPATH)/bin >> ~/.bashrc
```

Once you reload your terminal, you should have access to both `protoc` and `task` directly in your terminal (as long as the `bin` folder still exists with those executables inside).

## Running Specific Commands

Once you have installed the dependencies, you should have an executable in the `bin` folder called `task`. This is what you can use to run the `Taskfile.yml` file in the root of the project. 

### Starting the Development Server

The Taskfile has multiple commands you can run that can save time. For example, to start the development server, you can run this with `bin` not on your PATH variable: 
```
./bin/task run
```
Alternatively, you can run this if you have `bin` on your PATH: 
```
task run
```

For the rest of this README, I will refer to the commands as if you have `bin` in your PATH.

### Starting the Test Suite

You can also use this to start the test suite: 
```
task test
```

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

## What should be handled by this server and what should be handled by the client?

### Responsibilities of this Server
- Unit moves (though clients should implement their own algorithm for unit move submissions - the server will simply check if moves are valid and then perform them)
- World generation (the server will generate a world given a set of parameters and then send the world information - such as tiles, ruins, goods, etc. - back as a response)
- Combat (with both armies' information supplied and the tiles they are on, the server will calculate who is victorious)
- Fog of War (the server will only return information which the client can see, thus reducing cheating)
- Chatting
- Players losing connection, low connectivity, etc.
- Diplomacy
- AI at a later point (the goal is to have human players supported first)
- Minimal downtime
- Promises made within the "Goals" section

### Clients 
- Unit graphics and fields associated with them (DefaultIcon, DefaultSprite, etc. from the `Units.txt` file that shipped with the original game will not be present on the types provided by the server). 
    - The extra fields should be handled by extension of the types provided by the server's types (IE, importing the Abolitionist type from this server's package and then using an AbolitionistClient type which extends that for the extra graphics info)
- All graphics files / anything related to graphics or sound
    - Even though "diplomacy", for example, is listed in the "server will handle" section, the client is responsible for handling all graphical portions of the exchange (with the server being responsible for sending the message to the other player involved)
- Communication with the server (whether via gRPC or via HTTP/S) within the format specified by the API

### A Few Notes

It is notable that this server will NOT function exactly the same as the vanilla CTP / CTP2 servers did, so clients will need to be modified to actually utilize this server. This will likely include a large jump in version for original-source-based repos (such as the Spyroviper Edition and Apolyton Edition) since the protobuffers target newer versions of C++. 

Alternatively, new clients can be coded from scratch into any protobuffer-supported language and may use this server.

**Another very important note for client developers** - the default values set on the protos are the values the server is set to anticipate. Therefore, even though there are setters, do not set the values on units, structures or any other type within the client unless you really know what you want to do.

Doing so will cause the client to desync from the server and, in the case of bad data being sent to the server, your client may 
become disconnected from the server (as an attempt to thwart cheating).

## Tooling

See `TOOLING.md`.

## Contributing

See `CONTRIBUTING.md` for information on how to contribute to this effort. All skill levels are welcome!

## Can I Fork This Repo?

You absolutely can, so long as you follow the license in `LICENSE.md`. I am really looking forward to server-side modifications in the future!
