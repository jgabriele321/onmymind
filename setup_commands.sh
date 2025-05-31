#!/bin/bash

# Replace BOT_TOKEN with your actual token from .env
BOT_TOKEN=$(grep BOT_TOKEN .env | cut -d '=' -f2)

# Command definitions
COMMANDS='[
  {"command":"add","description":"Store new text"},
  {"command":"pull","description":"Get a random item"},
  {"command":"delete","description":"Delete the last pulled item"},
  {"command":"list","description":"Show all stored items"},
  {"command":"deleted","description":"Show deleted items"},
  {"command":"export","description":"Download a backup of all your data"},
  {"command":"time","description":"Calculate times, convert formats, or check time zones"},
  {"command":"undo","description":"Restore the last deleted item (within 1 hour)"},
  {"command":"help","description":"Show help message"}
]'

# Set commands
curl -X POST \
  https://api.telegram.org/bot${BOT_TOKEN}/setMyCommands \
  -H 'Content-Type: application/json' \
  -d "{\"commands\":${COMMANDS}}"

echo -e "\nCommands have been set up. They should now appear when you type / in your chat with the bot." 