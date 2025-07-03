package main

import (
	"context"
	firebase "firebase.google.com/go/v4"
	"firebase.google.com/go/v4/messaging"
	"fmt"
	"github.com/iikmaulana/mini-engine/custom_promo/engine"
	"github.com/iikmaulana/mini-engine/custom_promo/lib"
	"github.com/iikmaulana/mini-engine/custom_promo/models"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
	"github.com/robfig/cron/v3"
	"github.com/uzzeet/uzzeet-gateway/libs/utils/uttime"
	"google.golang.org/api/option"
	"log"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"
)

func main() {
	err := godotenv.Load("custom_promo/.env")
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
	tmpData, _ := engine.GetListCustomPromo()
	for _, v := range tmpData {
		tmpTime[v.ID] = fmt.Sprintf("%s %s", v.PengirimanBerikutnya, v.TimeCronjob)
		fmt.Println(fmt.Sprintf("%s ===> %s", v.ID, fmt.Sprintf("%s %s", v.PengirimanBerikutnya, v.TimeCronjob)))
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
	//runCront("faf92442-d591-4938-8066-ea8e63c657b9")
}

func runCront(tmpID string) {
	tmpCustomPromo, _ := engine.GetCustomPromo(tmpID)

	tmpType := ""
	tmpTitle := ""
	tmpText := ""
	switch tmpCustomPromo.TypePromo {
	case "sales":
		tmpTitle = "Promo Sales"
		tmpType = "promo_myfuso_customer"
		tmpText = fmt.Sprintf(`{"id":"%s","text":"%s","cover_image":"%s","description":"%s"}`, tmpCustomPromo.ID, tmpCustomPromo.Title, tmpCustomPromo.LinkImage, tmpCustomPromo.Description)
	case "service":
		tmpTitle = "Promo Service"
		tmpType = "promo_myfuso_customer"
		tmpText = fmt.Sprintf(`{"id":"%s","text":"%s","cover_image":"%s","description":"%s"}`, tmpCustomPromo.ID, tmpCustomPromo.Title, tmpCustomPromo.LinkImage, tmpCustomPromo.Description)
	case "sparepart":
		tmpTitle = "Promo Sparepart"
		tmpType = "promo_myfuso_customer"
		tmpText = fmt.Sprintf(`{"id":"%s","text":"%s","cover_image":"%s","description":"%s"}`, tmpCustomPromo.ID, tmpCustomPromo.Title, tmpCustomPromo.LinkImage, tmpCustomPromo.Description)
	default:
		tmpTitle = "Promo"
		tmpType = "promo_myfuso_customer"
		tmpText = fmt.Sprintf(`{"id":"%s","text":"%s","cover_image":"%s","description":"%s"}`, tmpCustomPromo.ID, tmpCustomPromo.Title, tmpCustomPromo.LinkImage, tmpCustomPromo.Description)
	}

	tmpSkipDB := false
	if tmpCustomPromo.DealerId != nil {
		tmpUserDealer, _ := engine.GetUserDealer(*tmpCustomPromo.DealerId)
		var wg sync.WaitGroup
		for _, v := range tmpUserDealer {
			wg.Add(1)
			go func(typeNotification, titleNotification, textNotification, tmpCustomPromoId, timeNext, timeCront, status string, tmpUser models.UserResult, wg *sync.WaitGroup) {
				tmpFormPromo := models.NotificationRequest{
					Title:            titleNotification,
					Text:             textNotification,
					Type:             typeNotification,
					SendTo:           "web",
					CreatedAt:        fmt.Sprintf("%s %s", timeNext, timeCront),
					SkipDB:           &tmpSkipDB,
					OrganizationID:   tmpUser.OrganizationId,
					UserID:           tmpUser.UserId,
					ReadStatus:       0,
					NotificationType: "individual",
				}
				if tmpFormPromo.CreatedAt != "" {
					_, errx := uttime.ParseWithFormat("2006-01-02 15:04:05", tmpFormPromo.CreatedAt)
					if errx != nil {
						fmt.Println(errx.Error())
					} else {
						_, err := engine.SendNotification(tmpFormPromo)
						if err != nil {
							fmt.Println(err.Error())
						}
						fmt.Println(fmt.Sprintf("Send notif at : %s", tmpFormPromo.CreatedAt))
					}
				}
				defer wg.Done()
			}(tmpType, tmpTitle, tmpText, tmpCustomPromo.ID, tmpCustomPromo.PengirimanBerikutnya, tmpCustomPromo.TimeCronjob, tmpCustomPromo.Status, v, &wg)
		}
		wg.Wait()
		if tmpCustomPromo.Status == "berlangsung" {
			_, _ = SendingFCMDealerContent(tmpType, tmpTitle, tmpCustomPromo.ID, tmpCustomPromo.Title, tmpText, *tmpCustomPromo.DealerId)
		}
	} else {
		tmpFormPromo := models.NotificationRequest{
			Title:            tmpTitle,
			Text:             tmpText,
			Type:             tmpType,
			SendTo:           "web",
			NotificationType: "broadcast",
			CreatedAt:        fmt.Sprintf("%s %s", tmpCustomPromo.PengirimanBerikutnya, tmpCustomPromo.TimeCronjob),
			SkipDB:           &tmpSkipDB,
		}
		if tmpFormPromo.CreatedAt != "" {
			_, errx := uttime.ParseWithFormat("2006-01-02 15:04:05", tmpFormPromo.CreatedAt)
			if errx != nil {
				fmt.Println(errx.Error())
			} else {
				_, err := engine.SendNotification(tmpFormPromo)
				if err != nil {
					fmt.Println(err.Error())
				}
				fmt.Println(fmt.Sprintf("Send notif at : %s", tmpFormPromo.CreatedAt))
				if tmpCustomPromo.Status == "berlangsung" {
					_, _ = SendingFCMContent(tmpType, tmpTitle, tmpCustomPromo.ID, tmpFormPromo.Title, tmpText)
				}
			}
		}
	}
}

func SendingFCMContent(tmpType, tmpTitle, tmpCustomeNotifID, tmpTitleCustom, tmpText string) (result string, err error) {

	fmt.Println("send notif")
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

		tmpLink := fmt.Sprintf("%s/promo/%s", os.Getenv("URL_MYFUSO"), tmpCustomeNotifID)

		message := &messaging.Message{
			Topic: tmpTopic,
			Data: map[string]string{
				"environment":     os.Getenv("FCM_ENVIRONMENT"),
				"id_custom_notif": tmpCustomeNotifID,
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

func SendingFCMDealerContent(tmpType, tmpTitle, tmpCustomeNotifID, tmpTitleCustom, tmpText, dealerId string) (result string, err error) {

	fmt.Println("send notif")
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
	tmpToken, _ := engine.GetTokenDealerFirebase(dealerId)
	for _, v := range tmpToken {
		_, errx := client.SubscribeToTopic(ctx, tmpToken, tmpTopic)
		if errx != nil {
			log.Fatalf("error getting Messaging client: %v", err)
		}

		tmpLink := fmt.Sprintf("%s/promo/%s", os.Getenv("URL_MYFUSO"), tmpCustomeNotifID)

		message := &messaging.Message{
			Token: v,
			Data: map[string]string{
				"environment":     os.Getenv("FCM_ENVIRONMENT"),
				"id_custom_notif": tmpCustomeNotifID,
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
