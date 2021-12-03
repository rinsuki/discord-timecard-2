package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/signal"
	"syscall"

	"github.com/bwmarrin/discordgo"
	"github.com/goccy/go-yaml"
)

type Config struct {
	Channels map[string]string `yaml:"channels"`
}

type Message struct {
	AfterChannel  string
	BeforeChannel string
	NotifyChannel string
}

func (m *Message) ToString(userID string) string {
	if m.AfterChannel == "" && m.BeforeChannel == "" {
		return "invalid state"
	} else if m.AfterChannel == "" {
		return fmt.Sprintf("[<#%s>] <@!%s> left.", m.BeforeChannel, userID)
	} else if m.BeforeChannel == "" {
		return fmt.Sprintf("[<#%s>] <@!%s> joined.", m.AfterChannel, userID)
	} else {
		return fmt.Sprintf("[<#%s>] <@!%s> moved from <#%s>.", m.AfterChannel, userID, m.BeforeChannel)
	}
}

func main() {
	configBytes, err := ioutil.ReadFile("config.yml")
	if err != nil {
		panic(err)
	}
	var config Config
	if err := yaml.Unmarshal(configBytes, &config); err != nil {
		panic(err)
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
		messages := []Message{}
		if m.BeforeUpdate != nil && m.GuildID != m.BeforeUpdate.GuildID {
			// サーバー間移動
			messages = append(messages, Message{
				BeforeChannel: m.BeforeUpdate.ChannelID,
				NotifyChannel: config.Channels[m.BeforeUpdate.GuildID],
			})
			messages = append(messages, Message{
				AfterChannel:  m.ChannelID,
				NotifyChannel: config.Channels[m.GuildID],
			})
		} else {
			beforeChannel := ""
			if m.BeforeUpdate != nil {
				beforeChannel = m.BeforeUpdate.ChannelID
			}
			messages = append(messages, Message{
				BeforeChannel: beforeChannel,
				AfterChannel:  m.ChannelID,
				NotifyChannel: config.Channels[m.GuildID],
			})
		}
		for _, message := range messages {
			if message.NotifyChannel == "" {
				continue
			}
			_, err := s.ChannelMessageSendComplex(message.NotifyChannel, &discordgo.MessageSend{
				Content:         message.ToString(m.UserID),
				AllowedMentions: &discordgo.MessageAllowedMentions{},
			})
			if err != nil {
				panic(err)
			}
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
