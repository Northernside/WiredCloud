package sqlite

import (
	"database/sql"
	"log"

	_ "github.com/mattn/go-sqlite3"
)

var db *sql.DB

func Init() {
	var err error
	db, err = sql.Open("sqlite3", "./wiredcloud.db")
	if err != nil {
		log.Fatal(err)
	}

	initTables()
}

func initTables() {
	_, err := db.Exec(`CREATE TABLE IF NOT EXISTS users (
		discord_id TEXT NOT NULL
	)`)
	if err != nil {
		log.Fatal(err)
	}
}

func CreateUser(discordId string) error {
	_, err := db.Exec("INSERT INTO users (discord_id) VALUES (?)", discordId)
	return err
}

// unsafe, dont allow key to be user input
func GetUser(key, value string) (string, error) {
	var discordId string
	err := db.QueryRow("SELECT discord_id FROM users WHERE "+key+" = ?", value).Scan(&discordId)
	return discordId, err
}

type DiscordUser struct {
	DiscordId string `json:"discord_id"`
}

func GetUsers() []DiscordUser {
	rows, err := db.Query("SELECT discord_id FROM users")
	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()

	users := []DiscordUser{}
	for rows.Next() {
		var discordId string
		err := rows.Scan(&discordId)
		if err != nil {
			log.Fatal(err)
		}

		users = append(users, DiscordUser{
			DiscordId: discordId,
		})
	}

	return users
}

func ChangeUserRole(discordId, role string) error {
	_, err := db.Exec("UPDATE users SET role = ? WHERE discord_id = ?", role, discordId)
	return err
}

func DeleteUser(key, value string) error {
	_, err := db.Exec("DELETE FROM users WHERE "+key+" = ?", value)
	return err
}

func Close() {
	db.Close()
}

func GetDB() *sql.DB {
	return db
}
