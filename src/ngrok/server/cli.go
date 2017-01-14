package server

import (
	"flag"
)

type Options struct {
	httpAddr   string
	httpsAddr  string
	tunnelAddr string
	domain     string
	tlsCrt     string
	tlsKey     string
	logto      string
	loglevel   string
}

func parseArgs() *Options {
	httpAddr := flag.String("httpAddr", ":8081", "监听的HTTP端口,设置空禁用")
	httpsAddr := flag.String("httpsAddr", "", "监听的HTTPS端口,设置空禁用")
	tunnelAddr := flag.String("tunnelAddr", ":8089", "客户端监听的端口")
	domain := flag.String("domain", "ngrok.lxwgo.com", "隧道的域名")
	tlsCrt := flag.String("tlsCrt", "", "TLS证书文件")
	tlsKey := flag.String("tlsKey", "", "TLS秘钥文件")
	logto := flag.String("log", "stdout", "输出日志文件. 'none'为不输出任何内容")
	loglevel := flag.String("log-level", "INFO", "日志文件级别. 可选值DEBUG, INFO, WARNING, ERROR")
	flag.Parse()

	return &Options{
		httpAddr:   *httpAddr,
		httpsAddr:  *httpsAddr,
		tunnelAddr: *tunnelAddr,
		domain:     *domain,
		tlsCrt:     *tlsCrt,
		tlsKey:     *tlsKey,
		logto:      *logto,
		loglevel:   *loglevel,
	}
}
