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

const fetchAuthorization = async () => {
  const { board } = window.miro;
  const token = await board.getIdToken();

  const path = `api/authorize`;

  const controller = new AbortController();
  const tid = setTimeout(() => controller.abort(), 3500);

  try {
    const response = await fetch(
      `${import.meta.env.VITE_MIRO_ONLYOFFICE_BACKEND}/${path}`,
      {
        method: 'GET',
        headers: {
          'Content-Type': 'application/json',
          'x-miro-signature': token,
        },
        signal: controller.signal,
      }
    );

    if (response.ok) {
      const data = await response.json();
      if (!data.token || !data.expires_at)
        throw new Error('failed to authorize the request');

      return {
        token: data.token,
        expiresAt: data.expires_at,
      };
    }

    if (response.status === 401) throw new Error('not authorized');

    if (response.status === 403) throw new Error('access denied');

    throw new Error('failed to authorize the request');
  } catch (error: unknown) {
    if (error instanceof Error && error.name === 'AbortError')
      throw new Error('request timeout');
    throw error;
  } finally {
    clearTimeout(tid);
  }
};

export default fetchAuthorization;
