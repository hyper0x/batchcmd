package main

import (
	"bytes"
	"flag"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"sort"
	"strings"
	"sync"
)

const (
	separator = "------------------------------------------------------------"
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

// findAllSubDirs find all of sub dirs of specified path base on depth-first.
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
		return nil
	}
	subPaths, err := file.Readdirnames(-1)
	if err != nil {
		log.Printf("ReaddirnamesError(%s): %s\n", filePath, err)
		return err
	}
	if len(subPaths) == 0 {
		log.Printf("Ignore EMPTY directory '%s'.\n", filePath)
		return nil
	}
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

// findAllTargetDirs check all of base paths, and find their sub dirs concurrently.
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
		wg.Add(1)
		go func() {
			defer func() {
				wg.Done()
			}()
			err := findAllSubDirs(absBasePath, depth, pathCh)
			if err != nil {
				log.Printf("FindAllDirsError (%s): %s\n", absBasePath, err)
			}
		}()
	}
	go func() {
		wg.Wait()
		close(pathCh)
	}()
}

func executeCommand(targetPath string, command string) (logContent string, err error) {
	logBuffer := new(bytes.Buffer)
	defer func() {
		logContent = logBuffer.String()
	}()
	bufferLog(logBuffer, separator)
	bufferLog(logBuffer, "\nEntry into target Path: %s\n", targetPath)
	err = os.Chdir(targetPath)
	if err != nil {
		bufferLog(logBuffer, "ChdirError (%s): %s\n", targetPath, err)
		return "", err
	}
	cmdWithArgs := strings.Split(command, " ")
	var cmd *exec.Cmd
	cmdLength := len(cmdWithArgs)
	realCmd := cmdWithArgs[0]
	args := cmdWithArgs[1:cmdLength]
	bufferLog(logBuffer, "Execute command (cmd=%s, args=%s)...\n", realCmd, args)
	if cmdLength > 1 {
		cmd = exec.Command(realCmd, args...)
	} else {
		cmd = exec.Command(realCmd)
	}
	result, err := cmd.Output()
	if err != nil {
		bufferLog(logBuffer, "CmdRunError (cmd=%s, args=%v): %s\n", realCmd, args, err)
		return "", err
	}
	bufferLog(logBuffer, "Output: %v\n", string(result))
	return "", nil
}

func bufferLog(buffer *bytes.Buffer, template string, args ...interface{}) {
	buffer.WriteString(fmt.Sprintf(template, args...))
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
	log.Printf("Parameters: \n  Command: %s\n  Base paths: %s\n  Depth: %d\n  Test: %v\n", command, basePaths, depth, isTest)

	pathCh := make(chan string, 5)
	findAllTargetDirs(basePaths, depth, pathCh)

	if isTest {
		var targetPaths []string
		for targetPath := range pathCh {
			targetPaths = append(targetPaths, targetPath)
		}
		sort.Strings(targetPaths)
		for _, targetPath := range targetPaths {
			log.Printf("Target path: %s\n", targetPath)
		}
		log.Println(separator)
		log.Println("The command(s) execution has been ignored in test mode.")
		return
	}

	var wg sync.WaitGroup
	exec := func() {
		defer wg.Done()
		for targetPath := range pathCh {
			logContent, _ := executeCommand(targetPath, command)
			log.Println(logContent)
		}
	}
	pNum := runtime.NumCPU()
	for i := 0; i < pNum; i++ {
		wg.Add(1)
		go exec()
	}
	wg.Wait()
	log.Println(separator)
	log.Println("The command(s) execution has been finished.")
}
