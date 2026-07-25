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

import { Sidebar } from '@wso2/oxygen-ui'
import {
  Blocks,
  Clock3,
  FileWarning,
  House,
  ListChecks,
  ShieldCheck,
  Target,
} from '@wso2/oxygen-ui-icons-react'
import { useTranslation } from 'react-i18next'
import { useLocation, useNavigate } from 'react-router-dom'
import { useScopes } from '../../../context/ScopeContext'
import { useDemoRole } from '../../../hooks/useDemoRole'

interface AppSidebarProps {
  collapsed: boolean
}

interface SidebarItem {
  id: string
  labelKey: string
  path: string
  icon: React.JSX.Element
}

const DASHBOARD_ITEMS: SidebarItem[] = [
  {
    id: 'dashboard',
    labelKey: 'sidebar.dashboard',
    path: '/dashboard',
    icon: <House size={18} />,
  },
]

const PENDING_CONSENT_ITEM: SidebarItem = {
  id: 'pending-consents',
  labelKey: 'sidebar.pendingConsents',
  path: '/consents?status=Pending',
  icon: <Clock3 size={18} />,
}

const GRIEVANCE_ITEMS: SidebarItem[] = [
  {
    id: 'my-grievances',
    labelKey: 'sidebar.myGrievances',
    path: '/grievances',
    icon: <FileWarning size={18} />,
  },
]

const GRIEVANCE_MANAGEMENT_ITEMS: SidebarItem[] = [
  {
    id: 'grievance-queue',
    labelKey: 'sidebar.grievanceQueue',
    path: '/grievance-management',
    icon: <ListChecks size={18} />,
  },
]

const PURPOSE_ITEM: SidebarItem = {
  id: 'purposes',
  labelKey: 'sidebar.purposes',
  path: '/purposes',
  icon: <Target size={18} />,
}

const ELEMENT_ITEM: SidebarItem = {
  id: 'elements',
  labelKey: 'sidebar.elements',
  path: '/elements',
  icon: <Blocks size={18} />,
}

function mapPathToMenuId(pathname: string, search: string): string {
  if (pathname.startsWith('/dashboard')) {
    return 'dashboard'
  }

  if (pathname.startsWith('/consents')) {
    const status = new URLSearchParams(search).get('status')

    if (status === 'Pending') {
      return 'pending-consents'
    }

    return 'all-consents'
  }

  if (pathname.startsWith('/grievances')) {
    return 'my-grievances'
  }

  if (pathname.startsWith('/grievance-management')) {
    return 'grievance-queue'
  }

  if (pathname.startsWith('/purposes')) {
    return 'purposes'
  }

  if (pathname.startsWith('/elements')) {
    return 'elements'
  }

  return 'dashboard'
}

function AppSidebar({ collapsed }: AppSidebarProps): React.JSX.Element {
  const { t } = useTranslation('common')
  const navigate = useNavigate()
  const location = useLocation()
  const { role } = useDemoRole()
  const { isAdmin, canReadPurposes, canReadElements, canReadGrievancesSelf } = useScopes()
  const showGrievances = role === 'dataPrincipal' && canReadGrievancesSelf
  const showGrievanceManagement = role === 'grievanceOfficer'

  const consentItem: SidebarItem = {
    id: 'all-consents',
    labelKey: isAdmin ? 'sidebar.allConsents' : 'sidebar.myConsents',
    path: '/consents',
    icon: <ShieldCheck size={18} />,
  }

  const consentItems: SidebarItem[] = [consentItem, PENDING_CONSENT_ITEM]
  const catalogItems: SidebarItem[] = [
    ...(canReadPurposes ? [PURPOSE_ITEM] : []),
    ...(canReadElements ? [ELEMENT_ITEM] : []),
  ]
  const allItems: SidebarItem[] = [
    ...DASHBOARD_ITEMS,
    ...consentItems,
    ...(showGrievances ? GRIEVANCE_ITEMS : []),
    ...(showGrievanceManagement ? GRIEVANCE_MANAGEMENT_ITEMS : []),
    ...catalogItems,
  ]

  const activeItem = mapPathToMenuId(location.pathname, location.search)

  return (
    <Sidebar
      collapsed={collapsed}
      activeItem={activeItem}
      onSelect={(id) => {
        const selectedItem = allItems.find((item) => item.id === id)
        if (selectedItem) {
          navigate(selectedItem.path)
        }
      }}
      aria-label={t('sidebar.ariaLabel')}
    >
      <Sidebar.Nav>
        <Sidebar.Category>
          {DASHBOARD_ITEMS.map((item) => (
            <Sidebar.Item key={item.id} id={item.id}>
              <Sidebar.ItemIcon>{item.icon}</Sidebar.ItemIcon>
              <Sidebar.ItemLabel>{t(item.labelKey)}</Sidebar.ItemLabel>
            </Sidebar.Item>
          ))}
        </Sidebar.Category>

        <Sidebar.Category>
          <Sidebar.CategoryLabel>{t('sidebar.consent')}</Sidebar.CategoryLabel>
          {consentItems.map((item) => (
            <Sidebar.Item key={item.id} id={item.id}>
              <Sidebar.ItemIcon>{item.icon}</Sidebar.ItemIcon>
              <Sidebar.ItemLabel>{t(item.labelKey)}</Sidebar.ItemLabel>
            </Sidebar.Item>
          ))}
        </Sidebar.Category>

        {showGrievances ? (
          <Sidebar.Category>
            <Sidebar.CategoryLabel>{t('sidebar.grievances')}</Sidebar.CategoryLabel>
            {GRIEVANCE_ITEMS.map((item) => (
              <Sidebar.Item key={item.id} id={item.id}>
                <Sidebar.ItemIcon>{item.icon}</Sidebar.ItemIcon>
                <Sidebar.ItemLabel>{t(item.labelKey)}</Sidebar.ItemLabel>
              </Sidebar.Item>
            ))}
          </Sidebar.Category>
        ) : null}

        {showGrievanceManagement ? (
          <Sidebar.Category>
            <Sidebar.CategoryLabel>{t('sidebar.grievanceManagement')}</Sidebar.CategoryLabel>
            {GRIEVANCE_MANAGEMENT_ITEMS.map((item) => (
              <Sidebar.Item key={item.id} id={item.id}>
                <Sidebar.ItemIcon>{item.icon}</Sidebar.ItemIcon>
                <Sidebar.ItemLabel>{t(item.labelKey)}</Sidebar.ItemLabel>
              </Sidebar.Item>
            ))}
          </Sidebar.Category>
        ) : null}

        {catalogItems.length > 0 && (
          <Sidebar.Category>
            <Sidebar.CategoryLabel>{t('sidebar.catalog')}</Sidebar.CategoryLabel>
            {catalogItems.map((item) => (
              <Sidebar.Item key={item.id} id={item.id}>
                <Sidebar.ItemIcon>{item.icon}</Sidebar.ItemIcon>
                <Sidebar.ItemLabel>{t(item.labelKey)}</Sidebar.ItemLabel>
              </Sidebar.Item>
            ))}
          </Sidebar.Category>
        )}
      </Sidebar.Nav>
    </Sidebar>
  )
}

export default AppSidebar
