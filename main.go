package main

import (
	"bufio"
	"database/sql"
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	"os"
)

const (
	username = "root"
	password = "root"
	ip = "localhost"
	port = "3306"
	db_name = "trial"
)

func checkErr(err error) {
	if err != nil {
		panic(err)
	}
}

func getCommand(reader *bufio.Reader) []byte {
	cmd, err := reader.ReadBytes('\n')
	checkErr(err)
	return cmd[0:len(cmd) - 1]
}

func commandLinePrompt() {
	fmt.Print("-> ")
}

func getInput(reader *bufio.Reader, prompt string) []byte {
	fmt.Println(prompt)
	commandLinePrompt()
	return getCommand(reader)
}

func getValidUserId(reader *bufio.Reader, db *sql.DB) []byte {
	userId := getInput(reader, "Enter the user id you want to delete:")
	rows, _ := db.Query("SELECT * FROM trial.users WHERE user_id = ?", userId)
	targetRow := rows.Next()
	for !targetRow {
		userId = getInput(reader, "Please enter a valid user id:")
		rows, _ := db.Query("SELECT * FROM trial.users WHERE user_id = ?", userId)
		targetRow = rows.Next()
	}

	return userId
}

func isDBEmpty(db *sql.DB) bool {
	var count int
	rows, _ := db.Query("SELECT COUNT(*) FROM trial.users")
	rows.Next()
	if err := rows.Scan(&count); err != nil {
		panic(err)
	}
	if count == 0 {
		fmt.Println("The table is empty")
		return true
	}

	return false
}

func main() {
	path := username + ":" + password + "@tcp(" + ip + ")/" + db_name + "?charset=utf8"

	db, err := sql.Open("mysql", path)
	checkErr(err)

	createTableQuery := "CREATE TABLE IF NOT EXISTS users(user_id int primary key auto_increment, username text, " +
		"password text, created_at datetime default CURRENT_TIMESTAMP)"
	//ctx, cancelFuc := context.WithTimeout(context.Background(), 5 * time.Second)
	//defer cancelFuc()
	//_, err = db.ExecContext(ctx, createTableQuery)

	_, err = db.Exec(createTableQuery)
	checkErr(err)

	reader := bufio.NewReader(os.Stdin)

	var cmdMap map[byte]bool
	cmdMap = make(map[byte]bool)

	cmdMap['q'] = true
	cmdMap['1'] = true
	cmdMap['2'] = true
	cmdMap['3'] = true
	cmdMap['4'] = true

	re:
	for {
		fmt.Println("Select the number for a database operation: ")
		fmt.Println("1. Print all users")
		fmt.Println("2. Add a user")
		fmt.Println("3. Delete a user")
		fmt.Println("4. Update user information")
		fmt.Println("Press q to exit the program")
		commandLinePrompt()
		cmd := getCommand(reader)

		for {
			if _, ok := cmdMap[cmd[0]]; ok && len(cmd) == 1 {
				break
			} else {
				fmt.Println("Please enter a valid option")
				commandLinePrompt()
				cmd = getCommand(reader)
			}
		}

		switch cmd[0] {
		case '1': {
			if isDBEmpty(db) {
				break
			}

			rows, _ := db.Query("SELECT * FROM trial.users")
			fmt.Printf("%-10s%-15s%-15s\n", "id", "username", "password")
			for rows.Next() {
				var userId int
				var username, password, timeStamp string
				if err := rows.Scan(&userId, &username, &password, &timeStamp); err != nil {
					checkErr(err)
				}

				fmt.Printf("%-10d%-15s%-15s\n", userId, username, password)
			}
		}
		case '2': {
			username := getInput(reader, "Enter your username:")
			password := getInput(reader, "Enter your password:")
			insertStmt, err := db.Prepare("INSERT INTO trial.users (`username`, `password`) values (?, ?)")
			if err == nil {
				_, err := insertStmt.Exec(username, password)
				if err ==nil {
					fmt.Println("User added")
				} else {
					panic(err)
				}
			} else {
				panic(err)
			}
		}
		case '3': {
			if isDBEmpty(db) {
				break
			}

			userId := getValidUserId(reader, db)

			if _, err = db.Query("DELETE FROM trial.users WHERE user_id = ?", userId); err != nil {
				panic(err)
			}
			fmt.Println("User deleted")
		}
		case '4': {
			if isDBEmpty(db) {
				break
			}

			userId := getValidUserId(reader, db)
			username := getInput(reader, "Enter your username:")
			password := getInput(reader, "Enter your password:")
			if _, err := db.Query("UPDATE trial.users SET username = ?, password = ? WHERE user_id = ?",
				username, password, userId); err != nil {
				panic(err)
			}
			fmt.Println("Updated successfully")
		}
		default: {
			fmt.Println("Bye")
			break re
		}
		}
	}
}
