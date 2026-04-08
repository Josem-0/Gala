package lastfm

import (
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"gala/config"
	"net/http"
	"net/url"
	"sort"
	"strings"
)

type SessionResponse struct {
	Session struct {
		Name string `json:"name"`
		Key  string `json:"key"`
	} `json:"session"`
}

func PostLastfm(params map[string]string) (map[string]interface{}, error) {
	c, _ := config.LoadConfig()

	params["api_key"] = c.LastFmApiKey
	params["api_sig"] = generateSignature(params, c.LastFmSecret)
	params["format"] = "json"

	data := url.Values{}
	for key, val := range params {
		data.Set(key, val)
	}

	req, err := http.NewRequest("POST", "https://ws.audioscrobbler.com/2.0/", strings.NewReader(data.Encode()))
	if err != nil {
		fmt.Println("Last.fm post error:", err)
		return nil, err
	}

	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")

	client := &http.Client{}
	res, err := client.Do(req)
	if err != nil {
		fmt.Println("Last.fm post error:", err)
		return nil, err
	}

	defer res.Body.Close()

	var result map[string]interface{}
	if err := json.NewDecoder(res.Body).Decode(&result); err != nil {
		fmt.Println("Last.fm POST error parsing JSON:", err)
		return nil, err
	}

	return result, nil

}

func generateSignature(params map[string]string, secret string) string {
	keys := make([]string, 0, len(params))
	for k := range params {
		if k != "format" && k != "callback" {
			keys = append(keys, k)
		}
	}
	sort.Strings(keys)

	var sb strings.Builder
	for _, k := range keys {
		sb.WriteString(k)
		sb.WriteString(params[k])
	}

	sb.WriteString(secret)

	hash := md5.Sum([]byte(sb.String()))
	return hex.EncodeToString(hash[:])
}

func FetchSessionKey(token, apiKey, secret string) (string, string, error) {
	params := map[string]string{
		"api_key": apiKey,
		"method":  "auth.getSession",
		"token":   token,
	}

	sig := generateSignature(params, secret)

	u, _ := url.Parse("https://ws.audioscrobbler.com/2.0/")
	query := u.Query()

	for k, v := range params {
		query.Set(k, v)
	}

	query.Set("api_sig", sig)
	query.Set("format", "json")
	u.RawQuery = query.Encode()

	resp, err := http.Get(u.String())
	if err != nil {
		return "", "", err
	}

	defer resp.Body.Close()

	var result SessionResponse
	json.NewDecoder(resp.Body).Decode(&result)

	return result.Session.Key, result.Session.Name, nil
}
