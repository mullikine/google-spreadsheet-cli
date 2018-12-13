package main

// https://docs.google.com/spreadsheets/d/1HD-8RW4YBKA1lmwaDveX5Hf-__6UTJ6p7LLoYjqIaho/edit#gid=0

// Spreadsheet ID:
// 1HD-8RW4YBKA1lmwaDveX5Hf-__6UTJ6p7LLoYjqIaho

import (
	"context"
	"fmt"
	"io/ioutil"
	"strconv"
	"strings"

	"github.com/juju/errors"
	"github.com/urfave/cli"
	"golang.org/x/oauth2/google"
	"gopkg.in/Iwark/spreadsheet.v2"
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

func cliMain(c *cli.Context) error {
	fmt.Printf("Hello %q", c.Args().Get(0))
	//fmt.Println("boom! I say!")

	os.Exit(0)

	// Need to connect spreadsheet to the service account
	spread, err := spread("1HD-8RW4YBKA1lmwaDveX5Hf-__6UTJ6p7LLoYjqIaho")
	if err != nil {
		panic(err)
	}

	if err != nil {
		panic(err)
	}

	segData, personas, err := segmentData(spread, func(s string) bool { return true })
	if err != nil {
		panic(err)
	}

	if err := setSegment(spread, segData, "segmented"); err != nil {
		panic(errors.ErrorStack(err))
	}

	if err := setPersonas(spread, personas); err != nil {
		panic(errors.ErrorStack(err))
	}

	kBenefits, err := keyBenefits(spread)
	if err != nil {
		panic(errors.ErrorStack(err))
	}

	keyBenData, err := keyBenefitData(spread, kBenefits)
	if err != nil {
		panic(errors.ErrorStack(err))
	}

	if err := setSegment(spread, keyBenData, "key_benefit_segmented"); err != nil {
		panic(errors.ErrorStack(err))
	}

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

func keyBenefitData(spread spreadsheet.Spreadsheet, keyBenefits []string) (map[string]int, error) {

	keyBenSegSheet, err := spread.SheetByTitle("key_benefit_segmented")
	if err != nil {
		return nil, errors.Trace(err)
	}

	responseSheet, err := spread.SheetByTitle("Form Responses 1")
	if err != nil {
		return nil, errors.Trace(err)
	}

	keyBenefitPersonas := make(map[string]bool)

	// clear benefits

	for i, _ := range keyBenSegSheet.Rows {
		if i == 0 {
			continue
		}

		keyBenSegSheet.Update(i, 3, "")
	}

	i := 1
	for _, benefit := range keyBenefits {
		keyBenSegSheet.Update(i, 3, benefit)
		i++

		for i, row := range responseSheet.Rows {
			if i == 0 {
				continue
			}

			// does the persona list the main benefit
			persona := row[7].Value
			print(row[3].Value)
			print(" : ")
			print(benefit)
			print(" : ")

			if !keyBenefitPersonas[persona] && strings.Contains(row[3].Value, benefit) {
				keyBenefitPersonas[persona] = true
			}
		}

		if err := keyBenSegSheet.Synchronize(); err != nil {
			return nil, errors.Trace(err)
		}

	}

	segData, _, err := segmentData(spread, func(persona string) bool {
		for keyBenPersona := range keyBenefitPersonas {
			if keyBenPersona == persona {
				return true
			}
		}
		return false
	})
	return segData, errors.Trace(err)
}

var sWords = []string{
	"",
	"make",
}

// remove the somewhats that don't have the popularWords
func keyBenefits(spread spreadsheet.Spreadsheet) ([]string, error) {

	benefitWordsSheet, err := spread.SheetByTitle("benefit_keywords")
	if err != nil {
		return nil, errors.Trace(err)
	}

	var topCount int

	wordsByCount := make(map[int][]string)

	for i, row := range benefitWordsSheet.Rows {
		if i == 0 {
			continue
		}

		count, err := strconv.Atoi(row[1].Value)
		if err != nil {
			return nil, errors.Trace(err)
		}

		if count > topCount {
			topCount = count
		}

		wordsByCount[count] = append(wordsByCount[count], row[0].Value)
	}

	return wordsByCount[topCount], nil
}

func segmentData(spread spreadsheet.Spreadsheet, filter func(string) bool) (map[string]int, map[string][]int, error) {

	responseSheet, err := spread.SheetByTitle("Form Responses 1")
	if err != nil {
		return nil, nil, errors.Trace(err)
	}

	// build segmented list
	segData := map[string]int{
		"very disappointed":     0,
		"somewhat disappointed": 0,
		"not disappointed":      0,
	}

	veryDisapointedPersonas := make(map[string]bool)

	// personas is a map of persona to a tuple of disappointment
	personas := make(map[string][]int)
	for i, row := range responseSheet.Rows {
		persona := row[7].Value

		if i == 0 || !filter(persona) {
			continue
		}

		disappointment := row[1].Value

		if len(personas[persona]) != 3 {
			personas[persona] = make([]int, 3)
		}

		switch disappointment {
		case "very disappointed":
			personas[persona][0]++
			veryDisapointedPersonas[persona] = true
		case "somewhat disappointed":
			personas[persona][1]++
		case "not disappointed":
			personas[persona][2]++
		default:
			return nil, nil, errors.Errorf("disappointment: %q, persona: %q", disappointment, persona)
		}

		if veryDisapointedPersonas[persona] {
			segData[disappointment]++
		}

	}

	return segData, personas, nil

}

func setSegment(spread spreadsheet.Spreadsheet, segData map[string]int, sheetTab string) error {

	segSheet, err := spread.SheetByTitle(sheetTab)
	if err != nil {
		return errors.Trace(err)
	}

	// clear existing segments
	tableLen := len(segSheet.Rows)
	for i := 1; i < tableLen; i++ {
		segSheet.Update(i, 0, "")
		segSheet.Update(i, 1, "")
	}

	// update segmented sheet
	i := 1
	for key, count := range segData {
		segSheet.Update(i, 0, key)
		segSheet.Update(i, 1, fmt.Sprintf("%d", count))
		i++
	}
	return errors.Trace(segSheet.Synchronize())
}

func setPersonas(spread spreadsheet.Spreadsheet, personas map[string][]int) error {
	personaSheet, err := spread.SheetByTitle("personas")
	if err != nil {
		return errors.Trace(err)
	}

	// clear existing personas
	tableLen := len(personaSheet.Rows)
	for i := 1; i < tableLen; i++ {
		personaSheet.Update(i, 0, "")
		personaSheet.Update(i, 1, "")
		personaSheet.Update(i, 2, "")
		personaSheet.Update(i, 3, "")
	}

	// update personas
	i := 1
	for persona, counts := range personas {
		personaSheet.Update(i, 0, persona)
		personaSheet.Update(i, 1, fmt.Sprintf("%d", counts[0]))
		personaSheet.Update(i, 2, fmt.Sprintf("%d", counts[1]))
		personaSheet.Update(i, 3, fmt.Sprintf("%d", counts[2]))
		i++
	}

	return errors.Trace(personaSheet.Synchronize())

}
