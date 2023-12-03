# Architecture

This document aims to explain the server architecture behind this project.

## Deployment

The server is automatically deployed to Fly.io (https://fly.io/) and is accessible at <TODO />.

Fly.io was chosen mainly because of ease-of-use - deploying is incredibly easy. They also have built-in Sentry integration for errors.

If you wish to have a setting changed on Fly.io, please message Ninjaboy on Discord. 

You are also free to host your own instance of this repo wherever you would like (with the only condition being that it cannot be sold or make money in any way, as per the license - it must be for personal, non-commercial use only). If you want to use this server code wholesale on your own server (or modify this server, then host it), that is fine for non-commercial use.

## Database

This server necessitates a Postgres database. Postgres was chosen in particular since a lot of the game information is highly relational.

The server deployed to Fly.io is connected to a Supabase instance. There are environment variables which correlate to the Supabase instance in the cloud, these are as follows: 
```
DATABASE_ENDPOINT=<Database Endpoint Here>
DATABASE_KEY=<Public Database Key Here>
DATABASE_SECRET=<Private Secret Here>
```

To work with Postgres for this repo, you must set up a local Postgres instance and ensure that is started before beginning the relevant development. From there, set DATABASE_ENDPOINT to your localhost URL of your development database. This will ensure you have the same database as supplied by Supabase without actually requiring any Supabase authentication. 


