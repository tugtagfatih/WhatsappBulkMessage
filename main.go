package main

import (
	"bufio"
	"fmt"
	"log"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/tebeka/selenium"
	"github.com/tebeka/selenium/chrome"
)

const (
	// chromeDriverPath: Path to the ChromeDriver executable.
	// If empty, chromedriver is expected to be in the system PATH.
	chromeDriverPath = "chromedriver-win64\\chromedriver.exe" // IMPORTANT: Change this to your ChromeDriver path

	// useProfile: Whether to use a Chrome profile.
	useProfile = true

	// profileFolder: Folder to store Chrome profile data.
	profileFolder = "whatsapp_profile_go" // IMPORTANT: Change this to your desired profile path. Relative paths are relative to the executable.

	// numbersFilePath: Name of the file containing phone numbers.
	numbersFilePath = "numbers.txt"
	// messageFilePath: Name of the file containing the message to send.
	messageFilePath = "text.txt"

	// logsDir: Directory to store log files.
	logsDir = "logs"

	// whatsAppURL: WhatsApp Web address.
	whatsAppURL = "https://web.whatsapp.com/"
	// seleniumPort: Port for ChromeDriver to run on.
	seleniumPort = 9515 // Common port for ChromeDriver, change if needed.
)

// readLines reads a file line by line and returns the lines as a slice.
func readLines(path string) ([]string, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var lines []string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		lines = append(lines, strings.TrimSpace(scanner.Text()))
	}
	return lines, scanner.Err()
}

// setupLogFile creates the logs directory if it doesn't exist and creates a new log file with a timestamp.
// It returns the file pointer and an error if any.
func setupLogFile() (*os.File, error) {
	if _, err := os.Stat(logsDir); os.IsNotExist(err) {
		if err := os.MkdirAll(logsDir, 0755); err != nil {
			return nil, fmt.Errorf("failed to create logs directory '%s': %w", logsDir, err)
		}
		fmt.Printf("Logs directory created: %s\n", logsDir)
	}

	timestamp := time.Now().Format("2006-01-02_15-04-05")
	logFileName := fmt.Sprintf("%s_log.txt", timestamp)
	logFilePath := filepath.Join(logsDir, logFileName)

	logFile, err := os.OpenFile(logFilePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return nil, fmt.Errorf("failed to open log file '%s': %w", logFilePath, err)
	}
	fmt.Printf("Logging to: %s\n", logFilePath)
	return logFile, nil
}

// logMessage writes a message to both the console and the log file.
func logMessage(logFile *os.File, format string, args ...interface{}) {
	message := fmt.Sprintf(format, args...)
	fmt.Print(message) // Print to console
	if logFile != nil {
		if _, err := logFile.WriteString(time.Now().Format("[2006-01-02 15:04:05] ") + message); err != nil {
			log.Printf("Failed to write to log file: %v", err) // Log to console if file write fails
		}
	}
}

func main() {
	// Setup logging
	logFile, err := setupLogFile()
	if err != nil {
		log.Fatalf("Error setting up log file: %v", err)
	}
	defer logFile.Close()

	logMessage(logFile, "Application started.\n")
	var service *selenium.Service

	if chromeDriverPath != "" {
		service, err = selenium.NewChromeDriverService(chromeDriverPath, seleniumPort)
		if err != nil {
			logMessage(logFile, "Failed to start ChromeDriver service from specified path ('%s'): %v. Trying ChromeDriver from PATH.\n", chromeDriverPath, err)
			service, err = selenium.NewChromeDriverService("", seleniumPort)
		}
	} else {
		service, err = selenium.NewChromeDriverService("", seleniumPort)
	}

	if err != nil {
		logMessage(logFile, "Error starting ChromeDriver service (tried specified path and PATH): %v\n", err)
		os.Exit(1)
	}
	defer service.Stop()

	// 2. Setup Chrome Capabilities
	caps := selenium.Capabilities{"browserName": "chrome"}
	chromeCaps := chrome.Capabilities{}

	if useProfile {
		// Ensure profile folder exists
		absProfileFolder, err := filepath.Abs(profileFolder)
		if err != nil {
			logMessage(logFile, "Could not get absolute path for profile folder ('%s'): %v\n", profileFolder, err)
			os.Exit(1)
		}

		if _, statErr := os.Stat(absProfileFolder); os.IsNotExist(statErr) {
			if mkdirErr := os.MkdirAll(absProfileFolder, 0755); mkdirErr != nil {
				logMessage(logFile, "Could not create profile folder ('%s'): %v\n", absProfileFolder, mkdirErr)
				os.Exit(1)
			}
			//logMessage(logFile, "Profile folder created: %s\n", absProfileFolder)
		} else if statErr != nil {
			logMessage(logFile, "Error checking profile folder ('%s'): %v\n", absProfileFolder, statErr)
			os.Exit(1)
		}
		chromeCaps.Args = append(chromeCaps.Args, fmt.Sprintf("user-data-dir=%s", absProfileFolder))
	}
	caps.AddChrome(chromeCaps)

	// 3. Connect to WebDriver
	wd, err := selenium.NewRemote(caps, fmt.Sprintf("http://localhost:%d/wd/hub", seleniumPort))
	if err != nil {
		logMessage(logFile, "Error connecting to WebDriver: %v\n", err)
		os.Exit(1)
	}
	defer wd.Quit()

	// 4. Navigate to WhatsApp Web and Login
	if err := wd.Get(whatsAppURL); err != nil {
		logMessage(logFile, "Error opening WhatsApp Web ('%s'): %v\n", whatsAppURL, err)
		os.Exit(1)
	}

	loginCheckSelector := "div[role='textbox']"
	//logMessage(logFile, "Please scan the QR code if prompted. Waiting for login (max 120 seconds for '%s' element to appear)...\n", loginCheckSelector)

	err = wd.WaitWithTimeoutAndInterval(func(wd selenium.WebDriver) (bool, error) {
		_, findErr := wd.FindElement(selenium.ByCSSSelector, loginCheckSelector)
		if findErr == nil {
			logMessage(logFile, "Login check element found.\n")
			return true, nil
		}
		return false, nil
	}, 120*time.Second, 1*time.Second)

	if err != nil {
		logMessage(logFile, "Login failed or timed out after 120 seconds! Error: %v. Please ensure you are logged into WhatsApp Web and the selector '%s' is correct.\n", err, loginCheckSelector)
		os.Exit(1)
	}
	//logMessage(logFile, "Successfully logged into WhatsApp Web!\n")
	time.Sleep(3 * time.Second)

	// 5. Read numbers from numbers.txt
	numbers, err := readLines(numbersFilePath)
	if err != nil {
		logMessage(logFile, "Error reading numbers from '%s': %v\n", numbersFilePath, err)
		os.Exit(1)
	}
	if len(numbers) == 0 {
		logMessage(logFile, "No numbers found in '%s'. Please add phone numbers to the file.\n", numbersFilePath)
		os.Exit(1)
	}
	//logMessage(logFile, "%d numbers found to process.\n", len(numbers))

	// 6. Read message from text.txt
	messageBytes, err := os.ReadFile(messageFilePath)
	if err != nil {
		logMessage(logFile, "Error reading message from '%s': %v\n", messageFilePath, err)
		os.Exit(1)
	}
	rawMessage := strings.TrimSpace(string(messageBytes))
	if rawMessage == "" {
		logMessage(logFile, "Message in '%s' is empty. Please write a message to send.\n", messageFilePath)
		os.Exit(1)
	}
	encodedMessage := url.QueryEscape(rawMessage)
	// logMessage(logFile, "Message to send (raw): %s\n", rawMessage)

	// 7. Send Messages
	sendButtonSelectors := []struct {
		Type string
		Path string
	}{
		{selenium.ByXPATH, "//span[@data-icon='wds-ic-send-filled']"},
		{selenium.ByXPATH, "//button[@aria-label='Send']"},
		{selenium.ByXPATH, "//button[@aria-label='Gönder']"},
		{selenium.ByCSSSelector, "span[data-testid='send']"},
		{selenium.ByCSSSelector, "button[data-testid='send']"},
	}
	messageBoxSelector := "div[data-testid='conversation-compose-box-input']"

	for _, number := range numbers {
		if number == "" {
			continue
		}
		//logMessage(logFile, "\nProcessing number: %s\n", number)

		sendURL := fmt.Sprintf("https://web.whatsapp.com/send?phone=%s&text=%s", number, encodedMessage)
		if err := wd.Get(sendURL); err != nil {
			logMessage(logFile, "FAILED to navigate to send URL for %s: %v\n", number, err)
			time.Sleep(2 * time.Second)
			continue
		}

		var sendButton selenium.WebElement
		buttonFound := false
		waitStartTime := time.Now()

		for time.Since(waitStartTime) < 30*time.Second {
			for _, selector := range sendButtonSelectors {
				element, findErr := wd.FindElement(selector.Type, selector.Path)
				if findErr == nil {
					displayed, _ := element.IsDisplayed()
					enabled, _ := element.IsEnabled()
					if displayed && enabled {
						sendButton = element
						buttonFound = true
						break
					}
				}
			}
			if buttonFound {
				break
			}
			time.Sleep(500 * time.Millisecond)
		}

		if !buttonFound {
			//logMessage(logFile, "FAILED: Send button for %s was not clickable within 30 seconds. Attempting to press Enter in message box.\n", number)
			msgBox, findErr := wd.FindElement(selenium.ByCSSSelector, messageBoxSelector)
			if findErr == nil {
				displayed, _ := msgBox.IsDisplayed()
				if displayed {
					//logMessage(logFile, "Message box found, sending Enter key for %s...\n", number)
					if errEnter := msgBox.SendKeys(selenium.EnterKey); errEnter == nil {
						//logMessage(logFile, "SUCCESS (probably): Message sent to %s by pressing Enter in message box.\n", number)
						time.Sleep(5 * time.Second)
						continue
					} else {
						//logMessage(logFile, "FAILED: Could not send Enter key to message box for %s: %v\n", number, errEnter)
					}
				} else {
					//logMessage(logFile, "FAILED: Message box found for %s but not displayed.\n", number)
				}
			} else {
				//logMessage(logFile, "FAILED: Message box ('%s') to press Enter not found for %s: %v\n", number, messageBoxSelector, findErr)
			}
			logMessage(logFile, "Skipping %s as message could not be sent.\n", number)
			time.Sleep(2 * time.Second)
			continue
		}

		if errClick := sendButton.Click(); errClick != nil {
			//logMessage(logFile, "FAILED: Error clicking send button for %s: %v. Attempting fallback Enter key.\n", number, errClick)
			msgBox, findErr := wd.FindElement(selenium.ByCSSSelector, messageBoxSelector)
			if findErr == nil {
				if errEnter := msgBox.SendKeys(selenium.EnterKey); errEnter == nil {
					logMessage(logFile, "SUCCESS (probably): Message sent to %s via Enter key after click failed.\n", number)
					time.Sleep(5 * time.Second)
					continue
				} else {
					logMessage(logFile, "FAILED: Fallback Enter key also failed to send to message box for %s: %v\n", number, errEnter)
				}
			}
			logMessage(logFile, "Skipping %s due to persistent send error.\n", number)
			time.Sleep(2 * time.Second)
			continue
		}

		logMessage(logFile, "SUCCESS: Message sent to: %s\n", number)
		time.Sleep(5 * time.Second)
	}

	logMessage(logFile, "\nAll message sending attempts completed.\n")
	logMessage(logFile, "Please check WhatsApp Web to confirm sent messages and the log file for details.\n")
	logMessage(logFile, "Browser window will close in 10 seconds...\n")
	time.Sleep(10 * time.Second)
	logMessage(logFile, "Application finished.\n")
}
