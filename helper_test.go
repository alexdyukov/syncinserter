package syncinserter_test

import (
	"strconv"
	"strings"
)

func insertString(rowsCount int) string {
	res := &strings.Builder{}
	res.WriteString(`INSERT INTO test (created_at, usr, diff) VALUES ($1, $2, $3)`)

	for index := 1; index < rowsCount; index++ {
		res.WriteString(", ($")
		res.WriteString(strconv.Itoa(index*3 + 1))
		res.WriteString(", $")
		res.WriteString(strconv.Itoa(index*3 + 2))
		res.WriteString(", $")
		res.WriteString(strconv.Itoa(index*3 + 3))
		res.WriteString(")")
	}

	res.WriteRune(';')

	return res.String()
}
