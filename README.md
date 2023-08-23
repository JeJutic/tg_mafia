# tg-mafia

[Telegram](https://telegram.org) bot for moderating offline [Mafia](https://en.wikipedia.org/wiki/Mafia_(party_game)) game.\
Currently 4 sides availible:
* Mafia side: Mafia role
* Peaceful side: Peaceful role, Doctor, Witness, Sheriff
* Maniac side: Maniac role
* Role Guesser side: Role Guesser role

Godoc for is available [here](https://pkg.go.dev/github.com/jejutic/tg_mafia/pkg).
Package `gameserver` can be used for adding mafia game system (server) on top of any
users-server communication as shown in [web_mafia](https://github.com/jejutic/web_mafia).

## Requirements

* **Golang** [1.20+]
* **Telegram bot token**
* **PostgreSQL** [14] running instance

## Libraries used

* [telegram-bot-api](https://github.com/go-telegram-bot-api/telegram-bot-api)\
Golang bindings for the Telegram Bot API
* [Mockery](https://github.com/vektra/mockery)\
A mock code autogenerator for Go
* [pgx](https://github.com/jackc/pgx)\
PostgreSQL Driver and Toolkit

## Usage

Program tries to obtain environment variables `TELEGRAM_APITOKEN`,
which will be used as a [bot token](https://core.telegram.org/bots/api#authorizing-your-bot),
and `POSTGRES_URI`, which will be used as [connection string to PostrgreSQL](https://www.postgresql.org/docs/current/libpq-connect.html#LIBPQ-CONNSTRING), at start-up.

Interaction with bot has not yet been translated from Russian. 
You can find commands [here](pkg/gameserver/startText.txt) or with `/start` command.
