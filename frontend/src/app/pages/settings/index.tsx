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
import { useLocation } from 'react-router-dom';
import { useTranslation } from 'react-i18next';

import Layout from '@components/Layout';

import Form from '@features/settings/components/Form';

import useSettingsStore from '@features/settings/stores/useSettingsStore';

import '@app/pages/settings/index.css';

interface SettingsPageProps {
  forceDisableBack?: boolean;
}

const SettingsPage: React.FC<SettingsPageProps> = ({
  forceDisableBack: propForceDisableBack,
}) => {
  const location = useLocation();
  const { t } = useTranslation();
  const { hasSettings } = useSettingsStore();

  const forceDisableBack =
    location.state?.forceDisableBack ?? propForceDisableBack ?? false;
  const backTo = !forceDisableBack && hasSettings ? '/' : undefined;

  return (
    <Layout
      title={t('pages.settings.title')}
      footerText={t('pages.settings.footer')}
      settings={false}
      help
      backTo={backTo}
    >
      <Form />
    </Layout>
  );
};

SettingsPage.displayName = 'SettingsPage';

export default SettingsPage;
