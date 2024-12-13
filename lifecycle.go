package openwechat

import (
	"io"
	"log"
	"os"
	"path/filepath"

	"github.com/eatmoreapple/openwechat"
	"github.com/gonebot-dev/gonebot/logging"
	"github.com/rs/zerolog"
)

// Create storage for hot login, returns true if it succeeds.
func tryCreateStorageFolder() bool {
	// Get current directory
	folderPath, err := filepath.Abs(filepath.Dir(os.Args[0]))
	if err != nil {
		logging.Logf(zerolog.FatalLevel, "OpenWechat", "Failed to get current directory: %s", err.Error())
		return false
	}
	folderPath = filepath.Join(folderPath, ".openwechat-hotlogin/")
	// Create folder if not exists
	if _, err = os.Stat(folderPath); os.IsNotExist(err) {
		err = os.MkdirAll(folderPath, os.ModePerm)
		if err != nil {
			logging.Logf(zerolog.FatalLevel, "OpenWechat", "Failed to create storage folder: %s", err.Error())
			return false
		}
		gitignore, err := os.Create(filepath.Join(folderPath, ".gitignore"))
		if err != nil {
			logging.Logf(zerolog.FatalLevel, "OpenWechat", "Failed to create .gitignore file: %s", err.Error())
			return false
		}
		_, err = gitignore.Write([]byte("*"))
		if err != nil {
			logging.Logf(zerolog.FatalLevel, "OpenWechat", "Failed to write to .gitignore file: %s", err.Error())
			return false
		}
		gitignore.Close()
	}
	return true
}

func start() {
	log.SetOutput(io.Discard)
	// Create a bot
	bot := openwechat.DefaultBot(openwechat.Desktop)

	// Set QRcode callback
	bot.UUIDCallback = openwechat.PrintlnQrcodeUrl
	// Set QRcode scan callback
	bot.ScanCallBack = func(body openwechat.CheckLoginResponse) {
		logging.Logf(zerolog.InfoLevel, "OpenWechat", "QRcode scanned, please confirm on your phone.")
	}
	// Register message handler
	bot.MessageHandler = receiveHandler

	logging.Logf(zerolog.InfoLevel, "OpenWechat", "Initializing, please scan the QRcode to login...")
	// Try set hot login storage and login
	if tryCreateStorageFolder() {
		reloadStorage := openwechat.NewFileHotReloadStorage(
			filepath.Join(
				filepath.Dir(os.Args[0]),
				".openwechat-hotlogin/storage.json",
			),
		)
		// Login
		if err := bot.PushLogin(reloadStorage, openwechat.NewRetryLoginOption()); err != nil {
			logging.Logf(zerolog.FatalLevel, "OpenWechat", "Login failed: %s", err.Error())
			logging.Logf(zerolog.InfoLevel, "OpenWechat", "Aborting...")
			return
		}
		reloadStorage.Close()
	} else if err := bot.Login(); err != nil {
		logging.Logf(zerolog.FatalLevel, "OpenWechat", "Login failed: %s", err.Error())
		logging.Logf(zerolog.InfoLevel, "OpenWechat", "Aborting...")
		return
	}
	logging.Logf(zerolog.InfoLevel, "OpenWechat", "Login successful!")
	Self, _ = bot.GetCurrentUser()

	go sendHandler()
	go actionHandler()

	// Block until finish
	bot.Block()
}

func finalize() {
	if Self.Bot() != nil && Self.Bot().Alive() {
		Self.Bot().Logout()
	}
	Self = nil
	logging.Log(zerolog.InfoLevel, "OpenWechat", "Shutdown complete!")
}
