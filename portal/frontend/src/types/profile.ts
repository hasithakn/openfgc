/*
 * Copyright (c) 2026, WSO2 LLC. (https://www.wso2.com).
 * Licensed under the Apache License, Version 2.0.
 */

export interface EmailAttr {
  value: string
  type?: string
  primary?: boolean
}

export interface PhoneAttr {
  value: string
  type?: string
  primary?: boolean
}

export interface AddressAttr {
  formatted?: string
  streetAddress?: string
  locality?: string
  region?: string
  postalCode?: string
  country?: string
  type?: string
  primary?: boolean
}

export interface ProfileResponse {
  username: string
  givenName: string
  familyName: string
  formattedName: string
  emails: EmailAttr[]
  phoneNumbers: PhoneAttr[]
  addresses: AddressAttr[]
}

export interface ProfileNameUpdate {
  givenName: string
  familyName: string
}

export interface ProfileUpdateRequest {
  name?: ProfileNameUpdate
  emails?: EmailAttr[]
  phoneNumbers?: PhoneAttr[]
  addresses?: AddressAttr[]
}
