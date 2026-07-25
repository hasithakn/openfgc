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

import {
  AppShell,
  Box,
  ColorSchemeToggle,
  Header,
  Tooltip,
  ToggleButton,
  ToggleButtonGroup,
} from '@wso2/oxygen-ui'
import { ShieldCheck, User } from '@wso2/oxygen-ui-icons-react'
import { useState } from 'react'
import { useTranslation } from 'react-i18next'
import { Outlet, useNavigate } from 'react-router-dom'
import type { DemoRole } from '../../../context/demoRole'
import { ROLE_HOME_PATH } from '../../../context/demoRole'
import { useDemoRole } from '../../../hooks/useDemoRole'
import AppSidebar from '../sidebar/AppSidebar'
import UserProfileMenu from './UserProfileMenu'

function MainLayout(): React.JSX.Element {
  const { t } = useTranslation('common')
  const navigate = useNavigate()
  const { role, setRole, canOverride } = useDemoRole()
  const [isSidebarCollapsed, setIsSidebarCollapsed] = useState<boolean>(false)

  return (
    <AppShell collapseOnSelectOnMobile>
      <AppShell.Navbar>
        <Header>
          <Header.Toggle
            collapsed={isSidebarCollapsed}
            onToggle={() => {
              setIsSidebarCollapsed((previous) => !previous)
            }}
          />
          <Header.Brand>
            <Header.BrandLogo>
              <Box
                component="img"
                src="/wso2-logo.png"
                alt=""
                aria-hidden="true"
                sx={{
                  width: 28,
                  height: 28,
                  objectFit: 'contain',
                }}
              />
            </Header.BrandLogo>
            <Header.BrandTitle>{t('app.title')}</Header.BrandTitle>
          </Header.Brand>
          <Header.Spacer />
          <Header.Actions>
            {canOverride ? (
              <ToggleButtonGroup
                size="small"
                exclusive
                value={role}
                onChange={(_, nextRole: DemoRole | null) => {
                  if (nextRole && nextRole !== role) {
                    setRole(nextRole)
                    navigate(ROLE_HOME_PATH[nextRole])
                  }
                }}
                aria-label={t('layout.demoRoleAriaLabel')}
              >
                <ToggleButton value="dataPrincipal" aria-label={t('layout.demoRoleDataPrincipal')}>
                  <Tooltip title={t('layout.demoRoleDataPrincipal')}>
                    <Box sx={{ display: 'flex', alignItems: 'center' }}>
                      <User size={16} />
                    </Box>
                  </Tooltip>
                </ToggleButton>
                <ToggleButton
                  value="grievanceOfficer"
                  aria-label={t('layout.demoRoleGrievanceOfficer')}
                >
                  <Tooltip title={t('layout.demoRoleGrievanceOfficer')}>
                    <Box sx={{ display: 'flex', alignItems: 'center' }}>
                      <ShieldCheck size={16} />
                    </Box>
                  </Tooltip>
                </ToggleButton>
              </ToggleButtonGroup>
            ) : null}
            <ColorSchemeToggle />
            <UserProfileMenu />
          </Header.Actions>
        </Header>
      </AppShell.Navbar>

      <AppShell.Sidebar>
        <AppSidebar collapsed={isSidebarCollapsed} />
      </AppShell.Sidebar>

      <AppShell.Main>
        <Box
          sx={{
            width: '100%',
            maxWidth: 'none',
            flex: 1,
          }}
        >
          <Outlet />
        </Box>
      </AppShell.Main>
    </AppShell>
  )
}

export default MainLayout
