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

import React, { forwardRef } from 'react';

import '@features/manager/components/empty.css';

interface EmptyProps extends React.HTMLAttributes<HTMLDivElement> {
  title: string;
  subtitle: string;
}

export const Empty = forwardRef<HTMLDivElement, EmptyProps>(
  ({ title, subtitle, className, ...props }, ref) => {
    return (
      <div ref={ref} className="empty-container" {...props}>
        <img
          className="empty-container__icon"
          src="/nodocs.svg"
          alt="No documents"
        />
        <span className="empty-container__title">{title}</span>
        <span className="empty-container__subtitle">{subtitle}</span>
      </div>
    );
  }
);

Empty.displayName = 'Empty';

export default Empty;
