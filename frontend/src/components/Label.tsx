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

import '@components/label.css';

interface LabelProps extends React.LabelHTMLAttributes<HTMLLabelElement> {
  children: React.ReactNode;
  className?: string;
  htmlFor: string;
}

const Label = forwardRef<HTMLLabelElement, LabelProps>(
  ({ children, className = '', htmlFor, ...props }, ref) => {
    return (
      <label
        ref={ref}
        className={`generic-label ${className}`}
        htmlFor={htmlFor}
        {...props}
      >
        {children}
      </label>
    );
  }
);

Label.displayName = 'Label';

export default Label;
