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
import { Link } from 'react-router-dom';
import { useTranslation } from 'react-i18next';

import Button from '@components/Button';
import Layout from '@components/Layout';
import Spinner from '@components/Spinner';

import Empty from '@features/manager/components/Empty';
import FilesList from '@features/file/components/List';
import Installation from '@features/manager/components/Installation';
import NotConfigured from '@features/manager/components/NotConfigured';
import Searchbar from '@features/file/components/Search';

import useFilesStore from '@features/file/stores/useFileStore';
import useApplicationStore from '@stores/useApplicationStore';

import SettingsPage from '@app/pages/settings';

import '@app/pages/manager/index.css';

const ManagerPage = () => {
  const { t } = useTranslation();
  const { admin, retriesExhausted, reloadAuthorization } =
    useApplicationStore();
  const {
    searchQuery,
    filteredDocuments,
    documents,
    loading,
    authError,
    serverConfigError,
    initialized,
    refreshDocuments,
  } = useFilesStore();
  const isInitialLoading = loading && !initialized;

  if (serverConfigError && admin) return <SettingsPage forceDisableBack />;

  const handleReload = async () => {
    if (retriesExhausted) await reloadAuthorization();
    await refreshDocuments();
  };

  const renderContent = () => {
    if (authError) return <Installation />;

    if (retriesExhausted)
      return (
        <Empty
          title={t('pages.manager.empty.max_retries_title')}
          subtitle={t('pages.manager.empty.max_retries_subtitle')}
        />
      );

    if (serverConfigError && !admin) return <NotConfigured />;

    if (loading && documents.length === 0)
      return (
        <div className="manager-container__main__loading">
          <Spinner size="large" />
        </div>
      );

    if (documents.length === 0)
      return (
        <Empty
          title={t('pages.manager.empty.title')}
          subtitle={t('pages.manager.empty.subtitle')}
        />
      );

    if (searchQuery !== '' && filteredDocuments.length === 0)
      return (
        <Empty
          title={t('pages.manager.empty.search_title')}
          subtitle={t('pages.manager.empty.search_subtitle')}
        />
      );

    return <FilesList />;
  };

  const showSettings =
    !isInitialLoading && !serverConfigError && !authError && !retriesExhausted;

  return (
    <Layout
      title={serverConfigError ? '' : t('pages.manager.title')}
      subtitle={
        !authError &&
        !serverConfigError &&
        !retriesExhausted &&
        documents.length > 0
          ? t('pages.manager.subtitle')
          : ''
      }
      footerText={t('pages.manager.footer')}
      reload
      settings={showSettings}
      onReload={handleReload}
    >
      <div className="manager-container">
        <div className="manager-container_shifted">
          {!authError &&
            !serverConfigError &&
            !retriesExhausted &&
            documents.length > 0 && <Searchbar />}
        </div>
        <div className="manager-container__main">
          <div className="manager-container_shifted manager-container__main__files">
            {renderContent()}
          </div>
        </div>
        <div className="manager-container_shifted manager-container__button-container">
          {!authError && !serverConfigError && !retriesExhausted && (
            <Link
              to="/create"
              state={{ isBack: false }}
              className={isInitialLoading ? 'disabled-link' : ''}
              onClick={(e) => isInitialLoading && e.preventDefault()}
            >
              <Button
                name={t('pages.manager.create')}
                variant="primary"
                disabled={isInitialLoading}
              />
            </Link>
          )}
        </div>
      </div>
    </Layout>
  );
};

ManagerPage.displayName = 'ManagerPage';

export default ManagerPage;
