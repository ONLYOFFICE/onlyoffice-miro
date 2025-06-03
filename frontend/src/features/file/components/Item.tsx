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

import React, { forwardRef, useEffect, useRef } from 'react';
import { useTranslation } from 'react-i18next';

import { Document } from '@features/file/lib/types';

import { openEditor } from '@features/file/api/file';

import getIcon from '@features/file/utils/icon';
import formatDate from '@features/file/utils/date';

import useFilesStore from '@features/file/stores/useFileStore';

import '@features/file/components/item.css';

interface FileItemProps extends React.HTMLAttributes<HTMLDivElement> {
  document: Document;
}

export const FileItem = forwardRef<HTMLDivElement, FileItemProps>(
  ({ document: fileDocument, className, ...props }, ref) => {
    const { t } = useTranslation();
    const {
      activeDropdown,
      toggleDropdown,
      converting,
      navigateDocument,
      downloadPdf,
      deleteDocument,
    } = useFilesStore();
    const dropdownRef = useRef<HTMLDivElement>(null);
    const isDropdownOpen = activeDropdown === fileDocument.id;

    useEffect(() => {
      const handleClickOutside = (event: MouseEvent) => {
        if (
          isDropdownOpen &&
          dropdownRef.current &&
          !dropdownRef.current.contains(event.target as Node)
        ) {
          toggleDropdown(null);
        }
      };

      window.document.addEventListener('mousedown', handleClickOutside);
      return () => {
        window.document.removeEventListener('mousedown', handleClickOutside);
      };
    }, [isDropdownOpen, toggleDropdown]);

    return (
      <div
        ref={ref}
        className="file-container"
        onClick={() =>
          openEditor(
            fileDocument,
            t('features.file.item.errors.failed_to_open')
          )
        }
        onKeyDown={(e) => {
          if (e.key === 'Enter' || e.key === ' ') {
            openEditor(
              fileDocument,
              t('features.file.item.errors.failed_to_open')
            );
          }
        }}
        role="button"
        tabIndex={0}
        {...props}
      >
        <img
          className="file-container__icon"
          src={getIcon(fileDocument.data?.title)}
          alt={fileDocument.data?.title}
        />
        <span className="file-container__title">
          {fileDocument.data?.title}
        </span>
        <span className="file-container__date file-container__text_secondary">
          {formatDate(fileDocument.createdAt)}
        </span>
        <span className="file-container__date file-container__text_secondary">
          {formatDate(fileDocument.modifiedAt)}
        </span>
        <div className="file-container__dropdown-wrapper">
          <button
            type="button"
            className="file-container__dropdown-button"
            onClick={(e) => {
              e.stopPropagation();
              toggleDropdown(isDropdownOpen ? null : fileDocument.id);
            }}
          >
            <div
              role="img"
              className="file-container__dropdown-button__icon"
              aria-label="Options"
            />
          </button>
          {isDropdownOpen && (
            <div
              ref={dropdownRef}
              className="file-container__dropdown-menu"
              onClick={(e) => e.stopPropagation()}
              onKeyDown={(e) => {
                if (e.key === 'Escape') {
                  toggleDropdown(null);
                }
              }}
              role="menu"
              tabIndex={-1}
            >
              <button
                type="button"
                className="file-container__dropdown-menu__item"
                onClick={(e) => {
                  e.stopPropagation();
                  navigateDocument(fileDocument);
                  toggleDropdown(null);
                }}
              >
                {t('features.file.item.navigate')}
              </button>
              <button
                type="button"
                className="file-container__dropdown-menu__item"
                onClick={(e) => {
                  e.stopPropagation();
                  downloadPdf(fileDocument);
                }}
                disabled={converting}
              >
                {t('features.file.item.download')}
              </button>
              <button
                type="button"
                className="file-container__dropdown-menu__item file-container__dropdown-menu__item_delete"
                onClick={(e) => {
                  e.stopPropagation();
                  deleteDocument(fileDocument);
                }}
                disabled={converting}
              >
                {t('features.file.item.delete')}
              </button>
            </div>
          )}
        </div>
      </div>
    );
  }
);

FileItem.displayName = 'FileItem';

export default FileItem;
