package hasura

import (
	"github.com/parnurzeal/gorequest"
	log "ngrok/log"
	"net/http"
	"errors"
	"encoding/json"
)

func SendQuery(m interface{}) (resp *http.Response, body []byte, err error) {
	var hError HasuraError
	request := gorequest.New()
	resp, body, errs := request.Post("https://data.beta.hasura.io/v1/query").
	Send(m).
	EndBytes()
	if errs != nil {
		err = errors.New(errs[0].Error())
		return
	}
	if resp.StatusCode != 200 {
		log.Warn("Status %d", resp.StatusCode)
		json.Unmarshal(body, &hError)
		log.Warn("Hasura Error %s", hError.Error)
		err = errors.New(hError.Error)
		return
	}
	return
}