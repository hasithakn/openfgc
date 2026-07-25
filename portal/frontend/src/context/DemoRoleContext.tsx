/*
 * Copyright (c) 2026, WSO2 LLC. (https://www.wso2.com).
 *
 * WSO2 LLC. licenses this file to you under the Apache License,
 * Version 2.0 (the "License"); you may not use this file except
 * in compliance with the License.
 * You may obtain a copy of the License at
 *
 * http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing,
 * software distributed under the License is distributed on an
 * "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY
 * KIND, either express or implied.  See the License for the
 * specific language governing permissions and limitations
 * under the License.
 */

import { useMemo, useState } from 'react'
import { isAuthEnabled } from '../utils/authClient'
import { useScopes } from './ScopeContext'
import { DemoRoleContext, type DemoRole } from './demoRole'

interface DemoRoleProviderProps {
  children: React.ReactNode
}

function DemoRoleProvider({ children }: DemoRoleProviderProps): React.JSX.Element {
  const { isGrievanceOfficer } = useScopes()
  const [manualRole, setManualRole] = useState<DemoRole | null>(null)
  const canOverride = !isAuthEnabled()
  const effectiveRole: DemoRole = isGrievanceOfficer ? 'grievanceOfficer' : 'dataPrincipal'
  const role = canOverride && manualRole ? manualRole : effectiveRole

  const value = useMemo(
    () => ({
      role,
      canOverride,
      setRole: (nextRole: DemoRole) => {
        if (canOverride) {
          setManualRole(nextRole)
        }
      },
    }),
    [role, canOverride],
  )

  return <DemoRoleContext.Provider value={value}>{children}</DemoRoleContext.Provider>
}

export default DemoRoleProvider
