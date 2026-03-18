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

import React, { forwardRef, useState, useEffect, ChangeEvent } from 'react';
import { useTranslation } from 'react-i18next';

import useFilesStore from '@features/file/stores/useFileStore';

import { sanitizeSearchQuery } from '@utils/sanitizer';

import '@features/file/components/search.css';

interface SearchbarProps extends React.HTMLAttributes<HTMLDivElement> {}

export const Searchbar = forwardRef<HTMLDivElement, SearchbarProps>(
  ({ className, ...props }, ref) => {
    const { searchQuery, setSearchQuery, initialized, loading } =
      useFilesStore();
    const [localQuery, setLocalQuery] = useState(searchQuery);
    const { t } = useTranslation();

    const disabled = loading && !initialized;

    useEffect(() => {
      const timer = setTimeout(() => {
        setSearchQuery(localQuery);
      }, 300);

      return () => clearTimeout(timer);
    }, [localQuery, setSearchQuery]);

    const handleSearchChange = (e: ChangeEvent<HTMLInputElement>) => {
      const sanitizedValue = sanitizeSearchQuery(e.target.value);
      setLocalQuery(sanitizedValue);
    };

    const handleClearSearch = () => {
      setLocalQuery('');
      setSearchQuery('');
    };

    return (
      <div
        className={`searchbar-container ${className || ''} ${disabled ? 'searchbar-container__disabled' : ''}`}
        ref={ref}
        {...props}
      >
        <div className="searchbar-container__main">
          <div className="searchbar-container__main__icon">
            <img src="/search.svg" alt="Search icon" />
          </div>
          <input
            className="searchbar-container__main__input"
            type="text"
            placeholder={t('features.file.search.placeholder')}
            value={localQuery}
            onChange={handleSearchChange}
            disabled={disabled}
          />
          {localQuery && (
            <button
              type="button"
              className="searchbar-container__main__clear"
              onClick={handleClearSearch}
              disabled={disabled}
              aria-label={t('features.file.search.clear')}
            >
              <img src="/cross.svg" alt="Clear search" />
            </button>
          )}
        </div>
      </div>
    );
  }
);

Searchbar.displayName = 'Searchbar';

export default Searchbar;
