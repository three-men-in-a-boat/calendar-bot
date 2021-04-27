package baseKeyboards

import tb "gopkg.in/tucnak/telebot.v2"

func getCommands() []string {
	return []string{"/today", "/next", "/date", "/create", "/about"}
}

func HelpCommandKeyboard() [][]tb.ReplyButton {

	var keyboard = make([][]tb.ReplyButton, 0)
	for idx, command := range getCommands() {
		if idx%2 == 0 {
			keyboard = append(keyboard, []tb.ReplyButton{})
		}
		keyboard[idx/2] = append(keyboard[idx/2], tb.ReplyButton{Text: command})
	}


	return keyboard
}
