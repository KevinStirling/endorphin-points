---
title: Discord Golang bot
description: A Discord bot written in Golang
tags:
- discordgo
- golang
---

## 💁‍♀️ How to use

- Install dependencies `go mod download`
- Connect to your Railway project `railway link`
- Start the bot `railway run go run main.go`

## 📝 Notes

When deploying via the cli, you may need to restart (not redloy!) from the web ui if a
command is registering as invalid. This could have something to do with the RemoveCommand
logic.