package model

// 观测集状态机：uploading → pending_attribution → pending_review → published → archived。
const (
	ObsUploading          = "uploading"
	ObsPendingAttribution = "pending_attribution"
	ObsPendingReview      = "pending_review"
	ObsPublished          = "published"
	ObsArchived           = "archived"
)

// 光谱峰状态机：raw → calibrated → suspected_artifact → excluded。
const (
	PeakRaw              = "raw"
	PeakCalibrated       = "calibrated"
	PeakSuspectedArtifact = "suspected_artifact"
	PeakExcluded         = "excluded"
)

// 归属候选状态机：generated → mutually_exclusive → sufficient → insufficient → rejected。
const (
	CandidateGenerated          = "generated"
	CandidateMutuallyExclusive  = "mutually_exclusive"
	CandidateSufficient         = "sufficient"
	CandidateInsufficient       = "insufficient"
	CandidateRejected           = "rejected"
)

// 归属版本状态机：draft → shared → frozen → superseded。
const (
	VersionDraft      = "draft"
	VersionShared     = "shared"
	VersionFrozen     = "frozen"
	VersionSuperseded = "superseded"
)

// 校准模型类型。
const (
	CalibModelOffset = "offset"
	CalibModelLinear = "linear"
)

// 校准状态。
const (
	CalibApplied    = "applied"
	CalibSuperseded = "superseded"
)

// 波长单位。
const (
	UnitAngstrom = "angstrom"
	UnitNanometer = "nm"
)

// TransitionLibVersion 元素跃迁库版本（代码内置库，随代码发布）。
const TransitionLibVersion = "2026.08"

// validObservationStatus 观测集合法状态集合。
var validObservationStatus = map[string]bool{
	ObsUploading:          true,
	ObsPendingAttribution: true,
	ObsPendingReview:      true,
	ObsPublished:          true,
	ObsArchived:           true,
}

// validPeakStatus 光谱峰合法状态集合。
var validPeakStatus = map[string]bool{
	PeakRaw:               true,
	PeakCalibrated:        true,
	PeakSuspectedArtifact: true,
	PeakExcluded:          true,
}

// validCandidateStatus 归属候选合法状态集合。
var validCandidateStatus = map[string]bool{
	CandidateGenerated:         true,
	CandidateMutuallyExclusive: true,
	CandidateSufficient:        true,
	CandidateInsufficient:      true,
	CandidateRejected:          true,
}

// validVersionStatus 归属版本合法状态集合。
var validVersionStatus = map[string]bool{
	VersionDraft:      true,
	VersionShared:     true,
	VersionFrozen:     true,
	VersionSuperseded: true,
}

// IsValidObservationStatus 报告状态是否为合法观测集状态。
func IsValidObservationStatus(s string) bool { return validObservationStatus[s] }

// IsValidPeakStatus 报告状态是否为合法光谱峰状态。
func IsValidPeakStatus(s string) bool { return validPeakStatus[s] }

// IsValidCandidateStatus 报告状态是否为合法候选状态。
func IsValidCandidateStatus(s string) bool { return validCandidateStatus[s] }

// IsValidVersionStatus 报告状态是否为合法版本状态。
func IsValidVersionStatus(s string) bool { return validVersionStatus[s] }

// observationTransitions 观测集允许的状态流转表。
var observationTransitions = map[string][]string{
	ObsUploading:          {ObsPendingAttribution, ObsArchived},
	ObsPendingAttribution: {ObsPendingReview, ObsUploading},
	ObsPendingReview:      {ObsPublished, ObsPendingAttribution},
	ObsPublished:          {ObsArchived},
	ObsArchived:           {},
}

// CanTransitionObservation 报告观测集能否从 from 流转到 to。
func CanTransitionObservation(from, to string) bool {
	for _, n := range observationTransitions[from] {
		if n == to {
			return true
		}
	}
	return false
}

// versionTransitions 归属版本允许的状态流转表。
var versionTransitions = map[string][]string{
	VersionDraft:      {VersionShared, VersionFrozen},
	VersionShared:     {VersionFrozen},
	VersionFrozen:     {},
	VersionSuperseded: {},
}

// CanTransitionVersion 报告归属版本能否从 from 流转到 to。
func CanTransitionVersion(from, to string) bool {
	for _, n := range versionTransitions[from] {
		if n == to {
			return true
		}
	}
	return false
}
