package main

import (
	"os"
	"path/filepath"
	"encoding/base64"
	"errors"
	"log"
	"flag"
	"database/sql"
	_ "github.com/tursodatabase/go-libsql"
	_ "embed"
)

//go:embed schema.sql
var ddl string

const configDirName = ".go-default-gen"

func configDir() string {
	dir, _ := os.UserHomeDir()
	return dir
}

type configFile struct {
	data []byte
	program string
	name string
	ext string
}

type config struct {
	files []configFile
	db *sql.DB
}

func (c *config) Close() {
	c.db.Close()
}

func initializeDB(configDir string) (*sql.DB, error) {
	db, err := sql.Open("libsql", "file:" + filepath.Join(configDir, "default-gen.db"))

	if err != nil {
		return nil, err
	}

	_, err = db.Exec(ddl)

	if err != nil {
		return nil, err
	}

	return db, nil
}

func main() {
	dir := configDir()

	path := filepath.Join(dir, configDirName)

	var dirNotFound bool

	_, err := os.Lstat(filepath.Join(dir, configDirName))

	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			dirNotFound = true
		} else {
			log.Fatal(err)
		}
	}

	if dirNotFound {
		if err := os.Mkdir(path, 0777); err != nil {
			log.Fatal(err)
		}
	}

	db, err := initializeDB(path)

	if err != nil {
		log.Fatal(err)
	}

	defer db.Close()

	var srcPath string
	var name string
	var program string

	addCmd := flag.NewFlagSet("add-config", flag.ExitOnError)
	addCmd.StringVar(&srcPath, "f", "", "path of config file to add")
	addCmd.StringVar(&name, "n", "", "user defined name for the config file")
	addCmd.StringVar(&program, "p", "", "name of the program that the config file is associated with")
	getCmd := flag.NewFlagSet("get-config", flag.ExitOnError)
	getCmd.StringVar(&name, "n", "", "user defined name for the config file")

	if len(os.Args) < 2 {
		flag.Usage()
		addCmd.Usage()
		getCmd.Usage()
		return
	}

	switch os.Args[1] {
	case "add-config":
		addCmd.Parse(os.Args[2:])
		if err := runAddCmd(db, srcPath, name, program); err != nil {
			log.Panic(err)
		}
	case "get-config":
		getCmd.Parse(os.Args[2:])
		if err := runGetCmd(db, name); err != nil {
			log.Panic(err)
		}
	}

}

const configInsertSqlStr = `
	INSERT INTO config_defaults
	(name, file_name, program, content)
	VALUES
	(?, ?, ?, ?)
`

type configInsertParams struct {
	name string
	fileName string
	program string
	content string
}

const getConfigByName = `
	SELECT content, name FROM config_defaults WHERE name = ?;
`

type configResult struct {
	content string
	name string
}

func getConfig(db *sql.DB, name string) (*configResult, error) {
	row := db.QueryRow(getConfigByName, name)

	c := configResult{}

	err := row.Scan(&c.content, &c.name)

	if err != nil {
		return nil, err
	}

	return &c, nil
}


func runAddCmd(db *sql.DB, srcPath string, name string, program string) error {
	if srcPath == "" {
		log.Panic("must provide a valid path to src config")
	}

	fullpath, err := filepath.Abs(srcPath)

	if err != nil {
		return err
	}

	data, err := os.ReadFile(fullpath)

	if err != nil {
		return err
	}

	encData := base64.StdEncoding.EncodeToString(data)

	params := &configInsertParams{
		program: program,
		name: name,
		content: encData,
	}

	_, err = db.Exec(
		configInsertSqlStr,
		params.name,
		params.fileName,
		params.program,
		params.content,
	)

	if err != nil {
		return err
	}

	return nil
}

func runGetCmd(db *sql.DB, name string) error {
	if name == "" {
		log.Panic("incorrect name. name cannot be empty")
	}

	res, err := getConfig(db, name)

	if err != nil {
		return err
	}

	data, err := base64.StdEncoding.DecodeString(res.content)

	if err != nil {
		return err
	}

	log.Println(string(data))

	return nil
}
