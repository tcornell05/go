package cycle

import (
	"fmt"
	"log"
	"strings"
	"sync"
	"time"

	"github.com/BurntSushi/xgb"
	"github.com/BurntSushi/xgb/xproto"
	"golang.design/x/hotkey"
)

type Keybinds struct {
	AddKeybind    string
	RemoveKeybind string
	CycleKeybind  string
}

type KeybindListener struct {
	cl           *CycleList
	preview      *Preview
	keybinds     Keybinds
	stopChan     chan struct{}
	addHotkey    *hotkey.Hotkey
	removeHotkey *hotkey.Hotkey
	cycleHotkey  *hotkey.Hotkey
	mu           sync.Mutex
	X            *xgb.Conn
}

func NewKeybindListener(cl *CycleList, keybinds Keybinds, preview *Preview) (*KeybindListener, error) {
	addModifiers, addKey := parseKeybind(keybinds.AddKeybind)
	removeModifiers, removeKey := parseKeybind(keybinds.RemoveKeybind)
	cycleModifiers, cycleKey := parseKeybind(keybinds.CycleKeybind)

	addHotkey := hotkey.New(addModifiers, addKey)
	removeHotkey := hotkey.New(removeModifiers, removeKey)
	cycleHotkey := hotkey.New(cycleModifiers, cycleKey)

	if err := addHotkey.Register(); err != nil {
		return nil, fmt.Errorf("failed to register add hotkey: %v", err)
	}
	if err := removeHotkey.Register(); err != nil {
		addHotkey.Unregister()
		return nil, fmt.Errorf("failed to register remove hotkey: %v", err)
	}
	if err := cycleHotkey.Register(); err != nil {
		addHotkey.Unregister()
		removeHotkey.Unregister()
		return nil, fmt.Errorf("failed to register cycle hotkey: %v", err)
	}

	X, err := xgb.NewConn()
	if err != nil {
		return nil, fmt.Errorf("failed to connect to X server: %v", err)
	}

	return &KeybindListener{
		cl:           cl,
		preview:      preview,
		keybinds:     keybinds,
		stopChan:     make(chan struct{}),
		addHotkey:    addHotkey,
		removeHotkey: removeHotkey,
		cycleHotkey:  cycleHotkey,
		X:            X,
	}, nil
}

func (kl *KeybindListener) Listen() {
	var cycleActive bool
	var lastCycleTime time.Time

	log.Println("Starting KeybindListener")

	altCheckTicker := time.NewTicker(50 * time.Millisecond)
	defer altCheckTicker.Stop()

	for {
		select {
		case <-kl.stopChan:
			log.Println("KeybindListener stopped")
			kl.X.Close()
			return
		case <-kl.addHotkey.Keydown():
			log.Println("Add hotkey pressed")
			handleAdd(kl.cl)
			kl.preview.updateContent()
		case <-kl.removeHotkey.Keydown():
			log.Println("Remove hotkey pressed")
			handleRemove(kl.cl)
			kl.preview.updateContent()
		case <-kl.cycleHotkey.Keydown():
			kl.mu.Lock()
			now := time.Now()
			if !cycleActive || now.Sub(lastCycleTime) > 200*time.Millisecond {
				log.Println("Cycle hotkey pressed")
				cycleActive = true
				lastCycleTime = now
				if !kl.preview.IsVisible() {
					log.Println("Cycle activated, showing preview")
					kl.preview.ShowPreview()
				}
				log.Printf("Focusing next window. Cycle active: %v", cycleActive)
				kl.cl.FocusNext()
				kl.preview.updateContent()
				kl.preview.ShowPreview()
			}
			kl.mu.Unlock()
		case <-altCheckTicker.C:
			kl.mu.Lock()
			altPressed := kl.isAltPressed()
			if cycleActive {
				if !altPressed {
					cycleActive = false
					log.Println("Alt released, hiding preview")
					kl.preview.HidePreview()
				} else {
					// Update preview to show current active window
					kl.preview.updateContent()
				}
			}
			kl.mu.Unlock()
		}
	}
}

func (kl *KeybindListener) isAltPressed() bool {
	state, err := xproto.QueryKeymap(kl.X).Reply()
	if err != nil {
		log.Printf("Failed to query keymap: %v", err)
		return false
	}

	// Check for left Alt (usually keycode 64) and right Alt (usually keycode 108)
	leftAlt := (state.Keys[64/8] & (1 << (64 % 8))) != 0
	rightAlt := (state.Keys[108/8] & (1 << (108 % 8))) != 0

	isPressed := leftAlt || rightAlt
	log.Printf("isAltPressed check result: %v", isPressed)
	return isPressed
}

func (kl *KeybindListener) Stop() {
	kl.mu.Lock()
	defer kl.mu.Unlock()

	if kl.stopChan != nil {
		close(kl.stopChan)
		kl.stopChan = nil
	}

	kl.addHotkey.Unregister()
	kl.removeHotkey.Unregister()
	kl.cycleHotkey.Unregister()
	kl.X.Close()
}

func parseKeybind(keybind string) ([]hotkey.Modifier, hotkey.Key) {
	modifiers := []hotkey.Modifier{}
	keys := strings.Split(keybind, "+")

	var key hotkey.Key
	for _, k := range keys {
		switch strings.ToLower(k) {
		case "alt":
			modifiers = append(modifiers, hotkey.Mod1)
		case "shift":
			modifiers = append(modifiers, hotkey.ModShift)
		case "ctrl":
			modifiers = append(modifiers, hotkey.ModCtrl)
		case "e":
			key = hotkey.KeyE
		case "d":
			key = hotkey.KeyD
		case "tab":
			key = tab
		}
	}

	return modifiers, key
}

func handleAdd(cl *CycleList) {
	windowID, processName, _, err := getActiveWindow()
	if err != nil {
		log.Printf("Failed to get active window: %v\n", err)
		return
	}
	cl.Add(processName)
	log.Printf("Added window: %s (ID: %s) to the cycle list.\n", processName, windowID)
	cl.PrintItems()
}

func handleRemove(cl *CycleList) {
	windowID, processName, _, err := getActiveWindow()
	if err != nil {
		log.Printf("Failed to get active window: %v\n", err)
		return
	}
	cl.Remove(processName)
	log.Printf("Removed window: %s (ID: %s) from the cycle list.\n", processName, windowID)
	cl.PrintItems()
}

const (
	tab = 0xff09
	alt = 0xffe9
)
