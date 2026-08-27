package sms

import (
	"context"
	"crypto/hmac"
	"crypto/sha1"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"sort"
	"strings"
	"time"

	"arlo-admin/internal/config"
)

type aliyunSender struct {
	cfg *config.AliyunSMSConfig
}

func newAliyunSender(cfg *config.AliyunSMSConfig) (*aliyunSender, error) {
	if cfg == nil {
		return nil, fmt.Errorf("阿里云短信配置为空")
	}
	if cfg.AccessKeyID == "" || cfg.AccessKeySecret == "" {
		return nil, fmt.Errorf("阿里云短信缺少 accessKeyId / accessKeySecret")
	}
	if cfg.SignName == "" || cfg.TemplateCode == "" {
		return nil, fmt.Errorf("阿里云短信缺少 signName / templateCode")
	}
	c := *cfg
	if c.TemplateParamKey == "" {
		c.TemplateParamKey = "code"
	}
	if c.RegionID == "" {
		c.RegionID = "cn-hangzhou"
	}
	return &aliyunSender{cfg: &c}, nil
}

func (s *aliyunSender) Name() string { return "aliyun" }

func (s *aliyunSender) Send(ctx context.Context, phone, code string) error {
	paramObj := map[string]string{s.cfg.TemplateParamKey: code}
	paramJSON, _ := json.Marshal(paramObj)

	params := map[string]string{
		"AccessKeyId":      s.cfg.AccessKeyID,
		"Action":           "SendSms",
		"Format":           "JSON",
		"PhoneNumbers":     phone,
		"RegionId":         s.cfg.RegionID,
		"SignName":         s.cfg.SignName,
		"SignatureMethod":  "HMAC-SHA1",
		"SignatureNonce":   fmt.Sprintf("%d", time.Now().UnixNano()),
		"SignatureVersion": "1.0",
		"TemplateCode":     s.cfg.TemplateCode,
		"TemplateParam":    string(paramJSON),
		"Timestamp":        time.Now().UTC().Format("2006-01-02T15:04:05Z"),
		"Version":          "2017-05-25",
	}

	keys := make([]string, 0, len(params))
	for k := range params {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	var canonical []string
	for _, k := range keys {
		canonical = append(canonical, percentEncode(k)+"="+percentEncode(params[k]))
	}
	canonicalQS := strings.Join(canonical, "&")
	stringToSign := "POST&" + percentEncode("/") + "&" + percentEncode(canonicalQS)

	mac := hmac.New(sha1.New, []byte(s.cfg.AccessKeySecret+"&"))
	mac.Write([]byte(stringToSign))
	signature := base64.StdEncoding.EncodeToString(mac.Sum(nil))
	params["Signature"] = signature

	form := url.Values{}
	for k, v := range params {
		form.Set(k, v)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, "https://dysmsapi.aliyuncs.com/", strings.NewReader(form.Encode()))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("阿里云短信请求失败: %w", err)
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(io.LimitReader(resp.Body, 1<<20))

	var out struct {
		Code    string `json:"Code"`
		Message string `json:"Message"`
		BizId   string `json:"BizId"`
	}
	if err := json.Unmarshal(body, &out); err != nil {
		return fmt.Errorf("阿里云短信响应解析失败: %s", string(body))
	}
	if !strings.EqualFold(out.Code, "OK") {
		return fmt.Errorf("阿里云短信发送失败: %s (%s)", out.Message, out.Code)
	}
	return nil
}

func percentEncode(s string) string {
	encoded := url.QueryEscape(s)
	encoded = strings.ReplaceAll(encoded, "+", "%20")
	encoded = strings.ReplaceAll(encoded, "*", "%2A")
	encoded = strings.ReplaceAll(encoded, "%7E", "~")
	return encoded
}
