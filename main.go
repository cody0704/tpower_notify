package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/joho/godotenv"
)

func sendTelegram(token, chatID, message string) error {
	apiURL := "https://api.telegram.org/bot" + token + "/sendMessage"
	resp, err := http.PostForm(apiURL, url.Values{
		"chat_id": {chatID},
		"text":    {message},
	})
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return fmt.Errorf("telegram API 回應錯誤: %d", resp.StatusCode)
	}

	return nil
}

// 將全形數字轉為半形
func toHalfWidthNumber(s string) string {
	replacer := strings.NewReplacer(
		"０", "0", "１", "1", "２", "2", "３", "3", "４", "4",
		"５", "5", "６", "6", "７", "7", "８", "8", "９", "9",
	)
	return replacer.Replace(s)
}

func main() {
	// load .env
	err := godotenv.Load()
	if err != nil {
		log.Fatal("載入 .env 檔案失敗:", err)
	}

	// 只保留爬蟲公告網站功能
	url := flag.String("url", "https://service.taipower.com.tw/branch/d117/xcnotice?xsmsid=0M242581310300276906", "公告網址")
	keywords := flag.String("keywords", os.Getenv("ADDRESS"), "地點關鍵字，多個以逗號分隔")
	flag.Parse()

	fmt.Println("公告網址:", *url)
	fmt.Println("地點關鍵字:", *keywords)

	// 讀取 Telegram Token 與 Chat ID
	telegramToken := os.Getenv("TELEGRAM_TOKEN")
	chatID := os.Getenv("TELEGRAM_CHAT_ID")
	if telegramToken == "" || chatID == "" {
		log.Fatal("請設定 TELEGRAM_TOKEN 與 TELEGRAM_CHAT_ID 環境變數")
	}

	var notifyMsgs []string
	keywordList := strings.Split(*keywords, ",")
	if err := crawlAndFindNotify(*url, keywordList, func(msg string) {
		notifyMsgs = append(notifyMsgs, msg)
	}); err != nil {
		log.Fatalf("爬取公告失敗: %v", err)
	}

	// 推送所有命中通知
	for _, msg := range notifyMsgs {
		err := sendTelegram(telegramToken, chatID, msg)
		if err != nil {
			log.Printf("推送 Telegram 失敗: %v", err)
		} else {
			fmt.Println("已推送 Telegram 通知:", msg)
		}
	}
}

// crawlAndFindNotify 會在每筆命中時呼叫 callback
func crawlAndFindNotify(url string, keywords []string, callback func(msg string)) error {
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		return err
	}

	found := false
	now := time.Now()

	// 只通知今天到一週內的停電公告
	weekLater := now.AddDate(0, 0, 7)
	doc.Find("table").Each(func(i int, table *goquery.Selection) {
		dateText := table.Find("caption").Text()
		// 嘗試解析日期，支援民國年格式（如 114 年 5 月 2 日）
		dateTextTrim := strings.TrimSpace(dateText)
		var date time.Time
		var dateErr error
		if dateTextTrim != "" {
			// 民國年格式：114 年 5 月 2 日
			if strings.Contains(dateTextTrim, "年") {
				// 解析民國年
				var y, m, d int
				re := `工作停電日期 \( 非限電 \)：(\d+) 年 (\d+) 月 (\d+) 日`
				matches := regexp.MustCompile(re).FindStringSubmatch(dateTextTrim)
				if len(matches) == 4 {
					y, _ = strconv.Atoi(matches[1])
					m, _ = strconv.Atoi(matches[2])
					d, _ = strconv.Atoi(matches[3])

					date = time.Date(y+1911, time.Month(m), d, 0, 0, 0, 0, time.Local)
				} else {
					dateErr = fmt.Errorf("無法解析日期: %s", dateTextTrim)
				}
			} else {
				// 嘗試西元格式
				date, dateErr = time.Parse("2006/01/02", dateTextTrim)
				if dateErr != nil {
					date, dateErr = time.Parse("2006-01-02", dateTextTrim)
				}
			}
		}
		if dateErr != nil || dateTextTrim == "" || date.IsZero() {
			// 無法解析日期，跳過
			fmt.Println("日期解析錯誤:", dateTextTrim, dateErr)
			return
		}
		if date.Before(now) || date.After(weekLater) {
			// 只通知今天到一週內的
			fmt.Println("日期不在範圍內:", dateTextTrim)
			return
		}
		table.Find("tr").Each(func(j int, tr *goquery.Selection) {
			tds := tr.Find("td")
			if tds.Length() < 2 {
				return
			}
			timeText := tds.Eq(0).Text()
			areaText := tds.Eq(1).Text()
			areaTextHalf := toHalfWidthNumber(areaText)
			for _, kw := range keywords {
				kw = strings.TrimSpace(kw)
				kwHalf := toHalfWidthNumber(kw)
				if kwHalf != "" && strings.Contains(areaTextHalf, kwHalf) {
					msg := fmt.Sprintf("[停電預告] %s\n%s\n停電時段：%s\n停電地區：%s", kw, strings.TrimSpace(dateText), strings.TrimSpace(timeText), strings.TrimSpace(areaText))
					callback(msg)
					found = true
				}
			}
		})
	})
	if !found {
		fmt.Println("查無指定地點停電公告")
	}
	return nil
}
