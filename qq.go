package main

import (
	"sync"

	"github.com/Logiase/MiraiGo-Template/bot"
	"github.com/Mrs4s/MiraiGo/client"
	"github.com/Mrs4s/MiraiGo/message"
)

type qq struct {
}

var instance *qq

type QQMsg struct {
	Type    int
	Content string
}

func init() {
	instance = &qq{}
	bot.RegisterModule(instance)
}

var validGroupId int64 = 0

func SetGroupID(n int64) {
	validGroupId = n
}

var externMsgHandler func(msg *message.GroupMessage)

func OnQQMsg(handler func(msg *message.GroupMessage)) {
	externMsgHandler = handler
}

func SendToQQGroup(content string, groupId int64) int32 {
	m := message.NewSendingMessage().Append(message.NewText(content))
	ret := bot.Instance.SendGroupMessage(groupId, m)
	return ret.Id
}

func (a *qq) MiraiGoModule() bot.ModuleInfo {
	return bot.ModuleInfo{
		ID:       "kook.route",
		Instance: instance,
	}
}

func (a *qq) Init() {
}

func (a *qq) PostInit() {
}

func (a *qq) Serve(b *bot.Bot) {
	b.GroupMessageEvent.Subscribe(func(c *client.QQClient, msg *message.GroupMessage) {
		go externMsgHandler(msg)
	})
}

func (a *qq) Start(bot *bot.Bot) {
}

func (a *qq) Stop(bot *bot.Bot, wg *sync.WaitGroup) {
	defer wg.Done()
}

func GroupMsgParse(msg *message.GroupMessage) (ok bool, text string) {
	ok = false
	for _, elem := range msg.Elements {
		switch e := elem.(type) {
		case *message.TextElement:
			text = text + e.Content
		case *message.AtElement:
			if e.Target == bot.Instance.Uin {
				ok = true
			}
		default:
		}
	}
	return
}
