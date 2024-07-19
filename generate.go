package main

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"

	_ "github.com/go-sql-driver/mysql"
	"github.com/joho/godotenv"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal(".env file is not present: ", err)
	}

	items, auswahl := Items()

	if auswahl == "y" || auswahl == "Y" || auswahl == "j" || auswahl == "J" {
		// Check if items are in Database
		for i := range items {
			check, err := CheckItem(items[i])
			if err != nil {
				fmt.Print(err)
				log.Fatalf(
					"Die Artikelnummer %s ist nicht in der Datenbank, bitte Datenbank aktualisieren oder Nummer prüfen!",
					items[i],
				)
			}
			if !check {
				log.Fatal(err)
			}
		}

		// Check if the array is too long. Maximum: 4 Items
		if len(items) > 4 {
			log.Fatal("Es können nur maximal 4 Artikelnummern genutzt werden, bitte anpassen!")
		}

		// Generate Batch File
		Artikelnummern := strings.Join(items, ",")

		installer, err := os.Create("install.bat")
		if err != nil {
			log.Fatal(err)
		}
		defer installer.Close()
		bat := fmt.Sprintf(
			"echo Powershell.exe -ExecutionPolicy Bypass -Command \"& 'C:\\Program Files (x86)\\Microsoft\\Edge\\Application\\msedge.exe' --kiosk aussteller.computer-extra.de/%s --edge-kios-type=fullscreen\" > \"%%USERPROFILE%%\\AppData\\Roaming\\Microsoft\\Windows\\Start Menu\\Programs\\Startup\\Aussteller.bat \"",
			Artikelnummern,
		)
		_, err = installer.WriteString(bat)
		if err != nil {
			log.Fatal(err)
		}

		// file, err := os.Create("Aussteller.ps1")
		// if err != nil {
		// 	log.Fatal(err)
		// }
		// defer file.Close()
		// url := fmt.Sprintf(
		// 	"& 'C:\\Program Files (x86)\\Microsoft\\Edge\\Application\\msedge.exe' --kiosk aussteller.computer-extra.de/%s --edge-kios-type=fullscreen",
		// 	Artikelnummern,
		// )
		// _, err = file.WriteString(url + "\n")
		// if err != nil {
		// 	log.Fatal(err)
		// }
		// file2, err := os.Create("Aussteller.bat")
		// defer file2.Close()
		// _, err = file2.WriteString(
		// 	"PowerShell.exe -ExecutionPolicy Bypass -Command \"& '%~dpn0.ps1'\"" + "\n",
		// )
		// if err != nil {
		// 	log.Fatal(err)
		// }

	} else {
		fmt.Println("Die Artikelnummern scheinen nicht zu stimmen. bitte neu starten!")
		fmt.Scan()
	}
}

func Items() ([]string, string) {
	var items string

	fmt.Println(
		"Bitte die Artikelnummern, mit Komma (,) getrennt eingeben und mit Enter bestätigen:",
	)
	fmt.Scan(&items)

	fmt.Println("Folgende Artikelnummern wurden eingegeben, passt das? y | n")
	Artikelnummern := strings.Split(items, ",")
	for i := range Artikelnummern {
		fmt.Println(Artikelnummern[i])
	}
	var auswahl string
	fmt.Scan(&auswahl)
	return strings.Split(items, ","), auswahl
}

func CheckItem(Artikelnummer string) (bool, error) {
	query := fmt.Sprintf("SELECT id FROM Aussteller WHERE Artikelnummer='%s'", Artikelnummer)

	port, err := strconv.ParseInt(os.Getenv("MYSQL_PORT"), 0, 64)
	if err != nil {
		return false, fmt.Errorf("SAGE_PORT not in .env: %s", err)
	}

	connstring := fmt.Sprintf(
		"%s:%s@tcp(%s:%d)/%s",
		os.Getenv("MYSQL_USER"),
		os.Getenv("MYSQL_PASS"),
		os.Getenv("MYSQL_SERVER"),
		port,
		os.Getenv("MYSQL_DB"),
	)
	conn, err := sql.Open("mysql", connstring)
	if err != nil {
		return false, fmt.Errorf("cannot connect to Database: %s", err)
	}

	rows, err := conn.Query(query)
	if err != nil {
		return false, fmt.Errorf("error while querying the db: %s", err)
	}

	for rows.Next() {
		var id sql.NullInt64
		if err := rows.Scan(&id); err != nil {
			return false, fmt.Errorf("error while scanning the row: %s", err)
		}
		if id.Valid {
			return true, nil
		} else {
			return false, fmt.Errorf("id is not valid")
		}
	}

	return false, fmt.Errorf("keine Rows gefunden?")
}
