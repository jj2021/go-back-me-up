package main

import (
	"flag"
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
var verbose *bool
var preserve *bool

var printFileName filepath.WalkFunc = func(path string, info os.FileInfo, err error) error {
	if err != nil {
		fmt.Println(err)
		return err
	}
	// Skip any excluded directories in the config
	for i := 0; i < len(configuration.Exclude); i++ {
		if info.IsDir() && path == configuration.Exclude[i] {
			if *verbose {
				fmt.Println("Skipped " + info.Name())
			}
			return filepath.SkipDir
		}
	}
	if *verbose {
		fmt.Println("Backing up: " + path)
	}
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

	verbose = flag.Bool("v", false, "Give verbose output")
	preserve = flag.Bool("p", false, "Preserve file attributes")
	flag.Parse()
	args := flag.Args()
	// Check for and execute configuration commands, otherwise execute the backup
	if len(args) > 0 {
		switch args[0] {
		case "loc":
			setLoc(args)
		case "dir":
			dir(args)
		case "exclude":
			exclude(args)
		case "config":
			showConfig()
		case "restore":
			fmt.Println("Restore not yet implemented")
		default:
			fmt.Println(args[0], " is not a command")
		}
	} else {
		// Loop over directories and back them up
		for i := 0; i < len(configuration.Dir); i++ {
			err := filepath.Walk(configuration.Dir[i], printFileName)
			if err != nil {
				fmt.Printf("Error walking path: %v \n", err)
				return
			}
		}
	}
}

func setLoc(args []string) {
	if len(args) > 1 {
		viper.Set("loc", args[1])
		viper.WriteConfig()
		fmt.Println(viper.Get("loc"))
	} else {
		fmt.Println("Usage: backmeup loc <backup path>")
	}
}

func dir(args []string) {
	if len(args) > 1 {
		switch args[1] {
		case "add":
			addDir(args)
		case "rm":
			removeDir(args)
		default:
			fmt.Println(args[1], " is not a command.")
		}
	} else {
		fmt.Println("Usage: backmeup dir <command>")
		fmt.Println("Commands:")
		fmt.Println("\t", "add: add directory to list of backed up directories")
		fmt.Println("\t", "rm: remove a directory from the list of backed up directories")
	}
}

func exclude(args []string) {
	if len(args) > 1 {
		switch args[1] {
		case "add":
			addExclusion(args)
		case "rm":
			removeExclusion(args)
		default:
			fmt.Println(args[1], " is not a command.")
		}
	} else {
		fmt.Println("Usage: backmeup exclude <command>")
		fmt.Println("Commands:")
		fmt.Println("\t", "add: add directory to list of excluded directories")
		fmt.Println("\t", "rm: remove a directory from the list of excluded directories")
	}

}

func addDir(args []string) {
	if len(args) > 2 {
		directories := append(viper.GetStringSlice("dir"), args[2:]...)
		viper.Set("dir", directories)
		viper.WriteConfig()
		fmt.Println("Adding Dir ", viper.Get("dir"))
	} else {
		fmt.Println("Usage: backmeup dir add <path>")
	}
}

func removeDir(args []string) {
	if len(args) > 2 {
		var directories []string
		removed := false
		for _, item := range viper.GetStringSlice("dir") {
			if item != args[2] {
				directories = append(directories, item)
			} else {
				removed = true
			}
		}
		viper.Set("dir", directories)
		viper.WriteConfig()
		if !removed {
			fmt.Println(args[2], "not removed, not in directory list")
		}
	} else {
		fmt.Println("Usage: backmeup dir rm <path>")
	}
}

func addExclusion(args []string) {
	if len(args) > 2 {
		exclusions := append(viper.GetStringSlice("exclude"), args[2:]...)
		viper.Set("exclude", exclusions)
		viper.WriteConfig()
		fmt.Println("Adding Exclusion", viper.Get("exclude"))
	} else {
		fmt.Println("Usage: backmeup exclude add <path>")
	}
}

func removeExclusion(args []string) {
	if len(args) > 2 {
		var exclusions []string
		removed := false
		for _, item := range viper.GetStringSlice("exclude") {
			if item != args[2] {
				exclusions = append(exclusions, item)
			} else {
				removed = true
			}
		}
		viper.Set("exclude", exclusions)
		viper.WriteConfig()
		if !removed {
			fmt.Println(args[2], "not removed, not in exclusion list")
		}
	} else {
		fmt.Println("Usage: backmeup exclude rm <path>")
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
	// Get relative path from drive (for windows devices)
	drive := filepath.VolumeName(src)
	//rel, err := filepath.Rel("C:\\", src)
	rel, err := filepath.Rel(drive+"\\", src)
	if err != nil {
		return bytes, err
	}
	// build the destination path from config loc property and relative path
	dest := filepath.Join(configuration.Loc, rel)
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

	bytes, err = io.Copy(destination, source)

	// close file before operating on file attributes
	// (if deferred, the chtimes function will not behave as expected)
	destination.Close()

	// copy file attributes if specified
	if *preserve {
		err := os.Chtimes(destination.Name(), srcStat.ModTime(), srcStat.ModTime())
		if err != nil {
			fmt.Printf("Error preserving time info for %s: %v\n", source.Name(), err.Error())
		}

		err = os.Chmod(destination.Name(), srcStat.Mode())
		if err != nil {
			fmt.Printf("Error preserving file mode for %s: %v\n", source.Name(), err.Error())
		}
	}

	return bytes, err
}
