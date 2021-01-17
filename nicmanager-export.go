package main

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"image/color"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"time"

	"fyne.io/fyne"
	"fyne.io/fyne/app"
	"fyne.io/fyne/canvas"
	"fyne.io/fyne/data/validation"
	"fyne.io/fyne/layout"
	"fyne.io/fyne/theme"
	"fyne.io/fyne/widget"
)

func main() {
	a := app.NewWithID("witte.io.nicmanager")
	w := a.NewWindow("Nicmanager Exporter") // main app name shown in process list

	ui_title := widget.NewLabel("Nicmanager Exporter")

	//TODO: Multiline text der die Verwendung dokumentiert

	/*
	   Nicmanager Credentials
	   Username [   ]   | Password [   ]
	   Checkbox Inventory cutoff [    ]
	   Output file  [    ]
	   [ Start ]
	   -----
	   Progress
	*/

	ui_cred_username := widget.NewEntry()
	ui_cred_username.SetPlaceHolder("account.user")
	ui_cred_username.Validator = validation.NewRegexp("^[a-z0-9]+$", "Darf nicht leer sein")
	ui_cred_password := widget.NewPasswordEntry()
	ui_cred_password.SetPlaceHolder("supergeheim")
	ui_cred_password.Validator = validation.NewRegexp("^[a-z0-9]+$", "Darf nicht leer sein")
	ui_cutoff_date := widget.NewEntry()
	ui_cutoff_date.SetPlaceHolder("2020-03-01")
	//TODO Validation mit Regex plus time.Parse bauen
	ui_cutoff_date.Validator = validation.NewRegexp("^20[0-9]{2}-[0-9]{2}-[0-9]{2}$", "Datum muss das Format YYYY-MM-DD haben")
	ui_file_name := widget.NewEntry()
	ui_file_name.SetPlaceHolder("Export_12345.csv")
	ui_file_name.Validator = validation.NewRegexp("^[a-zA-Z0-9_ -]+.csv$", "Der Dateiname muss auf .csv enden und die Datei darf noch nicht existieren")

	obscure_progress := widget.NewProgressBarInfinite()
	obscure_progress.Hide()

	status_message := canvas.NewText("", theme.TextColor())
	status_message.Hide()

	var ui_form = &widget.Form{
		Items: []*widget.FormItem{
			{Text: "Benutzer", Widget: ui_cred_username},
			{Text: "Passwort", Widget: ui_cred_password},
			{Text: "Stichtag", Widget: ui_cutoff_date},
			{Text: "Zieldatei", Widget: ui_file_name},
		},
		OnSubmit: func() {
			// show progressbar
			obscure_progress.Show()

			// open output file
			// TODO: more checks needed
			out_f, f_err := os.Create(ui_file_name.Text)
			if f_err != nil {
				log.Fatal(f_err)
			}
			defer out_f.Close()

			cutoff_date, dt_err := time.Parse("2006-01-02", ui_cutoff_date.Text)
			if dt_err != nil {
				log.Fatal(dt_err)
			}

			records_written, err := fetchNicmanager(
				ui_cred_username.Text,
				ui_cred_password.Text,
				cutoff_date,
				out_f,
			)

			if err != nil {
				log.Fatal(err)
			}

			status_message.Text = fmt.Sprintf("%d lines written", records_written)
			status_message.Show()

			// clear fields to disable submit button
			ui_cutoff_date.SetText("")

			// hide progressbar
			obscure_progress.Hide()
		},
	}

	w.SetContent(widget.NewVBox(
		ui_title,
		canvas.NewLine(theme.TextColor()),
		ui_form,

		obscure_progress,
		status_message,
		layout.NewSpacer(),
		canvas.NewText("© 2021", color.White),
	))
	w.Resize(fyne.NewSize(250, 500))

	w.ShowAndRun()
	tidyUp()
}

// tidyUp is called after the app quits
func tidyUp() {
	fmt.Println("Exit")

}

func fetchNicmanager(login string, password string, cutoff_date time.Time, out_f *os.File) (int, error) {

	// init vars
	var api_url string = "https://api.nicmanager.com/v1/domains?limit=100&page="
	var more_pages bool = true
	var page_no int = 1
	var records_written int = 0

	// contact API
	client := http.Client{}
	csv_writer := csv.NewWriter(out_f)
	defer csv_writer.Flush()

	for more_pages {
		req, r_err := http.NewRequest("GET", api_url+fmt.Sprintf("%d", page_no), nil)
		req.Header.Add("Accept", "application/json")
		req.SetBasicAuth(login, password)
		if r_err != nil {
			log.Fatal(r_err)
		}

		res, err := client.Do(req)
		if err != nil {
			log.Fatal(err)
		}
		defer res.Body.Close()

		if res.StatusCode != 200 {
			log.Fatalf("status code error: %d %s", res.StatusCode, res.Status)
		}

		// DEBUG remove later
		/*fmt.Println("Header output:")
		spew.Dump(res.Header)*/

		//DEBUG fulldoc := []byte("[{\"order_id\":263985,\"name\":\"1822com.de\",\"renewal_mode\":\"autorenew\",\"reference\":\"\",\"whoisprotection\":false,\"order_status\":\"active\",\"event_status\":\"done\",\"event_alias\":\"KK_IN_OK\",\"order_datetime\":\"2015-08-14T08:33:01Z\",\"start_datetime\":\"2015-08-14T08:33:11Z\",\"registration_datetime\":\"2015-08-14T08:33:11Z\",\"expiration_datetime\":\"2021-01-31T22:59:00Z\",\"delete_datetime\":null,\"handles\":{\"owner\":\"IX-GM5\",\"admin\":\"IX-DZ16\",\"tech\":\"IX-GM5\",\"zone\":\"IX-GM5\"},\"nameserver\":[{\"name\":\"ns1.parkingcrew.net\",\"addr\":null,\"type\":\"NS\"},{\"name\":\"ns2.parkingcrew.net\",\"addr\":null,\"type\":\"NS\"}]},{\"order_id\":305710,\"name\":\"26m.de\",\"renewal_mode\":\"autorenew\",\"reference\":\"\",\"whoisprotection\":false,\"order_status\":\"active\",\"event_status\":\"done\",\"event_alias\":\"KK_IN_OK\",\"order_datetime\":\"2016-09-02T15:23:42Z\",\"start_datetime\":\"2016-09-02T15:23:49Z\",\"registration_datetime\":\"2016-09-02T15:23:49Z\",\"expiration_datetime\":\"2021-01-31T22:59:00Z\",\"delete_datetime\":null,\"handles\":{\"owner\":\"IX-AG7\",\"admin\":\"IX-FB1\",\"tech\":\"IX-AG7\",\"zone\":\"IX-AG7\"},\"nameserver\":[{\"name\":\"ns1.parkingcrew.net\",\"addr\":null,\"type\":\"NS\"},{\"name\":\"ns2.parkingcrew.net\",\"addr\":null,\"type\":\"NS\"}]},{\"order_id\":305202,\"name\":\"2eaux.fr\",\"renewal_mode\":\"autorenew\",\"reference\":\"\",\"whoisprotection\":false,\"order_status\":\"closed\",\"event_status\":\"done\",\"event_alias\":\"CLOSE_OK\",\"order_datetime\":\"2016-08-23T09:25:54Z\",\"start_datetime\":\"2016-08-23T09:25:57Z\",\"registration_datetime\":\"2016-08-23T10:41:54Z\",\"expiration_datetime\":\"2018-08-23T10:41:53Z\",\"delete_datetime\":\"2018-01-31T15:13:39Z\",\"handles\":{\"owner\":\"IX-FB1\",\"admin\":\"IX-FB1\",\"tech\":\"IX-AG7\",\"zone\":\"IX-AG7\"},\"nameserver\":[{\"name\":\"ns1.parkingcrew.net\",\"addr\":null,\"type\":\"NS\"},{\"name\":\"ns2.parkingcrew.net\",\"addr\":null,\"type\":\"NS\"}]},{\"order_id\":462368,\"name\":\"2eaux.fr\",\"renewal_mode\":\"autorenew\",\"reference\":\"\",\"whoisprotection\":false,\"order_status\":\"active\",\"event_status\":\"done\",\"event_alias\":\"REG_OK\",\"order_datetime\":\"2019-05-03T13:19:12Z\",\"start_datetime\":\"2019-05-03T14:02:24Z\",\"registration_datetime\":\"2019-05-03T14:02:24Z\",\"expiration_datetime\":\"2021-05-03T14:02:24Z\",\"delete_datetime\":null,\"handles\":{\"owner\":\"IX-NM11\",\"admin\":\"IX-NM11\",\"tech\":\"IX-NM11\",\"zone\":\"IX-NM11\"},\"nameserver\":[{\"name\":\"ns1.parkingcrew.net\",\"addr\":null,\"type\":\"NS\"},{\"name\":\"ns2.parkingcrew.net\",\"addr\":null,\"type\":\"NS\"}]}]")

		// convert response into string
		fulldoc, err := ioutil.ReadAll(res.Body)

		type Domain struct {
			Name                  string
			Order_Status          string
			Order_DateTime        string
			Registration_DateTime string
			Delete_DateTime       string
		}
		var domain_list []Domain

		json_err := json.Unmarshal(fulldoc, &domain_list)
		if json_err != nil {
			log.Fatal(json_err)
		}

		for _, row_data := range domain_list {
			if records_written == 0 {
				csv_writer.Write([]string{
					"Domain",
					"Order Date",
					"Reg Date",
					"Close Date",
				})
			}

			date_ord, _ := time.Parse("2006-01-02T15:04:05Z", row_data.Order_DateTime)
			date_reg, _ := time.Parse("2006-01-02T15:04:05Z", row_data.Registration_DateTime)
			date_del, _ := time.Parse("2006-01-02T15:04:05Z", row_data.Delete_DateTime)
			/*fmt.Println(row_data.Delete_DateTime)
			fmt.Println(date_del)
			fmt.Println(date_del.Unix())
			fmt.Println(cutoff_date.Unix())*/

			// filter for records without delete date or with delete date after cutoff
			date_del_fmt := ""
			bool_write_record := false
			if row_data.Delete_DateTime != "" {
				date_del_fmt = date_del.Format("2006-01-02")
				if date_del.Unix() > cutoff_date.Unix() {
					bool_write_record = true
				}
			} else {
				bool_write_record = true
			}

			if bool_write_record {
				csv_writer.Write([]string{
					row_data.Name,
					date_ord.Format("2006-01-02"),
					date_reg.Format("2006-01-02"),
					date_del_fmt,
				})
				records_written++
				//fmt.Println("Written") // DEBUG
			}
			//fmt.Println("---") //DEBUG
		}

		// do we have more pages?
		more_pages = (len(domain_list) == 100)
	}

	return records_written, nil
}
