package main

import (
	"context"
	"fmt"
	Apriori "github.com/eMAGTechLabs/go-apriori"
	"github.com/jackc/pgx/v4"
	"github.com/manifoldco/promptui"
	"os"
	"strings"
)

func main() {
	conn, err := pgx.Connect(context.Background(), "postgres://postgres:oQyuEHiD8FF5DLJ1HDTwIWyhm9ie6csoIEvsEZz8GXhXE48RcS@zottelchin.de:8877/Masterarbeit")
	if err != nil {
		fmt.Printf("DB-Connect Error: %s\n", err.Error())
		os.Exit(1)
	}
	defer conn.Close(context.Background())

	tables := []string{}
	rows, _ := conn.Query(context.Background(), "select table_name from information_schema.tables where table_schema='public' and table_name not like '%_rules'")
	defer rows.Close()
	for rows.Next() {
		name := ""
		rows.Scan(&name)
		tables = append(tables, name)
	}

	promt := promptui.Select{Label: "Select Databasetable", Items: append(tables, "all", "exit"), Size: 10}
	_, res, _ := promt.Run()

	if res == "exit" {
		os.Exit(0)
	}

	if res == "all" {
		for _, t := range tables {
			calcApriori(t, conn)
			fmt.Println("")
		}
		os.Exit(0)
	}

	calcApriori(res, conn)
}

func calcApriori(table string, db *pgx.Conn) {
	fmt.Println("=================================================")
	fmt.Printf("=== %s ===\n", table)
	fmt.Println("=================================================")

	rows, err := db.Query(context.Background(), fmt.Sprintf("select hash, constants from %s", table))
	defer rows.Close()
	if err != nil {
		fmt.Println(err)
	}

	byCommit := map[string][]string{}
	for rows.Next() {
		var c, hash string
		rows.Scan(&hash, &c)
		consts := strings.Split(c, ",")
		byCommit[hash] = append(byCommit[hash], consts...)
	}

	if rows.Err() != nil {
		fmt.Println(rows.Err())
	}

	fmt.Printf("Number of Commits in DB: %d\n", len(byCommit))

	filtered := [][]string{}

	for hash, consts := range byCommit {
		byCommit[hash] = removeDuplicateStr(consts)
		filtered = append(filtered, byCommit[hash])
	}

	mConf := 0.85
	mSup := 15.0 / float64(len(byCommit))

	fmt.Printf("To be included, the features of the rule have to occure in %.0f of the commits = minSupoort: %.3f%%\n\n", float64(len(byCommit))*mSup, mSup)

	apri := Apriori.NewApriori(filtered)
	results := apri.Calculate(Apriori.NewOptions(mSup, mConf, 0.0, 0))

	pprintr(results)
	upladeRules(results, table, db)
	if len(results) == 0 {
		fmt.Println("No rule found in data")
	}
	fmt.Println("=================================================")
}

// from https://stackoverflow.com/questions/66643946/how-to-remove-duplicates-strings-or-int-from-slice-in-go
func removeDuplicateStr(strSlice []string) []string {
	allKeys := make(map[string]bool)
	list := []string{}
	for _, item := range strSlice {
		if _, value := allKeys[item]; !value {
			allKeys[item] = true
			list = append(list, item)
		}
	}
	return list
}

func pprintr(res []Apriori.RelationRecord) {
	for i, rr := range res {
		fmt.Printf("%+v\n", rr.GetSupportRecord())
		fmt.Printf("orderd Statistics\n")
		for _, stats := range rr.GetOrderedStatistic() {
			fmt.Printf("\t%+v\n", stats)
		}
		if i < len(res)-1 {
			fmt.Printf("-----------------\n")
		}
	}
}

func upladeRules(rules []Apriori.RelationRecord, table string, db *pgx.Conn) {
	for _, r := range rules {
		for _, stats := range r.GetOrderedStatistic() {
			db.Exec(context.Background(),
				fmt.Sprintf("insert into %s_rules (base, add, confidence, support) values ($1, $2, $3, $4)", table),
				strings.Join(stats.GetBase(), ";"),
				strings.Join(stats.GetAdd(), ";"),
				stats.GetConfidence(),
				r.GetSupportRecord().GetSupport())
		}
	}
}
