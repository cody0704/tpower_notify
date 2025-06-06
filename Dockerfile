# 使用官方 Golang 映像作為基礎
FROM golang:1.24.3

# 設定工作目錄
WORKDIR /app

# 複製 go.mod 和 go.sum 並下載依賴
COPY go.mod go.sum ./
RUN go mod download

# 複製其餘的應用程式程式碼
COPY . .

# 編譯應用程式
RUN go build -o main .

# 設定容器啟動時執行的命令
CMD ["./main"]