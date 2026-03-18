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

import React from 'react';
import { useTranslation } from 'react-i18next';

import Button from '@components/Button';

import '@features/manager/components/installation.css';

const Installation = () => {
  const { t } = useTranslation();
  return (
    <div className="installation-container">
      <img
        src="/notconfigured.svg"
        alt="Configuration Error"
        className="installation-container__icon"
      />
      <div className="installation-container__title">
        <span className="installation-container__title-text">
          {t('features.manager.installation.error')}
        </span>
      </div>
      <div className="installation-container__message">
        {t('features.manager.installation.description')}
      </div>
      <Button
        name={t('features.manager.installation.button')}
        variant="primary"
        onClick={() => {
          window.open(import.meta.env.VITE_MIRO_INSTALLATION_URL, '_blank');
          window.miro?.board.ui.closePanel();
        }}
        className="installation-container__button"
        title={t('features.manager.installation.button')}
      />
    </div>
  );
};

export default Installation;
