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

import { FileCreatedResponse } from '@features/manager/lib/types';

import { sanitizeFilename } from '@utils/sanitizer';

export const createFile = async (
  name: string,
  type: string
): Promise<FileCreatedResponse | null> => {
  try {
    const { board: miroBoard } = window.miro;
    const userPromise = miroBoard.getUserInfo();
    const boardPromise = miroBoard.getInfo();
    const tokenPromise = miroBoard.getIdToken();

    const [user, board, token] = await Promise.all([
      userPromise,
      boardPromise,
      tokenPromise,
    ]);
    const sanitizedName = sanitizeFilename(name);
    const path = `api/files/create?uid=${user.id}&bid=${board.id}`;
    const response = await fetch(
      `${import.meta.env.VITE_MIRO_ONLYOFFICE_BACKEND}/${path}`,
      {
        method: 'POST',
        body: JSON.stringify({
          board_id: board.id,
          file_name: sanitizedName,
          file_type: type,
          file_lang: board.locale,
        }),
        headers: {
          'Content-Type': 'application/json',
          'x-miro-signature': token,
        },
      }
    );

    if (!response.ok) throw new Error('Failed to create a new document');

    return (await response.json()).data;
  } catch {
    return null;
  }
};

export const fetchSupportedFileTypes = () => {
  return ['docx', 'xlsx', 'pptx'];
};
