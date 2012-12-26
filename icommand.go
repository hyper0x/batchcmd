package main

import (
	"flag"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
)

var (
	command    string
	parentDirs string
)

func init() {
	log.SetFlags(log.Lshortfile | log.LstdFlags)
	flag.StringVar(&command, "c", "", "The command that will be executed.")
	flag.StringVar(&parentDirs, "p", "", "The parent path of target directory. Note that multiple path needs to separated by commas ','.")
}

func checkBaseDir(basePath string) (targetPaths []string, err error) {
	printSegmentLine()
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
			if sort.SearchStrings(names, ".git") >= 0 {
				isTarget = true
			}
		}
		targetPath, err := filepath.Abs(subDirName)
		if err != nil {
			log.Printf("AbsError (%s): %s\n", targetPath, err)
			return targetPaths, err
		}
		if isTarget {
			log.Printf("Target: %s\n", targetPath)
			targetPaths = append(targetPaths, targetPath)
		} else {
			log.Printf("Non-target: %s\n", targetPath)
		}
	}
	printSegmentLine()
	return targetPaths, err
}

func executeCommand(targetPath string, command string) error {
	printSegmentLine()
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
	printSegmentLine()
	return nil
}

func printSegmentLine() {
	log.Println("------------------------------------------------------------")
}

func main() {
	flag.Parse()
	log.Printf("Command: %s\n", command)
	if len(command) == 0 {
		log.Println("The argument '-command' is NOT specified!")
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
	log.Printf("Parameters: \n  Command: %s\n  Base paths: %s", command, basePaths)
	printSegmentLine()
	targetPaths := make([]string, 0)
	for _, basePath := range basePaths {
		subTargetPaths, err := checkBaseDir(basePath)
		if err != nil {
			log.Println("CheckBaseDirError:", err)
			return
		}
		targetPaths = append(targetPaths, subTargetPaths...)
	}
	for _, targetPath := range targetPaths {
		executeCommand(targetPath, command)
	}
	printSegmentLine()
	log.Println("The command(s) execution has been finished.")
}
