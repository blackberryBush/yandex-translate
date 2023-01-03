package yandex_translate

import (
	"encoding/json"
	"errors"
	"io"
	"log"
	"net/http"
	"os/exec"
	"runtime"
	"strings"
	"time"
)

type YandexTranslator struct {
	folderID string        // your folderID in Yandex Cloud
	oauthKey string        // your OAuth token in Yandex.OAuth
	lastCall time.Time     // time of the last IAM token call
	apiKey   string        // current IAM token
	ttl      time.Duration // ttl - the maximum lifetime of the IAM token
}

// NewYandexTranslator : folderID - your folderID in Yandex Cloud, oauthKey - your OAuth token in Yandex.OAuth,
// ttl - the maximum lifetime of the IAM token, after which a new token will be requested when using the translator again
func NewYandexTranslator(folderID string, oauthKey string, ttl time.Duration) *YandexTranslator {
	t := YandexTranslator{
		folderID: folderID,
		oauthKey: oauthKey,
		lastCall: time.Now(),
		apiKey:   "",
		ttl:      ttl,
	}
	t.apiKey = t.getNewAPIKey()
	return &t
}

type inputText struct {
	TargetLanguageCode string `json:"targetLanguageCode"`
	Texts              string `json:"texts"`
	FolderID           string `json:"folderId"`
}

type outputTranslations struct {
	Translations []outputText `json:"translations"`
}
type outputText struct {
	Text     string `json:"text"`
	Language string `json:"detectedLanguageCode"`
}

func newInputText(targetLanguageCode string, texts string, folderID string) *inputText {
	return &inputText{TargetLanguageCode: targetLanguageCode, Texts: texts, FolderID: folderID}
}

func trace() string {
	pc := make([]uintptr, 10)
	runtime.Callers(2, pc)
	f := runtime.FuncForPC(pc[0])
	return f.Name()
}
func (tr *YandexTranslator) getNewAPIKey() string {
	cmd, err := exec.Command("powershell", "$yandexPassportOauthToken=\""+tr.oauthKey+"\"\n",
		"$Body=@{yandexPassportOauthToken=\"$yandexPassportOauthToken\"}|ConvertTo-Json", "-Compress\n",
		"Invoke-RestMethod", "-Method", "'POST'", "-Uri", "'https://iam.api.cloud.yandex.net/iam/v1/tokens'", "-Body", "$Body",
		"-ContentType", "'Application/json'|Select-Object", "-ExpandProperty", "iamToken").Output()
	if err != nil {
		log.Fatal(trace(), ": ", err)
	}
	return string(cmd[:len(cmd)-2])
}

func (tr *YandexTranslator) getAPIKey() string {
	if time.Since(tr.lastCall) > tr.ttl {
		tr.apiKey = tr.getNewAPIKey()
		tr.lastCall = time.Now()
	}
	return tr.apiKey
}

// TranslateByYandex : language - required language, text - source text
func (tr *YandexTranslator) TranslateByYandex(language string, text string) (string, error) {
	data := newInputText(language, text, tr.folderID)
	data1, _ := json.Marshal(data)
	client := &http.Client{}
	///
	r, _ := http.NewRequest(http.MethodPost, "https://translate.api.cloud.yandex.net/translate/v2/translate", strings.NewReader(string(data1)))
	r.Header.Add("Content-Type", "application/json")
	r.Header.Add("Authorization", "Bearer "+tr.getAPIKey())
	resp, err := client.Do(r)
	if err != nil {
		log.Println(trace(), ": ", err)
		return "", err
	}
	///
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			log.Println(trace(), ": ", err)
		}
	}(resp.Body)
	///
	translations := make([]byte, resp.ContentLength)
	n, err := resp.Body.Read(translations)
	if err != nil {
		log.Println(trace(), ": ", n, err)
		return "", err
	}
	translation := outputTranslations{}
	err = json.Unmarshal(translations, &translation)
	if err != nil {
		log.Println(trace(), ": ", n, err)
		return "", err
	}
	var result string
	for i, v := range translation.Translations {
		result += v.Text
		if i+1 < len(translation.Translations) {
			result += " "
		}
	}
	if result == "" {
		return "", errors.New("empty result")
	}
	return result, nil
}
