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

import { useEffect } from 'react'
import { Navigate, Route, Routes } from 'react-router-dom'
import MainLayout from './components/layout/main-layout/MainLayout'
import DemoRoleProvider from './context/DemoRoleContext'
import { ROLE_HOME_PATH } from './context/demoRole'
import { ScopeProvider, useScopes } from './context/ScopeContext'
import ElementDetailsPage from './features/catalog/ElementDetailsPage'
import ElementListPage from './features/catalog/ElementListPage'
import PurposeDetailsPage from './features/catalog/PurposeDetailsPage'
import PurposeListPage from './features/catalog/PurposeListPage'
import ConsentDetailsPage from './features/consent-registry/ConsentDetailsPage'
import ConsentRegistryPage from './features/consent-registry/ConsentRegistryPage'
import DashboardPage from './features/dashboard/DashboardPage'
import GrievanceCaseDetailPage from './features/grievance-management/GrievanceCaseDetailPage'
import GrievanceQueuePage from './features/grievance-management/GrievanceQueuePage'
import GrievanceDetailPage from './features/grievances/GrievanceDetailPage'
import GrievanceListPage from './features/grievances/GrievanceListPage'
import { useDemoRole } from './hooks/useDemoRole'
import { isAuthenticated, login } from './utils/authClient'

function AuthenticationGate({
  children,
}: {
  children: React.JSX.Element
}): React.JSX.Element | null {
  const authenticated = isAuthenticated()

  useEffect(() => {
    if (!authenticated) {
      login()
    }
  }, [authenticated])

  return authenticated ? children : null
}

function ScopeGuard({
  canAccess,
  redirectTo,
  children,
}: {
  canAccess: boolean
  redirectTo: string
  children: React.JSX.Element
}): React.JSX.Element {
  const { isLoading } = useScopes()
  if (isLoading) {
    return <></>
  }
  return canAccess ? children : <Navigate to={redirectTo} replace />
}

function AppRoutes(): React.JSX.Element {
  const { canReadElements, canReadPurposes, canReadGrievancesSelf } = useScopes()
  const { role } = useDemoRole()

  const canAccessGrievances = role === 'dataPrincipal' && canReadGrievancesSelf
  const canAccessGrievanceManagement = role === 'grievanceOfficer'
  const grievancesRedirect = canAccessGrievanceManagement
    ? ROLE_HOME_PATH.grievanceOfficer
    : '/consents'
  const grievanceManagementRedirect = canAccessGrievances
    ? ROLE_HOME_PATH.dataPrincipal
    : '/consents'

  return (
    <Routes>
      <Route element={<MainLayout />}>
        <Route path="/dashboard" element={<DashboardPage />} />
        <Route path="/consents" element={<ConsentRegistryPage />} />
        <Route path="/consents/:id" element={<ConsentDetailsPage />} />
        <Route
          path="/purposes"
          element={
            <ScopeGuard canAccess={canReadPurposes} redirectTo="/consents">
              <PurposeListPage />
            </ScopeGuard>
          }
        />
        <Route
          path="/purposes/:id"
          element={
            <ScopeGuard canAccess={canReadPurposes} redirectTo="/consents">
              <PurposeDetailsPage />
            </ScopeGuard>
          }
        />
        <Route
          path="/elements"
          element={
            <ScopeGuard canAccess={canReadElements} redirectTo="/consents">
              <ElementListPage />
            </ScopeGuard>
          }
        />
        <Route
          path="/elements/:id"
          element={
            <ScopeGuard canAccess={canReadElements} redirectTo="/consents">
              <ElementDetailsPage />
            </ScopeGuard>
          }
        />
        <Route
          path="/grievances"
          element={
            <ScopeGuard canAccess={canAccessGrievances} redirectTo={grievancesRedirect}>
              <GrievanceListPage />
            </ScopeGuard>
          }
        />
        <Route
          path="/grievances/:id"
          element={
            <ScopeGuard canAccess={canAccessGrievances} redirectTo={grievancesRedirect}>
              <GrievanceDetailPage />
            </ScopeGuard>
          }
        />
        <Route
          path="/grievance-management"
          element={
            <ScopeGuard
              canAccess={canAccessGrievanceManagement}
              redirectTo={grievanceManagementRedirect}
            >
              <GrievanceQueuePage />
            </ScopeGuard>
          }
        />
        <Route
          path="/grievance-management/:id"
          element={
            <ScopeGuard
              canAccess={canAccessGrievanceManagement}
              redirectTo={grievanceManagementRedirect}
            >
              <GrievanceCaseDetailPage />
            </ScopeGuard>
          }
        />
        <Route path="*" element={<Navigate to="/consents" replace />} />
      </Route>
    </Routes>
  )
}

function App(): React.JSX.Element {
  return (
    <AuthenticationGate>
      <ScopeProvider>
        <DemoRoleProvider>
          <AppRoutes />
        </DemoRoleProvider>
      </ScopeProvider>
    </AuthenticationGate>
  )
}

export default App
