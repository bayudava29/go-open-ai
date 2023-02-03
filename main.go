package main

import (
	"context"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"strconv"
	"time"

	"github.com/bayudava29/go-open-ai/config"
	"github.com/gin-gonic/gin"
	gogpt "github.com/sashabaranov/go-gpt3"
	"github.com/spf13/viper"
)

type Update struct {
	Id  int64   `json:"update_id"`
	Msg Message `json:"message"`
}

type Message struct {
	Id   int64  `json:"message_id"`
	Text string `json:"text"`
	Chat Chat   `json:"chat"`
	User User   `json:"from"`
}

type Chat struct {
	Id    int    `json:"id"`
	Title string `json:"title"`
}

type User struct {
	Id        int64  `json:"id"`
	FirstName string `json:"first_name"`
	Username  string `json:"username"`
}

type SendMessageRequest struct {
	ChatId string `json:"chat_id"`
	Text   string `json:"text"`
}

func main() {
	config.InitConfig()

	router := gin.Default()
	router.POST("/telegram/webhook", TelegramWebhook)
	router.GET("/ping", Ping)

	port := viper.GetString("PORT")
	srv := &http.Server{
		Addr:    ":" + port,
		Handler: router,
	}
	log.Printf("Server Initialized, listening at port: %s", port)

	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("listen: %s\n", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt)
	<-quit
	log.Print("Shutdown Server ...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		log.Fatal("Server Shutdown:", err)
	}
	log.Print("Server exiting")
}

func TelegramWebhook(c *gin.Context) {
	var request Update
	errBind := c.ShouldBindJSON(&request)
	if errBind != nil {
		log.Print(errBind.Error())
	}

	resGpt := GptRequest(request.Msg.Text)

	SendMessage(request.Msg.Chat.Id, resGpt)
}

func GptRequest(prompt string) string {
	c := gogpt.NewClient(viper.GetString("GPT_TOKEN"))
	ctx := context.Background()

	req := gogpt.CompletionRequest{
		Model:     gogpt.GPT3TextDavinci003,
		MaxTokens: 150,
		Prompt:    prompt,
	}
	resp, err := c.CreateCompletion(ctx, req)
	if err != nil {
		return "Ga ngerti bro, coba tanya bayu xixi"
	}
	return resp.Choices[0].Text
}

func SendMessage(chatId int, gpt string) {
	endpoint := fmt.Sprintf("%s/%s", viper.GetString("TELEGRAM_API"), "sendMessage")
	res, err := http.PostForm(
		endpoint,
		url.Values{
			"chat_id": {strconv.Itoa(chatId)},
			"text":    {gpt},
		},
	)
	if err != nil {
		log.Print(err.Error())
	}
	defer res.Body.Close()

	var bodyBytes, errRead = ioutil.ReadAll(res.Body)
	if errRead != nil {
		log.Printf("error in parsing telegram answer %s", errRead.Error())
	} else {
		bodyString := string(bodyBytes)
		log.Printf("Body of Telegram Response: %s", bodyString)
	}
}

func Ping(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"message": "pong",
	})
}
