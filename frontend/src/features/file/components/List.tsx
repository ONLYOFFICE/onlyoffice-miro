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

import React, { forwardRef, useCallback, useEffect } from 'react';
import { useTranslation } from 'react-i18next';
// eslint-disable-next-line import/no-extraneous-dependencies
import { ItemsUpdateEvent } from '@mirohq/websdk-types';
// eslint-disable-next-line import/no-extraneous-dependencies
import {
  ItemsCreateEvent,
  ItemsDeleteEvent,
} from '@mirohq/websdk-types/stable/api/ui';

import { Document } from '@features/file/lib/types';
import { FileCreatedEvent, FilesDeletedEvent } from '@lib/types';

import Spinner from '@components/Spinner';
import FileItem from '@features/file/components/Item';

import useFilesStore from '@features/file/stores/useFileStore';
import useEmitterStore, { EmitterEvents } from '@stores/useEmitterStore';

import '@features/file/components/list.css';

const documentType = 'document';

interface FilesListProps extends React.HTMLAttributes<HTMLDivElement> {}

export const FilesList = forwardRef<HTMLDivElement, FilesListProps>(
  ({ className, ...props }, ref) => {
    const { t } = useTranslation();
    const {
      searchQuery,
      filteredDocuments,
      documents,
      refreshDocuments,
      loading,
      cursor,
      setObserverRef,
      initialized,
      updateOnCreate,
      updateOnDelete,
      updateOnUpdate,
    } = useFilesStore();
    const { emitDocumentsAdded, emitDocumentsDeleted, emitNotification } =
      useEmitterStore();

    const listenDocumentAddedUI = useCallback(
      async (e: ItemsCreateEvent) => {
        const events = e.items
          .filter((doc) => doc.type === documentType)
          .map((doc) => {
            const documentItem = doc as unknown as Record<string, unknown>;
            return {
              id: documentItem?.id as string,
              name: (documentItem?.name as string) || '',
              createdAt: documentItem?.createdAt as string,
              modifiedAt: documentItem?.modifiedAt as string,
              links: {
                self:
                  ((documentItem?.links as Record<string, unknown>)
                    ?.self as string) || '',
              },
              type: documentType,
            } as FileCreatedEvent;
          });

        if (events.length > 0) {
          await emitDocumentsAdded(events);
          await emitNotification(
            t('features.file.list.notifications.documents_added')
          );
        }
      },
      [emitDocumentsAdded, emitNotification, t]
    );

    const listenDocumentDeletedUI = useCallback(
      async (e: ItemsDeleteEvent) => {
        const ids = e.items.map((item) => item.id);
        if (ids.length > 0) {
          await emitDocumentsDeleted(ids);
          updateOnDelete(ids);
        }
      },
      [emitDocumentsDeleted, updateOnDelete]
    );

    const listenDocumentUpdatedUI = useCallback(
      async (e: ItemsUpdateEvent) => {
        const docs = e.items.filter((doc) => doc.type === documentType);
        if (docs.length > 0) updateOnUpdate(docs as unknown as Document[]);
      },
      [updateOnUpdate]
    );

    const listenDocumentCreated = useCallback(
      async (event: FileCreatedEvent) => {
        const newDocument = {
          id: event.id,
          data: {
            title: event.name,
            documentUrl: event.links.self,
          },
          createdAt: event.createdAt,
          modifiedAt: event.modifiedAt,
          type: documentType,
        } as Document;
        updateOnCreate([newDocument]);
      },
      [updateOnCreate]
    );

    const listenDocumentsAdded = useCallback(async () => {
      await emitNotification(
        t('features.file.list.notifications.documents_added')
      );
    }, [emitNotification, t]);

    const listenDocumentsDeleted = useCallback(
      async (event: FilesDeletedEvent) => {
        updateOnDelete(event.ids);
      },
      [updateOnDelete]
    );

    useEffect(() => {
      if (!initialized) refreshDocuments();

      miro?.board.ui.on(
        EmitterEvents.MIRO_ITEMS_CREATED,
        listenDocumentAddedUI
      );
      miro?.board.ui.on(
        EmitterEvents.MIRO_ITEMS_DELETED,
        listenDocumentDeletedUI
      );
      miro?.board.ui.on(
        EmitterEvents.MIRO_ITEMS_UPDATED,
        listenDocumentUpdatedUI
      );

      miro?.board.events.on(
        EmitterEvents.DOCUMENT_CREATED,
        listenDocumentCreated
      );
      miro?.board.events.on(
        EmitterEvents.DOCUMENTS_ADDED,
        listenDocumentsAdded
      );
      miro?.board.events.on(
        EmitterEvents.DOCUMENTS_DELETED,
        listenDocumentsDeleted
      );

      return () => {
        miro?.board.ui.off(
          EmitterEvents.MIRO_ITEMS_CREATED,
          listenDocumentAddedUI
        );
        miro?.board.ui.off(
          EmitterEvents.MIRO_ITEMS_DELETED,
          listenDocumentDeletedUI
        );
        miro?.board.ui.off(
          EmitterEvents.MIRO_ITEMS_UPDATED,
          listenDocumentUpdatedUI
        );

        miro?.board.events.off(
          EmitterEvents.DOCUMENT_CREATED,
          listenDocumentCreated
        );
        miro?.board.events.off(
          EmitterEvents.DOCUMENTS_ADDED,
          listenDocumentsAdded
        );
        miro?.board.events.off(
          EmitterEvents.DOCUMENTS_DELETED,
          listenDocumentsDeleted
        );
      };
    }, [
      initialized,
      refreshDocuments,
      listenDocumentAddedUI,
      listenDocumentDeletedUI,
      listenDocumentUpdatedUI,
      listenDocumentCreated,
      listenDocumentsAdded,
      listenDocumentsDeleted,
    ]);

    const docs = searchQuery ? filteredDocuments : documents;

    return (
      <div
        ref={ref}
        className={`files-list-container ${className || ''}`}
        {...props}
      >
        {loading && !initialized && (
          <div className="files-list-container_overlay">
            <Spinner size="medium" />
          </div>
        )}
        {docs.length > 0 && (
          <div className="files-list-header">
            <span
              className="files-list-header__title"
              title={t('features.file.list.header.name')}
            >
              {t('features.file.list.header.name')}
            </span>
            <span className="files-list-header__icon" />
            <span
              className="files-list-header__date"
              title={t('features.file.list.header.created')}
              style={{ marginRight: '0.5rem' }}
            >
              {t('features.file.list.header.created')}
            </span>
            <span
              className="files-list-header__date"
              title={t('features.file.list.header.modified')}
            >
              {t('features.file.list.header.modified')}
            </span>
            <div className="files-list-header__dropdown-wrapper" />
          </div>
        )}
        {docs.map((doc) => (
          <FileItem key={doc.id} document={doc} />
        ))}
        {!searchQuery && documents.length > 0 && cursor && (
          <div
            ref={setObserverRef}
            style={{
              height: '10px',
              margin: '20px 0',
            }}
          >
            {loading && initialized && (
              <div style={{ textAlign: 'center', padding: '10px' }}>
                <Spinner size="small" />
              </div>
            )}
          </div>
        )}
      </div>
    );
  }
);

FilesList.displayName = 'FilesList';

export default FilesList;
