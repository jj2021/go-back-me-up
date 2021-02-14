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
	// Read in backup settings from configuration file
	viper.AddConfigPath(".")
	viper.SetConfigFile(".backmeup.yml")
	err := viper.ReadInConfig()
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
