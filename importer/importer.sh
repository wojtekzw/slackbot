#!/bin/bash
touch $1
> $1
# robots=(
#     "github.com/wojtekzw/slackbot/robots/decide"
#     "github.com/wojtekzw/slackbot/robots/bijin"
#     "github.com/wojtekzw/slackbot/robots/nihongo"
#     "github.com/wojtekzw/slackbot/robots/ping"
#     "github.com/wojtekzw/slackbot/robots/roll"
#     "github.com/wojtekzw/slackbot/robots/store"
#     "github.com/wojtekzw/slackbot/robots/wiki"
#     "github.com/wojtekzw/slackbot/robots/bot"
# )

robots=(
    "github.com/wojtekzw/slackbot/robots/ping"
    "github.com/wojtekzw/slackbot/robots/block"
    "github.com/wojtekzw/slackbot/robots/bots"
)

echo "package importer

import (" >> $1

for robot in "${robots[@]}"
do
    echo "    _ \"$robot\" // automatically generated import to register bot, do not change" >> $1
done
echo ")" >> $1

gofmt -w -s $1
