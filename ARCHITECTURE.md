# Architecture

This document aims to explain the server architecture behind this project.

## Deployment

The server is automatically deployed to Fly.io (https://fly.io/).

Fly.io was chosen mainly because of ease-of-use - deploying is incredibly easy. They also have built-in Sentry integration for errors.

If you wish to have a setting changed on Fly.io, please message Ninjaboy on Discord. 

You are also free to host your own instance of this repo wherever you would like (with the only condition being that it cannot be sold or make money in any way, as per the license - it must be for personal, non-commercial use only). If you want to use this server code wholesale on your own server (or modify this server, then host it), that is fine for non-commercial use.

## Database

This server uses Supabase for its database. Supabase uses Postgres.

For details on how to set up Supabase for local development (so you do not have to create an account), see README.md. If you would like to add tables to the database, please message Ninjaboy on Discord.