package utils

import (
	"bufio"
	"encoding/csv"
	"io"
	"log"
	"os"
)

type Records [][]string

func WriteCSVFile(filename string, columns []string, records Records) {
	f, err := os.Create(filename)
	if err != nil {
		log.Fatal("Cannot create file ", filename)
	}
	defer f.Close()
	writer := bufio.NewWriter(f)
	WriteCSV(writer, columns, records)
	_ = writer.Flush()
}

func WriteCSV(writer io.Writer, columns []string, records Records) {

	w := csv.NewWriter(writer)
	if err := w.Write(columns); err != nil {
		log.Fatalln("error writing column headers to csv:", err)
	}

	for _, record := range records {
		if err := w.Write(record); err != nil {
			log.Fatalln("error writing record to csv:", record, err)
		}
	}
	w.Flush()
	if err := w.Error(); err != nil {
		log.Fatal(err)
	}

}
