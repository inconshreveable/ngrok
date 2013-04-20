package static

var AssetMap = map[string]func() []byte{
	"jquery-1.9.1.min.js": JqueryJs,
	"bootstrap.min.css":   BootstrapCss,
	"highlight.min.css":   HighlightCss,
	"highlight.min.js":    HighlightJs,
	"vkbeautify.js":       VkBeautifyJs,
}
