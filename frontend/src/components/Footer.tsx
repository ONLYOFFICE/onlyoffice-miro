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

import React, { forwardRef } from 'react';

import '@components/footer.css';

interface FooterProps extends React.ButtonHTMLAttributes<HTMLDivElement> {
  text: string;
  reload?: boolean;
  settings?: boolean;
  help?: boolean;
  onReloadClick?: () => void;
  onSettingsClick?: () => void;
  onHelpClick?: () => void;
}

const Footer: React.FC<FooterProps> = forwardRef<HTMLDivElement, FooterProps>(
  (
    {
      id,
      text,
      disabled,
      reload = false,
      settings = true,
      help = true,
      onReloadClick,
      onSettingsClick,
      onHelpClick,
      ...props
    },
    ref
  ) => {
    const realId = id || Math.random().toString(36).substring(2, 9);

    return (
      <div id={realId} ref={ref} className="footer-container" {...props}>
        <span className="footer-container__title">
          {text || 'Developed by ONLYOFFICE'}
        </span>
        {reload && (
          <button
            className="footer-container__button"
            onClick={onReloadClick}
            aria-label="Reload"
            type="button"
          >
            <div role="img" className="footer-container__button__reload" />
          </button>
        )}
        {settings && (
          <button
            className="footer-container__button"
            onClick={onSettingsClick}
            aria-label="Settings"
            type="button"
          >
            <div role="img" className="footer-container__button__settings" />
          </button>
        )}
        {help && (
          <button
            className="footer-container__button"
            onClick={onHelpClick}
            aria-label="Help"
            type="button"
          >
            <div role="img" className="footer-container__button__help" />
          </button>
        )}
      </div>
    );
  }
);

Footer.displayName = 'Footer';

export default Footer;
