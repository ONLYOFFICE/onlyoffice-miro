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

import { FileCreatedResponse } from '@features/manager/lib/types';

import {
  createFile as createMiroFile,
  fetchSupportedFileTypes,
} from '@features/manager/api/file';
import { navigateDocument } from '@features/file/api/file';

interface CreatorState {
  selectedName: string;
  selectedType: string;

  loading: boolean;
  error: boolean;

  getSupportedTypes: () => string[];
  setSelectedName: (value: string) => void;
  setSelectedType: (value: string) => void;

  createFile: (navigateTo?: boolean) => Promise<FileCreatedResponse | null>;
  resetSelected: () => void;
}

const useCreatorStore = create<CreatorState>((set, get) => ({
  selectedName: 'New file',
  selectedType: '',

  loading: false,
  error: false,

  getSupportedTypes(): string[] {
    return fetchSupportedFileTypes();
  },
  setSelectedName: (value) => set({ selectedName: value }),
  setSelectedType: (value) => set({ selectedType: value }),

  createFile: async (
    navigateTo = true
  ): Promise<FileCreatedResponse | null> => {
    set({ loading: true, error: false });

    const { selectedName, selectedType } = get();
    if (!selectedName || !selectedType) return null;

    const createdFile = await createMiroFile(selectedName, selectedType);
    if (!createdFile) {
      set({ loading: false, error: true });
      setTimeout(() => set({ error: false }), 2500);
      return null;
    }

    set({ loading: false });

    if (navigateTo) await navigateDocument(createdFile.id);

    return createdFile;
  },
  resetSelected: () => {
    set({ selectedName: 'New file', selectedType: '' });
  },
}));

export default useCreatorStore;
