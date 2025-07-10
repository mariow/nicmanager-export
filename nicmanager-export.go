package main

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"image/color"
	"io"
	"log"
	"net/http"
	"os"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/data/validation"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

func main() {
	a := app.NewWithID("witte.io.nicmanager-export")
	w := a.NewWindow("Nicmanager Exporter") // main app name shown in process list

	uiTitle := widget.NewLabel("Nicmanager Exporter")

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

	// TODO: Validator in separate Funktionen auslagern
	uiCredUsername := widget.NewEntry()
	uiCredUsername.SetPlaceHolder("account.user")
	uiCredUsername.Validator = validation.NewRegexp("^[a-z0-9_.-]+$", "Darf nicht leer sein")
	uiCredPassword := widget.NewPasswordEntry()
	uiCredPassword.SetPlaceHolder("supergeheim")
	uiCredPassword.Validator = validation.NewRegexp("^.+$", "Darf nicht leer sein")
	uiCutoffDate := widget.NewEntry()
	uiCutoffDate.SetPlaceHolder("2020-03-01")
	//TODO Validation mit Regex plus time.Parse bauen
	uiCutoffDate.Validator = validation.NewRegexp("^20[0-9]{2}-[0-9]{2}-[0-9]{2}$", "Datum muss das Format YYYY-MM-DD haben")
	uiFilename := widget.NewEntry()
	uiFilename.SetPlaceHolder("Export_12345.csv")
	uiFilename.Validator = validation.NewRegexp("^[a-zA-Z0-9_ -]+.csv$", "Der Dateiname muss auf .csv enden und die Datei darf noch nicht existieren")

	obscureProgress := widget.NewProgressBarInfinite()
	obscureProgress.Hide()

	statusMessage := canvas.NewText("", theme.TextColor())
	statusMessage.Hide()

	var uiForm = &widget.Form{
		Items: []*widget.FormItem{
			{Text: "Benutzer", Widget: uiCredUsername},
			{Text: "Passwort", Widget: uiCredPassword},
			{Text: "Stichtag", Widget: uiCutoffDate},
			{Text: "Zieldatei", Widget: uiFilename},
		},
		OnSubmit: func() {
			// show progressbar
			obscureProgress.Show()

			// open output file
			// TODO: more checks needed
			outFile, fErr := os.Create(uiFilename.Text)
			if fErr != nil {
				log.Fatal(fErr)
			}
			defer outFile.Close()

			// TODO: Korrekterweise sollte das Datum immer mit 23:59:59 geparst werden
			cutoffDate, dtErr := time.Parse("2006-01-02", uiCutoffDate.Text)
			if dtErr != nil {
				log.Fatal(dtErr)
			}

			// fetch data from API and write to output file
			recordsWritten, err := fetchAndWrite(
				uiCredUsername.Text,
				uiCredPassword.Text,
				cutoffDate,
				outFile,
			)

			if err != nil {
				log.Fatal(err)
			}

			statusMessage.Text = fmt.Sprintf("%d Zeilen geschrieben", recordsWritten)
			statusMessage.Show()

			// clear fields to disable submit button
			uiCutoffDate.SetText("")

			// hide progressbar
			obscureProgress.Hide()
		},
	}

	w.SetContent(container.NewVBox(
		uiTitle,
		canvas.NewLine(theme.TextColor()),
		uiForm,

		obscureProgress,
		statusMessage,
		layout.NewSpacer(),
		canvas.NewText("© 2021-2025", color.White),
	))
	w.Resize(fyne.NewSize(300, 500))

	w.ShowAndRun()
}

func fetchAndWrite(login string, password string, cutoffDate time.Time, outFile *os.File) (int, error) {

	// init vars
	var morePages bool = true
	var recordsWritten int = 0
	var headerWritten bool = false

	// contact API
	client := http.Client{}
	csvWriter := csv.NewWriter(outFile)
	defer csvWriter.Flush()

	for pageNo := 1; morePages; pageNo++ {
		log.Println("requesting pageno " + fmt.Sprintf("%d", pageNo))

		fulldoc, err := fetchNicmanagerAPI(client, login, password, pageNo)
		if err != nil {
			log.Fatal(err)
		}

		// TODO: this needs to go into a separate function
		var domainList []Domain
		jsonErr := json.Unmarshal(fulldoc, &domainList)
		if jsonErr != nil {
			log.Fatal(jsonErr)
		}

		//TODO: Irgendwo hier gehen Daten verloren. Im JSON sind noch Daten drin, hier nicht mehr

		for _, rowData := range domainList {
			if !headerWritten {
				csvWriter.Write([]string{
					"Domain",
					"Order Date",
					"Reg Date",
					"Close Date",
				})
				headerWritten = true
			}

			// parse dates
			dateOrd, _ := parseAPIdate(rowData.OrderDateTime)
			dateReg, _ := parseAPIdate(rowData.RegistrationDateTime)

			// format Delete date for output
			dateDelFmt := ""
			if rowData.DeleteDateTime != "" {
				parsedDate, _ := parseAPIdate(rowData.DeleteDateTime)
				dateDelFmt = parsedDate.Format("2006-01-02")
			}

			//log.Printf("Dateldel: %s (%s) - Cutoff_Unix: %d", rowData.DeleteDateTime, dateDelFmt, cutoffDate.Unix())

			if rowData.IsBelowCutoff(cutoffDate) {
				csvWriter.Write([]string{
					rowData.Name,
					dateOrd.Format("2006-01-02"),
					dateReg.Format("2006-01-02"),
					dateDelFmt,
				})
				recordsWritten++
				log.Println("Written") // DEBUG
			}
			log.Println("---") //DEBUG
		}

		// do we have more pages?
		morePages = (len(domainList) == 100)
	}

	return recordsWritten, nil
}

func fetchNicmanagerAPI(client http.Client, login string, password string, pageNo int) ([]byte, error) {
	var apiURL string = "https://api.nicmanager.com/v1/domains?limit=100&page="
	req, rErr := http.NewRequest("GET", apiURL+fmt.Sprintf("%d", pageNo), nil)
	req.Header.Add("Accept", "application/json")
	req.SetBasicAuth(login, password)
	if rErr != nil {
		log.Fatal(rErr)
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
	//fmt.Println("Header output:")
	//spew.Dump(res.Header)

	// convert response into string
	return io.ReadAll(res.Body)

}
