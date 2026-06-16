package phrases

import (
	"github.com/TrixiS/goram"
	"github.com/TrixiS/goram/keyboards"
)

const (
	StartMessageTextFmt = `ID: <code>%d</code>
Username: <code>%s</code>`
	ChatButtonText    = "Чат"
	ChannelButtonText = "Канал"
	UserButtonText    = "Пользователь"
	IDFmt             = "<code>%d</code>"
	ChatIDRowFmt      = "ChatID: <code>%d</code>\n"
	UserIDRowFmt      = "UserID: <code>%d</code>"
)

const (
	ChatRequestID = iota
	ChannelRequestID
	UserRequestID
)

var (
	StartMarkup = &goram.ReplyKeyboardMarkup{
		Keyboard: keyboards.NewBuilder[goram.KeyboardButton]().
			Add(goram.KeyboardButton{
				Text: ChatButtonText,
				RequestChat: &goram.KeyboardButtonRequestChat{
					RequestID:     ChatRequestID,
					ChatIsChannel: false,
				},
			}).
			Add(goram.KeyboardButton{
				Text: ChannelButtonText,
				RequestChat: &goram.KeyboardButtonRequestChat{
					RequestID:     ChannelRequestID,
					ChatIsChannel: true,
				},
			}).
			Add(goram.KeyboardButton{
				Text: UserButtonText,
				RequestUsers: &goram.KeyboardButtonRequestUsers{
					RequestID:   UserRequestID,
					MaxQuantity: 1,
				},
			}).
			Adjust(2).
			Build(),
		IsPersistent:   true,
		ResizeKeyboard: true,
	}
)
