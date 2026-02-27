package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"

	"github.com/spf13/cobra"
)

var logsCmd = &cobra.Command{
	Use:   "logs <id>",
	Short: "Show worker output log",
	Args:  cobra.ExactArgs(1),
	RunE:  runLogs,
}

var (
	logsFollow bool
	logsJSON   bool
)

func init() {
	logsCmd.Flags().BoolVarP(&logsFollow, "follow", "f", false, "Tail log output")
	logsCmd.Flags().BoolVar(&logsJSON, "json", false, "Output raw NDJSON events")
	rootCmd.AddCommand(logsCmd)
}

func runLogs(cmd *cobra.Command, args []string) error {
	id := args[0]
	logPath := filepath.Join(stateDir, id+".log")

	f, err := os.Open(logPath)
	if err != nil {
		return fmt.Errorf("no log file for worker %s", id)
	}
	defer f.Close()

	if logsJSON {
		if logsFollow {
			return tailFile(cmd.OutOrStdout(), f)
		}
		io.Copy(cmd.OutOrStdout(), f)
		return nil
	}

	if logsFollow {
		return tailFileHuman(cmd.OutOrStdout(), f)
	}
	return renderHuman(cmd.OutOrStdout(), f)
}

func renderHuman(out io.Writer, r io.Reader) error {
	scanner := bufio.NewScanner(r)
	for scanner.Scan() {
		renderLine(out, scanner.Bytes())
	}
	return scanner.Err()
}

func renderLine(out io.Writer, line []byte) {
	var event map[string]interface{}
	if err := json.Unmarshal(line, &event); err != nil {
		fmt.Fprintf(out, "%s\n", line)
		return
	}

	typ, _ := event["type"].(string)
	switch typ {
	case "assistant":
		// Real stream-json nests content under .message.content[]
		// Test/simple format may have .content directly
		content := event["content"]
		if msg, ok := event["message"].(map[string]interface{}); ok {
			content = msg["content"]
		}
		renderContent(out, content)
	case "tool_use":
		name, _ := event["name"].(string)
		fmt.Fprintf(out, "[tool: %s]\n", name)
	case "result":
		sub, _ := event["subtype"].(string)
		fmt.Fprintf(out, "[result: %s]\n", sub)
	}
}

func renderContent(out io.Writer, content interface{}) {
	switch c := content.(type) {
	case string:
		if c != "" {
			fmt.Fprintf(out, "%s\n", c)
		}
	case []interface{}:
		for _, item := range c {
			m, ok := item.(map[string]interface{})
			if !ok {
				continue
			}
			itemType, _ := m["type"].(string)
			switch itemType {
			case "text":
				if text, ok := m["text"].(string); ok && text != "" {
					fmt.Fprintf(out, "%s\n", text)
				}
			case "tool_use":
				if name, ok := m["name"].(string); ok {
					fmt.Fprintf(out, "[tool: %s]\n", name)
				}
			}
		}
	}
}

func tailFile(out io.Writer, f *os.File) error {
	io.Copy(out, f)
	for {
		n, _ := io.Copy(out, f)
		if n == 0 {
			time.Sleep(200 * time.Millisecond)
		}
	}
}

func tailFileHuman(out io.Writer, f *os.File) error {
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		renderLine(out, scanner.Bytes())
	}
	for {
		if scanner.Scan() {
			renderLine(out, scanner.Bytes())
		} else {
			time.Sleep(200 * time.Millisecond)
			scanner = bufio.NewScanner(f)
		}
	}
}
