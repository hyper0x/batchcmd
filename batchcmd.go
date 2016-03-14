package main

import (
	"flag"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
)

var (
	command    string
	parentDirs string
	depth      int
	isTest     bool
)

func init() {
	log.SetFlags(log.Lshortfile | log.LstdFlags)
	flag.StringVar(&command, "c", "", "The command that will be executed.")
	flag.StringVar(&parentDirs, "p", "", "The parent path of target directory. Note that multiple path needs to separated by commas ','.")
	flag.IntVar(&depth, "d", 1, "The max depth  of target directory. ")
	flag.BoolVar(&isTest, "t", false, "Only test. (Do not execute the command)")
}

func findAllSubDirs(filePath string, depth int, ch chan<- string) error {
	if depth < 0 {
		return nil
	}
	file, err := os.Open(filePath)
	if err != nil {
		log.Printf("OpenError (%s): %s\n", filePath, err)
		return err
	}
	fileInfo, err := file.Stat()
	if err != nil {
		log.Printf("StatError (%s): %s\n", filePath, err)
		return err
	}
	if !fileInfo.IsDir() {
		log.Printf("Non-target: %s\n", filePath)
		return nil
	}
	subPaths, err := file.Readdirnames(-1)
	if err != nil {
		log.Printf("ReaddirnamesError(%s): %s\n", filePath, err)
		return err
	}
	if len(subPaths) == 0 {
		log.Printf("Note that Ignore EMPTY directory '%s'.\n", filePath)
		return nil
	}
	log.Printf("Target: %s\n", filePath)
	ch <- filePath
	if depth == 0 {
		return nil
	}
	newDepth := depth - 1
	for _, subPath := range subPaths {
		if strings.HasPrefix(subPath, ".") {
			continue
		}
		absSubPath := filepath.Join(filePath, subPath)
		err := findAllSubDirs(absSubPath, newDepth, ch)
		if err != nil {
			return err
		}
	}
	return nil
}

func findAllTargetDirs(basePaths []string, depth int, pathCh chan<- string) {
	var wg sync.WaitGroup
	var err error
	for _, basePath := range basePaths {
		log.Printf("Base Path: %s\n", basePath)
		var absBasePath string
		if filepath.IsAbs(basePath) {
			absBasePath = basePath
		} else {
			absBasePath, err = filepath.Abs(basePath)
			if err != nil {
				log.Printf("AbsBasePathError (%s): %s\n", basePath, err)
				close(pathCh)
				break
			}
		}

		go func() {
			wg.Add(1)
			defer func() {
				close(pathCh)
				wg.Done()
			}()
			err := findAllSubDirs(absBasePath, depth, pathCh)
			if err != nil {
				log.Printf("FindAllDirsError (%s): %s\n", absBasePath, err)
			}
		}()
	}
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
	var basePaths []string
	if len(parentDirs) > 0 {
		basePaths = strings.Split(parentDirs, ",")
	} else {
		defaultBasePath, err := os.Getwd()
		if err != nil {
			log.Println("GetwdError:", err)
			return
		}
		basePaths = []string{defaultBasePath}
	}
	log.Printf("Parameters: \n  Command: %s\n  Base paths: %s\n  Depth: %d\n Test: %v\n", command, basePaths, depth, isTest)

	pathCh := make(chan string, 5)
	findAllTargetDirs(basePaths, depth, pathCh)
	for targetPath := range pathCh {
		if !isTest {
			printSegmentLine()
			executeCommand(targetPath, command)
		}
	}
	printSegmentLine()
	log.Println("The command(s) execution has been finished.")
}
