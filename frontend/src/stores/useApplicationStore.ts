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

import fetchAuthorization from '@api/authorize';
import useSettingsStore from '@features/settings/stores/useSettingsStore';

interface ApplicationState {
  loading: boolean;
  authorized: boolean;
  admin: boolean;
  hasCookie: boolean;
  retriesExhausted: boolean;
  cookieExpiresAt: number | null;

  reloadAuthorization: () => Promise<void>;
  refreshAuthorization: () => Promise<void>;
  authorize: () => Promise<void>;
  shouldRefreshCookie: () => boolean;
}

const useApplicationStore = create<ApplicationState>((set, get) => ({
  loading: false,
  authorized: false,
  admin: false,
  hasCookie: false,
  retriesExhausted: false,
  cookieExpiresAt: null,

  reloadAuthorization: async () => {
    set({
      loading: true,
      authorized: false,
      admin: false,
      retriesExhausted: false,
    });
    try {
      const settingsStore = useSettingsStore.getState();
      await settingsStore.initializeSettings();
      set({
        loading: false,
        authorized: true,
        admin: true,
      });
    } catch (err) {
      const unauthorized =
        err instanceof Error && err.message === 'not authorized';
      const forbidden = err instanceof Error && err.message === 'access denied';
      const retries = err instanceof Error && err.message === 'max retries';
      set({
        loading: false,
        authorized: !unauthorized,
        admin: !unauthorized && !forbidden && !retries,
        retriesExhausted: retries,
      });
    } finally {
      const { admin } = get();
      const settingsStore = useSettingsStore.getState();
      const hasNoSettings = !settingsStore.hasSettings;
      if (hasNoSettings) {
        if (admin) window.location.hash = '#/settings';
        else window.location.hash = '#/';
      } else {
        window.location.hash = '#/';
      }
    }

    const settingsStore = useSettingsStore.getState();
    if (!settingsStore.hasSettings) {
      return undefined;
    }

    return undefined;
  },

  refreshAuthorization: async () => {
    try {
      set({ retriesExhausted: false });
      const settingsStore = useSettingsStore.getState();
      await settingsStore.initializeSettings();
      set({
        authorized: true,
        admin: true,
      });
    } catch (err) {
      const unauthorized =
        err instanceof Error && err.message === 'not authorized';
      const forbidden = err instanceof Error && err.message === 'access denied';
      const retries = err instanceof Error && err.message === 'max retries';
      set({
        authorized: !unauthorized,
        admin: !unauthorized && !forbidden,
        retriesExhausted: retries,
      });
    }
  },

  authorize: async () => {
    try {
      set({ hasCookie: false, retriesExhausted: false });
      const { expiresAt } = await fetchAuthorization();
      set({
        hasCookie: true,
        cookieExpiresAt: expiresAt,
      });
    } catch (err) {
      const unauthorized =
        err instanceof Error && err.message === 'not authorized';
      const forbidden = err instanceof Error && err.message === 'access denied';
      const retries = err instanceof Error && err.message === 'max retries';
      set({
        hasCookie: false,
        authorized: !unauthorized,
        admin: !unauthorized && !forbidden,
        retriesExhausted: retries,
        cookieExpiresAt: null,
      });
    }
  },

  shouldRefreshCookie: () => {
    const { hasCookie, cookieExpiresAt } = get();
    if (!hasCookie) return true;
    if (cookieExpiresAt === null) return true;
    return cookieExpiresAt * 1000 - Date.now() <= 30000;
  },
}));

export default useApplicationStore;
