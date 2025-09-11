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

import { create } from 'zustand';
import i18n from '@i18n/config';

import { Document } from '@features/file/lib/types';

import {
  convertDocument,
  deleteDocument,
  fetchDocuments,
  navigateDocument,
} from '@features/file/api/file';

import { useEmitterStore } from '@stores/useEmitterStore';

interface FilesState {
  documents: Document[];
  filteredDocuments: Document[];

  activeDropdown: string | null;
  cursor: string | null;
  searchQuery: string;

  initialized: boolean;
  loading: boolean;
  converting: boolean;

  authError: boolean;
  serverConfigError: boolean;

  setSearchQuery: (query: string) => void;
  setObserverRef: (node: HTMLElement | null) => void;
  toggleDropdown: (id: string | null) => void;

  loadMoreDocuments: () => Promise<void>;
  refreshDocuments: () => Promise<void>;

  navigateDocument: (document: Document) => void;
  downloadPdf: (document: Document) => void;
  deleteDocument: (document: Document) => Promise<void>;

  updateOnCreate: (documents: Document[]) => void;
  updateOnUpdate: (documents: Document[]) => void;
  updateOnDelete: (documentIds: string[]) => void;
}

const filterDocuments = (documents: Document[], query: string): Document[] => {
  if (!query.trim()) return documents;

  const lowerCaseQuery = query.toLowerCase();
  return documents.filter((doc) =>
    doc.data?.title?.toLowerCase().includes(lowerCaseQuery)
  );
};

export const useFilesStore = create<FilesState>((set, get) => ({
  documents: [],
  filteredDocuments: [],

  activeDropdown: null,
  cursor: null,
  searchQuery: '',

  initialized: false,
  loading: false,
  converting: false,

  authError: false,
  serverConfigError: false,

  setSearchQuery: (searchQuery: string) => {
    set({ searchQuery });
    const { documents } = get();
    const filteredDocuments = filterDocuments(documents, searchQuery);
    set({ filteredDocuments });
  },
  setObserverRef: (node: HTMLElement | null) => {
    if (!node) return undefined;

    const observer = new IntersectionObserver(
      (entries) => {
        if (entries[0].isIntersecting && !get().loading && get().cursor)
          get().loadMoreDocuments();
      },
      { threshold: 0.1, rootMargin: '100px' }
    );

    observer.observe(node);
    return () => observer.disconnect();
  },
  toggleDropdown: (id: string | null) => {
    const { activeDropdown } = get();
    if (activeDropdown === id) set({ activeDropdown: null });
    else set({ activeDropdown: id });
  },

  loadMoreDocuments: async () => {
    const { loading, cursor, initialized } = get();
    if (loading || !cursor || !initialized) return;

    set({ loading: true });
    try {
      const pageable = await fetchDocuments(cursor);
      if (!pageable.data.length) {
        set({ loading: false, cursor: null });
        return;
      }

      set((state) => ({
        documents: [...state.documents, ...pageable.data],
        loading: false,
        cursor: pageable.cursor,
      }));

      if (pageable.cursor) await get().loadMoreDocuments();
    } catch (error) {
      if (error instanceof Error) {
        if (
          error.message === 'not authorized' ||
          error.message === 'access denied'
        ) {
          set({ loading: false, authError: true });
        } else if (error.message === 'document server configuration error') {
          set({ loading: false, serverConfigError: true });
        } else {
          set({ loading: false });
        }
      } else {
        set({ loading: false });
      }
    }
  },
  refreshDocuments: async () => {
    const { loading, initialized } = get();
    if (loading) return;

    set({ loading: true, authError: false, serverConfigError: false });
    try {
      const pageable = await fetchDocuments();
      if (!initialized || !get().cursor) {
        set({
          documents: pageable.data,
          loading: false,
          cursor: pageable.cursor,
          initialized: true,
        });

        if (pageable.cursor) await get().loadMoreDocuments();
      } else {
        set((state) => ({
          documents: [...pageable.data, ...state.documents],
          loading: false,
          cursor: pageable.cursor || state.cursor,
        }));
      }
    } catch (error) {
      if (error instanceof Error) {
        if (
          error.message === 'not authorized' ||
          error.message === 'access denied'
        ) {
          set({ loading: false, authError: true });
        } else if (error.message === 'document server configuration error') {
          set({ loading: false, serverConfigError: true });
        } else {
          set({ loading: false });
        }
      } else {
        set({ loading: false });
      }
    }
  },

  navigateDocument: async (document: Document) => {
    await navigateDocument(document.id);
  },
  downloadPdf: async (document: Document) => {
    set({ converting: true });
    let url = '';
    try {
      // eslint-disable-next-line consistent-return
      const pollConversion = async (): Promise<void> => {
        const response = await convertDocument(document.id);
        if (response.percent < 100) {
          await new Promise((resolve) => {
            setTimeout(resolve, 125);
          });
          return pollConversion();
        }
        url = response.url;
      };

      await pollConversion();
      window.open(url, '_blank');
      set({ activeDropdown: null, converting: false });
    } catch (error) {
      set({ converting: false });
    }
  },
  deleteDocument: async (document: Document) => {
    const emitterStore = useEmitterStore.getState();
    try {
      await deleteDocument(document.id);
      await emitterStore.emitDocumentDeleted(document.id);
      set({ activeDropdown: null });
      get().updateOnDelete([document.id]);
    } catch (error) {
      if (error instanceof Error) {
        const errorMessage = error.message.toLowerCase();
        const isLockingError = 
          errorMessage.includes('another client is locking') ||
          errorMessage.includes('cannot remove the document') ||
          errorMessage.includes('locking item');
          
        if (isLockingError) {
          await emitterStore.emitNotification(
            i18n.t('features.file.item.errors.delete_locked'),
            'error'
          );
        } else {
          throw error;
        }
      } else {
        throw error;
      }
      set({ activeDropdown: null });
    }
  },

  updateOnCreate: (documents: Document[]) => {
    set((state) => {
      const existing = new Set(state.documents.map((doc) => doc.id));
      const docs = documents.filter((doc) => !existing.has(doc.id));

      if (docs.length === 0) return state;

      const merged = [...state.documents, ...docs];
      return {
        documents: merged,
        filteredDocuments: filterDocuments(merged, state.searchQuery),
      };
    });
  },
  updateOnUpdate: (documents: Document[]) => {
    set((state) => {
      const docsMap = new Map(documents.map((doc) => [doc.id, doc]));
      const docs = [...state.documents];
      docs.forEach((docItem) => {
        const updatedDoc = docsMap.get(docItem.id);
        if (updatedDoc) {
          const updatedDocItem = { ...docItem };
          updatedDocItem.createdAt = updatedDoc.createdAt || docItem.createdAt;
          updatedDocItem.modifiedAt =
            updatedDoc.modifiedAt || docItem.modifiedAt;
          Object.assign(docItem, updatedDocItem);
        }
      });
      return {
        documents: docs,
        filteredDocuments: filterDocuments(docs, state.searchQuery),
      };
    });
  },
  updateOnDelete: (ids: string[]) => {
    set((state) => {
      const docs = state.documents.filter((doc) => !ids.includes(doc.id));
      return {
        documents: docs,
        filteredDocuments: filterDocuments(docs, state.searchQuery),
      };
    });
  },
}));

export default useFilesStore;
