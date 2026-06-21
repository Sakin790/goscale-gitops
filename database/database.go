package database

import (
	"fmt"
	"log"
	"os"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/jmoiron/sqlx"
)

var DB *sqlx.DB

// 10,000 capacity-r buffer channel
var LogQueue = make(chan DbLog, 10000)

type DbLog struct {
	Timestamp     time.Time `db:"timestamp"`
	Method        string    `db:"method"`
	Proto         string    `db:"protocol"`
	Path          string    `db:"path"`
	RemoteAddress string    `db:"remote_address"`
}

func InitDB() {
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		dsn = "postgres://sakin:sakin123@localhost:5432/logs?sslmode=disable"
	}

	var err error
	DB, err = sqlx.Connect("pgx", dsn)
	if err != nil {
		log.Fatalf("[❌ DB ERROR] Connection failed: %v\n", err)
	}

	DB.SetMaxOpenConns(25)
	DB.SetMaxIdleConns(5)
	fmt.Println("✅ [DATABASE] Connected successfully via sqlx!")
}

// StartBufferLogWorker: 10 second and 10 items configuration for instant test
func StartBufferLogWorker() {
	go func() {
		// 🕒 [UPDATED] 1 minute-er jaygay 10 second kora holo
		ticker := time.NewTicker(10 * time.Second)
		var batch []DbLog

		for {
			select {
			case logItem := <-LogQueue:
				batch = append(batch, logItem)

				// 🚀 [UPDATED] 10 ta data hoye gele sathe sathe DB-te pathiye debe
				if len(batch) >= 10 {
					flushLogs(batch)
					batch = nil
				}

			case <-ticker.C:
				// Protiti 10 second por por auto-trigger hobe
				if len(batch) > 0 {
					flushLogs(batch)
					batch = nil
				}
			}
		}
	}()
}

func flushLogs(logs []DbLog) {
	query := `INSERT INTO request_logs (timestamp, method, protocol, path, remote_address) 
              VALUES (:timestamp, :method, :protocol, :path, :remote_address)`

	_, err := DB.NamedExec(query, logs)
	if err != nil {
		log.Printf("[❌ DB BATCH ERROR] Failed to insert logs: %v\n", err)
		return
	}
	fmt.Printf("[🚀 DB BATCH SAVED] Successfully inserted %d logs to PostgreSQL!\n", len(logs))
}
