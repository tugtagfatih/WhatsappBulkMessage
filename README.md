# WhatsApp Bulk Messenger (Go + Selenium)

A simple and efficient application that automates sending messages to multiple phone numbers via WhatsApp Web.  
Built with Go and powered by Selenium WebDriver.

---

## 🚀 Features

- Send custom text messages to multiple contacts
- Persistent login using Chrome user profiles (scan QR code only once)
- Detailed logging with timestamps
- Cross-platform support: Windows and Linux
- Lightweight, no external servers required
- Precompiled binaries available for instant use

---

## 📦 Quick Start (Using Precompiled Binaries)

You can download the latest compiled binaries from the [Releases](../../releases) section.

Available for:
- **Windows** (`whatsapp-bulk-messenger.exe`)
- **Linux** (`whatsapp-bulk-messenger`)

### Required Files (Place in the Same Directory)

- `chromedriver.exe` (Windows) or `chromedriver` (Linux)  
  > [Download the correct ChromeDriver version](https://sites.google.com/chromium.org/driver/) based on your Chrome version.
- `numbers.txt` → List of phone numbers (one per line, including country code)
- `text.txt` → The message content to be sent

### Example Files

**numbers.txt**
```
905xxxxxxxxx
441234567890
4915123456789
```

**text.txt**
```
Hello! This is an automated message sent via WhatsApp Bulk Messenger.
```

---

## 🛠️ Building from Source (Optional)

If you prefer to build from source:

1. Install [Go](https://golang.org/dl/).
2. Clone this repository:
   ```bash
   git clone https://github.com/yourusername/whatsapp-bulk-messenger.git
   cd whatsapp-bulk-messenger
   ```
3. Download dependencies:
   ```bash
   go get github.com/tebeka/selenium
   ```
4. Build the project:
   ```bash
   go build -o whatsapp-bulk-messenger main.go
   ```

---

## ⚙️ How It Works

1. Launches Chrome via ChromeDriver.
2. Loads your WhatsApp Web session using a local profile (`whatsapp_profile_go/`).
3. Reads phone numbers from `numbers.txt`.
4. Reads the message from `text.txt`.
5. Sends the message to each phone number sequentially.
6. Saves detailed logs under the `logs/` directory.

---

## 📂 Project Structure

```
/your-folder
│
├── chromedriver.exe / chromedriver
├── whatsapp-bulk-messenger.exe / whatsapp-bulk-messenger
├── numbers.txt
├── text.txt
├── logs/
│   └── YYYYMMDD-HHMMSS.log
└── whatsapp_profile_go/
```

---

## 📋 Important Notes

- **ChromeDriver Version:**  
  Ensure ChromeDriver matches your installed Chrome browser version.

- **Persistent Login:**  
  After the first QR scan, the session is saved in `whatsapp_profile_go/`. You do not need to scan the QR code again unless you delete this folder.

- **Error Handling:**  
  If ChromeDriver fails to start, the application attempts fallback methods using system PATH.

---

## 🚠 Troubleshooting

| Problem | Solution |
|:---|:---|
| QR Code keeps appearing | Delete `whatsapp_profile_go/` and re-scan |
| Send button not found | WhatsApp Web UI might have changed; update the element selectors |
| Chromedriver compatibility issues | Download the correct version of ChromeDriver matching your browser |

---

## 📄 License

This project is licensed under the [MIT License](LICENSE).

---

## 🙏 Acknowledgements

- [Go Programming Language](https://golang.org/)
- [Selenium WebDriver](https://www.selenium.dev/)

---
