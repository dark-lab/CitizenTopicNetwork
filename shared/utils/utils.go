package utils

import (
	"encoding/csv"
	"fmt"
	"os"
)

func IntInSlice(a int64, list []int64) bool {
	for _, b := range list {
		if b == a {
			return true
		}
	}
	return false
}

func WriteCSV(records [][]string, file string) {
	// Create a csv file
	f, err := os.Create(file)
	if err != nil {
		fmt.Println(err)
	}
	defer f.Close()
	// Write Unmarshaled json data to CSV file
	w := csv.NewWriter(f)
	w.WriteAll(records)
	w.Flush()
}
