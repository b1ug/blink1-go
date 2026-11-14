package blink1

import hid "github.com/b1ug/gid"

const (
	PatternKnowledgeUnknown = "unknown"
	PatternKnowledgeKnown   = "known"

	PatternEvidenceUnknown        = "unknown"
	PatternEvidenceSourceInferred = "source-inferred"
	PatternEvidenceDocumented     = "documented"

	PatternApplicabilityUnresolved      = "unresolved"
	PatternApplicabilityModelOnly       = "model-only"
	PatternApplicabilityFirmwareMatched = "firmware-matched"

	PatternQualificationUnverified = "unverified"
)

// PatternCapabilitySource identifies a fixed public source, not a hardware test.
type PatternCapabilitySource struct {
	ID  string `json:"id"`
	URL string `json:"url"`
}

// PatternCapabilityEvidence describes why a reference fact is known. An empty
// source set and the unknown basis do not establish that a feature is absent.
type PatternCapabilityEvidence struct {
	Basis   string                    `json:"basis"`
	Sources []PatternCapabilitySource `json:"sources"`
}

// PatternIntFact is a nullable integer fact. Unknown has knowledge "unknown"
// and value null; a known zero is distinct. The value describes the referenced
// source profile, whose applicability to the supplied firmware is separate.
type PatternIntFact struct {
	Knowledge string                    `json:"knowledge"`
	Value     *int                      `json:"value"`
	Evidence  PatternCapabilityEvidence `json:"evidence"`
}

// PatternBoolFact is a nullable boolean fact. Unknown is not a negative support
// claim: a known false value requires evidence just as a known true value does.
type PatternBoolFact struct {
	Knowledge string                    `json:"knowledge"`
	Value     *bool                     `json:"value"`
	Evidence  PatternCapabilityEvidence `json:"evidence"`
}

// PatternStorageCapabilities separates addressable RAM, physical flash, and the
// usable ordinary Pattern range. Unknown reserved lines are not counted as zero.
type PatternStorageCapabilities struct {
	RAMLines             PatternIntFact `json:"ram_lines"`
	FlashLines           PatternIntFact `json:"flash_lines"`
	UsableFlashLines     PatternIntFact `json:"usable_flash_lines"`
	StartupReservedLines PatternIntFact `json:"startup_reserved_lines"`
}

// PatternNotesCapabilities describes a separate data facility. Notes do not add
// playable Pattern rows. BytesPerSlot counts bytes, not Unicode characters.
type PatternNotesCapabilities struct {
	Supported         PatternBoolFact `json:"supported"`
	Slots             PatternIntFact  `json:"slots"`
	BytesPerSlot      PatternIntFact  `json:"bytes_per_slot"`
	WritePersistsPage PatternBoolFact `json:"write_persists_page"`
}

// PatternStartupCapabilities makes startup's collateral persistence explicit.
type PatternStartupCapabilities struct {
	Supported                   PatternBoolFact `json:"supported"`
	WritePersistsCurrentPattern PatternBoolFact `json:"write_persists_current_pattern"`
}

// PatternOperationQualification is independent of source facts and firmware
// matching. This resolver only emits unverified entries with no test evidence.
// A successful ordinary color command is not evidence for these operations.
type PatternOperationQualification struct {
	Operation string   `json:"operation"`
	Status    string   `json:"status"`
	Evidence  []string `json:"evidence"`
}

// PatternCapabilities is a fresh, caller-owned reference snapshot. Facts may be
// source-inferred or documented without being qualified on this device. Neither
// a known capacity nor firmware-matched applicability authorizes a device write.
// The resolver performs no enumeration, open, USB query, playback or storage I/O.
type PatternCapabilities struct {
	Identity                   DeviceProfile                   `json:"identity"`
	ReferenceFirmwareVersion   *int                            `json:"reference_firmware"`
	Applicability              string                          `json:"applicability"`
	Storage                    PatternStorageCapabilities      `json:"storage"`
	Notes                      PatternNotesCapabilities        `json:"notes"`
	Startup                    PatternStartupCapabilities      `json:"startup"`
	FlashCommitPersistsStartup PatternBoolFact                 `json:"flash_commit_persists_startup"`
	ArbitraryUserRAMSupported  PatternBoolFact                 `json:"arbitrary_user_ram_supported"`
	Qualification              []PatternOperationQualification `json:"qualification"`
	HardwareVerified           bool                            `json:"hardware_verified"`
	Warnings                   []string                        `json:"warnings"`
}

// IdentifyPatternCapabilities combines metadata-only identification with fixed
// public Pattern capability evidence. firmwareVersion uses GetVersion's integer
// encoding (major*100+minor, e.g. 204 or 304); pass 0 when it is unknown. The USB
// descriptor's bcdDevice is a separate field, not this argument.
//
// Firmware-matched means the supplied version matches the reviewed source's
// self-reported version. It is not hardware verification or proof that the
// supplied metadata came from a genuine device. Other recognized firmware keeps
// model-level reference capacity facts while application remains unresolved.
//
// MK2 firmware 206+ has a documented startup operation but an unresolved reserved
// layout; its usable flash range is never guessed as either 15 or 16 lines.
// This additive helper does not change existing Device or Controller behavior.
func IdentifyPatternCapabilities(info *hid.DeviceInfo, firmwareVersion int) PatternCapabilities {
	profile := unknownPatternCapabilities(IdentifyDeviceProfile(info, firmwareVersion))
	switch profile.Identity.Model {
	case ProfileModelMK2:
		referenceFirmware := 205
		profile.ReferenceFirmwareVersion = &referenceFirmware
		profile.Applicability = PatternApplicabilityModelOnly
		capacity := PatternCapabilitySource{
			ID:  "mk2-pattern-capacity-fw205",
			URL: "https://github.com/todbot/blink1/blob/101abe3599c661c4dd4d1204b4d50fb592293143/hardware/firmware_mk2/main.c#L140-L169",
		}
		persistence := PatternCapabilitySource{
			ID:  "mk2-pattern-flash-fw205",
			URL: "https://github.com/todbot/blink1/blob/101abe3599c661c4dd4d1204b4d50fb592293143/hardware/firmware_mk2/main.c#L330-L340",
		}
		profile.Storage.RAMLines = knownPatternInt(32, PatternEvidenceSourceInferred, capacity)
		profile.Storage.FlashLines = knownPatternInt(16, PatternEvidenceSourceInferred, capacity, persistence)
		if firmwareVersion == 205 {
			profile.Applicability = PatternApplicabilityFirmwareMatched
			profile.Storage.UsableFlashLines = knownPatternInt(16, PatternEvidenceSourceInferred, capacity, persistence)
			profile.Storage.StartupReservedLines = knownPatternInt(0, PatternEvidenceSourceInferred, capacity, persistence)
		} else if firmwareVersion >= 206 {
			profile.Applicability = PatternApplicabilityUnresolved
			profile.Startup.Supported = knownPatternBool(true, PatternEvidenceDocumented, PatternCapabilitySource{
				ID:  "mk2-startup-206-plus-documentation",
				URL: "https://github.com/todbot/blink1-tool/blob/f1547015b7f98d7f2e0b5f3b8bc32388c4015585/blink1-lib.h#L341-L352",
			})
			profile.Warnings = append(profile.Warnings, "MK2 startup is documented for firmware206+, but its implementation, reserved lines and persistence effects are unresolved.")
		} else {
			profile.Warnings = append(profile.Warnings, "Capacity evidence comes from public firmware205; it does not qualify the supplied firmware.")
		}
		profile.Warnings = append(profile.Warnings, "No Notes or arbitrary-user-RAM support was established; absence from reviewed code is not proof of unsupported hardware.")
	case ProfileModelMK3:
		referenceFirmware := 304
		profile.ReferenceFirmwareVersion = &referenceFirmware
		profile.Applicability = PatternApplicabilityModelOnly
		capacity := PatternCapabilitySource{
			ID:  "mk3-pattern-notes-capacity-fw304",
			URL: "https://github.com/todbot/blink1mk3/blob/ccf04e26726a4a41ef9689653ae12baa73dbc456/firmware/firmware-v30x/main.c#L189-L218",
		}
		persistence := PatternCapabilitySource{
			ID:  "mk3-persistent-pages-fw304",
			URL: "https://github.com/todbot/blink1mk3/blob/ccf04e26726a4a41ef9689653ae12baa73dbc456/firmware/firmware-v30x/main.c#L441-L508",
		}
		startup := PatternCapabilitySource{
			ID:  "mk3-startup-auto-persistence-fw304",
			URL: "https://github.com/todbot/blink1mk3/blob/ccf04e26726a4a41ef9689653ae12baa73dbc456/firmware/firmware-v30x/main.c#L1169-L1190",
		}
		notes := PatternCapabilitySource{
			ID:  "mk3-notes-page-persistence-fw304",
			URL: "https://github.com/todbot/blink1mk3/blob/ccf04e26726a4a41ef9689653ae12baa73dbc456/firmware/firmware-v30x/main.c#L1292-L1307",
		}
		profile.Storage.RAMLines = knownPatternInt(32, PatternEvidenceSourceInferred, capacity)
		profile.Storage.FlashLines = knownPatternInt(32, PatternEvidenceSourceInferred, capacity, persistence)
		profile.Notes.Supported = knownPatternBool(true, PatternEvidenceSourceInferred, notes)
		profile.Notes.Slots = knownPatternInt(20, PatternEvidenceSourceInferred, capacity)
		profile.Notes.BytesPerSlot = knownPatternInt(50, PatternEvidenceSourceInferred, capacity)
		profile.Notes.WritePersistsPage = knownPatternBool(true, PatternEvidenceSourceInferred, notes, persistence)
		profile.Startup.Supported = knownPatternBool(true, PatternEvidenceSourceInferred, startup)
		profile.Startup.WritePersistsCurrentPattern = knownPatternBool(true, PatternEvidenceSourceInferred, startup, persistence)
		profile.FlashCommitPersistsStartup = knownPatternBool(true, PatternEvidenceSourceInferred, persistence)
		if firmwareVersion == 304 {
			profile.Applicability = PatternApplicabilityFirmwareMatched
			profile.Storage.UsableFlashLines = knownPatternInt(32, PatternEvidenceSourceInferred, capacity, persistence)
			profile.Storage.StartupReservedLines = knownPatternInt(0, PatternEvidenceSourceInferred, capacity, startup)
		} else {
			profile.Warnings = append(profile.Warnings, "Capacity and persistence evidence comes from public firmware304; it does not qualify the supplied firmware.")
		}
		profile.Warnings = append(profile.Warnings, "Notes are separate byte storage, not extra playable Pattern lines; startup and Notes writes can persist complete affected pages.")
	default:
		profile.Warnings = append(profile.Warnings, "Metadata is missing, unfamiliar or conflicting; no reference capability profile was selected.")
	}
	return profile
}

func unknownPatternCapabilities(identity DeviceProfile) PatternCapabilities {
	profile := PatternCapabilities{
		Identity: identity, Applicability: PatternApplicabilityUnresolved,
		Storage: PatternStorageCapabilities{
			RAMLines: unknownPatternInt(), FlashLines: unknownPatternInt(),
			UsableFlashLines: unknownPatternInt(), StartupReservedLines: unknownPatternInt(),
		},
		Notes: PatternNotesCapabilities{
			Supported: unknownPatternBool(), Slots: unknownPatternInt(),
			BytesPerSlot: unknownPatternInt(), WritePersistsPage: unknownPatternBool(),
		},
		Startup: PatternStartupCapabilities{
			Supported: unknownPatternBool(), WritePersistsCurrentPattern: unknownPatternBool(),
		},
		FlashCommitPersistsStartup: unknownPatternBool(), ArbitraryUserRAMSupported: unknownPatternBool(),
		Qualification: []PatternOperationQualification{},
		Warnings:      []string{"Reference facts and firmware matching are not deployment permission or physical hardware qualification."},
	}
	for _, operation := range []string{"ram-upload", "ram-readback", "ram-playback", "flash-commit", "power-cycle-restore", "startup-write", "notes-write"} {
		profile.Qualification = append(profile.Qualification, PatternOperationQualification{
			Operation: operation, Status: PatternQualificationUnverified, Evidence: []string{},
		})
	}
	return profile
}

func unknownPatternInt() PatternIntFact {
	return PatternIntFact{Knowledge: PatternKnowledgeUnknown, Evidence: PatternCapabilityEvidence{Basis: PatternEvidenceUnknown, Sources: []PatternCapabilitySource{}}}
}

func unknownPatternBool() PatternBoolFact {
	return PatternBoolFact{Knowledge: PatternKnowledgeUnknown, Evidence: PatternCapabilityEvidence{Basis: PatternEvidenceUnknown, Sources: []PatternCapabilitySource{}}}
}

func knownPatternInt(value int, basis string, sources ...PatternCapabilitySource) PatternIntFact {
	return PatternIntFact{Knowledge: PatternKnowledgeKnown, Value: &value, Evidence: PatternCapabilityEvidence{Basis: basis, Sources: append([]PatternCapabilitySource{}, sources...)}}
}

func knownPatternBool(value bool, basis string, sources ...PatternCapabilitySource) PatternBoolFact {
	return PatternBoolFact{Knowledge: PatternKnowledgeKnown, Value: &value, Evidence: PatternCapabilityEvidence{Basis: basis, Sources: append([]PatternCapabilitySource{}, sources...)}}
}
