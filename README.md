## TPower Notify

這是一個用來通知台灣電力停電通知的幫手，透過提前一個禮拜將停電資訊通知到 Telegram，幫助你提前做好準備。

## 限制

- 目前只支援一個服務區域，不支援多個服務區域的關鍵字通知
- 目前只通知一週內的停電通知

## 使用方法

將 .env.example 複製成 .env 並填入相關資訊

```
TELEGRAM_TOKEN = "{填入你的 Telegram Bot Token}"
TELEGRAM_CHAT_ID = "{填入你的 Telegram Chat ID}"

TAIPOWER_URL= ""
# 你想通知的停電區域關建立，透過 `,` 分隔
ADDRESS = "新北市鶯歌區中山路,新北市新莊區新莊路"
```

TAIPOWER_URL 需要自行去確認所在區域的停電資訊網址，請參考 https://service.taipower.com.tw/branch

選擇對應的區域之後，從 「常用服務計」 -> 「計畫性工作停電公告」，並複製 計畫性工作停電公告 的 網址

- 範例，這是苗栗區的 https://service.taipower.com.tw/branch/d122/xcnotice?xsmsid=0M242581315483543322
- 範例，這是台北西區的 https://service.taipower.com.tw/branch/d117/xcnotice?xsmsid=0M242581310300276906
