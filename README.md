# Open CTP Server

[API Documentation](https://open-ctp-server.fly.dev/docs/)

## What is this?

This is the Open Call to Power Server project, or Open CTP Server for short.

This is a fan-developed server for the Call to Power series of games which intends to bring multiplayer to the franchise in a meaningful way (eventually supporting both Call to Power and Call to Power II). It is still a massive work-in-progress, as I only rarely have time to work on it. Contributions of all kinds are welcome, as long as they adhere to the guidelines outlined in `CONTRIBUTING.md` and as long as the contributions are done with other contributors in mind.

This project is **neither made nor sponsored by Activision**. This is a fan-made server with no affiliation to Activision. See `LICENSE.md` for more details.

## Endpoint

The server is currently deployed at https://open-ctp-server.fly.dev. To test connectivity, send a GET request to https://open-ctp-server.fly.dev/health and you should get back a JSON response of '{status:"OK"}'.

## Goals

There are multiple goals this project intends to tackle: 
- Create a proper backend for the Call to Power game series
- Support fan-made frontends (such as the Apolyton Edition and the Spyroviper Edition of the games) with minimal requirements
    - The server should be accessible and usable using cURL
- Allow for a vanilla+ experience for the games
    - Obvious bugs should not be preserved
    - Non-vanilla changes (such as tweaking unit health or adding extra units) should NOT be present
        - Feel free to fork this repo and modify it to your heart's content!
- Players will be able to choose a different username every game (the same as the base Call to Power multiplayer did)
    - Players will be able to password-protect their login to each game (so that others cannot join as them)
        - These will be temporary and will only last the duration of the game
        - No actual account will ever be required as this would mean a lot of work client-side for the Apolyton and Spyroviper Editions
- Additional game modes
    - Long-turn (with customizable turn timers) similar to Freeciv
    - Simultaneous Turns (everyone moves during one turn)
    - Hybrid Turns (Simultaneous Turns except for when nations are at war, where they take turns in-sequence)
    - Email game play will not be supported (as this gamemode has since become obsolete)

## Getting Started

### Requirements

- Environment with Bash
- Go Language Tools
- (Optional but highly recommended) Recommended Extensions

Quick Note: Please leave an issue on GitHub if you run into any problems with the setup process! This setup process is meant to be largely OS-agnostic (supporting Windows, MacOS and Debian-flavored Linux distros).

To get started, download the Go language tools and install them into an environment with Bash. Bash is necessary to run some of the scripts and is generally a good idea (though development is definitely possible without Bash).

If you are on Windows, one quick way to get Bash is by downloading and running Git Bash (part of Git for Windows): https://git-scm.com/download/win

If you are using VS Code, you should also get some extension recommendations. I highly recommend you install them, the spell checker is very nice and has saved many headaches!

### Getting Dependencies

Run `go install`. This will install all Go module dependencies.

You can also make use of the provided Taskfile by downloading Task from https://taskfile.dev if you would like (this is optional, but recommended). You can also just run the commands specified in the Taskfile directly, of course. For the rest of this README, the Task commands will be used in lieu of the commands they call. 

### Starting the Development Server

Before you can start the development server and effectively use it, you must either set up a Supabase account or use the Supabase CLI to mimic the database.

#### Using the Supabase CLI (recommended)

To use the Supabase CLI for database development on your local machine, first download the Supabase CLI app from here: https://supabase.com/docs/guides/cli/getting-started?queryGroups=platform&platform=linux

(optional) After you have downloaded the application and it is in your $PATH, log in if you have a Supabase account: `supabase login`

After that, run `supabase start` to start the Supabase dashboard for local development. This will allow you to interact with the database as a local copy without needing credentials to do so.

Once you have run `supabase start`, you can visit http://localhost:54323 to see your local Supabase dashboard.

In your `.env` file, fill out the following with the values gotten from the end of the `supabase start` command: 
```
SUPABASE_PROJECT_URL=(API URL)
SUPABASE_API_KEY=(anon key)
SUPABASE_SERVICE_ROLE=(service_role key)
SUPABASE_DB_URL=(DB URL)?sslmode=disable
```

Now, when you start the server, you should be able to successfully connect to the local database. Please take extra note of the "?sslmode=disable" at the end of SUPABASE_DB_URL. If you create a game, for example, you should see it show up in your local Supabase dashboard under the Table Editor.

#### Using the Supabase Dashboard

After setting up your Supabase account and project (both are free), you must add these values to a `.env` file located at the root of the project (next to `main.go`):

```
SUPABASE_PROJECT_URL=(your project URL here)
SUPABASE_API_KEY=(your API key here)
SUPABASE_SERVICE_ROLE=(your service role here)
```

Once you do this, you should be able to write to the database.

The Taskfile has multiple commands you can run that can save time. For example, to start the development server, you can run this with `bin` not on your PATH variable: 
```
task run
```

### Starting the Test Suite

You can also use this to start the test suite: 
```
task test
```

## Why Go?

To accomplish these goals with a small footprint, Go was chosen. Go was chosen in particular because Go is extremely fast, has very low memory requirements and is able to handle concurrency extremely well. It is also extremely easy to learn, which is important since the fans of this series whom are programmers are diverse in both preferred language and preferred stack. 

If you are trying to make a client and are struggling with using this server, feel free to reach out to Ninjaboy on Discord!

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
- Communication with the server (via HTTP/S) within the format specified by the API
- Addition of password fields for use with account creation (creating online accounts will require a password)

### A Few Notes

**This server will NOT work with either of the base, unmodified games**.

It is notable that this server will NOT function exactly the same as the vanilla CTP / CTP2 servers did, so clients will need to be modified to actually utilize this server.

## Tooling

See `TOOLING.md`.

## Contributing

See `CONTRIBUTING.md` for information on how to contribute to this effort. All skill levels are welcome!

## Can I Fork This Repo?

You absolutely can, so long as you follow the license in `LICENSE.md`. I am really looking forward to server-side modifications in the future!
