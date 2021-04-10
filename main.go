package main

import (
	"errors"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/bwmarrin/discordgo"
)

func main() {
	notifyChannelID := os.Getenv("DISCORD_NOTIFY_CHANNEL")
	if notifyChannelID == "" {
		panic(errors.New("DISCORD_NOTIFY_CHANNEL is required"))
	}

	discord, err := discordgo.New("Bot " + os.Getenv("DISCORD_TOKEN"))
	if err != nil {
		panic(err)
	}

	discord.AddHandler(func(s *discordgo.Session, m *discordgo.VoiceStateUpdate) {
		if m.BeforeUpdate != nil && m.BeforeUpdate.ChannelID == m.ChannelID {
			return
		}
		if m.BeforeUpdate == nil && m.ChannelID == "" {
			return
		}

		content := ":thinking_face:"
		if m.ChannelID == "" {
			content = fmt.Sprintf("[<#%s>] <@%s> left.", m.BeforeUpdate.ChannelID, m.UserID)
		} else if m.BeforeUpdate == nil {
			content = fmt.Sprintf("[<#%s>] <@%s> joined.", m.ChannelID, m.UserID)
		} else {
			content = fmt.Sprintf("[<#%s>] <@%s> moved from <#%s>.", m.ChannelID, m.UserID, m.BeforeUpdate.ChannelID)
		}
		_, err := s.ChannelMessageSendComplex(notifyChannelID, &discordgo.MessageSend{
			Content:         content,
			AllowedMentions: &discordgo.MessageAllowedMentions{},
		})
		if err != nil {
			panic(err)
		}
	})
	discord.Identify.Intents = discordgo.IntentsGuildVoiceStates

	if err := discord.Open(); err != nil {
		panic(err)
	}
	fmt.Println("Running")
	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt, os.Kill)
	<-sc
	discord.Close()
}
