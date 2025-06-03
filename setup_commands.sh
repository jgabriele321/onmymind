#!/bin/bash

# Replace BOT_TOKEN with your actual token from .env
BOT_TOKEN=$(grep BOT_TOKEN .env | cut -d '=' -f2)

# Command definitions
COMMANDS='[
  {"command":"remindme","description":"Set a one-time or recurring reminder, use -call for priority"},
  {"command":"reminders","description":"List all reminders with optional filter"},
  {"command":"delete","description":"Delete a specific reminder"},
  {"command":"complete","description":"Mark a reminder as completed"},
  {"command":"time","description":"Calculate times, convert formats, or check time zones"},
  {"command":"help","description":"Show help message"}
]'

# Set commands
curl -X POST \
  https://api.telegram.org/bot${BOT_TOKEN}/setMyCommands \
  -H 'Content-Type: application/json' \
  -d "{\"commands\":${COMMANDS}}"

echo -e "\nCommands have been set up. They should now appear when you type / in your chat with the bot." 