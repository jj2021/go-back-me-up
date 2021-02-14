package main

import (
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/spf13/viper"
)

type config struct {
	Loc     string
	Dir     []string
	Exclude []string
}

var configuration config
var printFileName filepath.WalkFunc = func(path string, info os.FileInfo, err error) error {
	if err != nil {
		fmt.Println(err)
		return err
	}
	// Skip any excluded directories in the config
	for i := 0; i < len(configuration.Exclude); i++ {
		if info.IsDir() && path == configuration.Exclude[i] {
			fmt.Println("Skipped " + info.Name())
			return filepath.SkipDir
		}
	}
	fmt.Println("Visited: " + path)
	if !info.IsDir() {
		_, err = copyfile(path)
		if err != nil {
			fmt.Println(err)
		}

	}
	return nil
}

func main() {
	viper.AddConfigPath(".")
	viper.SetConfigFile(".backmeup.yml")
	err := viper.ReadInConfig()

	args := os.Args
	fmt.Println(args)
	if len(args) > 1 {
		switch args[1] {
		case "loc":
			setLoc(args)
		case "dir":
			dir(args)
		case "exclude":
			exclude(args)
		case "config":
			showConfig()
		default:
			fmt.Println(args[1], " is not a command")
		}
	}
	//temporarily return for testing configuration commands
	return
	// Read in backup settings from configuration file
	if err != nil {
		fmt.Println(err)
	}
	viper.Unmarshal(&configuration)

	// Loop over directories and back them up
	for i := 0; i < len(configuration.Dir); i++ {
		err := filepath.Walk(configuration.Dir[i], printFileName)
		if err != nil {
			fmt.Printf("Error walking path: %v \n", err)
			return
		}
	}
}

func setLoc(args []string) {
	if len(args) > 2 {
		viper.Set("loc", args[2])
		viper.WriteConfig()
		fmt.Println(viper.Get("loc"))
	} else {
		fmt.Println("Usage: backmeup loc <backup path>")
	}
}

func dir(args []string) {
	if len(args) > 2 {
		switch args[2] {
		case "add":
			addDir(args)
		default:
			fmt.Println(args[2], " is not a command.")
		}
	} else {
		fmt.Println("Usage: backmeup dir <command>")
		fmt.Println("Commands:")
		fmt.Println("\t", "add: add directory to list of backed up directories")
		fmt.Println("\t", "rm: remove a directory from the list of backed up directories")
	}
}

func exclude(args []string) {
	if len(args) > 2 {
		switch args[2] {
		case "add":
			addExclusion(args)
		default:
			fmt.Println(args[2], " is not a command.")
		}
	} else {
		fmt.Println("Usage: backmeup exclude <command>")
		fmt.Println("Commands:")
		fmt.Println("\t", "add: add directory to list of excluded directories")
		fmt.Println("\t", "rm: remove a directory from the list of excluded directories")
	}

}

func addDir(args []string) {
	if len(args) > 3 {
		directories := append(viper.GetStringSlice("dir"), args[3:]...)
		viper.Set("dir", directories)
		viper.WriteConfig()
		fmt.Println("Adding Dir ", viper.Get("dir"))
	} else {
		fmt.Println("Usage: backmeup dir add <path>")
	}
}

func addExclusion(args []string) {
	if len(args) > 3 {
		exclusions := append(viper.GetStringSlice("exclude"), args[3:]...)
		viper.Set("exclude", exclusions)
		viper.WriteConfig()
		fmt.Println("Adding Exclusion", viper.Get("exclude"))
	} else {
		fmt.Println("Usage: backmeup exclude add <path>")
	}
}

func showConfig() {
	fmt.Println("Backup Location:")
	fmt.Println("\t", viper.Get("loc"))
	fmt.Println("Backed up directories:")
	for _, item := range viper.GetStringSlice("dir") {
		fmt.Println("\t", item)
	}
	fmt.Println("Excluded directories:")
	for _, item := range viper.GetStringSlice("exclude") {
		fmt.Println("\t", item)
	}
}

func copyfile(src string) (int64, error) {
	var err error
	var bytes int64
	// Get relative path from C: drive
	rel, err := filepath.Rel("C:\\", src)
	if err != nil {
		return bytes, err
	}
	fmt.Println("Relative path: " + rel)
	// build the destination path from config loc property and relative path
	dest := filepath.Join(configuration.Loc, rel)
	fmt.Println("Destination: " + dest)
	// check that the file exists and is a regular file
	srcStat, err := os.Stat(src)
	if err != nil {
		return 0, err
	}
	if !srcStat.Mode().IsRegular() {
		return 0, fmt.Errorf("%s is not a regular file", src)
	}

	// open the source file to be copied
	source, err := os.Open(src)
	if err != nil {
		return 0, err
	}
	defer source.Close()

	// copy the directory structure to backup location
	destDir := filepath.Dir(dest)
	var perm os.FileMode = 0777
	err = os.MkdirAll(destDir, perm)
	if err != nil {
		return 0, err
	}

	// create the destination file
	destination, err := os.Create(dest)
	if err != nil {
		return 0, err
	}
	defer destination.Close()

	bytes, err = io.Copy(destination, source)

	return bytes, err
}
