package config

type Config struct {
	OpenAI  OpenAIConfig  `json:"OpenAI"`
	Proxy   ProxyConfig   `json:"Proxy"`
	CQHttp  CQHttpConfig  `json:"CQHttp"`
	Service ServiceConfig `json:"Service"`
}

type OpenAIConfig struct {
	AuthorizationKey string
	ChatAPIPath      string
	Host             string
}

type ProxyConfig struct {
	Enable    bool
	ProxyHost string
}

type CQHttpConfig struct {
	Host string
	Port string
}

type ServiceConfig struct {
	Port string
}
