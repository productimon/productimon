package nlp

import (
	"log"
	"testing"
)

type TestData struct {
	app   string
	label string
}

func TestWikipedia(t *testing.T) {
	testdata := []TestData{
		{app: "Adobe Photoshop 2020", label: "Raster graphics editor"},
		{app: "App Store", label: "Digital distribution"},
		{app: "Chrome", label: "Web browser"},
		{app: "Cisco Webex Meetings", label: "Voice over IP"},
		{app: "Code", label: "Source code editor"},
		{app: "Find My", label: "Location aware"},
		{app: "Firefox", label: "Web browser"},
		{app: "Gedit", label: "Text editor"},
		{app: "Gnome-terminal", label: "Terminal Emulator"},
		{app: "Google Docs", label: "Collaborative software"},
		{app: "Messenger", label: "Instant messaging"},
		{app: "Microsoft PowerPoint", label: "Presentation program"},
		{app: "Microsoft Teams", label: "Collaborative software"},
		{app: "Signal", label: "voice calling"},
		{app: "Slack", label: "Collaborative software"},
		{app: "Spotify", label: "Music streaming"},
		{app: "Totem", label: "Media player"},
		{app: "VMWare Fusion", label: "Hypervisor"},
		{app: "VNC Viewer", label: "Remote administration"},
		{app: "Windows Explorer", label: "Shell"},
		{app: "Wireshark", label: "Packet analyzer"},
		{app: "XCode", label: "Integrated development environment"},
		{app: "iTerm2", label: "Terminal emulator"},
		{app: "ida", label: "Disassembler"},
		{app: "vim", label: "Text editor"},
		{app: "vlc", label: "Media player"},
		{app: "wHaTsApP", label: "Instant messaging"},
		{app: "zoom.us", label: "Videoconferencing"},
	}
	for _, td := range testdata {
		label := wikipediaLabel(td.app)
		if label != td.label {
			t.Fatalf("app %s: expected %s, got %s", td.app, td.label, label)
		}
		log.Printf("OK - %s - %s", td.app, td.label)
	}
}
