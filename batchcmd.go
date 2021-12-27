package main

import (
	"context"
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
	"time"

	loghelper "github.com/hyper0x/batchcmd/helper/log"
)

const (
	separator = "------------------------------------------------------------"
)

// logMap represents the log dictionary.
var logMap = loghelper.NewMap()

// appendLog is used to append the log.
func appendLog(path string, level loghelper.Level, content string) {
	if !verboseLog && level < loghelper.LEVEL_INFO {
		return
	}
	oneLog := loghelper.NewOne(level, content)
	logMap.Append(path, oneLog)
}

var (
	command    string
	parentDirs string
	depth      int
	isMock     bool
	timeout    int64
	verboseLog bool
)

func init() {
	flag.StringVar(&command, "c", "",
		"The command you want to execute")
	flag.StringVar(&parentDirs, "p", "",
		"The parent path of target directory(the command will be executed there). \n"+
			"Note that multiple path needs to separated by commas ','.")
	flag.IntVar(&depth, "d", 1,
		"The max search depth for the target directory in parent path(s). ")
	flag.BoolVar(&isMock, "m", false,
		"Only mock. (the command will not be executed)")
	flag.Int64Var(&timeout, "t", 30,
		"The seconds for timeout of per command.")
	flag.BoolVar(&verboseLog, "v", false,
		"Print the verbose logs.")
	if verboseLog {
		log.SetFlags(log.LstdFlags)
	} else {
		log.SetFlags(0)
	}
}

// findAllSubDirs find all of sub dirs of specified path base on depth-first.
func findAllSubDirs(filePath string, isParentDir bool, depth int, pathCh chan<- string) error {
	if depth < 0 {
		return nil
	}
	file, err := os.Open(filePath)
	if err != nil {
		return fmt.Errorf("open file: %s", err)
	}
	fileInfo, err := file.Stat()
	if err != nil {
		return fmt.Errorf("get file stat: %s", err)
	}
	if !fileInfo.IsDir() {
		return nil
	}
	subPaths, err := file.Readdirnames(-1)
	if err != nil {
		return fmt.Errorf("read subdir name: %s", err)
	}
	if len(subPaths) == 0 {
		one := loghelper.NewOne(loghelper.LEVEL_WARN,
			fmt.Sprintf("ignore empty dir '%s'.", filePath))
		log.Println(one)
		return nil
	}
	if !isParentDir {
		pathCh <- filePath
	}
	if depth == 0 {
		return nil
	}
	newDepth := depth - 1
	for _, subPath := range subPaths {
		if strings.HasPrefix(subPath, ".") {
			continue
		}
		absSubPath := filepath.Join(filePath, subPath)
		err := findAllSubDirs(absSubPath, false, newDepth, pathCh)
		if err != nil {
			return err
		}
	}
	return nil
}

// findAllTargetDirs check all of parent dirs, and find their sub dirs concurrently.
func findAllTargetDirs(parentDirs []string, depth int, dirCh chan<- string) {
	var absParentDirs []string
	for _, parentDir := range parentDirs {
		one := loghelper.NewOne(loghelper.LEVEL_INFO,
			fmt.Sprintf("check parent dir '%s'.", parentDir))
		log.Println(one)
		if filepath.IsAbs(parentDir) {
			absParentDirs = append(absParentDirs, parentDir)
		} else {
			absParentDir, err := filepath.Abs(parentDir)
			if err != nil {
				one := loghelper.NewOne(loghelper.LEVEL_ERROR,
					fmt.Sprintf("abs parent dir '%s': %s", parentDir, err))
				log.Println(one)
				continue
			}
			absParentDirs = append(absParentDirs, absParentDir)
		}
	}
	var wg sync.WaitGroup
	wg.Add(len(absParentDirs))
	for _, absParentDir := range absParentDirs {
		go func(dir string) {
			defer func() {
				wg.Done()
			}()
			err := findAllSubDirs(dir, true, depth, dirCh)
			if err != nil {
				appendLog(dir,
					loghelper.LEVEL_ERROR,
					fmt.Sprintf("find all subdirs: %s", err))
			}
		}(absParentDir)
	}
	go func() {
		wg.Wait()
		close(dirCh)
	}()
}

func executeCommand(
	targetDir string,
	command string,
	ctx context.Context,
	cancel context.CancelFunc) {
	appendLog(targetDir, loghelper.LEVEL_DEBUG, "entry into target dir.")
	err := os.Chdir(targetDir)
	if err != nil {
		appendLog(targetDir,
			loghelper.LEVEL_ERROR,
			fmt.Sprintf("change dir: %s", err))
		return
	}
	cmdWithArgs := strings.Split(command, " ")
	var cmd *exec.Cmd
	cmdLength := len(cmdWithArgs)
	realCmd := cmdWithArgs[0]
	args := cmdWithArgs[1:cmdLength]
	appendLog(targetDir, loghelper.LEVEL_DEBUG,
		fmt.Sprintf("execute command '%s'.", command))
	if cmdLength > 1 {
		cmd = exec.CommandContext(ctx, realCmd, args...)
	} else {
		cmd = exec.CommandContext(ctx, realCmd)
	}
	result, err := cmd.CombinedOutput()
	appendLog(targetDir, loghelper.LEVEL_INFO,
		fmt.Sprintf("output: \n%s", string(result)))
	if err != nil {
		appendLog(targetDir,
			loghelper.LEVEL_ERROR,
			fmt.Sprintf("run command '%s': %s", command, err))
	}
}

func main() {
	flag.Parse()
	if isMock {
		log.Println("Starting... (in mock environment)")
	} else {
		log.Println("Starting... (in formal environment)")
	}
	if len(command) == 0 {
		log.Println("The argument '-c' is NOT specified!")
		return
	}
	var allParentDirs []string
	if len(parentDirs) > 0 {
		allParentDirs = strings.Split(parentDirs, ",")
	} else {
		defaultParentDir, err := os.Getwd()
		if err != nil {
			log.Println("Work Dir Getting Error:", err)
			return
		}
		allParentDirs = []string{defaultParentDir}
	}
	paraDesc := "Parameters: " +
		"\n  Command: %s" +
		"\n  Parent Dirs: %s" +
		"\n  Depth: %d" +
		"\n  Mock: %v" +
		"\n  Timeout(per command): %ds" +
		"\n"
	log.Printf(paraDesc,
		command, allParentDirs, depth, isMock, timeout)
	log.Println("")

	dirCh := make(chan string, 100*len(allParentDirs))
	findAllTargetDirs(allParentDirs, depth, dirCh)

	if isMock {
		var targetDirs []string
		for targetDir := range dirCh {
			targetDirs = append(targetDirs, targetDir)
		}
		log.Println(separator)
		sort.Strings(targetDirs)
		log.Printf("Target dirs(%d): \n", len(targetDirs))
		for i, targetDir := range targetDirs {
			log.Printf("  %d. %s\n", i+1, targetDir)
		}
		log.Println(separator)
		log.Println("The command(s) execution has been ignored in mock mode.")
		return
	}

	workFunc := func(wg *sync.WaitGroup) {
		defer wg.Done()
		timeoutDuration := time.Duration(timeout) * time.Second
		for targetDir := range dirCh {
			ctx, cancel := context.WithTimeout(
				context.Background(), timeoutDuration)

			executeCommand(targetDir, command, ctx, cancel)
		}
	}
	pNum := runtime.GOMAXPROCS(-1)
	var wg sync.WaitGroup
	wg.Add(pNum)
	for i := 0; i < pNum; i++ {
		go workFunc(&wg)
	}
	wg.Wait()
	var summaries []string
	logMap.Range(func(key string, list loghelper.List) bool {
		log.Println(separator)
		log.Printf("Target Dir: %s\n", key)
		errorCount := 0
		for index, one := range list.GetAll() {
			log.Printf("  %d. %s\n", index+1, one.String())
			if one.Level() >= loghelper.LEVEL_ERROR {
				errorCount++
			}
		}
		var summary string
		if errorCount > 0 {
			summary = fmt.Sprintf("%s: failure(%d).", key, errorCount)
		} else {
			summary = fmt.Sprintf("%s: success.", key)
		}
		summaries = append(summaries, summary)
		return true
	})
	log.Println(separator)
	log.Println("Summary: ")
	for i, s := range summaries {
		log.Printf("  %d. %s\n", i+1, s)
	}
	log.Println(separator)
	log.Println("The command(s) execution has been finished.")
}
