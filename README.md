# --GPTBOT
LowOfCycles-GPTBOT

一个圆群的ZeroBot框架机器人。使用方式请参照 https://github.com/FloatTech/ZeroBot-Plugin

该机器人增加了一些功能

- 随机圆图功能，从图库中随机获取一些魔法少女小圆的图片。使用cos存储图库
- 修改了抽老婆功能
  - 抽老婆将原数据源中大量不认识的二游角色删除，改为世萌入围角色名单。
  - 允许群员添加，删除，查看老婆列表
  - 抽老婆卡池分为两个卡池，一个base卡池供全部群使用，里面是世萌入围角色名单。另一个是群私有卡池，群员添加和删除老婆均为操作群私有卡池，不会对其他群造成影响。
  - 添加抽老公功能，功能和抽老婆相同，只是卡池换成了男性角色
- chatgpt机器人，一个群chatgpt机器人。
  - 直接at机器人就会触发gpt聊天，gpt会有默认最近10句对话的记忆。
  - 可设置机器人参与群聊，群友发送文字信息会有概率触发机器人来附和你的对话

### How to start

- 准备一个OPEN AI账号，并取得你的API Key 
- 准备一个腾讯云COS账号作为数据存储用
- 将config/config.ini.template 去掉.template后缀名
- 填写里面需要的open ai以及cos的各项信息
- 运行该程序即可

