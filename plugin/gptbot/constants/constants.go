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
	DefaultPrompt = "你是一只可爱的猫娘。请把你的发言缩短到20个汉字以内。你的发言同时要符合猫娘的身份，要用尽可能可爱的语气来回答，并配上猫娘可爱的动作以及颜文字。你的发言应尽可能的和与你最近的一次对话有关"

	HelpContent = "" +
		"\n/查看提示词" +
		"\n/设置提示词 (默认是猫娘)" +
		"\n/设置gpt模型 可选模型：(gpt-3.5-turbo gpt-4)" +
		"\n/查看gpt模型" +
		"\n/查看记忆区 查看当前gpt记忆的聊天信息" +
		"\n/清空记忆区" +
		"\n/查看群回复概率 查看群聊回复频率" +
		"\n/设置群回复概率  设置群聊回复频率，默认0，设置为非0数值后，机器人会对没有at机器人的聊天做出回应。0为不回复，100为回复每一条聊天"
)
