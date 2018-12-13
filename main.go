package main

// https://docs.google.com/spreadsheets/d/1HD-8RW4YBKA1lmwaDveX5Hf-__6UTJ6p7LLoYjqIaho/edit#gid=0

// Spreadsheet ID:
// 1HD-8RW4YBKA1lmwaDveX5Hf-__6UTJ6p7LLoYjqIaho

import (
	"context"
	"fmt"
	"io/ioutil"

	"encoding/csv"
	"github.com/juju/errors"
	"github.com/urfave/cli"
	"golang.org/x/oauth2/google"
	"gopkg.in/Iwark/spreadsheet.v2"
	"io"
	"log"
	"os"
)

func spread(spreadsheetID string) (spreadsheet.Spreadsheet, error) {
	service, err := newService(c.String("service-creds"))
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
	// fmt.Printf("Hello %q", c.Args().Get(0))

	firstArg := c.Args().Get(0)

	switch {
	case  firstArg == "update":
		return nil
	}

	spread, err := spread(c.String("sheetid"))
	if err != nil {
		return errors.Trace(err)
	}

	sheet, err := spread.SheetByTitle(c.String("sheetname"))
	if err != nil {
		return errors.Trace(err)
	}

	m, err := LoadDataFromCSV(c.String("data"))
	if err != nil {
		return errors.Trace(err)
	}

	for i := 0; i < len(m); i++ {
		record := m[i]
		for j := 0; j < len(record); j++ {
			cell := record[j]
			fmt.Printf("%s\n", cell)
			sheet.Update(i, j, cell)
		}
	}
	sheet.Synchronize()

	// fmt.Printf("%s\n", m[76][2])

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
			Name:  "sheetid",
			Value: "1HD-8RW4YBKA1lmwaDveX5Hf-__6UTJ6p7LLoYjqIaho",
			Usage: "Google Spreadsheets ID (from url)",
		},
		cli.StringFlag{
			Name:  "sheetname",
			Value: "Pipeline",
			Usage: "Sheet Name (tab name)",
		},
		cli.StringFlag{
			Name:  "data",
			Value: "/home/shane/notes2018/remote/frontend/wizard/homicides.csv",
			Usage: "CSV file for uploading (to replace the contents of a sheet)",
		},
		cli.StringFlag{
			Name:  "col",
			Value: "0",
			Usage: "Column offset number (CSV is placed at this column)",
		},
		cli.StringFlag{
			Name:  "row",
			Value: "0",
			Usage: "Row offset number (CSV is placed at this row)",
		},
		cli.StringFlag{
			Name:  "service-creds",
			Value: "/home/shane/notes2018/ws/codelingo/pipeline/automate/codelingo-sheets-1543871274729-1c2d278df245.json",
			Usage: "Service Account json credentials (Sheet must be shared with the email defined inside this json file)",
		},
	}

	err := app.Run(os.Args)
	if err != nil {
		log.Fatal(err)
	}

	os.Exit(0)
}
