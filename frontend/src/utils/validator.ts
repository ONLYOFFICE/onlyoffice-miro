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

export const normalizeAddress = (value: string): string => {
  if (!value) return value;
  if (!/^https?:\/\//i.test(value)) {
    return `https://${value}`;
  }
  return value;
};

export const normalizeAddressForSave = (value: string): string => {
  if (!value) return value;

  let address = value;
  if (!/^https?:\/\//i.test(address)) {
    address = `https://${address}`;
  }

  return address.replace(/\/+$/, '');
};

export const validateAddress = (value: string): boolean => {
  if (!value) return false;

  let address = value;
  if (!/^https?:\/\//i.test(address)) {
    address = `https://${address}`;
  }

  try {
    const url = new URL(address);
    if (url.protocol !== 'https:') {
      return false;
    }
  } catch (e) {
    return false;
  }

  return true;
};

export const validateShortText = (value: string): boolean => {
  if (!value) return false;
  return value.length <= 255;
};
