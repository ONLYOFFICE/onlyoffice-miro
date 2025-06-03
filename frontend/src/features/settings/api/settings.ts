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

import {
  SettingsRequest,
  SettingsResponse,
} from '@features/settings/lib/types';

export const saveSettings = async (settings: SettingsRequest) => {
  try {
    const { board: miroBoard } = window.miro;
    const boardPromise = miroBoard.getInfo();
    const tokenPromise = miroBoard.getIdToken();

    const [board, token] = await Promise.all([boardPromise, tokenPromise]);
    const path = `api/settings`;
    const response = await fetch(
      `${import.meta.env.VITE_MIRO_ONLYOFFICE_BACKEND}/${path}`,
      {
        method: 'POST',
        body: JSON.stringify({
          board_id: board.id,
          ...settings,
        }),
        headers: {
          'Content-Type': 'application/json',
          'x-miro-signature': token,
        },
      }
    );

    if (response.ok) {
      return true;
    }

    if (response.status === 500 || response.status === 400 || response.status === 403) {
      let err;
      try {
        err = await response.json();
      } catch {
        err = null;
      }

      if (err && err.error) 
        throw { status: 500, error: err.error };
    }

    throw { status: 503, error: 'features.settings.form.errors.service_unavailable' };
  } catch (err) {
    throw err;
  }
};

export const fetchSettings: () => Promise<SettingsResponse> = async () => {
  const { board: miroBoard } = window.miro;
  const boardPromise = miroBoard.getInfo();
  const tokenPromise = miroBoard.getIdToken();

  const [board, token] = await Promise.all([boardPromise, tokenPromise]);
  const path = `api/settings?bid=${board.id}`;
  const response = await fetch(
    `${import.meta.env.VITE_MIRO_ONLYOFFICE_BACKEND}/${path}`,
    {
      method: 'GET',
      headers: {
        'Content-Type': 'application/json',
        'x-miro-signature': token,
      },
    }
  );

  if (response.ok) {
    const data = await response.json();
    return {
      ...data,
    };
  }

  if (response.status === 401) throw new Error('not authorized');

  if (response.status === 403) throw new Error('access denied');

  if (response.status === 404)
    return {
      address: '',
      header: '',
      secret: '',
    };

  const maxRetries = 3;
  const retryWithBackoff = async (
    retryCount = 0
  ): Promise<SettingsResponse> => {
    try {
      if (retryCount >= maxRetries) throw new Error('max retries');

      const backoffTime = 2 ** retryCount * 250;
      await new Promise((resolve) => {
        setTimeout(resolve, backoffTime);
      });

      const retryResponse = await fetch(
        `${import.meta.env.VITE_MIRO_ONLYOFFICE_BACKEND}/${path}`,
        {
          method: 'GET',
          headers: {
            'Content-Type': 'application/json',
            'x-miro-signature': token,
          },
        }
      );

      if (retryResponse.ok) {
        const data = await retryResponse.json();
        return {
          ...data,
        };
      }

      return await retryWithBackoff(retryCount + 1);
    } catch (error) {
      if (retryCount >= maxRetries)
        throw error;
      return retryWithBackoff(retryCount + 1);
    }
  };

  return retryWithBackoff();
};
