package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/hpcloud/tail"
)

const (
	defaultWalDir  = "./log-agent-wal"
	segmentSize    = 1 * 1024 * 1024 
	maxBatchLen    = 100
	maxBatchBytes  = 512 * 1024
)

var client = &http.Client{Timeout: 15 * time.Second}



type Entry struct {
	Offset int64  `json:"offset"`
	Log    string `json:"log"`
}

type LogBatchRequest struct {
	ServiceName string   `json:"service_name"`
	Logs        []string `json:"logs"`
}



func main() {
	walDir := getEnv("WAL_DIR", defaultWalDir)
	logFile := getEnv("LOG_FILE_PATH", "app.log")
	offsetFile := filepath.Join(walDir, "global_offset.dat")

	os.MkdirAll(walDir, 0755)

	
	var startOffset int64
	if data, err := os.ReadFile(offsetFile); err == nil {
		startOffset, _ = strconv.ParseInt(strings.TrimSpace(string(data)), 10, 64)
	}

	t, err := tail.TailFile(logFile, tail.Config{
		Follow:   true,
		ReOpen:   true,
		Location: &tail.SeekInfo{Offset: startOffset, Whence: os.SEEK_SET},
	})
	if err != nil {
		log.Fatalf("[-] Tail Error: %v", err)
	}

	
go func() {
    for line := range t.Lines {
        if line.Err != nil {
            continue
        }

        
        currentPos, _ := t.Tell()

        appendToSegment(walDir, Entry{
            Offset: currentPos, 
            Log:    formatLog(line.Text),
        })
    }
}()

	
	log.Printf("[+] Agent Online. Service: %s, WAL: %s", getEnv("SERVICE_NAME", "unknown"), walDir)
	for {
		processSegments(walDir, offsetFile)
		time.Sleep(1 * time.Second) 
	}
}

func formatLog(raw string) string {
	ip := getEnv("NODE_IP", "0.0.0.0")
	node := getEnv("NODE_NAME", "node")
	
	
	level := "INFO"
	upper := strings.ToUpper(raw)
	for _, l := range []string{"ERROR", "WARN", "DEBUG", "FATAL"} {
		if strings.Contains(upper, l) {
			level = l
			break
		}
	}

	return fmt.Sprintf("%s | %s | %s | %s", ip, node, level, strings.TrimSpace(raw))
}


func appendToSegment(walDir string, e Entry) {
	file := getActiveSegment(walDir)

	f, err := os.OpenFile(file, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return
	}
	defer f.Close()

	data, _ := json.Marshal(e)
	f.Write(append(data, '\n'))
}

func getActiveSegment(walDir string) string {
	files, _ := filepath.Glob(filepath.Join(walDir, "segment-*.log"))
	if len(files) == 0 {
		return newSegment(walDir, 1)
	}

	sort.Strings(files)
	last := files[len(files)-1]

	if info, err := os.Stat(last); err == nil && info.Size() >= segmentSize {
		return newSegment(walDir, extractID(last)+1)
	}
	return last
}

func newSegment(walDir string, id int) string {
	name := filepath.Join(walDir, fmt.Sprintf("segment-%06d.log", id))
	f, _ := os.Create(name)
	f.Close()
	return name
}

func extractID(path string) int {
	idStr := strings.TrimSuffix(strings.TrimPrefix(filepath.Base(path), "segment-"), ".log")
	id, _ := strconv.Atoi(idStr)
	return id
}



func processSegments(walDir, globalOffsetFile string) {
	files, _ := filepath.Glob(filepath.Join(walDir, "segment-*.log"))
	sort.Strings(files)

	for _, file := range files {
		currentOffset := loadOffset(file + ".offset")
		entries, bytesRead := readBatch(file, currentOffset)
		
		if len(entries) == 0 {
			continue 
		}

		if sendWithRetry(entries) {
			newOffset := currentOffset + int64(bytesRead)
			saveOffset(file+".offset", newOffset)

			info, _ := os.Stat(file)
			if newOffset >= info.Size() {
				os.Remove(file)
				os.Remove(file + ".offset")
			}
			
			saveOffset(globalOffsetFile, entries[len(entries)-1].Offset)
		} else {
			return 
		}
	}
}

func readBatch(file string, offset int64) ([]Entry, int) {
	f, err := os.Open(file)
	if err != nil {
		return nil, 0
	}
	defer f.Close()

	f.Seek(offset, 0)
	scanner := bufio.NewScanner(f)
	var batch []Entry
	totalBytes := 0

	for scanner.Scan() {
		line := scanner.Bytes()
		if len(batch) >= maxBatchLen || (totalBytes+len(line)) > maxBatchBytes {
			break
		}

		var e Entry
		if err := json.Unmarshal(line, &e); err == nil {
			batch = append(batch, e)
			totalBytes += len(line) + 1 
		}
	}
	return batch, totalBytes
}



func sendWithRetry(entries []Entry) bool {
	url := getEnv("SERVER_URL", "http://localhost:8080/api/v1/logs")
	token := getEnv("AGENT_ACCESS_KEY", "default_secret")
	
	var logs []string
	for _, e := range entries {
		logs = append(logs, e.Log)
	}

	body, _ := json.Marshal(LogBatchRequest{
		ServiceName: getEnv("SERVICE_NAME", "unknown"),
		Logs:        logs,
	})

	backoff := 1 * time.Second
	for {
		req, _ := http.NewRequest("POST", url, bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("X-Agent-Access-Key", token)

		resp, err := client.Do(req)
		if err == nil && resp.StatusCode == http.StatusOK {
			resp.Body.Close()
			return true
		}

		log.Printf("[!] Network/Server Error. Retrying in %v...", backoff)
		if resp != nil { resp.Body.Close() }

		time.Sleep(backoff)
		if backoff < 30*time.Second { backoff *= 2 }
	}
}


func loadOffset(path string) int64 {
	data, err := os.ReadFile(path)
	if err != nil { return 0 }
	val, _ := strconv.ParseInt(strings.TrimSpace(string(data)), 10, 64)
	return val
}

func saveOffset(path string, val int64) {
	tmp := path + ".tmp"
	os.WriteFile(tmp, []byte(strconv.FormatInt(val, 10)), 0644)
	os.Rename(tmp, path)
}

func getEnv(k, def string) string {
	if v := os.Getenv(k); v != "" { return v }
	return def
}