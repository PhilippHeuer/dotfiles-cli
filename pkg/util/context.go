package util

import (
	"bufio"
	"fmt"
	"log/slog"
	"os"
	"strconv"
	"strings"
)

func ParseContextValue(value string) interface{} {
	// bool
	if v, err := strconv.ParseBool(value); err == nil {
		return v
	}

	// int
	if v, err := strconv.ParseInt(value, 10, 64); err == nil {
		return v
	}

	// float
	if v, err := strconv.ParseFloat(value, 64); err == nil {
		return v
	}

	return value
}

func LoadContextFile(path string) (map[string]interface{}, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("failed to open context file: %w", err)
	}
	defer file.Close()

	ctx := make(map[string]interface{})
	scanner := bufio.NewScanner(file)
	for lineNo := 1; scanner.Scan(); lineNo++ {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		key, value, found := strings.Cut(line, "=")
		if !found {
			slog.Debug("skipping malformed context line", "line", lineNo)
			continue
		}

		key = strings.TrimSpace(key)
		value = strings.TrimSpace(value)

		if key == "" {
			continue
		}

		ctx[key] = ParseContextValue(value)
	}

	return ctx, scanner.Err()
}
