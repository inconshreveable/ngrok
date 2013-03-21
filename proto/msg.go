package proto

import (
	"encoding/json"
	"reflect"
)

var TypeMap map[string]reflect.Type

const (
	Version = "0.1"
)

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
	TypeMap["MetricsMsg"] = t((*MetricsMsg)(nil))
	TypeMap["MetricsRespMsg"] = t((*MetricsRespMsg)(nil))
}

type Message interface {
	GetType() string
	SetType(string)
}

type Envelope struct {
	Version string
	Type    string
	Payload json.RawMessage
}

type TypeEmbed struct {
	Type string
}

type RegMsg struct {
	TypeEmbed
	Protocol         string
	Hostname         string
	Subdomain        string
	ClientId         string
	HttpAuthUser     string
	HttpAuthPassword string
	User             string
	Password         string
	OS               string
	Arch             string
}

type RegAckMsg struct {
	TypeEmbed
	Type      string
	Url       string
	ProxyAddr string
	Error     string
}

type RegProxyMsg struct {
	TypeEmbed
	Url string
}

type ReqProxyMsg struct {
	TypeEmbed
}

type PingMsg struct {
	TypeEmbed
}

type PongMsg struct {
	TypeEmbed
}

type VersionMsg struct {
	TypeEmbed
}

type VersionRespMsg struct {
	TypeEmbed
	Version string
}

type MetricsMsg struct {
	TypeEmbed
}

type MetricsRespMsg struct {
	TypeEmbed
	Metrics string
}

func (m *TypeEmbed) GetType() string {
	return m.Type
}

func (m *TypeEmbed) SetType(typ string) {
	m.Type = typ
}
