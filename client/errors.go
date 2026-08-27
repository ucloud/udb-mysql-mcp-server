package client

import (
	"errors"
	"fmt"
	"strings"

	uerr "github.com/ucloud/ucloud-sdk-go/ucloud/error"
)

// UpstreamError indicates a UCloud API failure without leaking credentials.
type UpstreamError struct {
	Action  string
	RetCode int
	Message string
	Cause   error
}

func (e *UpstreamError) Error() string {
	if e.RetCode != 0 {
		return fmt.Sprintf("ucloud %s failed: ret_code=%d: %s", e.Action, e.RetCode, e.Message)
	}
	return fmt.Sprintf("ucloud %s failed: %s", e.Action, e.Message)
}

func (e *UpstreamError) Unwrap() error { return e.Cause }

// UpstreamRetCode returns the upstream RetCode when present.
func (e *UpstreamError) UpstreamRetCode() int { return e.RetCode }

// NotFoundError indicates the requested UDB instance does not exist.
type NotFoundError struct {
	InstanceID string
}

func (e *NotFoundError) Error() string {
	return fmt.Sprintf("udb: instance %q not found", e.InstanceID)
}

// BackupNotFoundError indicates the requested backup does not exist within a complete scan.
type BackupNotFoundError struct {
	BackupID int
}

func (e *BackupNotFoundError) Error() string {
	return fmt.Sprintf("udb: backup %d not found", e.BackupID)
}

// ScanIncompleteReason classifies why backup lookup could not complete.
type ScanIncompleteReason string

const (
	ScanIncompleteReasonLimit        ScanIncompleteReason = "limit"
	ScanIncompleteReasonCountChanged ScanIncompleteReason = "count_changed"
	ScanIncompleteReasonEmptyPage    ScanIncompleteReason = "empty_page"
	ScanIncompleteReasonIncomplete   ScanIncompleteReason = "incomplete"
)

// ScanIncompleteError indicates backup lookup stopped before the full API result set was scanned.
type ScanIncompleteError struct {
	BackupID      int
	APITotalCount int
	ScannedCount  int
	Reason        ScanIncompleteReason
}

func (e *ScanIncompleteError) Error() string {
	switch e.Reason {
	case ScanIncompleteReasonCountChanged:
		return fmt.Sprintf(
			"udb: backup %d lookup incomplete: upstream TotalCount changed during scan (scanned %d of %d)",
			e.BackupID, e.ScannedCount, e.APITotalCount,
		)
	case ScanIncompleteReasonEmptyPage:
		return fmt.Sprintf(
			"udb: backup %d lookup incomplete: empty page before scan covered stable TotalCount %d (scanned %d)",
			e.BackupID, e.APITotalCount, e.ScannedCount,
		)
	case ScanIncompleteReasonLimit:
		return fmt.Sprintf(
			"udb: backup %d lookup incomplete: scan cap reached (scanned %d of %d matching backups)",
			e.BackupID, e.ScannedCount, e.APITotalCount,
		)
	default:
		return fmt.Sprintf(
			"udb: backup %d lookup incomplete: scanned %d of %d matching backups",
			e.BackupID, e.ScannedCount, e.APITotalCount,
		)
	}
}

// InvalidInputError indicates caller-supplied parameters are invalid.
type InvalidInputError struct {
	Field   string
	Message string
}

func (e *InvalidInputError) Error() string {
	return fmt.Sprintf("udb: invalid %s: %s", e.Field, e.Message)
}

func mapSDKError(action string, err error) error {
	var serverErr uerr.ServerError
	if errors.As(err, &serverErr) {
		return &UpstreamError{
			Action:  action,
			RetCode: serverErr.Code(),
			Message: serverErr.Message(),
			Cause:   err,
		}
	}
	return &UpstreamError{Action: action, Message: genericUpstreamFailureMessage, Cause: err}
}

// projectScopeHintRetCodes lists upstream ret codes that usually mean the
// resource belongs to a different project than the one the call was scoped to.
// The upstream messages alone ("describe error", "permission error & resource
// not exist") do not point at the project mismatch.
var projectScopeHintRetCodes = map[int]struct{}{7009: {}, 7048: {}}

// enhanceProjectScopeError appends actionable guidance when an upstream failure
// is typically caused by addressing a resource that lives in another project.
// It returns the original error unchanged in every other case.
func enhanceProjectScopeError(err error, reqCtx CallContext) error {
	var upstream *UpstreamError
	if !errors.As(err, &upstream) {
		return err
	}
	if _, ok := projectScopeHintRetCodes[upstream.RetCode]; !ok {
		return err
	}
	if strings.Contains(upstream.Message, "hint:") {
		return err
	}
	scope := fmt.Sprintf("%q", reqCtx.ProjectID)
	if strings.TrimSpace(reqCtx.ProjectID) == "" {
		scope = "the default project"
	}
	hint := fmt.Sprintf(
		"; hint: resource not found under project %s — it may belong to another project; pass that project's project_id (candidates via udb_mysql_list_projects)",
		scope,
	)
	return &UpstreamError{
		Action:  upstream.Action,
		RetCode: upstream.RetCode,
		Message: upstream.Message + hint,
		Cause:   err,
	}
}
