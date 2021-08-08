package sms

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"github.com/gopherd/doge/query"
	"github.com/gopherd/doge/sms"
)

func init() {
	sms.Register("aliyun", open)
}

func open(source string) (sms.Provider, error) {
	var (
		options Options
		err     = parseSource(&options, source)
	)
	if err != nil {
		return nil, err
	}
	return NewClient(options), nil
}

type Options struct {
	Address  string `json:"address"`
	Key      string `json:"key"`
	TplId    string `json:"tpl_id"`
	TplValue string `json:"tpl_value"`
}

func (options Options) String() string {
	return options.Address + "?" + url.Values{
		"key":       {options.Key},
		"tpl_id":    {options.TplId},
		"tpl_value": {options.TplValue},
	}.Encode()
}

// parseSource parses options source string. Formats of source:
//
//	address?k1=v1&k2=v2&...&kn=vn
//
func parseSource(options *Options, source string) error {
	i := strings.IndexByte(source, '?')
	if i <= 0 {
		return errors.New("invalid source")
	}
	options.Address = source[:i]
	q, err := url.ParseQuery(source[i+1:])
	if err != nil {
		return err
	}
	return query.New(query.Query(q)).
		RequiredString(&options.Key, "key").
		RequiredString(&options.TplId, "tpl_id").
		RequiredString(&options.TplValue, "tpl_value").
		Err()
}

// Client implements sms.Provider
type Client struct {
	options Options
}

func NewClient(options Options) *Client {
	return &Client{
		options: options,
	}
}

func (c *Client) SendCode(phoneNumber string, code string) error {
	resp, err := http.PostForm(c.options.Address, url.Values{
		"apikey":    {c.options.Key},
		"mobile":    {phoneNumber},
		"tpl_id":    {c.options.TplId},
		"tpl_value": {fmt.Sprintf(c.options.TplValue, "%s", code)},
	})
	if err != nil {
		return err
	}
	if resp != nil {
		defer resp.Body.Close()
	}
	var result struct {
		HttpStatusCode int    `json:"http_status_code"`
		Code           int    `json:"code"`
		Msg            string `json:"msg"`
		Detail         string `json:"detail"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return err
	}
	if result.Code == 0 {
		return nil
	}
	return fmt.Errorf("(%d) %s", result.Code, result.Msg)
}
