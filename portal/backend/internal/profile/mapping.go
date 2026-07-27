/*
 * Copyright (c) 2026, WSO2 LLC. (https://www.wso2.com).
 * Licensed under the Apache License, Version 2.0.
 */

package profile

import "strings"

// EmailAttr is the BFF-facing shape of a SCIM email entry.
type EmailAttr struct {
	Value   string `json:"value"`
	Type    string `json:"type,omitempty"`
	Primary bool   `json:"primary,omitempty"`
}

// ProfileResponse is the flattened PII view returned to the frontend: identity, email,
// and age — deliberately kept to just these three, nothing else.
type ProfileResponse struct {
	Username      string      `json:"username"`
	GivenName     string      `json:"givenName"`
	FamilyName    string      `json:"familyName"`
	FormattedName string      `json:"formattedName"`
	Emails        []EmailAttr `json:"emails"`
	Age           *int        `json:"age,omitempty"`
}

// NameUpdate is submitted as a single unit since SCIM's "name" is one complex attribute —
// patching givenName without familyName (or vice versa) would otherwise drop the other.
type NameUpdate struct {
	GivenName  string `json:"givenName"`
	FamilyName string `json:"familyName"`
}

// ProfileUpdateRequest carries only the sections the client actually wants to change.
// UserName is intentionally absent: it's read-only from this view.
type ProfileUpdateRequest struct {
	Name   *NameUpdate  `json:"name,omitempty"`
	Emails *[]EmailAttr `json:"emails,omitempty"`
	Age    *int         `json:"age,omitempty"`
}

func toProfileResponse(u *ScimUser) ProfileResponse {
	resp := ProfileResponse{
		Username:      u.UserName,
		GivenName:     u.Name.GivenName,
		FamilyName:    u.Name.FamilyName,
		FormattedName: u.Name.Formatted,
		Emails:        make([]EmailAttr, 0, len(u.Emails)),
		Age:           u.Age,
	}
	for _, e := range u.Emails {
		resp.Emails = append(resp.Emails, EmailAttr{Value: e.Value, Type: e.Type, Primary: e.Primary})
	}
	return resp
}

// buildPatchOperations turns the sections present in req into one "replace" PatchOperation
// each, so a caller submitting e.g. only a changed email doesn't clobber name/age.
// ageAttributePath is the deployment-configured SCIM path for age (Service.AgeAttributePath());
// an Age update is silently dropped if it's not configured, since there's nowhere to send it.
func buildPatchOperations(req ProfileUpdateRequest, ageAttributePath string) []PatchOperation {
	var ops []PatchOperation

	if req.Name != nil {
		formatted := strings.TrimSpace(req.Name.GivenName + " " + req.Name.FamilyName)
		ops = append(ops, PatchOperation{
			Op:   "replace",
			Path: "name",
			Value: ScimName{
				GivenName:  req.Name.GivenName,
				FamilyName: req.Name.FamilyName,
				Formatted:  formatted,
			},
		})
	}
	if req.Emails != nil {
		emails := make([]ScimEmail, 0, len(*req.Emails))
		for _, e := range *req.Emails {
			emails = append(emails, ScimEmail{Value: e.Value, Type: e.Type, Primary: e.Primary})
		}
		ops = append(ops, PatchOperation{Op: "replace", Path: "emails", Value: emails})
	}
	if req.Age != nil && ageAttributePath != "" {
		ops = append(ops, PatchOperation{Op: "replace", Path: ageAttributePath, Value: *req.Age})
	}

	return ops
}
