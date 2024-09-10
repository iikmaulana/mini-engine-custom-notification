package engine

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/gearintellix/u2"
	"github.com/golang/protobuf/ptypes/any"
	"github.com/iikmaulana/mini-engine/custom_notification/models"
	"github.com/iikmaulana/mini-engine/custom_notification/service/grpc/packets"
	"github.com/jmoiron/sqlx"
	"github.com/uzzeet/uzzeet-gateway/libs"
	"github.com/uzzeet/uzzeet-gateway/libs/helper"
	"github.com/uzzeet/uzzeet-gateway/libs/helper/logger"
	"github.com/uzzeet/uzzeet-gateway/libs/helper/serror"
	"github.com/uzzeet/uzzeet-gateway/libs/utils/uttime"
	"google.golang.org/grpc"
	"log"
	"os"
	"strings"
	"time"
)

var ctx = context.Background()

func ConnectionCockroachDB() (*sqlx.DB, serror.SError) {
	sqlConn := helper.Env(libs.DBConnStr, `
        host=__host__
		port=__port__
        user=__user__
		port=__port__
        password=__pwd__
        dbname=__name__
        sslmode=__sslMode__
        application_name=__appKey__
    `)
	sqlConn = u2.Binding(sqlConn, map[string]string{
		"host":    helper.Env(libs.DBHost, "127.0.0.1"),
		"user":    helper.Env(libs.DBUser, "root"),
		"pwd":     helper.Env(libs.DBPwd, ""),
		"name":    helper.Env(libs.DBName, "asgard"),
		"sslMode": helper.Env(libs.DBSSLMode, "disable"),
		"appKey":  helper.Env(libs.AppKey, "device"),
		"appName": helper.Env(libs.AppName, "Device"),
		"port":    helper.Env(libs.DBPort, "54320"),
	})

	db, err := sqlx.Connect(helper.Env(libs.DBEngine, "impl"), sqlConn)
	if err != nil {
		return nil, serror.NewFromErrorc(err, fmt.Sprintf("failed connect to database %s", helper.Env(libs.DBName, "asgard")))
	}

	db.SetConnMaxLifetime(time.Minute * time.Duration(helper.StringToInt(helper.Env(libs.DBConnLifetime, "15"), 15)))
	db.SetMaxIdleConns(int(helper.StringToInt(helper.Env(libs.DBConnMaxIdle, "300"), 300)))
	db.SetMaxOpenConns(int(helper.StringToInt(helper.Env(libs.DBConnMaxOpen, "300"), 300)))

	return db, nil
}

func GetListCustomNotification() (result []models.CustomNotificationResult, serr serror.SError) {
	tmpQuery := `SELECT 
				id, 
				link_image, 
				title, 
				description, 
				category, 
				link_web, 
				type_notification, 
				frekuensi,
				coalesce(start_date :: TIMESTAMP(0) :: STRING, '') AS start_date,
				coalesce(end_date :: TIMESTAMP(0) :: STRING, '') AS end_date,
				coalesce(time_cronjob :: TIME(0) :: STRING, '') AS time_cronjob,
				coalesce(created_at :: TIMESTAMP(0) :: STRING, '') AS created_at,
				coalesce(updated_at :: TIMESTAMP(0) :: STRING, '') AS updated_at
				FROM db_myfuso.custom_notification`

	db, _ := ConnectionCockroachDB()

	rows, err := db.Queryx(tmpQuery)
	if err != nil {
		return result, serror.NewFromError(err)
	}

	defer rows.Close()
	tmpData := []models.CustomNotificationResult{}
	for rows.Next() {
		var tmpResult models.CustomNotificationResult
		if err := rows.StructScan(&tmpResult); err != nil {
			fmt.Println(err.Error())
			return result, serror.New("Failed scan for struct models")
		}
		tmpData = append(tmpData, tmpResult)
	}

	tmpArr := []models.CustomNotificationResult{}
	for _, v := range tmpData {
		if v.Category == "onetime" {
			tmpNow := uttime.Now()
			tmpEnd := fmt.Sprintf("%s %s", strings.Split(v.EndDate, " ")[0], v.TimeCronjob)
			endDate, _ := uttime.ParseFromString(tmpEnd)
			if tmpNow.Before(endDate) {
				v.PengirimanBerikutnya = endDate.Format("2006-01-02")
			}
		} else if v.Category == "periodic" {
			if v.Frekuensi == "harian" {
				tmpNow := uttime.Now()
				tmpStart := fmt.Sprintf("%s %s", strings.Split(v.StartDate, " ")[0], v.TimeCronjob)
				start, _ := uttime.ParseFromString(tmpStart)
				tmpEnd := fmt.Sprintf("%s %s", strings.Split(v.EndDate, " ")[0], v.TimeCronjob)
				endDate, _ := uttime.ParseFromString(tmpEnd)
				nextDay := start
				hasNextDay := nextDay.Before(endDate)
				for tmpNow.After(nextDay) && hasNextDay {
					nextDay = nextDay.AddDate(0, 0, 1)
					hasNextDay = nextDay.Before(endDate)
				}
				if hasNextDay {
					v.PengirimanBerikutnya = nextDay.Format("2006-01-02")
				}
			} else if v.Frekuensi == "mingguan" {
				tmpNow := uttime.Now()
				tmpStart := fmt.Sprintf("%s %s", strings.Split(v.StartDate, " ")[0], v.TimeCronjob)
				start, _ := uttime.ParseFromString(tmpStart)
				tmpEnd := fmt.Sprintf("%s %s", strings.Split(v.EndDate, " ")[0], v.TimeCronjob)
				endDate, _ := uttime.ParseFromString(tmpEnd)
				nextDay := start
				hasNextDay := nextDay.Before(endDate)
				for tmpNow.After(nextDay) && hasNextDay {
					nextDay = nextDay.AddDate(0, 0, 7)
					hasNextDay = nextDay.Before(endDate)
				}
				if hasNextDay {
					v.PengirimanBerikutnya = nextDay.Format("2006-01-02")
				}
			} else if v.Frekuensi == "bulanan" {
				tmpNow := uttime.Now()
				tmpStart := fmt.Sprintf("%s %s", strings.Split(v.StartDate, " ")[0], v.TimeCronjob)
				start, _ := uttime.ParseFromString(tmpStart)
				tmpEnd := fmt.Sprintf("%s %s", strings.Split(v.EndDate, " ")[0], v.TimeCronjob)
				endDate, _ := uttime.ParseFromString(tmpEnd)
				nextDay := start
				hasNextDay := nextDay.Before(endDate)
				for tmpNow.After(nextDay) && hasNextDay {
					nextDay = nextDay.AddDate(0, 0, 30)
					hasNextDay = nextDay.Before(endDate)
				}
				if hasNextDay {
					v.PengirimanBerikutnya = nextDay.Format("2006-01-02")
				}
			}
		}
		if v.PengirimanBerikutnya == "" {
			v.PengirimanBerikutnya = "-"
			v.Status = "selesai"
		} else {
			v.Status = "berlangsung"
		}
		tmpArr = append(tmpArr, v)
	}

	tmpRes := []models.CustomNotificationResult{}
	for _, v := range tmpArr {
		if v.PengirimanBerikutnya != "-" {
			tmpRes = append(tmpRes, v)
		}
	}

	defer db.Close()
	result = tmpRes
	return result, nil
}

func GetCustomNotification(tmpId string) (result models.CustomNotificationResult, serr serror.SError) {
	tmpQuery := `SELECT 
				id, 
				link_image, 
				title, 
				description, 
				category, 
				link_web, 
				type_notification, 
				frekuensi,
				coalesce(start_date :: TIMESTAMP(0) :: STRING, '') AS start_date,
				coalesce(end_date :: TIMESTAMP(0) :: STRING, '') AS end_date,
				coalesce(time_cronjob :: TIME(0) :: STRING, '') AS time_cronjob,
				coalesce(created_at :: TIMESTAMP(0) :: STRING, '') AS created_at,
				coalesce(updated_at :: TIMESTAMP(0) :: STRING, '') AS updated_at
				FROM db_myfuso.custom_notification where id = $1`

	db, _ := ConnectionCockroachDB()
	rows, err := db.Queryx(tmpQuery, tmpId)
	if err != nil {
		return result, serror.NewFromError(err)
	}

	defer rows.Close()
	tmpData := []models.CustomNotificationResult{}
	for rows.Next() {
		var tmpResult models.CustomNotificationResult
		if err := rows.StructScan(&tmpResult); err != nil {
			fmt.Println(err.Error())
			return result, serror.New("Failed scan for struct models")
		}
		tmpData = append(tmpData, tmpResult)
	}

	tmpArr := []models.CustomNotificationResult{}
	for _, v := range tmpData {
		if v.Category == "onetime" {
			duration := 30 * time.Second
			tmpNow := uttime.Now().Add(-duration)
			tmpEnd := fmt.Sprintf("%s %s", strings.Split(v.EndDate, " ")[0], v.TimeCronjob)
			endDate, _ := uttime.ParseFromString(tmpEnd)
			if tmpNow.Before(endDate) {
				v.PengirimanBerikutnya = endDate.Format("2006-01-02")
			}
		} else if v.Category == "periodic" {
			if v.Frekuensi == "harian" {
				duration := 30 * time.Second
				tmpNow := uttime.Now().Add(-duration)
				tmpStart := fmt.Sprintf("%s %s", strings.Split(v.StartDate, " ")[0], v.TimeCronjob)
				start, _ := uttime.ParseFromString(tmpStart)
				tmpEnd := fmt.Sprintf("%s %s", strings.Split(v.EndDate, " ")[0], v.TimeCronjob)
				endDate, _ := uttime.ParseFromString(tmpEnd)
				nextDay := start
				hasNextDay := nextDay.Before(endDate)
				for tmpNow.After(nextDay) && hasNextDay {
					nextDay = nextDay.AddDate(0, 0, 1)
					hasNextDay = nextDay.Before(endDate)
				}
				if hasNextDay {
					v.PengirimanBerikutnya = nextDay.Format("2006-01-02")
				}
			} else if v.Frekuensi == "mingguan" {
				duration := 30 * time.Second
				tmpNow := uttime.Now().Add(-duration)
				tmpStart := fmt.Sprintf("%s %s", strings.Split(v.StartDate, " ")[0], v.TimeCronjob)
				start, _ := uttime.ParseFromString(tmpStart)
				tmpEnd := fmt.Sprintf("%s %s", strings.Split(v.EndDate, " ")[0], v.TimeCronjob)
				endDate, _ := uttime.ParseFromString(tmpEnd)
				nextDay := start
				hasNextDay := nextDay.Before(endDate)
				for tmpNow.After(nextDay) && hasNextDay {
					nextDay = nextDay.AddDate(0, 0, 7)
					hasNextDay = nextDay.Before(endDate)
				}
				if hasNextDay {
					v.PengirimanBerikutnya = nextDay.Format("2006-01-02")
				}
			} else if v.Frekuensi == "bulanan" {
				duration := 30 * time.Second
				tmpNow := uttime.Now().Add(-duration)
				tmpStart := fmt.Sprintf("%s %s", strings.Split(v.StartDate, " ")[0], v.TimeCronjob)
				start, _ := uttime.ParseFromString(tmpStart)
				tmpEnd := fmt.Sprintf("%s %s", strings.Split(v.EndDate, " ")[0], v.TimeCronjob)
				endDate, _ := uttime.ParseFromString(tmpEnd)
				nextDay := start
				hasNextDay := nextDay.Before(endDate)
				for tmpNow.After(nextDay) && hasNextDay {
					nextDay = nextDay.AddDate(0, 0, 30)
					hasNextDay = nextDay.Before(endDate)
				}
				if hasNextDay {
					v.PengirimanBerikutnya = nextDay.Format("2006-01-02")
				}
			}
		}
		if v.PengirimanBerikutnya == "" {
			v.PengirimanBerikutnya = "-"
			v.Status = "selesai"
		} else {
			v.Status = "berlangsung"
		}
		tmpArr = append(tmpArr, v)
	}

	for _, v := range tmpArr {
		result = v
	}
	defer db.Close()
	return result, nil
}

func SendNotification(form models.NotificationRequest) (result string, serr serror.SError) {
	conn, err := grpc.Dial(os.Getenv("RPC_NOTIFICATION"), grpc.WithInsecure())
	if err != nil {
		log.Fatalf("Can't connect to the service : %v", err)
	}

	tmpByte, err := json.Marshal(form)
	if err != nil {
		return result, serror.NewFromError(err)
	}

	data := any.Any{
		Value: tmpByte,
	}

	client := packets.NewNotificationClient(conn)
	output, err := client.SendNotificationUsecase(context.Background(), &packets.SendNotificationRequest{
		Data: &data,
	})

	if err != nil {
		serrFmt := fmt.Sprintf("[service][repository][core][Notification] while grpc SendNotificationUsecase: %+v", err)
		logger.Err(serrFmt)
		return result, serror.NewFromErrorc(err, serrFmt)
	}
	if output.GetStatus() == 1 {
		result = string(output.GetData().Value)
	}

	return result, nil
}

func GetTokenFirebase() (result []string, serr serror.SError) {
	tmpQuery := `SELECT token from db_myfuso.firebase_token`

	db, _ := ConnectionCockroachDB()
	rows, err := db.Queryx(tmpQuery)
	if err != nil {
		return result, serror.NewFromError(err)
	}

	defer rows.Close()

	for rows.Next() {
		var tmpResult string
		if err := rows.Scan(&tmpResult); err != nil {
			fmt.Println(err.Error())
			return result, serror.New("Failed scan for struct models")
		}
		result = append(result, tmpResult)
	}

	return result, nil
}
