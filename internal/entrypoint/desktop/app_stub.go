//go:build !(windows || (linux && cgo) || (darwin && cgo))

package desktop

import "fmt"

// Run is a stub when CGO/WebKit is unavailable.
func Run(core CoreAPI) error {
	_ = core
	return fmt.Errorf("runtgine desktop requires CGO and a native WebView (Linux: libgtk-4-dev libwebkitgtk-6.0-dev)")
}
