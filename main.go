package main

import (
	"fmt"
	"os"
	"os/signal"
	"regexp"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/Mrs4s/MiraiGo/message"
	openaiezgo "github.com/Nigh/openai-ezgo"
	openai "github.com/sashabaranov/go-openai"
	"github.com/spf13/viper"
)

var baseurl string
var tokenLimiter int

var busyChannel map[string]bool

func init() {
	busyChannel = make(map[string]bool)
}

func main() {
	viper.SetDefault("gpttokenmax", 512)
	viper.SetDefault("gpttoken", "0")
	viper.SetDefault("token", "0")
	viper.SetDefault("baseurl", openai.DefaultConfig("").BaseURL)
	viper.SetConfigType("json")
	viper.SetConfigName("config")
	viper.AddConfigPath("./")
	err := viper.ReadInConfig()
	if err != nil {
		panic(fmt.Errorf("fatal error config file: %s", err))
	}
	gpttoken := viper.Get("gpttoken").(string)
	fmt.Println("gpttoken=" + gpttoken)
	tokenLimiter = viper.GetInt("gpttokenmax")
	fmt.Println("gpttokenmax=" + strconv.Itoa(tokenLimiter))
	baseurl = viper.Get("baseurl").(string)
	fmt.Println("baseurl=" + baseurl)

	qqbotInit()
	qqbotStart()

	cfg := openaiezgo.DefaultConfig(gpttoken)
	cfg.BaseURL = baseurl
	cfg.MaxTokens = tokenLimiter
	cfg.TimeoutCallback = func(from string, token int) {
		gid, _ := strconv.ParseInt(from, 10, 64)
		SendToQQGroup("连续对话已超时结束。共消耗token:`"+strconv.Itoa(token)+"`", gid)
	}
	openaiezgo.NewClientWithConfig(cfg)
	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt)
	<-sc
	fmt.Println("Bot will shutdown after 1 second.")

	<-time.After(time.Second * time.Duration(1))
}

func qqMsgHandler(msg *message.GroupMessage) {
	reply := func(words string) {
		SendToQQGroup(words, msg.GroupCode)
	}
	targetID := strconv.FormatInt(msg.GroupCode, 10)

	ok, text := GroupMsgParse(msg)
	if ok {
		words := strings.TrimSpace(text)
		if len(words) > 0 {
			if busyChannel[targetID] {
				reply("正在思考，请勿打扰。")
				return
			}
			busyChannel[targetID] = true
			defer func() {
				delete(busyChannel, targetID)
			}()
			rst := regexp.MustCompile(`重置对话.*`)
			if rst.MatchString(words) {
				reply(openaiezgo.EndSpeech(targetID))
				return
			}
			cmd := regexp.MustCompile(`调教\s*(.*)`)
			if cmd.MatchString(words) {
				reply(openaiezgo.NewCharacterSet(targetID, cmd.FindStringSubmatch(words)[1]))
				return
			}
			ans := openaiezgo.NewSpeech(targetID, words)
			if len(ans) > 0 {
				reply(ans)
			}
		}
		return
	}
}
