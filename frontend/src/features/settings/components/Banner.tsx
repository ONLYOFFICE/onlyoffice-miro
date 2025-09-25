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
import { useTranslation } from 'react-i18next';

import '@features/settings/components/banner.css';

interface BannerProps extends React.HTMLAttributes<HTMLDivElement> {
  url?: string;
}

export const Banner = forwardRef<HTMLDivElement, BannerProps>(
  (
    {
      className,
      url = 'https://www.onlyoffice.com/docs-registration.aspx?referer=miro',
      ...props
    },
    ref
  ) => {
    const { t } = useTranslation();

    const handleButtonClick = () => {
      window.open(url, '_blank', 'noopener,noreferrer');
    };

    return (
      <div ref={ref} className={`banner ${className || ''}`} {...props}>
        <div className="banner__content">
          <div className="banner__content__info">
            <h3 className="banner__content__info__title">
              {t('features.settings.banner.title')}
            </h3>
            <p className="banner__content__info__description">
              {t('features.settings.banner.description')}
            </p>
          </div>
          <button
            type="button"
            className="banner__content__button"
            onClick={handleButtonClick}
          >
            {t('features.settings.banner.button')}
          </button>
        </div>
      </div>
    );
  }
);

Banner.displayName = 'Banner';

export default Banner;
