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

import { Document, Pageable } from '@features/file/lib/types';

import useApplicationStore from '@stores/useApplicationStore';

export const openEditor = async (doc: Document, displayError?: string) => {
  const { board: miroBoard } = window.miro;
  const applicationStore = useApplicationStore.getState();
  const userPromise = miroBoard.getUserInfo();
  const boardPromise = miroBoard.getInfo();
  const [user, board] = await Promise.all([userPromise, boardPromise]);
  if (applicationStore.shouldRefreshCookie()) {
    await applicationStore.authorize();
    if (useApplicationStore.getState().shouldRefreshCookie()) {
      if (displayError) 
        await miroBoard.notifications.showError(displayError);
      return;
    }
  }

  const path = `api/editor?uid=${user.id}&fid=${doc.id}&bid=${board.id}&lang=${board.locale}`;
  window.open(
    `${import.meta.env.VITE_MIRO_ONLYOFFICE_BACKEND}/${path}`,
    '_blank'
  );
};

export const fetchDocuments = async (
  cursor?: string | null,
  retryCount = 0
): Promise<Pageable<Document>> => {
  const { board: miroBoard } = window.miro;
  const boardPromise = miroBoard.getInfo();
  const tokenPromise = miroBoard.getIdToken();
  const maxRetries = 3;

  const [board, token] = await Promise.all([boardPromise, tokenPromise]);
  const path = `api/files`;
  const cursorParam = cursor ? `&cursor=${cursor}` : '';
  const response = await fetch(
    `${import.meta.env.VITE_MIRO_ONLYOFFICE_BACKEND}/${path}?bid=${board.id}${cursorParam}`,
    {
      method: 'GET',
      headers: {
        'Content-Type': 'application/json',
        'x-miro-signature': token,
      },
    }
  );

  if (response.status === 401) throw new Error('not authorized');
  if (response.status === 403) throw new Error('access denied');
  if (response.status === 409)
    throw new Error('document server configuration error');

  if (response.status !== 200) {
    if (retryCount < maxRetries) {
      const delay = 2 ** retryCount * 1000;
      await new Promise((resolve) => {
        setTimeout(resolve, delay);
      });
      return fetchDocuments(cursor, retryCount + 1);
    }

    throw new Error('could not fetch documents information');
  }

  return response.json();
};

export const navigateDocument = async (id: string): Promise<void> => {
  const { board: miroBoard } = window.miro;

  await miroBoard.deselect();
  const target = await miroBoard.getById(id);
  await miroBoard.viewport.zoomTo(target);
  await miroBoard.select({ id });
};

export const convertDocument = async (
  id: string
): Promise<{ url: string; token: string }> => {
  const { board: miroBoard } = window.miro;
  const boardPromise = miroBoard.getInfo();
  const tokenPromise = miroBoard.getIdToken();

  const [board, token] = await Promise.all([boardPromise, tokenPromise]);
  const path = `api/files/convert`;
  const response = await fetch(
    `${import.meta.env.VITE_MIRO_ONLYOFFICE_BACKEND}/${path}?fid=${id}&bid=${board.id}`,
    {
      method: 'GET',
      headers: {
        'Content-Type': 'application/json',
        'x-miro-signature': token,
      },
    }
  );

  if (response.status !== 200)
    throw new Error('could not get converted document');

  return response.json();
};

export const deleteDocument = async (id: string): Promise<void> => {
  const { board: miroBoard } = window.miro;

  const target = await miroBoard.getById(id);
  if (target) await miroBoard.remove(target);
};
