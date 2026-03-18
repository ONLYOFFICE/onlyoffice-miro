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

import DOMPurify from 'dompurify';

const SANITIZER_CONFIGS = {
  strict: {
    ALLOWED_TAGS: [] as string[],
    ALLOWED_ATTR: [] as string[],
    KEEP_CONTENT: true,
  },
  basic: {
    ALLOWED_TAGS: ['b', 'i', 'em', 'strong', 'span'] as string[],
    ALLOWED_ATTR: ['class'] as string[],
    KEEP_CONTENT: true,
  },
  display: {
    ALLOWED_TAGS: [] as string[],
    ALLOWED_ATTR: [] as string[],
    KEEP_CONTENT: true,
    RETURN_DOM: false,
    RETURN_DOM_FRAGMENT: false,
  },
};

export const sanitizeFormInput = (input: string): string => {
  if (!input || typeof input !== 'string') return '';

  return DOMPurify.sanitize(input, SANITIZER_CONFIGS.strict);
};

export const sanitizeForDisplay = (content: string): string => {
  if (!content || typeof content !== 'string') return '';

  return DOMPurify.sanitize(content, SANITIZER_CONFIGS.display);
};

export const sanitizeUrl = (url: string): string => {
  if (!url || typeof url !== 'string') return '';

  const sanitized = sanitizeFormInput(url);
  const lowercaseUrl = sanitized.toLowerCase().trim();
  const dangerousProtocols = [
    // eslint-disable-next-line no-script-url
    'javascript:',
    'data:',
    'vbscript:',
    'file:',
    'about:',
  ];

  const hasDangerousProtocol = dangerousProtocols.some((protocol) =>
    lowercaseUrl.startsWith(protocol)
  );

  if (hasDangerousProtocol) {
    return '';
  }

  return sanitized;
};

export const sanitizeFilename = (filename: string): string => {
  if (!filename || typeof filename !== 'string') return '';

  let sanitized = sanitizeFormInput(filename);

  const invalidCharsRegex = new RegExp(
    `[<>:"/\\\\|?*${String.fromCharCode(0)}-${String.fromCharCode(31)}]`,
    'g'
  );
  sanitized = sanitized.replace(invalidCharsRegex, '');
  sanitized = sanitized.replace(/^[\s.]+|[\s.]+$/g, '');
  if (sanitized.length > 255) {
    sanitized = sanitized.substring(0, 255);
  }

  return sanitized;
};

export const sanitizeSearchQuery = (query: string): string => {
  if (!query || typeof query !== 'string') return '';

  const sanitized = sanitizeFormInput(query);

  return sanitized.replace(/[<>]/g, '').trim();
};

export const sanitizeObject = <T extends Record<string, unknown>>(
  obj: T,
  sanitizer: (value: string) => string = sanitizeFormInput
): T => {
  if (!obj || typeof obj !== 'object') return obj;

  const sanitized = { ...obj } as Record<string, unknown>;

  Object.keys(sanitized).forEach((key) => {
    if (Object.prototype.hasOwnProperty.call(sanitized, key)) {
      const value = sanitized[key];

      if (typeof value === 'string') {
        sanitized[key] = sanitizer(value);
      } else if (value && typeof value === 'object') {
        sanitized[key] = sanitizeObject(
          value as Record<string, unknown>,
          sanitizer
        );
      }
    }
  });

  return sanitized as T;
};

export const sanitizeConfig = (config: {
  address?: string;
  secret?: string;
  header?: string;
  demo?: boolean;
}) => {
  return {
    address: config.address ? sanitizeUrl(config.address) : '',
    secret: config.secret ? sanitizeFormInput(config.secret) : '',
    header: config.header ? sanitizeFormInput(config.header) : '',
    demo: config.demo ?? false,
  };
};

export const containsSuspiciousContent = (input: string): boolean => {
  if (!input || typeof input !== 'string') return false;

  const suspiciousPatterns = [
    /<script/i,
    /javascript:/i,
    /on\w+\s*=/i,
    /<iframe/i,
    /<object/i,
    /<embed/i,
    /<link/i,
    /<meta/i,
    /data:\s*text\/html/i,
    /vbscript:/i,
  ];

  return suspiciousPatterns.some((pattern) => pattern.test(input));
};

export const securitySanitize = (
  input: string
): {
  sanitized: string;
  wasSuspicious: boolean;
  originalLength: number;
  sanitizedLength: number;
} => {
  if (!input || typeof input !== 'string') {
    return {
      sanitized: '',
      wasSuspicious: false,
      originalLength: 0,
      sanitizedLength: 0,
    };
  }

  const wasSuspicious = containsSuspiciousContent(input);
  const sanitized = sanitizeFormInput(input);

  return {
    sanitized,
    wasSuspicious,
    originalLength: input.length,
    sanitizedLength: sanitized.length,
  };
};
