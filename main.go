package main

import (
	"context"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"
	"time"
)

const (
	defaultCheckIntervalMin = 1
	defaultIPURLv4          = "https://api.ipify.org"
	defaultIPURLv6          = "https://api64.ipify.org"
)

var defaultIPFile = filepath.Join(os.TempDir(), "last_ip")

func getEnv(key, fallback string) string {
	val := os.Getenv(key)
	if val == "" {
		return fallback
	}
	return val
}

func getCheckInterval() time.Duration {
	val := getEnv("CHECK_INTERVAL_MINUTES", fmt.Sprintf("%d", defaultCheckIntervalMin))
	mins, err := strconv.Atoi(val)
	if err != nil || mins <= 0 {
		log.Printf("Invalid CHECK_INTERVAL_MINUTES: %s, using default of %d", val, defaultCheckIntervalMin)
		return time.Duration(defaultCheckIntervalMin) * time.Minute
	}
	return time.Duration(mins) * time.Minute
}

func getCurrentIP(ctx context.Context, url string) (string, error) {
	var lastErr error
	backoffs := []time.Duration{2 * time.Second, 4 * time.Second, 8 * time.Second}

	client := &http.Client{}

	for i, delay := range backoffs {
		req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
		if err != nil {
			return "", err
		}
		resp, err := client.Do(req)
		if err != nil {
			lastErr = err
			log.Printf("Attempt %d: error contacting IP service: %v", i+1, err)
			time.Sleep(delay)
			continue
		}
		body, err := io.ReadAll(resp.Body)
		_ = resp.Body.Close()
		if err != nil {
			lastErr = err
			log.Printf("Attempt %d: error reading IP response: %v", i+1, err)
			time.Sleep(delay)
			continue
		}
		ip := strings.TrimSpace(string(body))
		if ip != "" {
			return ip, nil
		}
		lastErr = fmt.Errorf("empty IP response")
		log.Printf("Attempt %d: received empty IP", i+1)
		time.Sleep(delay)
	}
	return "", fmt.Errorf("failed to get current IP after retries: %w", lastErr)
}

func readLastIP(ipFile string) string {
	data, err := os.ReadFile(ipFile)
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(data))
}

func writeLastIP(ipFile, ip string) {
	_ = os.WriteFile(ipFile, []byte(ip), 0644)
}

func runUpdateCycle(ctx context.Context, ipFile, ipURLv4, ipURLv6 string) {
	ipV4, err := getCurrentIP(ctx, ipURLv4)
	if err != nil {
		log.Printf("Failed to get IPv4: %v", err)
	}
	ipV6, err := getCurrentIP(ctx, ipURLv6)
	if err != nil {
		log.Printf("Failed to get IPv6: %v", err)
	}
	if ipV4 == "" && ipV6 == "" {
		log.Printf("No usable IPs found, skipping update cycle")
		return
	}

	lastIP := readLastIP(ipFile)
	combinedIP := fmt.Sprintf("%s|%s", ipV4, ipV6)
	if combinedIP != lastIP {
		log.Printf("IP change detected: %s -> %s", lastIP, combinedIP)
		log.Printf("Would update DNS here if Cloudflare code were active")
		writeLastIP(ipFile, combinedIP)
	} else {
		log.Printf("No IP change detected: %s", combinedIP)
	}
}

func main() {
	ipFile := getEnv("IP_FILE", defaultIPFile)
	ipURLv4 := getEnv("IP_CHECK_URL_V4", defaultIPURLv4)
	ipURLv6 := getEnv("IP_CHECK_URL_V6", defaultIPURLv6)
	checkInterval := getCheckInterval()

	log.Printf("\nCurrent Configuration:\n%s\n%s\n%s\n%s\n",
		fmt.Sprintf("  %-24s = %s", "IP_FILE", ipFile),
		fmt.Sprintf("  %-24s = %s", "IP_CHECK_URL_V4", ipURLv4),
		fmt.Sprintf("  %-24s = %s", "IP_CHECK_URL_V6", ipURLv6),
		fmt.Sprintf("  %-24s = %.0f", "CHECK_INTERVAL_MINUTES", checkInterval.Minutes()),
	)

	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, os.Interrupt, syscall.SIGTERM)
	done := make(chan struct{})

	go func() {
		ticker := time.NewTicker(checkInterval)
		defer ticker.Stop()
		ctx := context.Background()
		for {
			runUpdateCycle(ctx, ipFile, ipURLv4, ipURLv6)
			select {
			case <-ticker.C:
				continue
			case <-done:
				log.Println("Graceful shutdown triggered")
				return
			}
		}
	}()

	<-signalChan
	log.Println("Shutdown signal received")
	close(done)
	// Let goroutine finish if it’s in the middle of work
	time.Sleep(1 * time.Second)
	log.Println("Shutdown complete")
}
