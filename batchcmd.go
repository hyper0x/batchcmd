package main

import (
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

	loghelper "github.com/hyper0x/batchcmd/helper/log"
)

const (
	separator = "------------------------------------------------------------"
)

// logMap 代表日志字典。
var logMap = loghelper.NewMap()

// appendLog 用于追加日志。
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
	isTest     bool
	verboseLog bool
)

func init() {
	flag.StringVar(&command, "c", "", "The command that will be executed.")
	flag.StringVar(&parentDirs, "p", "", "The parent path of target directory. Note that multiple path needs to separated by commas ','.")
	flag.IntVar(&depth, "d", 1, "The max depth  of target directory. ")
	flag.BoolVar(&isTest, "t", false, "Only test. (Do not execute the command)")
	flag.BoolVar(&verboseLog, "v", false, "Print verbose log.")
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

func executeCommand(targetDir string, command string) {
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
		cmd = exec.Command(realCmd, args...)
	} else {
		cmd = exec.Command(realCmd)
	}
	result, err := cmd.Output()
	if err != nil {
		appendLog(targetDir,
			loghelper.LEVEL_ERROR,
			fmt.Sprintf("run command '%s': %s", command, err))
		return
	}
	appendLog(targetDir, loghelper.LEVEL_INFO,
		fmt.Sprintf("output: \n%s", string(result)))
	return
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
	log.Printf("Parameters: \n  Command: %s\n  Parent Dirs: %s\n  Depth: %d\n  Test: %v\n",
		command, allParentDirs, depth, isTest)
	log.Println("")

	dirCh := make(chan string, 50)
	findAllTargetDirs(allParentDirs, depth, dirCh)

	if isTest {
		var targetDirs []string
		for targetDir := range dirCh {
			targetDirs = append(targetDirs, targetDir)
		}
		log.Println("")
		log.Println(separator)
		sort.Strings(targetDirs)
		log.Printf("Target dirs(%d): \n", len(targetDirs))
		for i, targetDir := range targetDirs {
			log.Printf("  %d. %s\n", i+1, targetDir)
		}
		log.Println("")
		log.Println(separator)
		log.Println("The command(s) execution has been ignored in test mode.")
		return
	}

	workFunc := func(wg *sync.WaitGroup) {
		defer wg.Done()
		for targetDir := range dirCh {
			executeCommand(targetDir, command)
		}
	}
	pNum := runtime.NumCPU()
	var wg sync.WaitGroup
	wg.Add(pNum)
	for i := 0; i < pNum; i++ {
		go workFunc(&wg)
	}
	wg.Wait()
	var summaries []string
	logMap.Range(func(key string, list loghelper.List) bool {
		log.Println("")
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
	log.Println("")
	log.Println(separator)
	log.Println("Summary: ")
	for i, s := range summaries {
		log.Printf("  %d. %s\n", i+1, s)
	}
	log.Println("")
	log.Println(separator)
	log.Println("The command(s) execution has been finished.")
}
