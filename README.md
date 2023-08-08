# tg-mafia

[Telegram](https://telegram.org) bot for moderating offline [Mafia](https://en.wikipedia.org/wiki/Mafia_(party_game)) game.\
Currently 4 sides availible:
* Mafia side: Mafia role
* Peaceful side: Peaceful role, Doctor, Witness, Sheriff
* Maniac side: Maniac role
* Role Guesser side: Role Guesser role

## Requirements

* **Golang** [1.20+]

## Libraries used

* [telegram-bot-api](https://github.com/go-telegram-bot-api/telegram-bot-api)\
Golang bindings for the Telegram Bot API
* [Mockery](https://github.com/vektra/mockery)\
A mock code autogenerator for Go

## Usage

Program tries to obtain environment variable `TELEGRAM_APITOKEN`
at start-up which will be used as a [bot token](https://core.telegram.org/bots/api#authorizing-your-bot).

_Interaction with bot has not yet been translated from Russian._

Команды для ввода боту:

* `/create [тэг роли [тэг роли [тэг роли [...]]]]`\
  Создает игровую комнату с _картами ролей_ соответствующими 
  указанным тэгам.
  
  Тэги:
  - Мафия: мафия, маф
  - Мирный: мирный, мир
  - Врач: врач, доктор, док
  - Свидетельница: свидетельница, свид
  - Комиссар: комиссар, ком, шериф
  - Маньяк: маньяк, ман, убийца
  - Разгадыватель: разгадыватель, раз
</br>

* `/join _код-комнаты_ [никнейм]` (или простая форма - `_код-комнаты_ [никнейм]`)\
  Присодиняет пользователя к игровой комнате с еще не начавшейся
  игрой. Когда никнейм не указан, используется telegram username.
  Когда число игроков становится равно числу _карт ролей_, указанных
  при создании, игра автоматически начинается. Карты ролей
  распределяются случайным образом.

* `/stop`\
  Останавливает комнату игры, в которой сейчас находится игрок.
  Возобновление на данный момент не предусмотрено.