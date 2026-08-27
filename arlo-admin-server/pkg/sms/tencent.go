package sms

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"arlo-admin/internal/config"
)

type tencentSender struct {
	cfg *config.TencentSMSConfig
}

func newTencentSender(cfg *config.TencentSMSConfig) (*tencentSender, error) {
	if cfg == nil {
		return nil, fmt.Errorf("腾讯云短信配置为空")
	}
	if cfg.SecretID == "" || cfg.SecretKey == "" {
		return nil, fmt.Errorf("腾讯云短信缺少 secretId / secretKey")
	}
	if cfg.AppID == "" || cfg.SignName == "" || cfg.TemplateID == "" {
		return nil, fmt.Errorf("腾讯云短信缺少 appId / signName / templateId")
	}
	c := *cfg
	if c.Region == "" {
		c.Region = "ap-guangzhou"
	}
	return &tencentSender{cfg: &c}, nil
}

func (s *tencentSender) Name() string { return "tencent" }

func (s *tencentSender) Send(ctx context.Context, phone, code string) error {
	phoneNumber := phone
	if !strings.HasPrefix(phoneNumber, "+") {
		phoneNumber = "+86" + phoneNumber
	}

	payload := map[string]interface{}{
		"PhoneNumberSet":   []string{phoneNumber},
		"SmsSdkAppId":      s.cfg.AppID,
		"SignName":         s.cfg.SignName,
		"TemplateId":       s.cfg.TemplateID,
		"TemplateParamSet": []string{code},
	}
	body, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	host := "sms.tencentcloudapi.com"
	service := "sms"
	action := "SendSms"
	version := "2021-01-11"
	timestamp := time.Now().Unix()
	date := time.Unix(timestamp, 0).UTC().Format("2006-01-02")

	hashedPayload := sha256Hex(body)
	canonicalHeaders := fmt.Sprintf("content-type:application/json; charset=utf-8\nhost:%s\nx-tc-action:%s\n",
		host, strings.ToLower(action))
	signedHeaders := "content-type;host;x-tc-action"
	canonicalRequest := strings.Join([]string{
		"POST",
		"/",
		"",
		canonicalHeaders,
		signedHeaders,
		hashedPayload,
	}, "\n")

	credentialScope := fmt.Sprintf("%s/%s/tc3_request", date, service)
	stringToSign := strings.Join([]string{
		"TC3-HMAC-SHA256",
		fmt.Sprintf("%d", timestamp),
		credentialScope,
		sha256Hex([]byte(canonicalRequest)),
	}, "\n")

	secretDate := hmacSHA256([]byte("TC3"+s.cfg.SecretKey), date)
	secretService := hmacSHA256(secretDate, service)
	secretSigning := hmacSHA256(secretService, "tc3_request")
	signature := hex.EncodeToString(hmacSHA256(secretSigning, stringToSign))

	authorization := fmt.Sprintf("TC3-HMAC-SHA256 Credential=%s/%s, SignedHeaders=%s, Signature=%s",
		s.cfg.SecretID, credentialScope, signedHeaders, signature)

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, "https://"+host, bytes.NewReader(body))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json; charset=utf-8")
	req.Header.Set("Host", host)
	req.Header.Set("X-TC-Action", action)
	req.Header.Set("X-TC-Version", version)
	req.Header.Set("X-TC-Timestamp", fmt.Sprintf("%d", timestamp))
	req.Header.Set("X-TC-Region", s.cfg.Region)
	req.Header.Set("Authorization", authorization)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("腾讯云短信请求失败: %w", err)
	}
	defer resp.Body.Close()
	respBody, _ := io.ReadAll(io.LimitReader(resp.Body, 1<<20))

	var out struct {
		Response struct {
			Error *struct {
				Code    string `json:"Code"`
				Message string `json:"Message"`
			} `json:"Error"`
			SendStatusSet []struct {
				Code    string `json:"Code"`
				Message string `json:"Message"`
			} `json:"SendStatusSet"`
			RequestId string `json:"RequestId"`
		} `json:"Response"`
	}
	if err := json.Unmarshal(respBody, &out); err != nil {
		return fmt.Errorf("腾讯云短信响应解析失败: %s", string(respBody))
	}
	if out.Response.Error != nil {
		return fmt.Errorf("腾讯云短信发送失败: %s (%s)", out.Response.Error.Message, out.Response.Error.Code)
	}
	if len(out.Response.SendStatusSet) == 0 {
		return fmt.Errorf("腾讯云短信无发送结果")
	}
	st := out.Response.SendStatusSet[0]
	if !strings.EqualFold(st.Code, "Ok") {
		return fmt.Errorf("腾讯云短信发送失败: %s (%s)", st.Message, st.Code)
	}
	return nil
}

func sha256Hex(b []byte) string {
	sum := sha256.Sum256(b)
	return hex.EncodeToString(sum[:])
}

func hmacSHA256(key []byte, data string) []byte {
	mac := hmac.New(sha256.New, key)
	mac.Write([]byte(data))
	return mac.Sum(nil)
}
