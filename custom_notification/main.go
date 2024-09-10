package custom_notification

import (
	"context"
	models2 "github.com/iikmaulana/mini-engine/custom_notification/models"
	"github.com/iikmaulana/mini-engine/custom_promo/lib"
	"github.com/robfig/cron/v3"
	"os"
	"os/signal"
	"syscall"
	"time"

	//"encoding/json"
	firebase "firebase.google.com/go/v4"
	"firebase.google.com/go/v4/messaging"
	"fmt"
	"github.com/iikmaulana/mini-engine/custom_notification/engine"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
	"github.com/uzzeet/uzzeet-gateway/libs/utils/uttime"
	"google.golang.org/api/option"
	"log"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal(err)
	}

	tmpCront()
	jakartaTime, _ := time.LoadLocation("Asia/Jakarta")
	scheduler := cron.New(cron.WithLocation(jakartaTime))
	defer scheduler.Stop()
	scheduler.AddFunc("59 23 * * *", tmpCront)
	go scheduler.Start()
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)
	<-sig
}

func tmpCront() {
	fmt.Println(fmt.Sprintf("Date : %s", uttime.Now().Format("2006-01-02")))

	tmpTime := map[string]string{}
	tmpData, _ := engine.GetListCustomNotification()
	for _, v := range tmpData {
		tmpTime[v.Id] = fmt.Sprintf("%s %s", v.PengirimanBerikutnya, v.TimeCronjob)
		fmt.Println(fmt.Sprintf("%s ===> %s", v.Id, fmt.Sprintf("%s %s", v.PengirimanBerikutnya, v.TimeCronjob)))
	}

	jakartaTime, _ := time.LoadLocation("Asia/Jakarta")
	scheduler := cron.New(cron.WithLocation(jakartaTime))

	for k, v := range tmpTime {
		tmpK := k
		tmpV := v
		tmpTimex, _ := uttime.ParseFromString(tmpV)
		tmpCrontab := lib.ToCrontab(tmpTimex)
		scheduler.AddFunc(tmpCrontab, func() {
			runCront(tmpK)
		})
	}
	scheduler.Run()
	//runCront("63f65c03-1d60-46bd-816c-5468c9d94d79")
}

func runCront(tmpId string) {
	tmpCustomNotif, _ := engine.GetCustomNotification(tmpId)

	tmpType := ""
	tmpTitle := ""
	tmpText := ""
	switch tmpCustomNotif.TypeNotification {
	case "promosi":
		tmpTitle = "Promosi"
		tmpType = "promo_myfuso"
		tmpText = fmt.Sprintf(`{"id":"%s","text":"%s","cover_image":"%s","description":"%s","link_web":"%s"}`, tmpCustomNotif.Id, tmpCustomNotif.Title, tmpCustomNotif.LinkImageWeb, tmpCustomNotif.Description, tmpCustomNotif.LinkWeb)
	case "aktivitas":
		tmpTitle = "Aktivitas"
		tmpType = "activity_myfuso"
		tmpText = fmt.Sprintf(`{"id":"%s","text":"%s","cover_image":"%s","description":"%s","link_web":"%s"}`, tmpCustomNotif.Id, tmpCustomNotif.Title, tmpCustomNotif.LinkImageWeb, tmpCustomNotif.Description, tmpCustomNotif.LinkWeb)
	default:
		tmpTitle = "Informasi"
		tmpType = "info_myfuso"
		tmpText = fmt.Sprintf(`{"id":"%s","text":"%s","cover_image":"%s","description":"%s","link_web":"%s"}`, tmpCustomNotif.Id, tmpCustomNotif.Title, tmpCustomNotif.LinkImageWeb, tmpCustomNotif.Description, tmpCustomNotif.LinkWeb)
	}

	tmpSkipDB := false
	tmpFormNotif := models2.NotificationRequest{
		Title:            tmpTitle,
		Text:             tmpText,
		Type:             tmpType,
		SendTo:           "web",
		NotificationType: "broadcast",
		CreatedAt:        fmt.Sprintf("%s %s", tmpCustomNotif.PengirimanBerikutnya, tmpCustomNotif.TimeCronjob),
		SkipDB:           &tmpSkipDB,
	}

	if tmpFormNotif.CreatedAt != "" {
		_, errx := uttime.ParseWithFormat("2006-01-02 15:04:05", tmpFormNotif.CreatedAt)
		if errx != nil {
			fmt.Println(errx.Error())
		} else {
			_, err := engine.SendNotification(tmpFormNotif)
			if err != nil {
				fmt.Println(err.Error())
			}
			fmt.Println(fmt.Sprintf("Send notif at : %s", tmpFormNotif.CreatedAt))
			_, _ = SendingFCMContent(tmpType, tmpTitle, tmpFormNotif.ID, tmpFormNotif.Title, tmpText)
		}
	}
	//_, _ = SendingFCMContent(tmpType, tmpTitle, tmpCustomNotif.Id, tmpFormNotif.Title, tmpText)
}

func SendingFCMContent(tmpType, tmpTitle, tmpCustomeNotifId, tmpTitleCustom, tmpText string) (result string, err error) {

	ctx := context.Background()
	opt := option.WithCredentialsFile("my-fuso-81375-firebase-adminsdk-aihny-c26072971b.json")

	app, err := firebase.NewApp(ctx, nil, opt)
	if err != nil {
		log.Fatalf("error initializing app: %v", err)
	}

	client, err := app.Messaging(ctx)
	if err != nil {
		log.Fatalf("error getting Messaging client: %v", err)
	}

	tmpTopic := os.Getenv("FCM_TOPIC")
	tmpToken, _ := engine.GetTokenFirebase()
	if len(tmpToken) > 0 {
		_, errx := client.SubscribeToTopic(ctx, tmpToken, tmpTopic)
		if errx != nil {
			log.Fatalf("error getting Messaging client: %v", err)
		}

		tmpLink := fmt.Sprintf("%s/?globalNotifId=%s", os.Getenv("URL_MYFUSO"), tmpCustomeNotifId)

		message := &messaging.Message{
			Topic: tmpTopic,
			Data: map[string]string{
				"environment":     os.Getenv("FCM_ENVIRONMENT"),
				"id_custom_notif": tmpCustomeNotifId,
				"title":           tmpTitle,
				"type_name":       tmpType,
				"text":            tmpText,
			},
			Webpush: &messaging.WebpushConfig{
				Notification: &messaging.WebpushNotification{
					Title: tmpTitle,
					Body:  tmpTitleCustom,
					Icon:  "https://devvisa.ktbfuso.id/images/ktb_logo.png",
				},
				FCMOptions: &messaging.WebpushFCMOptions{
					Link: tmpLink,
				},
			},
		}

		response, err := client.Send(ctx, message)
		if err != nil {
			log.Fatalf("error sending message: %v", err)
		}

		fmt.Println(fmt.Sprintf("==========> %s FCM response: %s", uttime.Now().Format("2006-01-02 15:04:00"), response))
	}
	return "", nil
}
