package automation

// AutomationDraft is the result of importing a recorder export. It is not yet
// persisted; the frontend shows it in the editor and the user explicitly saves.
type AutomationDraft struct {
	Title       string
	Description string
	Steps       []Step
	Warnings    []string // non-fatal: unknown step types, dropped no-op steps, etc.
}

// RecorderImporter parses a single recorder format into an AutomationDraft.
// First-cut implementation: ChromeRecorderImporter (Chrome DevTools Recorder JSON).
// Future: SeleniumIDEImporter, PlaywrightCodegenImporter — same interface, no
// orchestration changes.
type RecorderImporter interface {
	Format() string                                                 // identifier, e.g. "chrome-devtools-recorder"
	Detect(raw []byte) bool                                         // sniff to auto-pick the importer
	Import(raw []byte, platformID string) (AutomationDraft, error)
}

// PickImporter returns the first registered importer whose Detect returns true,
// or nil if none match. New importers register themselves via init().
func PickImporter(raw []byte) RecorderImporter {
	for _, imp := range registeredImporters {
		if imp.Detect(raw) {
			return imp
		}
	}
	return nil
}

var registeredImporters []RecorderImporter

// RegisterImporter is called from init() in implementation files.
func RegisterImporter(i RecorderImporter) {
	registeredImporters = append(registeredImporters, i)
}
