package main

import (
	"flag"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

var (
	command    string
	parentDirs string
	isTest     bool
)

func init() {
	log.SetFlags(log.Lshortfile | log.LstdFlags)
	flag.StringVar(&command, "c", "", "The command that will be executed.")
	flag.StringVar(&parentDirs, "p", "", "The parent path of target directory. Note that multiple path needs to separated by commas ','.")
	flag.BoolVar(&isTest, "t", false, "Only test. (Do not execute the command)")
}

func checkBaseDir(basePath string) (targetPaths []string, err error) {
	log.Printf("Base Path: %s\n", basePath)
	baseDir, err := os.Open(basePath)
	if err != nil {
		log.Printf("OpenBaseDirError (%s): %s\n", baseDir, err)
		return nil, err
	}
	subDirs, err := baseDir.Readdir(-1)
	if err != nil {
		log.Println("ReaddirError:", err)
		return nil, err
	}
	for _, v := range subDirs {
		isTarget := false
		fileInfo := v.(os.FileInfo)
		subDirName := fileInfo.Name()
		absSubDirName, err := filepath.Abs(subDirName)
		if err != nil {
			log.Printf("AbsError (%s): %s\n", absSubDirName, err)
			return targetPaths, err
		}
		if fileInfo.IsDir() {
			subDir, err := os.Open(subDirName)
			if err != nil {
				log.Println("OpenSubDirError:", err)
				return targetPaths, err
			}
			names, err := subDir.Readdirnames(-1)
			if err != nil {
				log.Println("ReaddirnamesError:", err)
				return targetPaths, err
			}
			if len(names) == 0 {
				log.Printf("Note that Ignore EMPTY directory '%s'.\n", absSubDirName)
			} else {
				isTarget = true
			}
		}
		if isTarget {
			log.Printf("Target: %s\n", absSubDirName)
			targetPaths = append(targetPaths, absSubDirName)
		} else {
			log.Printf("Non-target: %s\n", absSubDirName)
		}
	}
	return targetPaths, err
}

func executeCommand(targetPath string, command string) error {
	log.Printf("Entry into target Path: %s\n", targetPath)
	err := os.Chdir(targetPath)
	if err != nil {
		log.Printf("ChdirError (%s): %s\n", targetPath, err)
		return err
	}
	cmdWithArgs := strings.Split(command, " ")
	var cmd *exec.Cmd
	cmdLength := len(cmdWithArgs)
	realCmd := cmdWithArgs[0]
	args := cmdWithArgs[1:cmdLength]
	log.Printf("Execute command (cmd=%s, args=%s)...\n", realCmd, args)
	if cmdLength > 1 {
		cmd = exec.Command(realCmd, args...)
	} else {
		cmd = exec.Command(realCmd)
	}
	result, err := cmd.Output()
	if err != nil {
		log.Printf("CmdRunError (cmd=%s, args=%v): %s\n", realCmd, args, err)
		return err
	}
	log.Printf("Output (dir=%s, cmd=%s, agrs=%v): \n%v\n", targetPath, realCmd, args, string(result))
	return nil
}

func printSegmentLine() {
	log.Println("------------------------------------------------------------")
}

func main() {
	flag.Parse()
	if isTest {
		log.Println("Starting... (in test enviroment)")
	} else {
		log.Println("Starting... (in formal enviroment)")
	}
	if len(command) == 0 {
		log.Println("The argument '-c' is NOT specified!")
		return
	}
	basePaths := make([]string, 0)
	if len(parentDirs) > 0 {
		basePaths = strings.Split(parentDirs, ",")
	} else {
		defaultBasePath, err := os.Getwd()
		if err != nil {
			log.Println("GetwdError:", err)
			return
		}
		basePaths = append(basePaths, defaultBasePath)
	}
	log.Printf("Parameters: \n  Command: %s\n  Base paths: %s\n  Test: %v\n", command, basePaths, isTest)
	targetPaths := make([]string, 0)
	for _, basePath := range basePaths {
		printSegmentLine()
		subTargetPaths, err := checkBaseDir(basePath)
		if err != nil {
			log.Println("CheckBaseDirError:", err)
			return
		}
		targetPaths = append(targetPaths, subTargetPaths...)
	}
	if !isTest {
		for _, targetPath := range targetPaths {
			printSegmentLine()
			executeCommand(targetPath, command)
		}
		printSegmentLine()
		log.Println("The command(s) execution has been finished.")
	}
}
