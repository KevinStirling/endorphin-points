package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/bwmarrin/discordgo"
	"github.com/redis/go-redis/v9"
)

var (
	RemoveCommands = flag.Bool("rmcmd", true, "Remove all commands after shutting down or not")
	AppId          = flag.String("appid", os.Getenv("APP_ID"), "The registered discord app id")
	DiscordToken   = flag.String("token", os.Getenv("DISCORD_TOKEN"), "Auth token for discord api")
	RedisHost      = flag.String("redis_host", os.Getenv("REDISHOST"), "Redis host addr")
	RedisPort      = flag.String("redis_port", os.Getenv("REDISPORT"), "Redis port")
	RedisUser      = flag.String("redis_user", os.Getenv("REDISUSER"), "Redis store username")
	RedisPass      = flag.String("redis_pass", os.Getenv("REDISPASSWORD"), "Redis store password")
)

func main() {
	// Create a new session using the DISCORD_TOKEN environment variable from Railway
	dg, err := discordgo.New("Bot " + *DiscordToken)
	if err != nil {
		fmt.Printf("Error while starting bot: %s", err)
		return
	}

	store := redis.NewClient(&redis.Options{
		Addr:     *RedisHost + ":" + *RedisPort,
		Username: *RedisUser,
		Password: *RedisPass,
	})

	ctx := context.Background()

	if ping := store.Ping(ctx); ping.Val() != "PONG" {
		fmt.Printf("Failed to connect to Redis: %s", ping)
		panic(ping)
	} else {
		fmt.Println("Redis connection established")
	}

	// ---------------
	// Server Commands
	// ---------------

	testCmd := &discordgo.ApplicationCommand{Name: "test", Description: "just a test"}
	_, err = dg.ApplicationCommandCreate(*AppId, "", testCmd)
	if err != nil {
		fmt.Printf("Could not create command: %s ", err)
		return
	}

	commands := []*discordgo.ApplicationCommand{

		{
			Name:        "newbet",
			Description: "create new bet",
			Options: []*discordgo.ApplicationCommandOption{
				{
					Type:        discordgo.ApplicationCommandOptionString,
					Name:        "title",
					Description: "What are we bettin on, boys",
					Required:    true,
				},
				{
					Type:        discordgo.ApplicationCommandOptionString,
					Name:        "outcome-1",
					Description: "Outcome option 1",
					Required:    true,
				},
				{
					Type:        discordgo.ApplicationCommandOptionString,
					Name:        "outcome-2",
					Description: "Outcome option 2",
					Required:    true,
				},
			},
		},
	}

	registeredCommands := make([]*discordgo.ApplicationCommand, len(commands))
	for i, v := range commands {
		cmd, err := dg.ApplicationCommandCreate(*AppId, "", v)
		if err != nil {
			fmt.Printf("Could not create new bet: %s  ", err)
			return
		}
		registeredCommands[i] = cmd
	}

	// -------------------
	// End Server Commands
	// -------------------

	// Add the command handler
	dg.AddHandler(commandCreate)

	// Add the message handler
	dg.AddHandler(messageCreate)

	// Add the Guild Messages intent
	dg.Identify.Intents = discordgo.IntentsGuildMessages

	// Connect to the gateway
	err = dg.Open()
	if err != nil {
		fmt.Printf("Error while connecting to gateway: %s", err)
		return
	}
	// Wait until Ctrl+C or another signal is received
	fmt.Println("The bot is now running. Press Ctrl+C to exit.")
	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt, os.Kill)
	<-sc

	if *RemoveCommands {
		fmt.Println("Removing commands...")
		for _, v := range registeredCommands {
			err := dg.ApplicationCommandDelete(dg.State.User.ID, "", v.ID)
			if err != nil {
				panic("Cannot delete " + v.Name + " command: " + err.Error())
			}
		}
	}

	// Close the Discord session
	dg.Close()
}

func commandCreate(s *discordgo.Session, i *discordgo.InteractionCreate) {

	data := i.ApplicationCommandData()
	options := i.ApplicationCommandData().Options
	switch data.Name {
	case "test":
		fmt.Println("Hit test command")
		err := s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: "Shut Up",
			},
		})
		if err != nil {
			panic(err.Error())
		}
	case "newbet":
		optionMap := make(map[string]*discordgo.ApplicationCommandInteractionDataOption, len(options))
		for _, opt := range options {
			optionMap[opt.Name] = opt
		}
		fmt.Println("Hit newbet command")
		margs := make([]interface{}, 0, len(options))
		msgformat := i.Member.User.Username + " started a bet!" + "\n"
		title, ok := optionMap["title"]
		outcome1, ok := optionMap["outcome-1"]
		outcome2, ok := optionMap["outcome-2"]
		if ok {
			margs = append(margs, title.StringValue(), outcome1.StringValue(), outcome2.StringValue())
			msgformat += "> %s\n > %s OR %s"
		}

		if err := s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: fmt.Sprintf(
					msgformat,
					margs...,
				),
			},
		}); err != nil {
			panic(err.Error())
		}

	}
	return

}

func messageCreate(s *discordgo.Session, m *discordgo.MessageCreate) {
	// Don't proceed if the message author is a bot
	if m.Author.Bot {
		return
	}

	if m.Content == "the body is bad" {
		s.ChannelMessageSend(m.ChannelID, "the body is not g0000ooood!")
		return
	}

}
