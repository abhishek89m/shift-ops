package main

import (
	"errors"
	"fmt"
	"slices"
	"strings"
)

var (
	errTaskNotFound      = errors.New("task not found")
	errInvalidTransition = errors.New("invalid task status transition")
	errValidation        = errors.New("invalid task update")
)

var allowedResolutionCodes = []string{
	"battery_swapped",
	"reparked",
	"checked_ok",
	"collected_for_repair",
	"unable_to_locate",
	"blocked_access",
	"duplicate_task",
	"other",
}

func validatePatchRequest(req taskPatchRequest) error {
	if req.Status == nil && req.Notes == nil && req.ChecklistState == nil {
		return fmt.Errorf("%w: include at least one mutable field", errValidation)
	}

	if req.Status != nil {
		if *req.Status != statusPending && *req.Status != statusInProgress && *req.Status != statusCompleted && *req.Status != statusSkipped {
			return fmt.Errorf("%w: status must be pending, in_progress, completed, or skipped", errValidation)
		}

		if *req.Status == statusCompleted || *req.Status == statusSkipped {
			if req.CompletedBy == nil || strings.TrimSpace(*req.CompletedBy) == "" {
				return fmt.Errorf("%w: completed_by is required for terminal updates", errValidation)
			}
			if req.ResolutionCode == nil || strings.TrimSpace(*req.ResolutionCode) == "" {
				return fmt.Errorf("%w: resolution_code is required for terminal updates", errValidation)
			}
			// Keep reporting/metrics clean by constraining terminal outcomes to known codes.
			if !slices.Contains(allowedResolutionCodes, strings.TrimSpace(*req.ResolutionCode)) {
				return fmt.Errorf("%w: resolution_code must be one of %s", errValidation, strings.Join(allowedResolutionCodes, ", "))
			}

			return nil
		}
	}

	if req.CompletedBy != nil || req.ResolutionCode != nil {
		return fmt.Errorf("%w: completed_by and resolution_code are only valid for terminal updates", errValidation)
	}

	return nil
}

func canTransition(from, to taskStatus) bool {
	allowed := map[taskStatus][]taskStatus{
		statusPending:    {statusInProgress},
		statusInProgress: {statusPending, statusCompleted, statusSkipped},
	}

	return slices.Contains(allowed[from], to)
}
