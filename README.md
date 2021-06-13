# steam-discount-notif-bot

## Available commands
* `/sub [game_id]` &mdash; subscribe by game id
	* `/sub 620`
* `/sub [steam store game link]` &mdash; subscribe by game link
	* `/sub https://store.steampowered.com/app/620/Portal_2/`
* `/subs` &mdash; show your subscriptions

## How it works?
Every 4 hours the bot checks if there are any discounts and notifies all corresponding users (by
sending a message to the original chat_id).

## Is it online?
The bot is hosted on Heroku, here is a link https://t.me/steam_discount_notif_bot
