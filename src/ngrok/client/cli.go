package client

import (
	"flag"
	"fmt"
	"ngrok/version"
	"os"
)

const usage1 string = `使用: %s [subdomain] <port>
选项:
	subdomain	你需要映射的子域名, 例如: test
	port		你需要映射的本地端口或者地址,例如: 8080
`

const usage2 string = `
示例:
	ngrok test 8080
	映射一个本地 8080 端口到 test.ngrok.lxwgo.com

	ngrok demo 127.0.0.1:8080
	映射一个本地 8080 端口到 demo.ngrok.lxwgo.com


其他使用: %s <command>
选项:
	help	打印帮助信息
	version	打印当前版本

示例:
	ngrok help
	ngrok version

`

type Options struct {
	config    string
	logto     string
	loglevel  string
	authtoken string
	httpauth  string
	hostname  string
	protocol  string
	subdomain string
	port      string
	command   string
	args      []string
}

func ParseArgs() (opts *Options, err error) {
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, usage1, os.Args[0])
		flag.PrintDefaults()
		fmt.Fprintf(os.Stderr, usage2, os.Args[0])
	}

	flag.Parse()

	opts = &Options{
		config:    "",
		logto:     "none",
		loglevel:  "INFO",
		httpauth:  "",
		subdomain: "",
		protocol:  "http",
		authtoken: "",
		hostname:  "",
		port:      "",
		command:   flag.Arg(0),
	}

	switch opts.command {
	case "version":
		fmt.Println(version.MajorMinor())
		os.Exit(0)
	case "help", "":
		flag.Usage()
		os.Exit(0)

	default:
		if len(flag.Args()) > 1 && len(flag.Args()) < 3 {
			opts.subdomain = flag.Arg(0)
			opts.port = flag.Arg(1)
			opts.command = "default"
			return
		}

		err = fmt.Errorf("错误的命令, 请使用 %s help 查看帮助", os.Args[0])
	}

	return
}
