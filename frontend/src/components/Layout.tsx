/**
 *
 * (c) Copyright Ascensio System SIA 2025
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 *
 */

import React, { ReactNode } from 'react';
import { Link, useNavigate } from 'react-router-dom';

import Footer from '@components/Footer';

import useApplicationStore from '@stores/useApplicationStore';

import '@components/layout.css';

interface LayoutProps {
  title: string;
  subtitle?: string;
  footerText: string;
  backTo?: string;
  reload?: boolean;
  settings?: boolean;
  help?: boolean;
  onReload?: () => void;
  onSettings?: () => void;
  onHelp?: () => void;
  children: ReactNode;
}

const Layout: React.FC<LayoutProps> = ({
  title,
  subtitle,
  footerText,
  backTo,
  reload,
  settings,
  help,
  onReload,
  onSettings,
  onHelp = () =>
    window.open('https://github.com/ONLYOFFICE/onlyoffice-miro', '_blank'),
  children,
}) => {
  const { admin } = useApplicationStore();
  const navigate = useNavigate();

  const handleSettings = onSettings || (() => navigate('/settings'));

  const showSettings = settings !== false && admin;
  return (
    <div className="layout-container">
      <div className="layout-header">
        <div className="layout-shifted">
          {backTo ? (
            <div className="layout-title">
              <Link to={backTo} className="layout-title__back">
                <img
                  src="/arrow-left.svg"
                  alt="Back"
                  className="layout-title__back-icon"
                />
                <span className="layout-title__back-text" title={title}>
                  {title}
                </span>
              </Link>
            </div>
          ) : (
            <div className="layout-title" title={title}>
              {title}
            </div>
          )}
          {subtitle && <div className="layout-subtitle">{subtitle}</div>}
        </div>
      </div>
      <div className="layout-main">{children}</div>
      <div className="layout-footer">
        <Footer
          text={footerText}
          reload={reload}
          settings={showSettings}
          help={help}
          onReloadClick={onReload}
          onSettingsClick={handleSettings}
          onHelpClick={onHelp}
        />
      </div>
    </div>
  );
};

Layout.displayName = 'Layout';

export default Layout;
