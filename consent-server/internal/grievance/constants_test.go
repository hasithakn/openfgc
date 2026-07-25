/*
 * Copyright (c) 2026, WSO2 LLC. (https://www.wso2.com).
 * Licensed under the Apache License, Version 2.0.
 */

package grievance

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestPriorityMappingsRoundTrip(t *testing.T) {
	require.Len(t, priorityDBToAPI, 4)
	for dbValue, apiValue := range priorityDBToAPI {
		require.Equal(t, dbValue, priorityAPIToDB[apiValue], "priority round-trip mismatch for %q", dbValue)
	}
}

func TestStatusMappingsRoundTrip(t *testing.T) {
	require.Len(t, statusDBToAPI, 5)
	for dbValue, apiValue := range statusDBToAPI {
		require.Equal(t, dbValue, statusAPIToDB[apiValue], "status round-trip mismatch for %q", dbValue)
	}
}

func TestEntryTypeMappingsCoverAllDBValues(t *testing.T) {
	require.Len(t, entryTypeDBToAPI, 5)
}

func TestVisibilityMappingsRoundTrip(t *testing.T) {
	for dbValue, apiValue := range visibilityDBToAPI {
		require.Equal(t, dbValue, visibilityAPIToDB[apiValue])
	}
}

func TestActorRoleMappingsRoundTrip(t *testing.T) {
	for dbValue, apiValue := range actorRoleDBToAPI {
		require.Equal(t, dbValue, actorRoleAPIToDB[apiValue])
	}
}

func TestCategoryToPriorityDBHasTenCategoriesMappingToKnownPriorities(t *testing.T) {
	require.Len(t, categoryToPriorityDB, 10)
	for category, priority := range categoryToPriorityDB {
		_, ok := priorityDBToAPI[priority]
		require.True(t, ok, "category %q maps to unrecognized DB priority %q", category, priority)
	}
}

func TestNextStatusesDBOnlyReferencesKnownStatuses(t *testing.T) {
	for from, tos := range nextStatusesDB {
		_, ok := statusDBToAPI[from]
		require.True(t, ok, "transition table has unknown from-status %q", from)
		for _, to := range tos {
			_, ok := statusDBToAPI[to]
			require.True(t, ok, "transition table has unknown to-status %q for from-status %q", to, from)
		}
	}
}

func TestResolvedStatusIsTerminalExceptViaReopenException(t *testing.T) {
	require.Empty(t, nextStatusesDB[statusDBResolved])
}

func TestWaitingOnDPOReachableOnlyViaAutoTransition(t *testing.T) {
	// No entry in the transition table targets Waiting on DPO — it is only ever entered
	// via the auto-transition-on-reply exception in service.go's PostMessage.
	for from, tos := range nextStatusesDB {
		for _, to := range tos {
			require.NotEqual(t, statusDBWaitingOnDPO, to, "transition table should not target Waiting on DPO from %q", from)
		}
	}
}
