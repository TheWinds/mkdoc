package insomnia

import (
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"github.com/thewinds/mkdoc"
	"github.com/thewinds/mkdoc/generators/objmock"
	"math/rand"
	"strings"
	"time"
)

type Generator struct{}

func init() {
	mkdoc.RegisterGenerator(&Generator{})
}

func (g *Generator) Gen(ctx *mkdoc.DocGenContext) (output []byte, err error) {
	data := &insomniaExport{
		Type:   "export",
		Format: 4,
		Date:   time.Now().Format(time.RFC3339),
		Source: "mkdoc",
	}

	wrk := &workspace{
		ID:          genResID("wrk"),
		Created:     time.Now().Unix(),
		Description: ctx.Config.Name,
		Modified:    time.Now().Unix(),
		Name:        fmt.Sprintf("%s-%s", ctx.Tag, ctx.Config.Name),
		Type:        "workspace",
	}

	envs := &environment{
		ID:          genResID("env"),
		Created:     time.Now().Unix(),
		Data:        map[string]string{"base_url": ctx.Config.APIBaseURL},
		MetaSortKey: time.Now().Unix(),
		Modified:    time.Now().Unix(),
		Name:        "env",
		ParentID:    wrk.ID,
		Type:        "environment",
	}

	data.Resources = append(data.Resources, wrk, envs)

	var commonHeaders []*requestHeader

	for _, header := range ctx.Config.CommonHeader {
		commonHeaders = append(commonHeaders, &requestHeader{
			ID:    genResID("pair"),
			Name:  header.Name,
			Value: header.Default,
		})
	}

	for _, api := range ctx.APIs {
		now := time.Now().Unix()
		req := &request{
			ID:                              genResID("req"),
			Authentication:                  reqAuth{},
			Body:                            nil,
			Created:                         now,
			Description:                     api.Desc,
			Headers:                         make([]*requestHeader, 0, len(commonHeaders)),
			MetaSortKey:                     -now,
			Method:                          strings.ToUpper(api.Method),
			Modified:                        now,
			Name:                            api.Name,
			Parameters:                      []interface{}{},
			ParentID:                        wrk.ID,
			SettingDisableRenderRequestBody: true,
			SettingEncodeURL:                true,
			SettingFollowRedirects:          "global",
			SettingRebuildPath:              true,
			SettingSendCookies:              true,
			SettingStoreCookies:             true,
			URL:                             fmt.Sprintf("{{base_url}}%s", api.Path),
			Type:                            "request",
		}

		req.Headers = append(req.Headers, commonHeaders...)
		for k := range api.Header {
			req.Headers = append(req.Headers, &requestHeader{
				ID:    genResID("pair"),
				Name:  k,
				Value: "",
			})
		}
		switch api.InArgEncoder {
		case "json":
			body := &textReqBody{
				MimeType: "application/json",
			}
			body.Text, err = objmock.NewJSONMocker().MockPretty(api.InArgument, ctx.RefObj)
			if err != nil {
				return nil, err
			}
			req.Headers = append(req.Headers, &requestHeader{
				ID:    genResID("pair"),
				Name:  "Content-Type",
				Value: "application/json",
			})
			req.Body = body
		default:
			body := &structuredReqBody{
				MimeType: "multipart/form-data",
			}
			for _, field := range api.InArgument.Fields {
				if field.Tag.GetValue("json") == "-" {
					continue
				}
				param := &reqParam{
					Description: field.Desc,
					ID:          genResID("pair"),
					Name:        field.Tag.GetFirstValue("json", ","),
					Value:       "",
				}

				body.Params = append(body.Params, param)
			}

			req.Headers = append(req.Headers, &requestHeader{
				ID:    genResID("pair"),
				Name:  "Content-Type",
				Value: "multipart/form-data",
			})
			req.Body = body
		}
		data.Resources = append(data.Resources, req)

	}

	return json.Marshal(data)
}

func (g *Generator) Name() string {
	return "insomnia"
}

func (g *Generator) FileExt() string {
	return "json"
}

type insomniaExport struct {
	Type      string        `json:"_type"`
	Format    int           `json:"__export_format"`
	Date      string        `json:"__export_date"`
	Source    string        `json:"__export_source"`
	Resources []interface{} `json:"resources"`
}

type request struct {
	ID                              string           `json:"_id"`
	Authentication                  reqAuth          `json:"authentication"`
	Body                            interface{}      `json:"body"`
	Created                         int64            `json:"created"`
	Description                     string           `json:"description"`
	Headers                         []*requestHeader `json:"headers"`
	IsPrivate                       bool             `json:"isPrivate"`
	MetaSortKey                     int64            `json:"metaSortKey"`
	Method                          string           `json:"method"`
	Modified                        int64            `json:"modified"`
	Name                            string           `json:"name"`
	Parameters                      []interface{}    `json:"parameters"`
	ParentID                        interface{}      `json:"parentId"`
	SettingDisableRenderRequestBody bool             `json:"settingDisableRenderRequestBody"`
	SettingEncodeURL                bool             `json:"settingEncodeUrl"`
	SettingFollowRedirects          string           `json:"settingFollowRedirects"`
	SettingRebuildPath              bool             `json:"settingRebuildPath"`
	SettingSendCookies              bool             `json:"settingSendCookies"`
	SettingStoreCookies             bool             `json:"settingStoreCookies"`
	URL                             string           `json:"url"`
	Type                            string           `json:"_type"`
}
type reqAuth struct{}

type reqParam struct {
	Description string `json:"description"`
	ID          string `json:"id"`
	Name        string `json:"name"`
	Value       string `json:"value"`
}

type structuredReqBody struct {
	MimeType string      `json:"mimeType"`
	Params   []*reqParam `json:"params"`
}

type textReqBody struct {
	MimeType string `json:"mimeType"`
	Text     string `json:"text"`
}

type requestHeader struct {
	ID    string `json:"id"`
	Name  string `json:"name"`
	Value string `json:"value"`
}

type workspace struct {
	ID          string      `json:"_id"`
	Created     int64       `json:"created"`
	Description string      `json:"description"`
	Modified    int64       `json:"modified"`
	Name        string      `json:"name"`
	ParentID    interface{} `json:"parentId"`
	Type        string      `json:"_type"`
}

type environment struct {
	ID                string            `json:"_id"`
	Color             interface{}       `json:"color"`
	Created           int64             `json:"created"`
	Data              map[string]string `json:"data"`
	DataPropertyOrder interface{}       `json:"dataPropertyOrder"`
	IsPrivate         bool              `json:"isPrivate"`
	MetaSortKey       int64             `json:"metaSortKey"`
	Modified          int64             `json:"modified"`
	Name              string            `json:"name"`
	ParentID          interface{}       `json:"parentId"`
	Type              string            `json:"_type"`
}

type cookieJar struct {
	ID       string        `json:"_id"`
	Cookies  []interface{} `json:"cookies"`
	Created  int64         `json:"created"`
	Modified int64         `json:"modified"`
	Name     string        `json:"name"`
	ParentID interface{}   `json:"parentId"`
	Type     string        `json:"_type"`
}

func init() {
	rand.Seed(time.Now().Unix())
}

func genResID(typ string) string {
	k := fmt.Sprintf("%s%d%d", typ, time.Now().UnixNano(), rand.Int31())
	sum := md5.Sum([]byte(k))
	return fmt.Sprintf("typ_%s", hex.EncodeToString(sum[:]))
}