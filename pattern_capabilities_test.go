package blink1

import (
	"encoding/json"
	"reflect"
	"strings"
	"testing"

	hid "github.com/b1ug/gid"
)

func patternTestMetadata(model string) *hid.DeviceInfo {
	if model == ProfileModelMK3 {
		return &hid.DeviceInfo{VendorID: b1VendorID, ProductID: b1ProductID, VersionNumber: 0x0101, Product: "blink(1) mk3", SerialNumber: "30001234"}
	}
	return &hid.DeviceInfo{VendorID: b1VendorID, ProductID: b1ProductID, VersionNumber: 2, Product: "blink(1) mk2", SerialNumber: "20005601"}
}

func assertPatternInt(t *testing.T, fact PatternIntFact, value int, known bool) {
	t.Helper()
	if !known {
		if fact.Knowledge != PatternKnowledgeUnknown || fact.Value != nil || fact.Evidence.Basis != PatternEvidenceUnknown || len(fact.Evidence.Sources) != 0 {
			t.Fatalf("unknown integer fact became a claim: %#v", fact)
		}
		return
	}
	if fact.Knowledge != PatternKnowledgeKnown || fact.Value == nil || *fact.Value != value || fact.Evidence.Basis == PatternEvidenceUnknown || len(fact.Evidence.Sources) == 0 {
		t.Fatalf("integer fact = %#v, want known %d", fact, value)
	}
}

func assertPatternBool(t *testing.T, fact PatternBoolFact, value bool, known bool) {
	t.Helper()
	if !known {
		if fact.Knowledge != PatternKnowledgeUnknown || fact.Value != nil || fact.Evidence.Basis != PatternEvidenceUnknown || len(fact.Evidence.Sources) != 0 {
			t.Fatalf("unknown boolean fact became a negative/positive claim: %#v", fact)
		}
		return
	}
	if fact.Knowledge != PatternKnowledgeKnown || fact.Value == nil || *fact.Value != value || fact.Evidence.Basis == PatternEvidenceUnknown || len(fact.Evidence.Sources) == 0 {
		t.Fatalf("boolean fact = %#v, want known %v", fact, value)
	}
}

func assertPatternUnverified(t *testing.T, got PatternCapabilities) {
	t.Helper()
	want := []string{"ram-upload", "ram-readback", "ram-playback", "flash-commit", "power-cycle-restore", "startup-write", "notes-write"}
	if got.HardwareVerified || len(got.Qualification) != len(want) {
		t.Fatalf("source facts promoted hardware qualification: %#v", got)
	}
	for i, operation := range want {
		entry := got.Qualification[i]
		if entry.Operation != operation || entry.Status != PatternQualificationUnverified || len(entry.Evidence) != 0 {
			t.Fatalf("operation qualification = %#v", entry)
		}
	}
}

func TestPatternCapabilitiesFirmwareApplicability(t *testing.T) {
	for _, tc := range []struct {
		name, model, applicability string
		firmware, reference        int
		ram, flash, usable         int
		usableKnown                bool
		startupKnown               bool
	}{
		{"mk2 unknown firmware", ProfileModelMK2, PatternApplicabilityModelOnly, 0, 205, 32, 16, 0, false, false},
		{"mk2 connected204 is not205", ProfileModelMK2, PatternApplicabilityModelOnly, 204, 205, 32, 16, 0, false, false},
		{"mk2 public205", ProfileModelMK2, PatternApplicabilityFirmwareMatched, 205, 205, 32, 16, 16, true, false},
		{"mk2 documented206 unresolved layout", ProfileModelMK2, PatternApplicabilityUnresolved, 206, 205, 32, 16, 0, false, true},
		{"mk2 later209 not silently matched", ProfileModelMK2, PatternApplicabilityUnresolved, 209, 205, 32, 16, 0, false, true},
		{"mk3 unknown firmware", ProfileModelMK3, PatternApplicabilityModelOnly, 0, 304, 32, 32, 0, false, true},
		{"mk3 older303 is not304", ProfileModelMK3, PatternApplicabilityModelOnly, 303, 304, 32, 32, 0, false, true},
		{"mk3 public304", ProfileModelMK3, PatternApplicabilityFirmwareMatched, 304, 304, 32, 32, 32, true, true},
		{"mk3 later305 is not304", ProfileModelMK3, PatternApplicabilityModelOnly, 305, 304, 32, 32, 0, false, true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			metadata := patternTestMetadata(tc.model)
			before := *metadata
			got := IdentifyPatternCapabilities(metadata, tc.firmware)
			if !reflect.DeepEqual(got.Identity, IdentifyDeviceProfile(metadata, tc.firmware)) || !reflect.DeepEqual(before, *metadata) {
				t.Fatal("capability resolution changed or replaced metadata-only identity")
			}
			if got.Applicability != tc.applicability || got.ReferenceFirmwareVersion == nil || *got.ReferenceFirmwareVersion != tc.reference || got.Identity.FirmwareVersion != tc.firmware {
				t.Fatalf("source/target firmware applicability conflated: %#v", got)
			}
			assertPatternInt(t, got.Storage.RAMLines, tc.ram, true)
			assertPatternInt(t, got.Storage.FlashLines, tc.flash, true)
			assertPatternInt(t, got.Storage.UsableFlashLines, tc.usable, tc.usableKnown)
			assertPatternInt(t, got.Storage.StartupReservedLines, 0, tc.usableKnown)
			assertPatternBool(t, got.Startup.Supported, true, tc.startupKnown)
			assertPatternBool(t, got.ArbitraryUserRAMSupported, false, false)
			if tc.model == ProfileModelMK2 {
				assertPatternBool(t, got.Notes.Supported, false, false)
				assertPatternInt(t, got.Notes.Slots, 0, false)
				assertPatternInt(t, got.Notes.BytesPerSlot, 0, false)
				assertPatternBool(t, got.Notes.WritePersistsPage, false, false)
				assertPatternBool(t, got.Startup.WritePersistsCurrentPattern, false, false)
				assertPatternBool(t, got.FlashCommitPersistsStartup, false, false)
				if tc.startupKnown && got.Startup.Supported.Evidence.Basis != PatternEvidenceDocumented {
					t.Fatal("startup documentation was relabeled as firmware implementation")
				}
			} else {
				assertPatternBool(t, got.Notes.Supported, true, true)
				assertPatternInt(t, got.Notes.Slots, 20, true)
				assertPatternInt(t, got.Notes.BytesPerSlot, 50, true)
				assertPatternBool(t, got.Notes.WritePersistsPage, true, true)
				assertPatternBool(t, got.Startup.WritePersistsCurrentPattern, true, true)
				assertPatternBool(t, got.FlashCommitPersistsStartup, true, true)
				if got.Identity.USBDeviceVersion != 0x0101 {
					t.Fatal("raw MK3 USB bcdDevice was changed to generation/firmware")
				}
			}
			assertPatternUnverified(t, got)
		})
	}
}

func TestPatternCapabilitiesUnknownIdentityNeverGainsFacts(t *testing.T) {
	mk2, mk3 := patternTestMetadata(ProfileModelMK2), patternTestMetadata(ProfileModelMK3)
	changed := func(change func(*hid.DeviceInfo)) *hid.DeviceInfo {
		copy := *mk2
		change(&copy)
		return &copy
	}
	for _, tc := range []struct {
		name     string
		metadata *hid.DeviceInfo
		firmware int
	}{
		{"nil metadata", nil, 205},
		{"foreign VID", changed(func(i *hid.DeviceInfo) { i.VendorID++ }), 205},
		{"foreign PID", changed(func(i *hid.DeviceInfo) { i.ProductID++ }), 205},
		{"missing serial", changed(func(i *hid.DeviceInfo) { i.SerialNumber = "" }), 205},
		{"invalid serial", changed(func(i *hid.DeviceInfo) { i.SerialNumber = "2zzzzzzz" }), 205},
		{"unknown prefix", changed(func(i *hid.DeviceInfo) { i.SerialNumber = "40001234" }), 0},
		{"missing product", changed(func(i *hid.DeviceInfo) { i.Product = "" }), 205},
		{"conflicting product", changed(func(i *hid.DeviceInfo) { i.Product = "blink(1) mk3" }), 205},
		{"missing USB version", changed(func(i *hid.DeviceInfo) { i.VersionNumber = 0 }), 205},
		{"conflicting USB version", changed(func(i *hid.DeviceInfo) { i.VersionNumber = 0x0101 }), 205},
		{"negative firmware", mk2, -1},
		{"non digit minor 217", mk2, 217},
		{"unsupported wire minor210", mk2, 210},
		{"wrong firmware generation", mk2, 304},
		{"raw bcdDevice is not firmware", mk3, 0x0101},
		{"large firmware", mk2, 0x7fffffff},
	} {
		t.Run(tc.name, func(t *testing.T) {
			got := IdentifyPatternCapabilities(tc.metadata, tc.firmware)
			if got.Identity.Model != ProfileModelUnknown || got.ReferenceFirmwareVersion != nil || got.Applicability != PatternApplicabilityUnresolved {
				t.Fatalf("unidentified target obtained reference applicability: %#v", got)
			}
			for _, fact := range []PatternIntFact{got.Storage.RAMLines, got.Storage.FlashLines, got.Storage.UsableFlashLines, got.Storage.StartupReservedLines, got.Notes.Slots, got.Notes.BytesPerSlot} {
				assertPatternInt(t, fact, 0, false)
			}
			for _, fact := range []PatternBoolFact{got.Notes.Supported, got.Notes.WritePersistsPage, got.Startup.Supported, got.Startup.WritePersistsCurrentPattern, got.FlashCommitPersistsStartup, got.ArbitraryUserRAMSupported} {
				assertPatternBool(t, fact, false, false)
			}
			assertPatternUnverified(t, got)
		})
	}
}

func TestPatternCapabilitiesOwnAllMutableState(t *testing.T) {
	info := patternTestMetadata(ProfileModelMK3)
	a := IdentifyPatternCapabilities(info, 304)
	before, err := json.Marshal(a)
	if err != nil {
		t.Fatal(err)
	}
	*a.Storage.RAMLines.Value = 999
	a.Storage.RAMLines.Evidence.Sources[0].ID = "changed"
	if *a.Storage.FlashLines.Value != 32 || a.Storage.FlashLines.Evidence.Sources[0].ID == "changed" {
		t.Fatal("different facts share mutable pointer/evidence state")
	}
	*a.ReferenceFirmwareVersion = 999
	*a.Notes.Supported.Value = false
	a.Notes.Slots.Evidence.Sources[0].URL = "changed"
	a.Startup.Supported.Evidence.Sources[0].URL = "changed"
	a.Identity.Capabilities[0].HardwareVerified = true
	a.Warnings[0] = "changed"
	a.Qualification[0].Status = "changed"
	a.Qualification[0].Evidence = append(a.Qualification[0].Evidence, "changed")
	b := IdentifyPatternCapabilities(info, 304)
	after, err := json.Marshal(b)
	if err != nil || string(before) != string(after) {
		t.Fatalf("caller mutation changed later resolution: %s; %v", after, err)
	}
	if a.Identity.FirmwareVersion != 304 || info.VersionNumber != 0x0101 {
		t.Fatal("reference fields share input/legacy identity state")
	}
}

func TestPatternCapabilitiesJSONKeepsUnknownZeroAndFalseDistinct(t *testing.T) {
	knownZero := IdentifyPatternCapabilities(patternTestMetadata(ProfileModelMK3), 304).Storage.StartupReservedLines
	unknown := IdentifyPatternCapabilities(patternTestMetadata(ProfileModelMK2), 206).Storage.StartupReservedLines
	knownFalse := knownPatternBool(false, PatternEvidenceDocumented, PatternCapabilitySource{ID: "test-negative-evidence", URL: "https://example.invalid/fixed-test"})
	data, err := json.Marshal(struct {
		KnownZero  PatternIntFact  `json:"known_zero"`
		Unknown    PatternIntFact  `json:"unknown"`
		KnownFalse PatternBoolFact `json:"known_false"`
	}{knownZero, unknown, knownFalse})
	if err != nil {
		t.Fatal(err)
	}
	var got map[string]struct {
		Knowledge string      `json:"knowledge"`
		Value     interface{} `json:"value"`
	}
	if err := json.Unmarshal(data, &got); err != nil {
		t.Fatal(err)
	}
	if got["known_zero"].Knowledge != PatternKnowledgeKnown || got["known_zero"].Value != float64(0) || got["known_false"].Value != false || got["unknown"].Knowledge != PatternKnowledgeUnknown || got["unknown"].Value != nil {
		t.Fatalf("unknown/zero/false were flattened or omitted: %s", data)
	}
}

func TestPatternCapabilitiesEvidenceUsesImmutablePublicSources(t *testing.T) {
	for _, model := range []string{ProfileModelMK2, ProfileModelMK3} {
		firmware := 206
		if model == ProfileModelMK3 {
			firmware = 304
		}
		profile := IdentifyPatternCapabilities(patternTestMetadata(model), firmware)
		data, err := json.Marshal(profile)
		if err != nil {
			t.Fatal(err)
		}
		var value interface{}
		if err := json.Unmarshal(data, &value); err != nil {
			t.Fatal(err)
		}
		var inspect func(interface{})
		inspect = func(value interface{}) {
			switch item := value.(type) {
			case []interface{}:
				for _, child := range item {
					inspect(child)
				}
			case map[string]interface{}:
				if url, ok := item["url"].(string); ok {
					if !strings.HasPrefix(url, "https://github.com/todbot/") || strings.Contains(url, "/main/") || strings.Contains(url, "/master/") || !strings.Contains(url, "#L") {
						t.Fatalf("mutable/unattributed source URL: %q", url)
					}
				}
				for _, child := range item {
					inspect(child)
				}
			}
		}
		inspect(value)
	}
}
