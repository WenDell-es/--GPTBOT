# --GPTBOT
LowOfCycles-GPTBOT

一个简单的Http server 调用官方的gpt接口

## main分支为单独的gpt机器人
### How to start

- 准备一个OPEN AI账号，并取得你的API Key 
- 准备能够连接国外网络的网络环境.若使用代理服务器，请按下面配置Proxy
- 准备一个QQ号，并部署go-CQHttp。 参考 [go-CQHttp](https://github.com/Mrs4s/go-cqhttp)
- 准备配置文件，模板为configuration/conf.template，将后缀名改为.ini。并依次修改以下内容
  - 将[OpenAi] AuthorizationKey替换为你的open ai api key
  - [Proxy]一栏中，若你不需要代理，则将enable设置为false，否则请将Host设置为你的代理服务器地址(若需要自行搭建代理服务器，可直接编译proxy/main.go。这是一个简单的Http代理)
  - [CQHttp]一栏，填写你部署go-CQHttp的IP和监听端口。
  - [Service]本服务的监听端口，请确保与go-CQHttp中的通信地址一致。默认5701
- 运行该程序即可

## zeroBot分支为使用zeroBot框架将gpt bot作为一个子插件
使用方式请参照 https://github.com/FloatTech/ZeroBot-Plugin
data/gptbot中存在配置文件模板，将其名改为conf.ini后填写你的api key即可
