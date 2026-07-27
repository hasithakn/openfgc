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

// PhoneAttr is the BFF-facing shape of a SCIM phone number entry.
type PhoneAttr struct {
	Value   string `json:"value"`
	Type    string `json:"type,omitempty"`
	Primary bool   `json:"primary,omitempty"`
}

// AddressAttr is the BFF-facing shape of a SCIM address entry, flattened to the single
// "formatted" line — nothing here needs street/city/region/postal broken out separately.
type AddressAttr struct {
	Formatted string `json:"formatted"`
	Type      string `json:"type,omitempty"`
	Primary   bool   `json:"primary,omitempty"`
}

// ProfileResponse is the flattened PII view returned to the frontend: identity, nickname,
// email, phone, address, age, and birthday — deliberately kept to just these, nothing else.
type ProfileResponse struct {
	Username      string        `json:"username"`
	GivenName     string        `json:"givenName"`
	FamilyName    string        `json:"familyName"`
	FormattedName string        `json:"formattedName"`
	NickName      string        `json:"nickName,omitempty"`
	Emails        []EmailAttr   `json:"emails"`
	PhoneNumbers  []PhoneAttr   `json:"phoneNumbers"`
	Addresses     []AddressAttr `json:"addresses"`
	Age           *int          `json:"age,omitempty"`
	Birthday      string        `json:"birthday,omitempty"`
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
	Name         *NameUpdate    `json:"name,omitempty"`
	NickName     *string        `json:"nickName,omitempty"`
	Emails       *[]EmailAttr   `json:"emails,omitempty"`
	PhoneNumbers *[]PhoneAttr   `json:"phoneNumbers,omitempty"`
	Addresses    *[]AddressAttr `json:"addresses,omitempty"`
	Age          *int           `json:"age,omitempty"`
	Birthday     *string        `json:"birthday,omitempty"`
}

func toProfileResponse(u *ScimUser) ProfileResponse {
	resp := ProfileResponse{
		Username:      u.UserName,
		GivenName:     u.Name.GivenName,
		FamilyName:    u.Name.FamilyName,
		FormattedName: u.Name.Formatted,
		NickName:      u.NickName,
		Emails:        make([]EmailAttr, 0, len(u.Emails)),
		PhoneNumbers:  make([]PhoneAttr, 0, len(u.PhoneNumbers)),
		Addresses:     make([]AddressAttr, 0, len(u.Addresses)),
		Age:           u.Age,
		Birthday:      u.Birthday,
	}
	for _, e := range u.Emails {
		resp.Emails = append(resp.Emails, EmailAttr{Value: e.Value, Type: e.Type, Primary: e.Primary})
	}
	for _, p := range u.PhoneNumbers {
		resp.PhoneNumbers = append(resp.PhoneNumbers, PhoneAttr{Value: p.Value, Type: p.Type, Primary: p.Primary})
	}
	for _, a := range u.Addresses {
		resp.Addresses = append(resp.Addresses, AddressAttr{Formatted: a.Formatted, Type: a.Type, Primary: a.Primary})
	}

	// "formatted" is a separate SCIM sub-attribute from givenName/familyName — a PATCH
	// that only sets givenName/familyName (as callers commonly do) leaves it blank, so
	// derive it from those before falling further back to userName below.
	if strings.TrimSpace(resp.FormattedName) == "" {
		resp.FormattedName = strings.TrimSpace(resp.GivenName + " " + resp.FamilyName)
	}
	// A freshly-registered user has no SCIM name/emails set at all yet. userName is
	// always present and, for this deployment, is itself the user's email — so it's a
	// safe fallback for both rather than showing the caller a blank identity.
	if resp.FormattedName == "" {
		resp.FormattedName = u.UserName
	}
	if len(resp.Emails) == 0 && strings.Contains(u.UserName, "@") {
		resp.Emails = append(resp.Emails, EmailAttr{Value: u.UserName, Primary: true})
	}

	return resp
}

// buildPatchOperations turns the sections present in req into one "replace" PatchOperation
// each, so a caller submitting e.g. only a changed email doesn't clobber name/age.
// ageAttributePath/birthdayAttributePath are the deployment-configured SCIM paths
// (Service.AgeAttributePath()/BirthdayAttributePath()); an update is silently dropped if
// its attribute isn't configured, since there's nowhere to send it.
func buildPatchOperations(req ProfileUpdateRequest, ageAttributePath, birthdayAttributePath string) []PatchOperation {
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
	if req.NickName != nil {
		ops = append(ops, PatchOperation{Op: "replace", Path: "nickName", Value: *req.NickName})
	}
	if req.Emails != nil {
		emails := make([]ScimEmail, 0, len(*req.Emails))
		for _, e := range *req.Emails {
			emails = append(emails, ScimEmail{Value: e.Value, Type: e.Type, Primary: e.Primary})
		}
		ops = append(ops, PatchOperation{Op: "replace", Path: "emails", Value: emails})
	}
	if req.PhoneNumbers != nil {
		phones := make([]ScimPhoneNumber, 0, len(*req.PhoneNumbers))
		for _, p := range *req.PhoneNumbers {
			phones = append(phones, ScimPhoneNumber{Value: p.Value, Type: p.Type, Primary: p.Primary})
		}
		ops = append(ops, PatchOperation{Op: "replace", Path: "phoneNumbers", Value: phones})
	}
	if req.Addresses != nil {
		addresses := make([]ScimAddress, 0, len(*req.Addresses))
		for _, a := range *req.Addresses {
			addresses = append(addresses, ScimAddress{Formatted: a.Formatted, Type: a.Type, Primary: a.Primary})
		}
		ops = append(ops, PatchOperation{Op: "replace", Path: "addresses", Value: addresses})
	}
	if req.Age != nil && ageAttributePath != "" {
		ops = append(ops, PatchOperation{Op: "replace", Path: ageAttributePath, Value: *req.Age})
	}
	if req.Birthday != nil && birthdayAttributePath != "" {
		ops = append(ops, PatchOperation{Op: "replace", Path: birthdayAttributePath, Value: *req.Birthday})
	}

	return ops
}
