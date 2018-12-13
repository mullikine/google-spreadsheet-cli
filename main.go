package main

// https://docs.google.com/spreadsheets/d/1HD-8RW4YBKA1lmwaDveX5Hf-__6UTJ6p7LLoYjqIaho/edit#gid=0

// Spreadsheet ID:
// 1HD-8RW4YBKA1lmwaDveX5Hf-__6UTJ6p7LLoYjqIaho

import (
	"context"
	"fmt"
	"io/ioutil"

	"encoding/csv"
	"github.com/d4l3k/go-pry/pry"
	"github.com/juju/errors"
	"github.com/urfave/cli"
	"golang.org/x/oauth2/google"
	"gopkg.in/Iwark/spreadsheet.v2"
	"io"
	"log"
	"os"
)

func spread(spreadsheetID string) (spreadsheet.Spreadsheet, error) {
	service, err := newService("/home/shane/notes2018/ws/codelingo/pipeline/automate/codelingo-sheets-1543871274729-1c2d278df245.json")
	if err != nil {
		return spreadsheet.Spreadsheet{}, errors.Trace(err)
	}

	return service.FetchSpreadsheet(spreadsheetID)
}

func newService(creds string) (*spreadsheet.Service, error) {
	data, err := ioutil.ReadFile(creds)
	if err != nil {
		return nil, err
	}
	conf, err := google.JWTConfigFromJSON(data, spreadsheet.Scope)
	if err != nil {
		return nil, err
	}
	client := conf.Client(context.TODO())
	return spreadsheet.NewServiceWithClient(client), nil
}

func LoadDataFromCSV(fileName string) (cells [][]string, err error) {
	file, err := os.Open(fileName)
	if err != nil {
		fmt.Println("Error:", err)
		return nil, err
	}
	defer file.Close()

	reader := csv.NewReader(file)
	for {
		record, err := reader.Read()

		if err == io.EOF {
			break
		} else if err != nil {
			fmt.Println("Error:", err)
			return nil, err
		}

		cells = append(cells, record)
	}

	return cells, nil
}

func cliMain(c *cli.Context) error {
	fmt.Printf("Hello %q", c.Args().Get(0))

	//fmt.Println("boom! I say!")

	spread, err := spread("1HD-8RW4YBKA1lmwaDveX5Hf-__6UTJ6p7LLoYjqIaho")
	if err != nil {
		return errors.Trace(err)
	}

	_, err = spread.SheetByTitle("Pipeline")
	if err != nil {
		return errors.Trace(err)
	}

	pry.Pry()

	LoadDataFromCSV("/home/shane/notes2018/remote/frontend/wizard/homicides.csv")

	// row := 1
	// column := 2
	// sheet.Update(row, column, "hogehoge")
	// sheet.Update(3, 2, "fugafuga")

	//ap.Synchronize()

	// if err != nil {
	// 	return errors.Trace(err)
	// }

	// //	automatedPipeline, err := spreadsheet.SheetByTitle("Automated Pipeline")

	// err := spread.Synchronize()
	// if err != nil {
	// 	return errors.Trace(err)
	// }

	os.Exit(0)

	return nil
}

func main() {
	app := cli.NewApp()
	app.Name = "gss"
	app.Usage = "Interact with Google Spreadsheets"

	app.Action = cliMain

	// app.Action = func(c *cli.Context) error {
	// 	fmt.Printf("Hello %q", c.Args().Get(0))
	// 	//fmt.Println("boom! I say!")
	// 	return nil
	// }

	app.Flags = []cli.Flag{
		cli.StringFlag{
			Name:  "upload",
			Value: "path/to/my_data.csv",
			Usage: "CSV file for uploading (to replace the contents of a sheet)",
		},
	}

	err := app.Run(os.Args)
	if err != nil {
		log.Fatal(err)
	}

	os.Exit(0)
}
