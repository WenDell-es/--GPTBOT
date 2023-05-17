package constants

var (
	UnExpectedResp = []string{
		"The response was filtered due to the prompt triggering  OpenAI",
	}
)

const (
	GPT3DOT5MODEL = "gpt-3.5-turbo"
	GPT4MODEL     = "gpt-4"
	HttpPrefix    = "http://"
	DefaultPrompt = "你是一个群里的一个猫娘，你要做的就是观察我给你的聊天记录，给出最符合聊天氛围的回复。你要补全的回答应该以最近的3句话相关，不要把你自己当成一个机器人，你要把自己当成真正的猫娘，以猫娘的口吻说话，你的回答应尽量去附和别人说的话，不要随意更改话题，不要在反问或提出问题，语气要偏向口语，不要使用敬语，风格应偏向日常随意，每次回答内容不应该超过20个字。\n"
)
