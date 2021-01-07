package main

import (
	"example.com/functions"
	"encoding/json"
	"encoding/xml"
	"io/ioutil"
	"os"
	"time"
)

type XmlExport struct {
	XMLName xml.Name `xml:"NewDataSet"`
	Records []Record `xml:"Table"`
}

type Record struct {
	XMLName xml.Name `xml:"Table"`
	Id      int64    `xml:"comment_id"`
	Name    string   `xml:"name"`
	Email   string   `xml:"email_address"`
	Comment string   `xml:"comment"`
	Date    string   `xml:"date_posted"`
	Status  string   `xml:"status"`
}

func convert(xmlFile string) {
	comments, err := loadFromXml(xmlFile)
	if err != nil {
		panic(err)
	}
	err = writeCommentsJson(comments)
	if err != nil {
		panic(err)
	}
}

func loadFromXml(xmlFile string) ([]functions.Comment, error) {
	loc, err := time.LoadLocation("America/New_York")
	if err != nil {
		panic(err)
	}

	bytes, err := ioutil.ReadFile(xmlFile)
	if err != nil {
		panic(err)
	}

	var xmlExport XmlExport
	// we unmarshal our byteArray which contains our
	// xmlFiles content into 'users' which we defined above
	err = xml.Unmarshal(bytes, &xmlExport)
	if err != nil {
		panic(err)
	}

	comments := make([]functions.Comment, 0)
	for _, record := range xmlExport.Records {
		if record.Status != "approved" {
			continue
		}
		timestamp, err := time.ParseInLocation("1/2/2006 3:04:05 PM", record.Date, loc)
		if err != nil {
			panic("time parse failed")
		}
		//outStr := timestamp.String()
		//fmt.Println(outStr)

		comment := functions.Comment{
			Name:      record.Name,
			Email:     record.Email,
			Msg:       record.Comment,
			Timestamp: timestamp,
		}
		comments = append(comments, comment)
	}

	return comments, nil
}

func writeCommentsJson(comments []functions.Comment) error {

	delim := []byte{0x0A}
	f, err := os.Create("comments.dat")
	if err != nil {
		// TODO
	}
	defer f.Close()

	for _, comment := range comments {
		bytes, err := json.Marshal(&comment)
		if err != nil {
			// TODO
		}
		f.Write(bytes)
		f.Write(delim)
	}
	return nil
}
