package utils

import (
	"database/sql"
	"fmt"
	"strconv"
	"time"
)

// SessionEvicter evicts sessions older than 14 days
func SessionEvicter(db *sql.DB) {
	for {
		res, err := db.Exec("DELETE FROM sessions WHERE modified < $1",
			time.Now().AddDate(0, 0, -14))
		if err != nil {
			fmt.Println("[ERROR] utils.SessionEvicter: " + err.Error())
		} else {
			numAffected, _ := res.RowsAffected()
			if numAffected > 0 {
				fmt.Println("[INFO] utils.SessionEvicter: " +
					strconv.FormatInt(numAffected, 10) + " sessions evicted")
			}
		}
		time.Sleep(time.Hour * 1)
	}
}
