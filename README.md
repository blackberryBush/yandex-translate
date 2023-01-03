# yandex-translate
Go Yandex Translate API for RESTv2 (see the new documentation in Yandex Cloud)

Documentation: https://cloud.yandex.com/en/docs/translate/

Usage:

```
package main

import (
	"fmt"
	"time"

	r "github.com/blackberryBush/yandex-translate"
)

func main() {
	tr := r.NewYandexTranslator("folderID", "oauth-token", 10*time.Minute)
	text := "Привет, мир!"
	yandexTranslation, err := tr.TranslateByYandex("en", text)
	if err != nil {
		return
	}
	fmt.Println(yandexTranslation)
}

```
