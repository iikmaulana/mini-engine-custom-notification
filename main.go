package main

import (
	"fmt"
	"github.com/iikmaulana/mini-engine/engine"
	"github.com/iikmaulana/mini-engine/lib"
	"github.com/iikmaulana/mini-engine/models"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
	"github.com/robfig/cron/v3"
	"github.com/uzzeet/uzzeet-gateway/libs/utils/uttime"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"
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
		tmpText = fmt.Sprintf(`{"id":"%s","text":"%s","cover_image":"%s"}`, tmpCustomNotif.Id, tmpCustomNotif.Title, tmpCustomNotif.LinkImageWeb)
	case "aktivitas":
		tmpTitle = "Aktivitas"
		tmpType = "aktivitas_myfuso"
		tmpText = fmt.Sprintf(`{"id":"%s","text":"%s","cover_image":"%s"}`, tmpCustomNotif.Id, tmpCustomNotif.Title, tmpCustomNotif.LinkImageWeb)
	default:
		tmpTitle = "Informasi"
		tmpType = "info_myfuso"
		tmpText = fmt.Sprintf(`{"id":"%s","text":"%s","cover_image":"%s"}`, tmpCustomNotif.Id, tmpCustomNotif.Title, tmpCustomNotif.LinkImageWeb)
	}

	tmpSkipDB := false
	tmpFormNotif := models.NotificationRequest{
		Title:            tmpTitle,
		Text:             tmpText,
		Type:             tmpType,
		SendTo:           "web",
		NotificationType: "broadcast",
		CreatedAt:        fmt.Sprintf("%s %s", tmpCustomNotif.PengirimanBerikutnya, tmpCustomNotif.TimeCronjob),
		SkipDB:           &tmpSkipDB,
	}

	res, err := engine.SendNotification(tmpFormNotif)
	if err != nil {
		fmt.Println(err.Error())
	}
	fmt.Println(res)
}
