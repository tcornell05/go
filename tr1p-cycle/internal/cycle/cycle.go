package cycle

import (
	"fmt"
	"log"
	"os/exec"
	"strconv"
	"strings"
	"sync"
	"time"
)

type CycleList struct {
	mu      sync.Mutex
	head    *CycleItem
	current *CycleItem
	track   map[int]*CycleItem
}

type CycleItem struct {
	next    *CycleItem
	prev    *CycleItem
	title   string
	process int
	name    string
	appName string
}

func NewCycleList() *CycleList {
	return &CycleList{track: make(map[int]*CycleItem)}
}

func (c *CycleList) Add(title string) {
	c.mu.Lock()
	defer c.mu.Unlock()

	windowID, windowTitle, processID, err := getActiveWindow()
	if err != nil {
		log.Printf("Failed to get active window: %v\n", err)
		return
	}

	pid, _ := strconv.Atoi(processID)
	if _, exists := c.track[pid]; exists {
		log.Printf("Item already in list: %s\n", windowTitle)
		return
	}

	appName, err := getApplicationName(pid)
	if err != nil {
		log.Printf("Failed to get application name: %v\n", err)
		appName = "Unknown"
	}

	newItem := &CycleItem{title: windowTitle, process: pid, name: windowTitle, appName: appName}

	if c.head == nil {
		c.head = newItem
		c.current = newItem
		newItem.next = newItem
		newItem.prev = newItem
	} else {
		newItem.prev = c.current
		newItem.next = c.current.next
		c.current.next.prev = newItem
		c.current.next = newItem
	}

	c.track[pid] = newItem
	log.Printf("Added item: %s (Window ID: %s, App: %s)\n", windowTitle, windowID, appName)
}

func (c *CycleList) Remove(title string) {
	c.mu.Lock()
	defer c.mu.Unlock()

	_, windowTitle, processID, err := getActiveWindow()
	if err != nil {
		log.Printf("Failed to get active window: %v\n", err)
		return
	}

	pid, _ := strconv.Atoi(processID)
	if curr, exists := c.track[pid]; exists {
		c.removeItem(curr)
		log.Printf("Removed item: %s (Process ID: %d)\n", windowTitle, pid)
	} else {
		log.Printf("%v is not in the cycle list, so it won't be removed!\n", windowTitle)
	}
}

func (c *CycleList) removeItem(item *CycleItem) {
	if item.next == item {
		c.head = nil
		c.current = nil
	} else {
		item.prev.next = item.next
		item.next.prev = item.prev
		if item == c.head {
			c.head = item.next
		}
		if item == c.current {
			c.current = item.next
		}
	}
	delete(c.track, item.process)
}

func (c *CycleList) FocusNext() {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.current == nil {
		log.Println("No items in the list.")
		return
	}

	startItem := c.current
	for {
		c.current = c.current.next
		if c.isWindowOpen(c.current) {
			break
		}
		if c.current == startItem {
			log.Println("No open windows in the list.")
			return
		}
	}

	windowID, err := findWindowID(c.current.process)
	if err != nil {
		log.Printf("Could not find window for item: %s\n", c.current.title)
		return
	}

	err = exec.Command("wmctrl", "-ia", windowID).Run()
	if err != nil {
		log.Printf("Error focusing window: %s\n", err)
	} else {
		log.Printf("Focused on window: %s\n", c.current.title)
	}
}

func (c *CycleList) isWindowOpen(item *CycleItem) bool {
	_, err := findWindowID(item.process)
	return err == nil
}

func (c *CycleList) MonitorActiveWindow() {
	for {
		_, _, processID, err := getActiveWindow()
		if err != nil {
			time.Sleep(1 * time.Second)
			continue
		}
		pid, _ := strconv.Atoi(processID)
		c.mu.Lock()
		if item, exists := c.track[pid]; exists {
			c.current = item
			log.Printf("Updated current to window: %s (Process ID: %d, App: %s)\n", item.title, item.process, item.appName)
		}
		c.mu.Unlock()

		time.Sleep(500 * time.Millisecond)
	}
}

func (c *CycleList) GetItems() []CycleItem {
	c.mu.Lock()
	defer c.mu.Unlock()

	var items []CycleItem
	if c.current != nil {
		item := c.current
		for {
			items = append(items, *item)
			item = item.next
			if item == c.current {
				break
			}
		}
	}

	return items
}

func (c *CycleList) GetCurrentItem() *CycleItem {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.current
}

func (c *CycleList) PrintItems() {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.head == nil {
		fmt.Println("The cycle list is empty.")
		return
	}

	fmt.Println("Current items in the cycle list:")
	current := c.head
	for {
		fmt.Printf("Item: %s (Process ID: %d, App: %s)\n", current.name, current.process, current.appName)
		current = current.next
		if current == c.head {
			break
		}
	}
}

func findProcessID(processName string) (int, error) {
	out, err := exec.Command("pgrep", "-f", processName).Output()
	if err != nil {
		return 0, fmt.Errorf("error running pgrep command: %v", err)
	}

	lines := strings.Split(strings.TrimSpace(string(out)), "\n")
	if len(lines) > 0 {
		return strconv.Atoi(lines[0])
	}

	return 0, fmt.Errorf("no process found for name: %s", processName)
}

func findWindowID(processID int) (string, error) {
	out, err := exec.Command("wmctrl", "-lp").Output()
	if err != nil {
		return "", fmt.Errorf("error running wmctrl command: %v", err)
	}

	lines := strings.Split(string(out), "\n")
	for _, line := range lines {
		if strings.Contains(line, fmt.Sprintf(" %d ", processID)) {
			fields := strings.Fields(line)
			if len(fields) > 0 {
				return fields[0], nil
			}
		}
	}

	return "", fmt.Errorf("no window found for process ID: %d", processID)
}

func getActiveWindow() (string, string, string, error) {
	idBytes, err := exec.Command("xdotool", "getactivewindow").Output()
	if err != nil {
		return "", "", "", fmt.Errorf("could not get active window ID: %v", err)
	}
	windowID := strings.TrimSpace(string(idBytes))

	titleBytes, err := exec.Command("xdotool", "getwindowname", windowID).Output()
	if err != nil {
		return "", "", "", fmt.Errorf("could not get window title: %v", err)
	}
	windowTitle := strings.TrimSpace(string(titleBytes))

	pidBytes, err := exec.Command("xdotool", "getwindowpid", windowID).Output()
	if err != nil {
		return "", "", "", fmt.Errorf("could not get PID for window: %v", err)
	}
	pid := strings.TrimSpace(string(pidBytes))

	return windowID, windowTitle, pid, nil
}

func getApplicationName(pid int) (string, error) {
	cmd := exec.Command("ps", "-p", strconv.Itoa(pid), "-o", "comm=")
	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("error getting application name: %v", err)
	}
	return strings.TrimSpace(string(output)), nil
}
