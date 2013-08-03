package msg

import (
	"encoding/json"
	"reflect"
)

var TypeMap map[string]reflect.Type

func init() {
	TypeMap = make(map[string]reflect.Type)

	t := func(obj interface{}) reflect.Type { return reflect.TypeOf(obj).Elem() }
	TypeMap["RegMsg"] = t((*RegMsg)(nil))
	TypeMap["RegAckMsg"] = t((*RegAckMsg)(nil))
	TypeMap["RegProxyMsg"] = t((*RegProxyMsg)(nil))
	TypeMap["ReqProxyMsg"] = t((*ReqProxyMsg)(nil))
	TypeMap["PingMsg"] = t((*PingMsg)(nil))
	TypeMap["PongMsg"] = t((*PongMsg)(nil))
	TypeMap["VerisonMsg"] = t((*VersionMsg)(nil))
	TypeMap["VersionRespMsg"] = t((*VersionRespMsg)(nil))
}

type Message interface{}

type Envelope struct {
	Type    string
	Payload json.RawMessage
}

type RegMsg struct {
	Version   string
	MmVersion string
	Protocol  string
	Hostname  string
	Subdomain string
	ClientId  string
	HttpAuth  string
	User      string
	Password  string
	OS        string
	Arch      string
}

type RegAckMsg struct {
	Version   string
	MmVersion string
	Url       string
	ProxyAddr string
	Error     string
}

type RegProxyMsg struct {
	Url string
}

type ReqProxyMsg struct {
}

type PingMsg struct {
}

type PongMsg struct {
}

type VersionMsg struct {
}

type VersionRespMsg struct {
	Version   string
	MmVersion string
}
