package app

import (
	"slices"
	"iter"
	"encoding/base64"
	"context"
	"fmt"
	"os"
	"errors"
	"database/sql"
	"path/filepath"
	"flag"
	schema "default-gen/sql"
	"default-gen/database"
	_ "github.com/tursodatabase/go-libsql"
)

const defaultConfigDir = ".go-default-gen"
const defaultDbName = "default-gen.db"

type commandHandler func(printUsage bool, args ...string) error

type App struct {
	configDirPath string
	dbUrl string
	db *sql.DB
	cmdHandlers map[string] commandHandler

	inputFile string
	programName string
	configName string
	outDir string
}

func createConfigDirIfNotExist(dir string) error {
	var dirNotFound bool

	_, err := os.Lstat(dir)

	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			dirNotFound = true
		} else {
			return err
		}
	}

	if !dirNotFound {
		return nil
	}

	return os.Mkdir(dir, 0777)
}

func initializeDb(dburl string) (*sql.DB, error) {
	db, err := sql.Open("libsql", dburl)
	
	if err != nil {
		return nil, err
	}

	ddl := schema.Get()

	_, err = db.Exec(ddl)

	if err != nil {
		return nil, err
	}

	return db, nil
}

func createDbUrl(configDir string, dbName string) string {
	return "file:" + filepath.Join(configDir, dbName)
}

func New() *App {
	return &App{
		cmdHandlers: make(map[string]commandHandler),
	}
}

func (a *App) Init() error {
	userHomeDir, err := os.UserHomeDir()

	if err != nil {
		return err
	}

	configDirPath := filepath.Join(userHomeDir, defaultConfigDir)

	if err := createConfigDirIfNotExist(configDirPath); err != nil {
		return err
	}

	dbUrl := createDbUrl(configDirPath, defaultDbName)

	db, err := initializeDb(dbUrl)

	if err != nil {
		return err
	}

	a.configDirPath = configDirPath
	a.dbUrl = dbUrl
	a.db = db

	a.registerHandler("add-config", addConfig(a))
	a.registerHandler("get-config", getConfigCmd(a))
	a.registerHandler("list", listConfigCmd(a))

	return nil
}

func (a *App) registerHandler(name string, h commandHandler) {
	a.cmdHandlers[name] = h
}

func sortedMap(m map[string]commandHandler) iter.Seq[commandHandler] {
	return func(yield func(h commandHandler) bool) {
		keys := make([]string, 0, len(m))

		for key := range m {
			keys = append(keys, key)
		}

		slices.Sort(keys)

		for _, key := range keys {
			if !yield(m[key]) {
				return
			}
		}
	}
}

func (a *App) Close() {
	a.db.Close()
}

func (a *App) printUsage() {
	for h := range sortedMap(a.cmdHandlers) {
		h(true)
		fmt.Println()
	}
}

func (a *App) Run() error {
	defer a.Close()

	args := os.Args

	if len(args) < 2 {
		fmt.Fprintln(os.Stdout, "incorrect number of flags")
		fmt.Println()
		a.printUsage()
		return nil
	}

	cmdName := args[1]

	handler, ok := a.cmdHandlers[cmdName]

	if !ok {
		a.printUsage()
		return nil
	}

	return handler(false, args[2:]...)

}

var ErrMissingInputFile = errors.New("missing input file")
var ErrMissingProgramName = errors.New("missing program name")
var ErrMissingConfigName = errors.New("missing config name")
var ErrInputFileDNE = errors.New("provided input file does not exist")

func addConfig(a *App) commandHandler {
	return func(printUsage bool, args ...string) error {
		addCmd := flag.NewFlagSet("add-config", flag.ExitOnError)
		addCmd.StringVar(&a.inputFile, "f", "", "config file to add to database")
		addCmd.StringVar(&a.configName, "n", "", "user defined name for the config")
		addCmd.StringVar(&a.programName, "p", "", "the name for the program that the config is associated with")

		if printUsage {
			addCmd.Usage()
			return nil
		}

		if len(args) < 3 {
			addCmd.Usage()
			return nil
		}

		if err := addCmd.Parse(args); err != nil {
			return err
		}

		if a.inputFile == "" {
			return ErrMissingInputFile
		}

		if a.programName == "" {
			return ErrMissingProgramName
		}

		if a.configName == "" {
			return ErrMissingConfigName
		}

		configPath, err  := filepath.Abs(a.inputFile)

		if err != nil {
			return err
		}

		content, err := encodeConfigFileContent(configPath)

		if err != nil {
			return err
		}

		queries := database.New(a.db)

		insertParams := database.InsertConfigDefaultsParams{
			Name: a.configName,
			Program: a.programName,
			FileName: a.inputFile,
			Content: content,
		}

		if err := queries.InsertConfigDefaults(context.Background(), insertParams); err != nil {
			return err
		}

		return nil
	}
}

func getConfigCmd(a *App) commandHandler {
	return func(printUsage bool, args ...string) error {
		getCmd := flag.NewFlagSet("get-config", flag.ExitOnError)
		getCmd.StringVar(&a.configName, "n", "", "the name of the config to retrieve")
		getCmd.StringVar(&a.outDir, "d", "", "directory to write the retrieved config")

		if printUsage {
			getCmd.Usage()
			return nil
		}

		if err := getCmd.Parse(args); err != nil {
			return err
		}

		if a.configName == "" {
			return ErrMissingConfigName
		}

		queries := database.New(a.db)

		row, err := queries.GetConfigDefaultByName(context.Background(), a.configName)

		if err != nil {
			return err
		}

		data, err := decodeConfigFileContent(row.Content)

		if err != nil {
			return err
		}

		if a.outDir != "" {
			return writeConfigToFile(a.outDir, row.FileName, data)
		}

		fmt.Fprintln(os.Stdout, string(data))

		return nil
	}
}

func writeConfigToFile(dir string, fileName string, content []byte) error {
	dir, err  := filepath.Abs(dir)

	if err != nil {
		return err
	}

	file, err := os.Create(filepath.Join(dir, fileName))

	if err != nil {
		return err
	}

	defer file.Close()

	file.Write(content)

	return nil
}

func listConfigCmd(a *App) commandHandler {
	return func(printUsage bool, args ...string) error {
		listCmd := flag.NewFlagSet("list", flag.ExitOnError)

		if printUsage {
			listCmd.Usage()
			return nil
		}

		if err := listCmd.Parse(args); err != nil {
			return err
		}

		queries := database.New(a.db)

		rows, err := queries.GetAllConfigs(context.Background())

		if err != nil {
			return err
		}

		for _, row := range rows {
			fmt.Fprintln(os.Stdout, row.Name)
		}

		return nil
	}
}

func encodeConfigFileContent(filePath string) (string, error) {
	data, err := os.ReadFile(filePath)

	if err != nil {
		return "", err
	}

	content := base64.StdEncoding.EncodeToString(data)

	return content, nil
}


func decodeConfigFileContent(content string) ([]byte, error) {
	return base64.StdEncoding.DecodeString(content)
}
