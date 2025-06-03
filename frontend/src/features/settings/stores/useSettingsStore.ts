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

import { fetchSettings, saveSettings } from '@features/settings/api/settings';

interface SettingsState {
  address: string;
  header: string;
  secret: string;
  demoStarted: string;

  demo: boolean;
  persistedCredentials: boolean;

  loading: boolean;
  hasSettings: boolean;
  error: string | null;

  setAddress: (value: string) => void;
  setHeader: (value: string) => void;
  setSecret: (value: string) => void;
  setDemo: (value: boolean) => void;
  saveSettings: () => Promise<void>;
  initializeSettings: () => Promise<void>;
}

const useSettingsStore = create<SettingsState>((set, get) => ({
  address: '',
  header: '',
  secret: '',
  demoStarted: '',

  demo: false,
  persistedCredentials: false,

  loading: false,
  hasSettings: false,
  error: null,

  setAddress: (value) => set({ address: value }),
  setHeader: (value) => set({ header: value }),
  setSecret: (value) => set({ secret: value }),
  setDemo: (value) => set({ demo: value }),
  initializeSettings: async () => {
    set({ loading: true, error: null });
    try {
      const settings = await fetchSettings();
      set({
        address: settings.address || '',
        header: settings.header || '',
        secret: settings.secret || '',
        demo: settings.demo.enabled,
        demoStarted: settings.demo.started,
        persistedCredentials: !!(
          settings.address &&
          settings.header &&
          settings.secret
        ),
        loading: false,
        error: null,
      });
    } catch (error) {
      const isAccessDenied =
        error instanceof Error &&
        (error.message === 'access denied' ||
          error.message === 'not authorized');
      if (isAccessDenied) throw error;

      set({
        loading: false,
        error: isAccessDenied ? 'access denied' : null,
      });
    } finally {
      const { demoStarted, persistedCredentials } = get();
      const isDemoExpired = demoStarted
        ? (() => {
            const startTime = new Date(demoStarted).getTime();
            const expiryDays = parseInt(
              import.meta.env.VITE_ASC_DEMO_EXPIRATION_DAYS || '30',
              10
            );
            const expiryTime = startTime + expiryDays * 24 * 60 * 60 * 1000;
            const currentTime = Date.now();
            return currentTime > expiryTime;
          })()
        : false;
      set({ hasSettings: !isDemoExpired || persistedCredentials });
    }
  },

  saveSettings: async () => {
    const { address, header, secret, demo, demoStarted } = get();
    if (demoStarted && (address === '' || header === '' || secret === ''))
      return;

    set({ loading: true, error: null });
    try {
      await saveSettings({ address, header, secret, demo });
      set({
        address,
        header,
        secret,
        persistedCredentials: address !== '' && header !== '' && secret !== '',
        demoStarted:
          (demoStarted === '' || !demoStarted) && demo
            ? new Date().toISOString()
            : demoStarted,
        demo,
        loading: false,
      });
    } catch (error) {
      const isAccessDenied =
        error instanceof Error && error.message === 'access denied';
      set({
        loading: false,
        error: isAccessDenied ? 'access denied' : null,
      });
      throw error;
    }
  },
}));

export default useSettingsStore;
